import os

CLICKHOUSE_HOST     = os.environ.get("CLICKHOUSE_HOST", "localhost")
CLICKHOUSE_PORT     = int(os.environ.get("CLICKHOUSE_PORT", "9000"))
CLICKHOUSE_DB       = os.environ.get("CLICKHOUSE_DB", "cinema")
CLICKHOUSE_USER     = os.environ.get("CLICKHOUSE_USER", "default")
CLICKHOUSE_PASSWORD = os.environ.get("CLICKHOUSE_PASSWORD", "")

POSTGRES_DSN = os.environ.get(
    "POSTGRES_DSN",
    "postgresql://analytics:analytics@localhost:5432/analytics",
)

AGGREGATION_INTERVAL_MINUTES = int(os.environ.get("AGGREGATION_INTERVAL_MINUTES", "5"))
APP_PORT = int(os.environ.get("APP_PORT", "8082"))

S3_ENDPOINT   = os.environ.get("S3_ENDPOINT",   "http://localhost:9002")
S3_ACCESS_KEY = os.environ.get("S3_ACCESS_KEY", "minioadmin")
S3_SECRET_KEY = os.environ.get("S3_SECRET_KEY", "minioadmin")
S3_BUCKET     = os.environ.get("S3_BUCKET",     "movie-analytics")
