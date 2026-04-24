import logging
import time
from datetime import date, datetime

import psycopg2
import psycopg2.extras

import config

log = logging.getLogger(__name__)

_UPSERT = """
INSERT INTO daily_metrics (date, metric_name, metric_value, computed_at)
VALUES (%s, %s, %s, %s)
ON CONFLICT (date, metric_name) DO UPDATE
    SET metric_value = EXCLUDED.metric_value,
        computed_at  = EXCLUDED.computed_at
"""


def _conn():
    return psycopg2.connect(config.POSTGRES_DSN)


def upsert_metrics(target_date: date, metrics: dict[str, float]) -> None:
    computed_at = datetime.utcnow()
    rows = [(target_date, name, value, computed_at) for name, value in metrics.items()]

    for attempt in range(3):
        try:
            with _conn() as conn, conn.cursor() as cur:
                cur.execute(
                    "DELETE FROM daily_metrics WHERE date = %s AND metric_name LIKE 'top_movie_%%'",
                    (target_date,),
                )
                psycopg2.extras.execute_batch(cur, _UPSERT, rows)
            log.info("Upserted %d metrics for %s", len(rows), target_date)
            return
        except Exception as exc:
            log.warning("Postgres write failed (attempt %d/3): %s", attempt + 1, exc)
            time.sleep(2 ** attempt)
    log.error("Failed to write metrics for %s after 3 attempts", target_date)
