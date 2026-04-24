import logging
import time
from datetime import date, timedelta

import ch_storage
import queries
import s3_export
import storage

log = logging.getLogger(__name__)


def run(target_date: date) -> int:
    log.info("Aggregation started for %s", target_date)
    t0 = time.monotonic()

    dau_val    = queries.dau(target_date)
    watch_rows = queries.avg_watch_time_detail(target_date)
    top        = queries.top_movies(target_date)
    conv       = queries.conversion_detail(target_date)
    ret_data   = queries.retention_all_days(target_date)

    pg_metrics: dict[str, float] = {
        "dau":            dau_val,
        "avg_watch_time": watch_rows["avg_seconds"],
        "conversion":     conv["rate"],
        "retention_d1":   round(ret_data["retained"][1] / ret_data["cohort_size"], 4) if ret_data["cohort_size"] else 0.0,
        "retention_d7":   round(ret_data["retained"][7] / ret_data["cohort_size"], 4) if ret_data["cohort_size"] else 0.0,
    }
    for rank, (movie_id, views) in enumerate(top, 1):
        pg_metrics[f"top_movie_{rank}_{movie_id}"] = float(views)

    storage.upsert_metrics(target_date, pg_metrics)

    ch_storage.write_dau(target_date, dau_val)
    ch_storage.write_avg_watch_time(target_date, watch_rows["avg_seconds"], watch_rows["total_views"])
    ch_storage.write_top_movies(target_date, top)
    ch_storage.write_conversion(target_date, conv["started"], conv["finished"], conv["rate"])
    if ret_data["cohort_size"] > 0:
        ch_storage.write_retention(target_date, ret_data["cohort_size"], ret_data["retained"])

    try:
        s3_export.export(target_date)
    except Exception:
        log.exception("S3 export failed for %s (will retry next cycle)", target_date)

    elapsed = time.monotonic() - t0
    n = len(pg_metrics)
    log.info("Aggregation done for %s: %d metrics in %.2fs", target_date, n, elapsed)
    return n


def run_last_days(days: int = 10) -> None:
    today = date.today()
    for i in range(days):
        run(today - timedelta(days=i))
