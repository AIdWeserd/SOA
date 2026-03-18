CREATE TABLE flights (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_number   VARCHAR(10)     NOT NULL,
    airline         VARCHAR(100)    NOT NULL,
    origin          CHAR(3)         NOT NULL,
    destination     CHAR(3)         NOT NULL,
    departure_at    TIMESTAMPTZ     NOT NULL,
    arrival_at      TIMESTAMPTZ     NOT NULL,
    total_seats     INT             NOT NULL CHECK (total_seats > 0),
    available_seats INT             NOT NULL CHECK (available_seats >= 0),
    price           NUMERIC(10, 2)  NOT NULL CHECK (price > 0),
    status          VARCHAR(20)     NOT NULL DEFAULT 'SCHEDULED',
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_available_lte_total CHECK (available_seats <= total_seats),
    CONSTRAINT chk_status CHECK (status IN ('SCHEDULED','DEPARTED','CANCELLED','COMPLETED'))
);

CREATE UNIQUE INDEX uq_flight_number_date ON flights (flight_number, departure_at);
