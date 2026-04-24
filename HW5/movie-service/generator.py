import logging
import random
import time
import uuid
from datetime import date, datetime, timedelta, timezone

from config import GENERATOR_INTERVAL_MS
from models import DeviceType
from producer import publish

log = logging.getLogger(__name__)

MOVIES  = [f"movie_{i}" for i in range(1, 21)]
DEVICES = [d.value for d in DeviceType]

HISTORY_DAYS     = 10
USERS_PER_COHORT = 5


def _cohort_start() -> date:
    return date.today() - timedelta(days=HISTORY_DAYS - 1)


def _eligible_users(target_date: date) -> list[str]:
    start = _cohort_start()
    users = []
    d = start
    while d <= target_date:
        day_offset = (d - start).days
        for n in range(1, USERS_PER_COHORT + 1):
            users.append(f"user_{day_offset * USERS_PER_COHORT + n}")
        d += timedelta(days=1)
    return users


def _ts(target_date: date) -> str:
    dt = datetime(
        target_date.year, target_date.month, target_date.day,
        random.randint(0, 23), random.randint(0, 59), random.randint(0, 59),
        tzinfo=timezone.utc,
    )
    return dt.isoformat()


def make_session(target_date: date) -> list[dict]:
    user   = random.choice(_eligible_users(target_date))
    movie  = random.choice(MOVIES)
    device = random.choice(DEVICES)
    sid    = str(uuid.uuid4())
    events = []

    def ev(etype: str, prog: int = 0) -> dict:
        return {
            "event_id":         str(uuid.uuid4()),
            "user_id":          user,
            "movie_id":         movie,
            "event_type":       etype,
            "timestamp":        _ts(target_date),
            "device_type":      device,
            "session_id":       sid,
            "progress_seconds": prog,
        }

    if random.random() < 0.3:
        events.append(ev("SEARCHED"))

    events.append(ev("VIEW_STARTED"))

    duration = random.randint(300, 7200)
    if random.random() < 0.4:
        pause_at = random.randint(60, duration - 60)
        events.append(ev("VIEW_PAUSED", pause_at))
        events.append(ev("VIEW_RESUMED", pause_at))

    if random.random() < 0.7:
        events.append(ev("VIEW_FINISHED", duration))

    if random.random() < 0.2:
        events.append(ev("LIKED", duration))

    return events


def generator_loop() -> None:
    log.info("Event generator started (interval=%dms)", GENERATOR_INTERVAL_MS)
    today = date.today()
    while True:
        try:
            target_date = today - timedelta(days=random.randint(0, 9))
            for record in make_session(target_date):
                publish(record)
                time.sleep(GENERATOR_INTERVAL_MS / 1000 * random.uniform(0.5, 1.5))
        except Exception:
            log.exception("Generator error, retrying in 5s")
            time.sleep(5)
