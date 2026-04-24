import csv
import io
import logging
from datetime import date

import boto3
import psycopg2

import config

log = logging.getLogger(__name__)


def _s3():
    return boto3.client(
        "s3",
        endpoint_url=config.S3_ENDPOINT,
        aws_access_key_id=config.S3_ACCESS_KEY,
        aws_secret_access_key=config.S3_SECRET_KEY,
    )


def export(target_date: date) -> None:
    with psycopg2.connect(config.POSTGRES_DSN) as conn, conn.cursor() as cur:
        cur.execute(
            "SELECT metric_name, metric_value, computed_at FROM daily_metrics WHERE date = %s ORDER BY metric_name",
            (target_date,),
        )
        rows = cur.fetchall()

    if not rows:
        log.warning("No metrics for %s, skipping S3 export", target_date)
        return

    buf = io.StringIO()
    writer = csv.writer(buf)
    writer.writerow(["date", "metric_name", "metric_value", "computed_at"])
    for metric_name, metric_value, computed_at in rows:
        writer.writerow([target_date, metric_name, metric_value, computed_at])

    key = f"daily/{target_date}/aggregates.csv"
    _s3().put_object(
        Bucket=config.S3_BUCKET,
        Key=key,
        Body=buf.getvalue().encode(),
        ContentType="text/csv",
    )
    log.info("Exported %d metrics to s3://%s/%s", len(rows), config.S3_BUCKET, key)
