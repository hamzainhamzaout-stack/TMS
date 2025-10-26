package driver

import (
	"context"
	"log/slog"

	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/domain"
)

type DriverService struct {
	logger  *slog.Logger
	queries dm_svc_db.Queries
	cfg     config.Config
}

func NewDriverService(logger *slog.Logger, queries dm_svc_db.Queries, cfg config.Config) *DriverService {
	return &DriverService{
		logger:  logger,
		queries: queries,
		cfg:     cfg,
	}
}

func (s *DriverService) CreateDriver(ctx context.Context, driver *domain.Driver) (dm_svc_db.Driver, error) {
	s.logger.Info("Creating driver", "driver", driver)
	return s.queries.CreateDriver(ctx, dm_svc_db.CreateDriverParams{
		FirstName:     driver.FirstName,
		LastName:      driver.LastName,
		PhoneNumber:   driver.PhoneNumber,
		LicenseNumber: driver.Licenses,
	})
}

func (s *DriverService) UpdateDriverStatus(ctx context.Context, driver *domain.Driver) (dm_svc_db.Driver, error) {
	s.logger.Info("Updating driver status", "driver", driver)
	return s.queries.UpdateDriver(ctx, dm_svc_db.UpdateDriverParams{
		ID:            int32(driver.ID),
		FirstName:     driver.FirstName,
		LastName:      driver.LastName,
		PhoneNumber:   driver.PhoneNumber,
		LicenseNumber: driver.Licenses,
	})
}

func (s *DriverService) ListDrivers(ctx context.Context) ([]dm_svc_db.Driver, error) {
	s.logger.Info("Listing drivers")
	return s.queries.ListDrivers(ctx)
}

func (s *DriverService) DeleteDriver(ctx context.Context, driverID int32) error {
	s.logger.Info("Deleting driver", "driverID", driverID)
	return s.queries.DeleteDriver(ctx, driverID)
}

func (s *DriverService) ListCars(ctx context.Context) ([]dm_svc_db.Car, error) {
	s.logger.Info("Listing cars")
	return s.queries.ListCars(ctx)
}

func (s *DriverService) CreateCar(ctx context.Context, car *dm_svc_db.Car) (dm_svc_db.Car, error) {
	s.logger.Info("Creating car", "car", car)
	return s.queries.CreateCar(ctx, dm_svc_db.CreateCarParams{
		Make:         car.Make,
		Model:        car.Model,
		Year:         car.Year,
		Color:        car.Color,
		LicensePlate: car.LicensePlate,
	})
}

func (s *DriverService) UpdateCar(ctx context.Context, car *dm_svc_db.Car) (dm_svc_db.Car, error) {
	s.logger.Info("Updating car", "car", car)
	return s.queries.UpdateCar(ctx, dm_svc_db.UpdateCarParams{
		ID:           car.ID,
		Make:         car.Make,
		Model:        car.Model,
		Year:         car.Year,
		Color:        car.Color,
		LicensePlate: car.LicensePlate,
	})
}

func (s *DriverService) UpdateCarStatus(ctx context.Context, carID int32, carStatus domain.CarStatus) (dm_svc_db.Car, error) {
	s.logger.Info("Updating car status", "carStatus", carStatus)
	status := &carStatus
	return s.queries.UpdateCarStatus(ctx, dm_svc_db.UpdateCarStatusParams{
		ID:     carID,
		Status: (*string)(status),
	})
}

func (s *DriverService) DeleteCar(ctx context.Context, carID int32) error {
	s.logger.Info("Deleting car", "carID", carID)
	return s.queries.DeleteCar(ctx, carID)
}

func (s *DriverService) ChangeCarDriver(ctx context.Context, changeDriver *dm_svc_db.ChangeCarDriverParams) error {
	s.logger.Info("Changing car driver", "changeDriver", changeDriver)
	return s.queries.ChangeCarDriver(ctx, *changeDriver)
}

func (s *DriverService) CreateTrip(ctx context.Context, trip *dm_svc_db.CreateTripParams) error {
	s.logger.Info("Creating trip", "trip", trip)
	return s.queries.CreateTrip(ctx, *trip)
}

func (s *DriverService) ListTrips(ctx context.Context) ([]dm_svc_db.Trip, error) {
	s.logger.Info("Listing trips")
	return s.queries.GetTrips(ctx)
}

func (s *DriverService) UpdateTrip(ctx context.Context, trip *dm_svc_db.UpdateTripParams) (dm_svc_db.Trip, error) {
	s.logger.Info("Updating trip", "trip", trip)
	return s.queries.UpdateTrip(ctx, *trip)
}

func (s *DriverService) DeleteTrip(ctx context.Context, tripID int32) error {
	s.logger.Info("Deleting trip", "tripID", tripID)
	return s.queries.DeleteTrip(ctx, tripID)
}

func (s *DriverService) ChangeTripDriver(ctx context.Context, changeDriver *dm_svc_db.UpdateTripDriverParams) (dm_svc_db.Trip, error) {
	s.logger.Info("Changing trip driver", "changeDriver", changeDriver)
	return s.queries.UpdateTripDriver(ctx, *changeDriver)
}
