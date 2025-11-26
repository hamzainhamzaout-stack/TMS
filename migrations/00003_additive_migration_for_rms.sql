-- +goose Up
-- +goose StatementBegin

CREATE TYPE payment_status AS ENUM (
  'PENDING',
  'PARTIAL',
  'PAID'
);

CREATE TYPE trip_status AS ENUM (
  'PLANNED',
  'CONFIRMED',
  'COMPLETED',
  'CANCELLED'
);

CREATE TYPE transaction_category AS ENUM(
  'INCOME_DEPOSIT',
  'INCOME_FINAL',
  'EXPENSE_TRIP',
  'EXPENSE_OVERHEAD',
  'EXPENSE_MAINTENANCE',
  'EXPENSE_DRIVER_PAYROLL'
);

CREATE TYPE maintenance_type AS ENUM(
  'PREVENTIVE',
  'CORRECTIVE'
);

CREATE TYPE ledger_type AS ENUM(
  'SALARY_CREDIT',
  'ADVANCE_DEBIT',
  'TRIP_ALLOWANCE'
);


ALTER TABLE trip
ADD COLUMN IF NOT EXISTS client_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS total_quote NUMERIC DEFAULT 0,
ADD COLUMN IF NOT EXISTS paid_amount NUMERIC DEFAULT 0,
ADD COLUMN IF NOT EXISTS payment_status payment_status DEFAULT 'PENDING',
ADD COLUMN IF NOT EXISTS status trip_status DEFAULT 'PLANNED';


CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    trip_id INTEGER REFERENCES trip(id) ON DELETE SET NULL,
    amount NUMERIC NOT NULL,
    category transaction_category NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    handled_by uuid REFERENCES users(id),
    handler_name VARCHAR(255),
    notes TEXT,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS maintenance_logs (
    id SERIAL PRIMARY KEY,
    car_id INTEGER NOT NULL,
    maintenance_type maintenance_type NOT NULL,
    maintenance_name VARCHAR(50) NOT NULL,
    cost NUMERIC NOT NULL,
    description TEXT NOT NULL,
    service_provider VARCHAR(255),
    service_date DATE DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS driver_ledger (
    id SERIAL PRIMARY KEY,
    driver_id INTEGER NOT NULL REFERENCES driver(id) ON DELETE CASCADE,
    entry_type VARCHAR(50) NOT NULL,
    amount NUMERIC NOT NULL,
    related_trip_id INTEGER REFERENCES trip(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS driver_ledger;
DROP TABLE IF EXISTS maintenance_logs;
DROP TABLE IF EXISTS transactions;

ALTER TABLE trip
    DROP COLUMN IF EXISTS client_name,
    DROP COLUMN IF EXISTS total_quote,
    DROP COLUMN IF EXISTS paid_amount,
    DROP COLUMN IF EXISTS payment_status,
    DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS ledger_type;
DROP TYPE IF EXISTS maintenance_type;
DROP TYPE IF EXISTS transaction_category;
DROP TYPE IF EXISTS trip_status;
DROP TYPE IF EXISTS payment_status;
-- +goose StatementEnd
