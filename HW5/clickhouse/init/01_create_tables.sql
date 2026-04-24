CREATE DATABASE IF NOT EXISTS cinema;

-- ─── Raw events ───────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cinema.kafka_movie_events
(
    event_id         String,
    user_id          String,
    movie_id         String,
    event_type       String,
    timestamp        String,
    device_type      String,
    session_id       String,
    progress_seconds Int32
)
ENGINE = Kafka
SETTINGS
    kafka_broker_list               = 'kafka-1:9092,kafka-2:9092',
    kafka_topic_list                = 'movie-events',
    kafka_group_name                = 'clickhouse-consumer',
    kafka_format                    = 'AvroConfluent',
    format_avro_schema_registry_url = 'http://schema-registry:8081',
    kafka_num_consumers             = 1;

CREATE TABLE IF NOT EXISTS cinema.movie_events
(
    event_id         String,
    user_id          String,
    movie_id         String,
    event_type       LowCardinality(String),
    event_timestamp  DateTime,
    event_date       Date MATERIALIZED toDate(event_timestamp),
    device_type      LowCardinality(String),
    session_id       String,
    progress_seconds Int32
)
ENGINE = ReplacingMergeTree
PARTITION BY toYYYYMM(event_timestamp)
ORDER BY (event_date, user_id, movie_id, event_type, event_id);

CREATE MATERIALIZED VIEW IF NOT EXISTS cinema.mv_kafka_to_events
TO cinema.movie_events
AS
SELECT
    event_id,
    user_id,
    movie_id,
    event_type,
    parseDateTimeBestEffort(timestamp) AS event_timestamp,
    device_type,
    session_id,
    progress_seconds
FROM cinema.kafka_movie_events;

-- ─── Aggregation tables ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cinema.agg_dau
(
    event_date  Date,
    dau         UInt64,
    computed_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(computed_at)
ORDER BY event_date;

CREATE TABLE IF NOT EXISTS cinema.agg_avg_watch_time
(
    event_date   Date,
    avg_seconds  Float64,
    total_views  UInt64,
    computed_at  DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(computed_at)
ORDER BY event_date;

CREATE TABLE IF NOT EXISTS cinema.agg_top_movies
(
    event_date  Date,
    movie_id    String,
    view_count  UInt64,
    rank        UInt32,
    computed_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(computed_at)
ORDER BY (event_date, movie_id);

CREATE TABLE IF NOT EXISTS cinema.agg_conversion
(
    event_date      Date,
    started         UInt64,
    finished        UInt64,
    conversion_rate Float64,
    computed_at     DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(computed_at)
ORDER BY event_date;

CREATE TABLE IF NOT EXISTS cinema.agg_retention
(
    cohort_date   Date,
    day_number    UInt8,
    cohort_size   UInt64,
    retained      UInt64,
    retention_pct Float64,
    computed_at   DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(computed_at)
ORDER BY (cohort_date, day_number);
