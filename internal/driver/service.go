package driver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
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
	driverData, err := s.queries.CreateDriver(ctx, dm_svc_db.CreateDriverParams{
		FirstName:     driver.FirstName,
		LastName:      driver.LastName,
		PhoneNumber:   driver.PhoneNumber,
		LicenseNumber: driver.Licenses,
	})
	if err != nil {
		return dm_svc_db.Driver{}, errors.New("error creating driver")
	}
	return dm_svc_db.Driver{
		FirstName:     driverData.FirstName,
		LastName:      driverData.LastName,
		LicenseNumber: driverData.LicenseNumber,
		PhoneNumber:   driverData.PhoneNumber,
	}, nil
}

func (s *DriverService) UpdateDriverStatus(ctx context.Context, driverID int32, req *domain.Driver) (dm_svc_db.Driver, error) {
	s.logger.Info("Updating driver status", "driver", req.LastName)
	existing, err := s.queries.GetDriverByID(ctx, driverID)
	if err != nil {
		return dm_svc_db.Driver{}, err
	}

	if req.FirstName != "" {
		existing.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existing.LastName = req.LastName
	}
	if req.PhoneNumber != "" {
		existing.PhoneNumber = req.PhoneNumber
	}
	if req.Licenses != "" {
		existing.LicenseNumber = req.Licenses
	}

	return s.queries.UpdateDriver(ctx, dm_svc_db.UpdateDriverParams{
		ID:            existing.ID,
		FirstName:     existing.FirstName,
		LastName:      existing.LastName,
		PhoneNumber:   existing.PhoneNumber,
		LicenseNumber: existing.LicenseNumber,
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

func (s *DriverService) ChangeCarDriver(ctx context.Context, driverID int32, carID int32) error {
	s.logger.Info("Changing car driver", "driverID", driverID)
	return s.queries.ChangeCarDriver(ctx, dm_svc_db.ChangeCarDriverParams{
		CurrentDriverID: &driverID,
		ID:              carID,
	})
}

func (s *DriverService) CreateTrip(ctx context.Context, trip domain.CreateTripParams) error {
	s.logger.Info("Creating trip", "trip", trip)

	totalCost := pgtype.Numeric{}
	if err := totalCost.Scan(strconv.FormatFloat(trip.TotalCost, 'f', 2, 64)); err != nil {
		return fmt.Errorf("failed to convert total_cost: %w", err)
	}
	tripReturning, err := s.queries.CreateTrip(ctx, dm_svc_db.CreateTripParams{
		DriverID:   trip.DriverID,
		TotalCost:  totalCost,
		ApprovedBy: trip.ApprovedBy,
	})
	if err != nil {
		return fmt.Errorf("failed to create trip: %w", err)
	}

	for _, expense := range trip.ExpenseIDs {
		if err := s.queries.AddExpensesToTrip(ctx, dm_svc_db.AddExpensesToTripParams{
			TripID:     tripReturning.ID,
			ExpensesID: expense,
		}); err != nil {
			return fmt.Errorf("failed to add expense: %w", err)
		}
	}

	return nil
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
