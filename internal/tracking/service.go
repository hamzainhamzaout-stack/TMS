package tracking

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUnknownDevice  = errors.New("unknown or revoked tracking device")
	ErrDriverNotFound = errors.New("driver not found")
)

const (
	maxFutureSkew   = 10 * time.Minute
	maxFixAge       = 7 * 24 * time.Hour // offline buffers older than this are dropped
	maxAccuracyM    = 1000
	defaultMaxPoint = 5000
)

type Service struct {
	logger  *slog.Logger
	pool    *pgxpool.Pool
	queries *dm_svc_db.Queries
	now     func() time.Time
}

func NewService(logger *slog.Logger, pool *pgxpool.Pool) *Service {
	return &Service{
		logger:  logger,
		pool:    pool,
		queries: dm_svc_db.New(pool),
		now:     time.Now,
	}
}

// EnableTracking issues a new device token for a driver, revoking any
// previous one. The plain token is returned once and never stored.
func (s *Service) EnableTracking(ctx context.Context, driverID int32, createdBy pgtype.UUID) (string, dm_svc_db.TrackingDevice, error) {
	if _, err := s.queries.GetDriverByID(ctx, driverID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", dm_svc_db.TrackingDevice{}, ErrDriverNotFound
		}
		return "", dm_svc_db.TrackingDevice{}, err
	}

	token, err := newDeviceToken()
	if err != nil {
		return "", dm_svc_db.TrackingDevice{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", dm_svc_db.TrackingDevice{}, err
	}
	defer tx.Rollback(ctx)

	q := s.queries.WithTx(tx)
	if err := q.RevokeActiveTrackingDevices(ctx, driverID); err != nil {
		return "", dm_svc_db.TrackingDevice{}, err
	}
	device, err := q.CreateTrackingDevice(ctx, dm_svc_db.CreateTrackingDeviceParams{
		DriverID:  driverID,
		TokenHash: hashToken(token),
		CreatedBy: createdBy,
	})
	if err != nil {
		return "", dm_svc_db.TrackingDevice{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", dm_svc_db.TrackingDevice{}, err
	}
	return token, device, nil
}

func (s *Service) DisableTracking(ctx context.Context, driverID int32) error {
	return s.queries.RevokeActiveTrackingDevices(ctx, driverID)
}

// TrackingStatus returns the driver's active device, or nil if tracking is off.
func (s *Service) TrackingStatus(ctx context.Context, driverID int32) (*dm_svc_db.TrackingDevice, error) {
	d, err := s.queries.GetActiveTrackingDeviceByDriver(ctx, driverID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Ingest stores the fixes sent by a device. It returns how many were new.
func (s *Service) Ingest(ctx context.Context, token string, fixes []Fix) (int, error) {
	device, err := s.queries.GetActiveTrackingDeviceByTokenHash(ctx, hashToken(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrUnknownDevice
	}
	if err != nil {
		return 0, err
	}

	now := s.now().UTC()
	var tripID *int32
	if id, err := s.queries.GetActiveTripIDForDriver(ctx, device.DriverID); err == nil {
		tripID = &id
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	stored := 0
	var latest time.Time
	for _, f := range fixes {
		if reason := rejectReason(f, now); reason != "" {
			s.logger.Debug("dropping gps fix", "driver_id", device.DriverID, "reason", reason)
			continue
		}
		n, err := s.queries.InsertDriverLocation(ctx, dm_svc_db.InsertDriverLocationParams{
			DriverID:   device.DriverID,
			DeviceID:   device.ID,
			TripID:     tripID,
			Latitude:   f.Latitude,
			Longitude:  f.Longitude,
			SpeedKmh:   f.SpeedKmh,
			Heading:    f.Heading,
			AltitudeM:  f.AltitudeM,
			AccuracyM:  f.AccuracyM,
			BatteryPct: f.BatteryPct,
			IsMoving:   f.IsMoving,
			RecordedAt: ts(f.RecordedAt),
		})
		if err != nil {
			return stored, fmt.Errorf("insert location: %w", err)
		}
		stored += int(n)
		if f.RecordedAt.After(latest) {
			latest = f.RecordedAt
		}
	}

	if !latest.IsZero() {
		if err := s.queries.TouchTrackingDevice(ctx, dm_svc_db.TouchTrackingDeviceParams{
			SeenAt: ts(latest),
			ID:     device.ID,
		}); err != nil {
			return stored, err
		}
	}
	return stored, nil
}

func rejectReason(f Fix, now time.Time) string {
	switch {
	case f.Latitude < -90 || f.Latitude > 90 || f.Longitude < -180 || f.Longitude > 180:
		return "coordinates out of range"
	case f.Latitude == 0 && f.Longitude == 0:
		return "null island"
	case f.RecordedAt.After(now.Add(maxFutureSkew)):
		return "timestamp in the future"
	case f.RecordedAt.Before(now.Add(-maxFixAge)):
		return "timestamp too old"
	case f.AccuracyM != nil && *f.AccuracyM > maxAccuracyM:
		return "accuracy too low"
	}
	return ""
}

func (s *Service) LivePositions(ctx context.Context) ([]dm_svc_db.ListLiveDriverPositionsRow, error) {
	return s.queries.ListLiveDriverPositions(ctx)
}

func (s *Service) DriverLocations(ctx context.Context, driverID int32, from, to time.Time) ([]dm_svc_db.ListDriverLocationsRow, error) {
	return s.queries.ListDriverLocations(ctx, dm_svc_db.ListDriverLocationsParams{
		DriverID:  driverID,
		FromTime:  ts(from),
		ToTime:    ts(to),
		MaxPoints: defaultMaxPoint,
	})
}

func (s *Service) TripLocations(ctx context.Context, tripID int32) ([]dm_svc_db.ListTripLocationsRow, error) {
	return s.queries.ListTripLocations(ctx, dm_svc_db.ListTripLocationsParams{
		TripID:    &tripID,
		MaxPoints: defaultMaxPoint,
	})
}

// RunRetention deletes points older than retention once a day until ctx ends.
func (s *Service) RunRetention(ctx context.Context, retention time.Duration) {
	purge := func() {
		n, err := s.queries.DeleteDriverLocationsBefore(ctx, ts(s.now().Add(-retention)))
		if err != nil {
			s.logger.Error("gps retention purge failed", "error", err)
			return
		}
		if n > 0 {
			s.logger.Info("gps retention purge", "deleted", n)
		}
	}
	purge()
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			purge()
		}
	}
}

// Tokens avoid look-alike characters because drivers may type them by hand.
const tokenAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"
const tokenLength = 16 // ~79 bits

func newDeviceToken() (string, error) {
	b := make([]byte, tokenLength)
	max := big.NewInt(int64(len(tokenAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = tokenAlphabet[n.Int64()]
	}
	return string(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ts(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
