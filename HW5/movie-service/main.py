import logging
import threading
import uuid
from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException
from confluent_kafka import KafkaException

import config
from models import MovieEvent
from producer import get_schema_id, publish
from generator import generator_loop

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)

app = FastAPI(title="Movie Service")


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/events")
def ingest_event(event: MovieEvent):
    if not event.event_id:
        event.event_id = str(uuid.uuid4())
    if not event.timestamp:
        event.timestamp = datetime.now(timezone.utc).isoformat()

    record = {
        "event_id":         event.event_id,
        "user_id":          event.user_id,
        "movie_id":         event.movie_id,
        "event_type":       event.event_type.value,
        "timestamp":        event.timestamp,
        "device_type":      event.device_type.value,
        "session_id":       event.session_id,
        "progress_seconds": event.progress_seconds,
    }
    try:
        publish(record)
    except KafkaException as exc:
        raise HTTPException(status_code=503, detail=str(exc))

    log.info("Ingested event_id=%s event_type=%s", event.event_id, event.event_type)
    return {"event_id": event.event_id}


@app.on_event("startup")
def startup():
    threading.Thread(target=get_schema_id, daemon=True).start()
    if config.GENERATOR_ENABLED:
        threading.Thread(target=generator_loop, daemon=True).start()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=config.APP_PORT, log_level="info")
