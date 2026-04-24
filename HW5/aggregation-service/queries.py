from datetime import date

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


def dau(target_date: date) -> float:
    rows = _ch().execute(
        "SELECT uniqExact(user_id) FROM cinema.movie_events WHERE event_date = %(d)s",
        {"d": target_date},
    )
    return float(rows[0][0]) if rows else 0.0


def avg_watch_time_detail(target_date: date) -> dict:
    rows = _ch().execute(
        """
        SELECT avg(progress_seconds), count()
        FROM cinema.movie_events
        WHERE event_date = %(d)s AND event_type = 'VIEW_FINISHED'
        """,
        {"d": target_date},
    )
    if not rows or rows[0][0] is None:
        return {"avg_seconds": 0.0, "total_views": 0}
    return {"avg_seconds": float(rows[0][0]), "total_views": int(rows[0][1])}


def avg_watch_time(target_date: date) -> float:
    return avg_watch_time_detail(target_date)["avg_seconds"]


def top_movies(target_date: date, limit: int = 10) -> list[tuple[str, int]]:
    rows = _ch().execute(
        """
        SELECT movie_id, count() AS views
        FROM cinema.movie_events
        WHERE event_date = %(d)s AND event_type = 'VIEW_STARTED'
        GROUP BY movie_id
        ORDER BY views DESC
        LIMIT %(limit)s
        """,
        {"d": target_date, "limit": limit},
    )
    return [(row[0], row[1]) for row in rows]


def conversion_detail(target_date: date) -> dict:
    rows = _ch().execute(
        """
        SELECT
            countIf(event_type = 'VIEW_STARTED')  AS started,
            countIf(event_type = 'VIEW_FINISHED') AS finished
        FROM cinema.movie_events
        WHERE event_date = %(d)s
        """,
        {"d": target_date},
    )
    if not rows:
        return {"started": 0, "finished": 0, "rate": 0.0}
    started, finished = rows[0]
    rate = round(finished / started, 4) if started > 0 else 0.0
    return {"started": int(started), "finished": int(finished), "rate": rate}


def conversion(target_date: date) -> float:
    return conversion_detail(target_date)["rate"]


def retention_all_days(target_date: date) -> dict:
    rows = _ch().execute(
        """
        WITH cohort AS (
            SELECT user_id
            FROM cinema.movie_events
            GROUP BY user_id
            HAVING minIf(event_date, event_type = 'VIEW_STARTED') = %(d)s
        ),
        user_days AS (
            SELECT c.user_id,
                maxIf(1, e.event_date = toDate(%(d)s) + 0) AS d0,
                maxIf(1, e.event_date = toDate(%(d)s) + 1) AS d1,
                maxIf(1, e.event_date = toDate(%(d)s) + 2) AS d2,
                maxIf(1, e.event_date = toDate(%(d)s) + 3) AS d3,
                maxIf(1, e.event_date = toDate(%(d)s) + 4) AS d4,
                maxIf(1, e.event_date = toDate(%(d)s) + 5) AS d5,
                maxIf(1, e.event_date = toDate(%(d)s) + 6) AS d6,
                maxIf(1, e.event_date = toDate(%(d)s) + 7) AS d7
            FROM cohort c
            LEFT JOIN cinema.movie_events e ON c.user_id = e.user_id
            GROUP BY c.user_id
        )
        SELECT count() AS cohort_size,
               countIf(d0=1), countIf(d1=1), countIf(d2=1), countIf(d3=1),
               countIf(d4=1), countIf(d5=1), countIf(d6=1), countIf(d7=1)
        FROM user_days
        """,
        {"d": target_date},
    )
    if not rows or rows[0][0] == 0:
        return {"cohort_size": 0, "retained": [0] * 8}

    cohort_size = rows[0][0]
    retained = list(rows[0][1:])
    return {"cohort_size": cohort_size, "retained": retained}


def retention(target_date: date) -> dict[str, float]:
    data = retention_all_days(target_date)
    cohort_size = data["cohort_size"]
    retained = data["retained"]
    if cohort_size == 0:
        return {"retention_d1": 0.0, "retention_d7": 0.0}
    return {
        "retention_d1": round(retained[1] / cohort_size, 4),
        "retention_d7": round(retained[7] / cohort_size, 4),
    }
