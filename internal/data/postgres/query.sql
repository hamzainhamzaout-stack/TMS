-- name: CreateUser :one
INSERT INTO users(first_name, last_name, phone_number, password_hash, role_id)
VALUES($1, $2, $3, $4, $5)
RETURNING *;

-- name: UserProfile :one
SELECT
  u.id,
  u.first_name,
  u.last_name,
  u.phone_number,
  u.status,
  u.created_at,
  u.updated_at,
  r.id AS role_id,
  r.name AS role_name
FROM users u
JOIN role r ON u.role_id = r.id
WHERE u.id = $1;

-- name: IsTokenBlacklisted :one
SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token = $1);

-- name: CreateRole :one
INSERT INTO role(name)
VALUES($1)
RETURNING *;

-- name: GetUserByPhoneNumber :one
SELECT *
FROM users
WHERE phone_number = $1;

-- name: GetUsers :many
SELECT *
FROM users
WHERE status <> 'deleted';

-- name: GetOTPByUserID :one
SELECT * FROM otps
WHERE user_id = $1 AND status = 'active'
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateUserStatus :one
UPDATE users
SET status = 'active'
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $1, status='active', updated_at = NOW()
WHERE id = $2;

-- name: BlacklistToken :exec
INSERT INTO blacklisted_tokens (token) VALUES ($1);

-- name: CreateOTP :one
INSERT INTO otps (user_id, otp_hash, expires_at, delivery_method)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateOTPStatus :exec
UPDATE otps
SET status = $1
WHERE id = $2;

-- name: InvalidateActiveOTPsForUser :exec
UPDATE otps
SET status = 'expired'
WHERE user_id = $1 AND status = 'active';

-- name: GetMyRole :one
SELECT
  r.id,
  r.name
FROM users u
JOIN role r ON u.role_id = r.id
WHERE u.id = $1;

-- name: GetRoles :many
SELECT
  r.id,
  r.name
FROM role r;

-- name: UpdateRole :one
UPDATE role
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM role
WHERE id = $1;

-- name: CreateDriver :one
INSERT INTO driver(first_name, last_name, phone_number, license_number)
VALUES($1, $2, $3, $4)
RETURNING *;

-- name: UpdateDriver :one
UPDATE driver
SET
  first_name    = COALESCE($2, first_name),
  last_name     = COALESCE($3, last_name),
  phone_number  = COALESCE($4, phone_number),
  license_number= COALESCE($5, license_number),
  updated_at    = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteDriver :exec
UPDATE driver
SET status = 'inactive'
WHERE id = $1;

-- name: ListDrivers :many
SELECT *
FROM driver;

-- name: GetDriverByID :one
SELECT *
FROM driver
WHERE id = $1;

-- name: ListCars :many
SELECT *
FROM cars;

-- name: CreateCar :one
INSERT INTO cars(current_driver_id, license_plate, make, model, year, color, vin, fuel_type, transmission, mileage, status)
VALUES( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateCar :one
UPDATE cars
SET current_driver_id = $2, license_plate = $3, make = $4, model = $5, year = $6, color = $7, vin = $8, fuel_type = $9, transmission = $10, mileage = $11, status = $12
WHERE id = $1
RETURNING *;

-- name: UpdateCarPartial :one
UPDATE cars
SET
    current_driver_id = COALESCE(sqlc.narg('current_driver_id'), current_driver_id),
    license_plate     = COALESCE(sqlc.narg('license_plate'), license_plate),
    make              = COALESCE(sqlc.narg('make'), make),
    model             = COALESCE(sqlc.narg('model'), model),
    year              = COALESCE(sqlc.narg('year'), year),
    color             = COALESCE(sqlc.narg('color'), color),
    vin               = COALESCE(sqlc.narg('vin'), vin),
    fuel_type         = COALESCE(sqlc.narg('fuel_type'), fuel_type),
    transmission      = COALESCE(sqlc.narg('transmission'), transmission),
    mileage           = COALESCE(sqlc.narg('mileage'), mileage),
    status            = COALESCE(sqlc.narg('status'), status),
    updated_at        = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: GetCarByID :one
SELECT *
FROM cars
WHERE id = $1;

-- name: UpdateCarStatus :one
UPDATE cars
SET status = $1
WHERE id = $2
RETURNING *;

-- name: DeleteCar :exec
UPDATE cars
SET status = 'inactive'
WHERE id = $1;

-- name: ChangeCarDriver :exec
UPDATE cars
SET current_driver_id = $1
WHERE id = $2
RETURNING *;

-- name: CreateTrip :one
INSERT INTO trip(driver_id, total_cost, approved_by, origin, destination, departure_time, arrival_time, items, total_quote, item_weight, co_driver_name, client_name, client_phone_number, settlement_due_date, status)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: AddExpensesToTrip :exec
INSERT INTO trip_expenses(trip_id, expenses_id)
VALUES($1, $2);

-- name: GetTrips :many
SELECT *
FROM trip;

-- name: GetTripsByDriver :many
SELECT *
FROM trip
WHERE driver_id = $1;

-- name: UpdateTripDriver :one
UPDATE trip
SET driver_id = $2
WHERE id = $1
RETURNING *;

-- -- name: updatetrip :one
-- UPDATE trip
-- SET total_cost = $2, approved_by = $3, created_at = $4, updated_at = $5
-- WHERE id = $1
-- RETURNING *;

-- name: DeleteTrip :exec
UPDATE trip
SET isActive = false
WHERE id = $1;

-- name: UpdateTripReceiptURL :exec
UPDATE trip
SET receipt_url = $2
WHERE id = $1;


-- name: CreateExpenseType :one
INSERT INTO expenses(name)
VALUES($1)
RETURNING *;

-- name: GetExpenseTypes :many
SELECT *
FROM expenses;

-- name: UpdateExpenseType :one
UPDATE expenses
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteExpenseType :exec
DELETE FROM expenses
WHERE id = $1;

-- name: GetTripDetail :one
SELECT
  t.*,
  d.first_name AS driver_first_name,
  d.last_name AS driver_last_name,
  COALESCE(
    (SELECT json_agg(
      json_build_object(
        'expense_name', e.name,
        'cost', te.cost,
        'receipt_url', te.receipt_url,
        'remark', te.remark
      )
    )
    FROM trip_expenses te
    JOIN expenses e ON te.expenses_id = e.id
    WHERE te.trip_id = t.id),
    '[]'::json
  ) AS expenses,
  COALESCE(
    (SELECT json_agg(
      json_build_object(
        'id', tr.id,
        'amount', tr.amount,
        'category', tr.category,
        'payment_method', tr.payment_method,
        'notes', tr.notes,
        'created_at', tr.created_at
      )
    )
    FROM transactions tr
    WHERE tr.trip_id = t.id AND tr.category IN ('INCOME_DEPOSIT', 'INCOME_FINAL')),
    '[]'::json
  ) AS incomes
FROM trip t
JOIN driver d ON t.driver_id = d.id
WHERE t.id = $1;

-- name: GetDashboardCounts :one
SELECT
    COALESCE((SELECT COUNT(*) FROM driver), 0) AS driver_count,
    COALESCE((SELECT COUNT(*) FROM cars), 0) AS car_count,
    COALESCE((SELECT COUNT(*) FROM expenses), 0) AS expense_type_count,
    COALESCE((SELECT COUNT(*) FROM trip), 0) AS trip_count,
    COALESCE((SELECT COUNT(*) FROM driver WHERE status = 'active'), 0) AS active_driver_count,
    COALESCE((SELECT COUNT(*) FROM trip WHERE isActive), 0) AS active_trips_count;

-- name: UpdateUserProfile :one
UPDATE users
SET
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name)
WHERE
    id = $1
RETURNING *;

-- name: CreateTripExpense :one
INSERT INTO trip_expenses(trip_id, expenses_id, cost, receipt_url, remark)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateTransaction :one
INSERT INTO transactions (
    trip_id,
    amount,
    category,
    payment_method,
    handled_by,
    handler_name,
    notes,
    transaction_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;
-- name: UpdateTripPaymentState :exec
UPDATE trip
SET
    paid_amount = paid_amount + $2,
    payment_status = CASE
        WHEN (paid_amount + $2) >= total_quote THEN 'PAID'::payment_status
        ELSE 'PARTIAL'::payment_status
    END,
    status = CASE
        WHEN status = 'PLANNED'::trip_status THEN 'CONFIRMED'::trip_status
        ELSE status
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
-- name: CreateMaintenance :one
INSERT INTO maintenance_logs (car_id, maintenance_type, cost, description, maintenance_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFinancialSummary :one
-- TODO: maintenance_cost should be calculated from the maintenance_logs table.
SELECT
    COALESCE(SUM(CASE WHEN category = 'INCOME_DEPOSIT' OR category = 'INCOME_FINAL' THEN amount ELSE 0 END), 0) AS total_revenue,
    COALESCE(SUM(CASE WHEN category = 'EXPENSE_MAINTENANCE' THEN amount ELSE 0 END), 0) AS maintenance_cost,
    COALESCE(SUM(CASE WHEN category = 'EXPENSE_TRIP' THEN amount ELSE 0 END), 0) AS trip_overheads,
    COALESCE(SUM(CASE WHEN category = 'EXPENSE_DRIVER_PAYROLL' THEN amount ELSE 0 END), 0) AS driver_payroll
FROM
    transactions;

-- name: GetTripProfitability :one
SELECT
    t.id AS trip_id,
    COALESCE(
        SUM(CASE WHEN tx.category IN ('INCOME_DEPOSIT', 'INCOME_FINAL') THEN tx.amount ELSE 0 END),
        0
    )::float8 AS total_revenue,

    (
        COALESCE(SUM(CASE WHEN tx.category IN ('EXPENSE_TRIP', 'EXPENSE_OVERHEAD') THEN tx.amount ELSE 0 END), 0)
        - t.returned_amount
    )::float8 AS total_expenses

FROM
    trip t
LEFT JOIN
    transactions tx ON t.id = tx.trip_id
WHERE
    t.id = $1
GROUP BY
    t.id;


-- name: GetTripTransactionDetail :one
SELECT *
FROM transactions
WHERE trip_id = $1;

-- name: DriverLedger :one
INSERT INTO driver_ledger (driver_id, entry_type, amount, related_trip_id, notes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDriverLedger :many
SELECT *
FROM driver_ledger
WHERE driver_id = $1;

-- name: GetDriverDetail :one
SELECT
    d.*,
    c.license_plate,
    COALESCE(
        JSON_AGG(dl ORDER BY dl.created_at)
        FILTER (WHERE dl.id IS NOT NULL),
        '[]'
    ) AS ledger_entries
FROM driver d
LEFT JOIN cars c ON d.id = c.current_driver_id
LEFT JOIN driver_ledger dl ON d.id = dl.driver_id
WHERE d.id = $1
GROUP BY d.id, c.license_plate;

-- name: RegisterReturnPayment :one
UPDATE trip
SET
    returned_amount = returned_amount + $2,
    total_cost = total_cost - $2
WHERE id = $1
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
  first_name    = COALESCE($2, first_name),
  last_name     = COALESCE($3, last_name),
  phone_number  = COALESCE($4, phone_number),
  updated_at    = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
UPDATE users
SET status = 'deleted'
WHERE id = $1;

-- -- name: UpdateTrip :one
-- UPDATE trip
-- SET
--     total_cost = COALESCE(total_cost, $2),
--     approved_by = $3,
--     created_at = $4,
--     updated_at = $5
-- WHERE id = $1
-- RETURNING *;

-- name: UpdateTrip :one
UPDATE trip
SET
    driver_id           = COALESCE($2, driver_id),
    total_cost          = COALESCE($3, total_cost),
    receipt_url         = COALESCE($4, receipt_url),
    approved_by         = COALESCE($5, approved_by),
    isactive            = COALESCE($6, isactive),
    client_name         = COALESCE($7, client_name),
    total_quote         = COALESCE($8, total_quote),
    paid_amount         = COALESCE($9, paid_amount),
    payment_status      = COALESCE($10, payment_status),
    status              = COALESCE($11, status),
    origin              = COALESCE($12, origin),
    destination         = COALESCE($13, destination),
    departure_time      = COALESCE($14, departure_time),
    arrival_time        = COALESCE($15, arrival_time),
    items               = COALESCE($16, items),
    returned_amount     = COALESCE($17, returned_amount),
    client_phone_number = COALESCE($18, client_phone_number),
    updated_at          = NOW(),
    item_weight         = COALESCE($19, item_weight),
    co_driver_name      = COALESCE($20, co_driver_name)
WHERE id = $1
RETURNING *;
