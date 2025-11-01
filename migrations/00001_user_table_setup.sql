-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS role (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL
);

CREATE TYPE user_state AS ENUM (
    'uninitialized',
    'active',
    'unactive',
    'deleted'
);

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    password_hash bytea,
    phone_number VARCHAR(50) UNIQUE NOT NULL,
    role_id INTEGER NOT NULL REFERENCES role(id) ON DELETE CASCADE,
    status user_state NOT NULL DEFAULT 'uninitialized',
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TYPE otp_status AS ENUM ('active', 'used', 'expired');

CREATE TABLE IF NOT EXISTS otps (
    id SERIAL PRIMARY KEY,
    user_id uuid NOT NULL,
    otp_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    delivery_method VARCHAR(20),
    attempt_count INTEGER DEFAULT 0,
    status otp_status DEFAULT 'active',
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS driver (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    license_number VARCHAR(150) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS expenses (
id SERIAL PRIMARY KEY,
name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS trip(
  id SERIAL PRIMARY KEY,
  driver_id INTEGER NOT NULL REFERENCES driver (id) ON DELETE CASCADE,
  total_cost NUMERIC NOT NULL,
  approved_by uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  isActive BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS trip_expenses(
  trip_id INTEGER NOT NULL REFERENCES trip (id) ON DELETE CASCADE,
  expenses_id INTEGER NOT NULL REFERENCES expenses (id) ON DELETE CASCADE,
  PRIMARY KEY (trip_id, expenses_id)
);

CREATE TABLE IF NOT EXISTS cars(
  id SERIAL PRIMARY KEY,
  current_driver_id INTEGER REFERENCES driver (id) ON DELETE SET NULL,
  license_plate VARCHAR(20) UNIQUE NOT NULL,
  make VARCHAR(100) NOT NULL,        -- e.g., Toyota
  model VARCHAR(100) NOT NULL,       -- e.g., Corolla
  year INTEGER CHECK (year >= 1900), -- manufacturing year
  color VARCHAR(50),
  vin VARCHAR(50) UNIQUE,            -- vehicle identification number
  fuel_type VARCHAR(20) CHECK (fuel_type IN ('petrol', 'diesel', 'electric', 'hybrid')),
  transmission VARCHAR(20) CHECK (transmission IN ('manual', 'automatic')),
  mileage NUMERIC,                   -- total km/miles driven
  status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'in_use', 'maintenance', 'inactive')),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blacklisted_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role;
DROP TABLE IF EXISTS otps;
DROP TABLE IF EXISTS trip;
DROP TABLE IF EXISTS cars;
DROP TYPE IF EXISTS otp_status;
DROP TYPE IF EXISTS user_state;
DROP TABLE IF EXISTS blacklisted_tokens;
DROP TABLE IF EXISTS driver;
-- +goose StatementEnd
