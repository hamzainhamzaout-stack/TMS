-- name: RevokeActiveTrackingDevices :exec
UPDATE tracking_device
SET revoked_at = NOW()
WHERE driver_id = $1 AND revoked_at IS NULL;

-- name: CreateTrackingDevice :one
INSERT INTO tracking_device (driver_id, token_hash, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetActiveTrackingDeviceByDriver :one
SELECT *
FROM tracking_device
WHERE driver_id = $1 AND revoked_at IS NULL;

-- name: GetActiveTrackingDeviceByTokenHash :one
SELECT *
FROM tracking_device
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: TouchTrackingDevice :exec
UPDATE tracking_device
SET last_seen_at = GREATEST(COALESCE(last_seen_at, sqlc.arg(seen_at)::timestamptz), sqlc.arg(seen_at)::timestamptz)
WHERE id = sqlc.arg(id);

-- name: GetActiveTripIDForDriver :one
SELECT id
FROM trip
WHERE driver_id = $1
  AND isactive = TRUE
  AND status IN ('PLANNED', 'CONFIRMED')
ORDER BY created_at DESC
LIMIT 1;

-- name: InsertDriverLocation :execrows
INSERT INTO driver_location (
    driver_id, device_id, trip_id, latitude, longitude, speed_kmh, heading,
    altitude_m, accuracy_m, battery_pct, is_moving, recorded_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
ON CONFLICT ON CONSTRAINT driver_location_device_time_unique DO NOTHING;

-- name: ListLiveDriverPositions :many
-- Every driver with tracking enabled, plus their most recent point (if any).
SELECT
    d.id AS driver_id,
    d.first_name,
    d.last_name,
    d.phone_number,
    c.license_plate,
    td.last_seen_at,
    dl.trip_id,
    dl.latitude,
    dl.longitude,
    dl.speed_kmh,
    dl.heading,
    dl.accuracy_m,
    dl.battery_pct,
    dl.recorded_at
FROM tracking_device td
JOIN driver d ON d.id = td.driver_id
LEFT JOIN cars c ON c.id = (
    SELECT cars.id FROM cars WHERE cars.current_driver_id = d.id ORDER BY cars.id LIMIT 1
)
LEFT JOIN driver_location dl ON dl.id = (
    SELECT x.id FROM driver_location x WHERE x.driver_id = d.id ORDER BY x.recorded_at DESC LIMIT 1
)
WHERE td.revoked_at IS NULL
ORDER BY d.first_name, d.last_name;

-- name: ListDriverLocations :many
SELECT id, trip_id, latitude, longitude, speed_kmh, heading, accuracy_m, battery_pct, recorded_at
FROM driver_location
WHERE driver_id = sqlc.arg(driver_id)
  AND recorded_at >= sqlc.arg(from_time)::timestamptz
  AND recorded_at <= sqlc.arg(to_time)::timestamptz
ORDER BY recorded_at
LIMIT sqlc.arg(max_points);

-- name: ListTripLocations :many
SELECT id, trip_id, latitude, longitude, speed_kmh, heading, accuracy_m, battery_pct, recorded_at
FROM driver_location
WHERE trip_id = sqlc.arg(trip_id)
ORDER BY recorded_at
LIMIT sqlc.arg(max_points);

-- name: DeleteDriverLocationsBefore :execrows
DELETE FROM driver_location
WHERE recorded_at < sqlc.arg(cutoff)::timestamptz;
