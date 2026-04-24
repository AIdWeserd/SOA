import time
import uuid
from datetime import datetime, timezone

import pytest
import requests
from clickhouse_driver import Client as CHClient

MOVIE_SERVICE_URL = "http://localhost:8080"
CH_HOST  = "localhost"
CH_DB    = "cinema"
TIMEOUT  = 60
POLL_INT = 3


def _wait_for_service(url: str, retries: int = 20, delay: float = 3.0):
    for _ in range(retries):
        try:
            r = requests.get(f"{url}/health", timeout=5)
            if r.status_code == 200:
                return
        except requests.exceptions.ConnectionError:
            pass
        time.sleep(delay)
    pytest.skip(f"Service not ready: {url}")


@pytest.fixture(scope="session", autouse=True)
def wait_for_services():
    _wait_for_service(MOVIE_SERVICE_URL)


@pytest.fixture
def ch():
    client = CHClient(host=CH_HOST, port=9000, database=CH_DB,
                      user="default", password="")
    yield client
    client.disconnect()


def publish_event(**overrides) -> dict:
    payload = {
        "user_id":          "test-user",
        "movie_id":         "test-movie",
        "event_type":       "VIEW_STARTED",
        "device_type":      "DESKTOP",
        "session_id":       str(uuid.uuid4()),
        "progress_seconds": 0,
        **overrides,
    }
    resp = requests.post(f"{MOVIE_SERVICE_URL}/events", json=payload, timeout=10)
    assert resp.status_code == 200, f"Status {resp.status_code}: {resp.text}"
    data = resp.json()
    assert "event_id" in data
    payload["event_id"] = data["event_id"]
    return payload


def poll_clickhouse(ch: CHClient, event_id: str) -> list:
    deadline = time.time() + TIMEOUT
    while time.time() < deadline:
        rows = ch.execute(
            "SELECT event_id, user_id, movie_id, event_type, device_type, progress_seconds "
            "FROM cinema.movie_events WHERE event_id = %(eid)s",
            {"eid": event_id},
        )
        if rows:
            return rows
        time.sleep(POLL_INT)
    pytest.fail(f"Event {event_id} did not appear in ClickHouse within {TIMEOUT}s")


class TestProducerAPI:
    def test_health(self):
        r = requests.get(f"{MOVIE_SERVICE_URL}/health")
        assert r.status_code == 200
        assert r.json()["status"] == "ok"

    def test_publish_returns_valid_uuid(self):
        event = publish_event()
        assert uuid.UUID(event["event_id"])

    def test_invalid_event_type_rejected(self):
        r = requests.post(f"{MOVIE_SERVICE_URL}/events", json={
            "user_id": "u1", "movie_id": "m1",
            "event_type": "INVALID", "device_type": "DESKTOP", "session_id": "s1",
        })
        assert r.status_code == 422

    def test_negative_progress_rejected(self):
        r = requests.post(f"{MOVIE_SERVICE_URL}/events", json={
            "user_id": "u1", "movie_id": "m1",
            "event_type": "VIEW_STARTED", "device_type": "DESKTOP",
            "session_id": "s1", "progress_seconds": -1,
        })
        assert r.status_code == 422

    def test_all_event_types_accepted(self):
        for etype in ["VIEW_STARTED", "VIEW_FINISHED", "VIEW_PAUSED",
                      "VIEW_RESUMED", "LIKED", "SEARCHED"]:
            event = publish_event(event_type=etype)
            assert uuid.UUID(event["event_id"])

    def test_all_device_types_accepted(self):
        for dtype in ["MOBILE", "DESKTOP", "TV", "TABLET"]:
            event = publish_event(device_type=dtype)
            assert uuid.UUID(event["event_id"])


class TestEndToEndPipeline:
    def test_event_reaches_clickhouse(self, ch):
        event = publish_event(
            user_id="e2e-user",
            movie_id="e2e-movie",
            event_type="VIEW_FINISHED",
            progress_seconds=3600,
        )
        rows = poll_clickhouse(ch, event["event_id"])
        row = rows[0]
        assert row[0] == event["event_id"]
        assert row[1] == "e2e-user"
        assert row[2] == "e2e-movie"
        assert row[3] == "VIEW_FINISHED"
        assert row[5] == 3600

    def test_event_fields_integrity(self, ch):
        event = publish_event(
            user_id="field-user",
            movie_id="field-movie",
            device_type="TABLET",
            event_type="LIKED",
        )
        rows = poll_clickhouse(ch, event["event_id"])
        assert rows[0][3] == "LIKED"
        assert rows[0][4] == "TABLET"

    def test_session_events_all_arrive(self, ch):
        user_id    = f"session-user-{uuid.uuid4()}"
        session_id = str(uuid.uuid4())
        sequence   = [
            ("VIEW_STARTED",  0),
            ("VIEW_PAUSED",   120),
            ("VIEW_RESUMED",  120),
            ("VIEW_FINISHED", 3600),
        ]
        for etype, prog in sequence:
            publish_event(user_id=user_id, movie_id="session-movie",
                          event_type=etype, progress_seconds=prog,
                          session_id=session_id)

        deadline = time.time() + TIMEOUT
        while time.time() < deadline:
            rows = ch.execute(
                "SELECT count() FROM cinema.movie_events "
                "WHERE user_id = %(uid)s AND session_id = %(sid)s",
                {"uid": user_id, "sid": session_id},
            )
            if rows and int(rows[0][0]) >= len(sequence):
                return
            time.sleep(POLL_INT)
        pytest.fail("Not all session events arrived in ClickHouse")
