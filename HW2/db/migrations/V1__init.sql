CREATE TYPE product_status AS ENUM ('ACTIVE', 'INACTIVE', 'ARCHIVED');

CREATE TABLE products (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255)    NOT NULL,
    description TEXT,
    price       NUMERIC(19, 4)  NOT NULL CHECK (price > 0),
    stock       INTEGER         NOT NULL CHECK (stock >= 0),
    category    VARCHAR(255)    NOT NULL,
    status      product_status  NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_status ON products (status);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
