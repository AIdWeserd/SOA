from datetime import date, datetime

from clickhouse_driver import Client as CHClient

import config


def _ch() -> CHClient:
    return CHClient(
        host=config.CLICKHOUSE_HOST,
        port=config.CLICKHOUSE_PORT,
        database=config.CLICKHOUSE_DB,
        user=config.CLICKHOUSE_USER,
        password=config.CLICKHOUSE_PASSWORD,
    )


def write_dau(target_date: date, dau: float) -> None:
    _ch().execute(
        "INSERT INTO cinema.agg_dau (event_date, dau) VALUES",
        [(target_date, int(dau))],
    )


def write_avg_watch_time(target_date: date, avg_seconds: float, total_views: int) -> None:
    _ch().execute(
        "INSERT INTO cinema.agg_avg_watch_time (event_date, avg_seconds, total_views) VALUES",
        [(target_date, avg_seconds, total_views)],
    )


def write_top_movies(target_date: date, movies: list[tuple[str, int]]) -> None:
    rows = [(target_date, movie_id, views, rank) for rank, (movie_id, views) in enumerate(movies, 1)]
    _ch().execute(
        "INSERT INTO cinema.agg_top_movies (event_date, movie_id, view_count, rank) VALUES",
        rows,
    )


def write_conversion(target_date: date, started: int, finished: int, rate: float) -> None:
    _ch().execute(
        "INSERT INTO cinema.agg_conversion (event_date, started, finished, conversion_rate) VALUES",
        [(target_date, started, finished, rate)],
    )


def write_retention(target_date: date, cohort_size: int, retained: list[int]) -> None:
    rows = []
    for day_number, ret_count in enumerate(retained):
        pct = round(ret_count / cohort_size, 4) if cohort_size > 0 else 0.0
        rows.append((target_date, day_number, cohort_size, ret_count, pct))
    _ch().execute(
        "INSERT INTO cinema.agg_retention (cohort_date, day_number, cohort_size, retained, retention_pct) VALUES",
        rows,
    )
