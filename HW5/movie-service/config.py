import os

KAFKA_BROKERS         = os.environ["KAFKA_BOOTSTRAP_SERVERS"]
SCHEMA_REGISTRY_URL   = os.environ["SCHEMA_REGISTRY_URL"].rstrip("/")
KAFKA_TOPIC           = os.environ.get("KAFKA_TOPIC", "movie-events")
GENERATOR_ENABLED     = os.environ.get("GENERATOR_ENABLED", "false").lower() == "true"
GENERATOR_INTERVAL_MS = int(os.environ.get("GENERATOR_INTERVAL_MS", "1000"))
APP_PORT              = int(os.environ.get("APP_PORT", "8080"))
SCHEMA_PATH           = "schema/movie_event.avsc"
