import io
import json
import logging
import struct
import threading
import time

import fastavro
import requests
from confluent_kafka import Producer

from config import KAFKA_BROKERS, KAFKA_TOPIC, SCHEMA_PATH, SCHEMA_REGISTRY_URL

log = logging.getLogger(__name__)

with open(SCHEMA_PATH) as f:
    _schema_dict = json.load(f)
_schema = fastavro.parse_schema(_schema_dict)

_schema_id: int | None = None
_schema_id_lock = threading.Lock()


def get_schema_id() -> int:
    global _schema_id
    if _schema_id is not None:
        return _schema_id
    with _schema_id_lock:
        if _schema_id is not None:
            return _schema_id
        subject = f"{KAFKA_TOPIC}-value"
        payload = {"schema": json.dumps(_schema_dict), "schemaType": "AVRO"}
        for attempt in range(10):
            try:
                resp = requests.post(
                    f"{SCHEMA_REGISTRY_URL}/subjects/{subject}/versions",
                    json=payload,
                    timeout=5,
                )
                resp.raise_for_status()
                _schema_id = resp.json()["id"]
                log.info("Schema registered: id=%d subject=%s", _schema_id, subject)
                return _schema_id
            except Exception as exc:
                delay = min(2 ** attempt, 60)
                log.warning("Schema Registry not ready (%s), retry %d/10 in %ds", exc, attempt + 1, delay)
                time.sleep(delay)
        raise RuntimeError("Cannot connect to Schema Registry after 10 attempts")


def encode(record: dict) -> bytes:
    schema_id = get_schema_id()
    buf = io.BytesIO()
    buf.write(b"\x00")
    buf.write(struct.pack(">I", schema_id))
    fastavro.schemaless_writer(buf, _schema, record)
    return buf.getvalue()


_producer: Producer | None = None
_producer_lock = threading.Lock()


def get_producer() -> Producer:
    global _producer
    if _producer is None:
        with _producer_lock:
            if _producer is None:
                _producer = Producer({
                    "bootstrap.servers": KAFKA_BROKERS,
                    "acks": "all",
                    "retries": 5,
                    "retry.backoff.ms": 300,
                    "delivery.timeout.ms": 30000,
                })
    return _producer


def _on_delivery(err, msg):
    if err:
        log.error("Delivery error: %s", err)
    else:
        log.info("Published event_id=%s partition=%d", msg.key().decode() if msg.key() else "?", msg.partition())


def publish(record: dict) -> None:
    payload = encode(record)
    key = record["user_id"].encode()
    get_producer().produce(KAFKA_TOPIC, value=payload, key=key, on_delivery=_on_delivery)
    get_producer().poll(0)
