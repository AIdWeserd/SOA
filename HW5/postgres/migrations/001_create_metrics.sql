CREATE TABLE IF NOT EXISTS daily_metrics (
    date         DATE        NOT NULL,
    metric_name  VARCHAR(64) NOT NULL,
    metric_value FLOAT       NOT NULL,
    computed_at  TIMESTAMP   NOT NULL DEFAULT now(),
    PRIMARY KEY (date, metric_name)
);
