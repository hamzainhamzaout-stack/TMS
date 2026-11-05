-- +goose Up
-- +goose StatementBegin

-- A phone running Traccar Client, linked to one driver. Only a SHA-256 hash
-- of the device token is stored; the plain token is shown once at creation.
CREATE TABLE IF NOT EXISTS tracking_device (
    id           SERIAL PRIMARY KEY,
    driver_id    INTEGER NOT NULL REFERENCES driver (id) ON DELETE CASCADE,
    token_hash   TEXT NOT NULL UNIQUE,
    created_by   uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ
);

-- At most one active device per driver.
CREATE UNIQUE INDEX IF NOT EXISTS tracking_device_one_active_per_driver
    ON tracking_device (driver_id) WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS driver_location (
    id          BIGSERIAL PRIMARY KEY,
    driver_id   INTEGER NOT NULL REFERENCES driver (id) ON DELETE CASCADE,
    device_id   INTEGER NOT NULL REFERENCES tracking_device (id) ON DELETE CASCADE,
    trip_id     INTEGER REFERENCES trip (id) ON DELETE SET NULL,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    speed_kmh   REAL,
    heading     REAL,
    altitude_m  REAL,
    accuracy_m  REAL,
    battery_pct REAL,
    is_moving   BOOLEAN,
    recorded_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- The app re-sends buffered points after losing signal; ignore repeats.
    CONSTRAINT driver_location_device_time_unique UNIQUE (device_id, recorded_at)
);

CREATE INDEX IF NOT EXISTS driver_location_driver_time_idx
    ON driver_location (driver_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS driver_location_trip_time_idx
    ON driver_location (trip_id, recorded_at) WHERE trip_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS driver_location;
DROP TABLE IF EXISTS tracking_device;
-- +goose StatementEnd
