CREATE TABLE seat_reservations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id   UUID            NOT NULL REFERENCES flights(id),
    booking_id  UUID            NOT NULL UNIQUE,
    seat_count  INT             NOT NULL CHECK (seat_count > 0),
    status      VARCHAR(20)     NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_reservation_status CHECK (status IN ('ACTIVE','RELEASED','EXPIRED'))
);
