package driver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Robotech-Org/driver-management/internal/config"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/domain"
	"github.com/google/uuid"
	"github.com/greatcloak/decimal"
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

func (s *DriverService) GetDriverDetail(ctx context.Context, driverID int32) (dm_svc_db.GetDriverDetailRow, error) {
	s.logger.Info("Getting driver detail", "driverID", driverID)
	return s.queries.GetDriverDetail(ctx, driverID)
}

func (s *DriverService) DeleteDriver(ctx context.Context, driverID int32) error {
	s.logger.Info("Deleting driver", "driverID", driverID)
	return s.queries.DeleteDriver(ctx, driverID)
}

func (s *DriverService) GetDriverLedger(ctx context.Context, driverID int32) ([]dm_svc_db.DriverLedger, error) {
	s.logger.Info("Getting driver ledger", "driverID", driverID)
	return s.queries.GetDriverLedger(ctx, driverID)
}

func (s *DriverService) AddPayrollService(ctx context.Context, driverID int32, req *domain.DriverLedger) error {
	s.logger.Info("Adding payroll", "driverID", driverID)
	_, err := s.queries.DriverLedger(ctx, dm_svc_db.DriverLedgerParams{
		DriverID:      req.DriverID,
		EntryType:     req.EntryType,
		Amount:        req.Amount,
		RelatedTripID: nil,
		Notes:         &req.Notes,
	})
	if err != nil {
		return err
	}

	txnData, err := s.queries.CreateTransaction(ctx, dm_svc_db.CreateTransactionParams{
		Category: dm_svc_db.TransactionCategoryEXPENSEDRIVERPAYROLL,
		Amount:   req.Amount,
	})
	if err != nil {
		return err
	}
	s.logger.Info("Transaction created", "txnID", txnData.ID)
	return nil
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

func (s *DriverService) CreateTrip(ctx context.Context, trip dm_svc_db.CreateTripParams) error {
	s.logger.Info("Creating trip", "trip", trip)
	tripReturning, err := s.queries.CreateTrip(ctx, dm_svc_db.CreateTripParams{
		DriverID:      trip.DriverID,
		TotalCost:     trip.TotalCost,
		ApprovedBy:    trip.ApprovedBy,
		Origin:        trip.Origin,
		Destination:   trip.Destination,
		DepartureTime: trip.DepartureTime,
		ArrivalTime:   trip.ArrivalTime,
		Items:         trip.Items,
	})
	if err != nil {
		return fmt.Errorf("failed to create trip: %w", err)
	}

	// for _, expense := range trip.ExpenseIDs {
	// 	if err := s.queries.AddExpensesToTrip(ctx, dm_svc_db.AddExpensesToTripParams{
	// 		TripID:     tripReturning.ID,
	// 		ExpensesID: expense,
	// 	}); err != nil {
	// 		return fmt.Errorf("failed to add expense: %w", err)
	// 	}
	// }

	tripID := tripReturning.ID
	notes := "Initial trip cost"

	var handlerName string
	userProfile, err := s.queries.UserProfile(ctx, trip.ApprovedBy)
	if err != nil {
		s.logger.Error("failed to get user profile for handler name", "error", err)
	} else {
		handlerName = userProfile.FirstName + " " + userProfile.LastName
	}

	_, err = s.queries.CreateTransaction(ctx, dm_svc_db.CreateTransactionParams{
		TripID:        &tripID,
		Amount:        trip.TotalCost,
		Category:      dm_svc_db.TransactionCategoryEXPENSETRIP,
		PaymentMethod: "system",
		HandledBy: pgtype.UUID{
			Bytes: tripReturning.ApprovedBy,
			Valid: true,
		},
		HandlerName: &handlerName,
		Notes:       &notes,
	})
	if err != nil {
		return fmt.Errorf("failed to create transaction for trip cost: %w", err)
	}

	s.logger.Info("Successfully created", "trip", tripReturning)
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

func (s *DriverService) GetTripDetailService(ctx context.Context, tripID int32) (dm_svc_db.GetTripDetailRow, error) {
	s.logger.Info("Getting trip detail", "tripID", tripID)
	return s.queries.GetTripDetail(ctx, tripID)
}

func (s *DriverService) UploadReceiptService(ctx context.Context, tripID int32, receiptURL string) error {
	// TODO: Implement image upload to a cloud storage provider like Cloudinary or S3.
	// The `receiptURL` parameter will be the URL returned by the storage provider.

	return s.queries.UpdateTripReceiptURL(ctx, dm_svc_db.UpdateTripReceiptURLParams{
		ID:         tripID,
		ReceiptUrl: &receiptURL,
	})
}

func (s *DriverService) GetExpenseTypesService(ctx context.Context) ([]dm_svc_db.Expense, error) {
	s.logger.Info("Getting expense types")
	return s.queries.GetExpenseTypes(ctx)
}

func (s *DriverService) CreateExpenseService(ctx context.Context, name string) (dm_svc_db.Expense, error) {
	s.logger.Info("Creating expense", "expense", name)
	return s.queries.CreateExpenseType(ctx, name)
}

func (s *DriverService) UpdateExpenseService(ctx context.Context, expenseID int32, name string) (dm_svc_db.Expense, error) {
	s.logger.Info("Updating expense", "expenseID", expenseID, "name", name)
	return s.queries.UpdateExpenseType(ctx, dm_svc_db.UpdateExpenseTypeParams{
		ID:   expenseID,
		Name: name,
	})
}

func (s *DriverService) DeleteExpenseService(ctx context.Context, expenseID int32) error {
	s.logger.Info("Deleting expense", "expenseID", expenseID)
	return s.queries.DeleteExpenseType(ctx, expenseID)
}

func (s *DriverService) GetStatsService(ctx context.Context) (dm_svc_db.GetDashboardCountsRow, error) {
	s.logger.Info("Getting stats")
	return s.queries.GetDashboardCounts(ctx)
}

func (s *DriverService) CreateTripExpenseService(
	ctx context.Context,
	tripExpense domain.TripExpense,
) (dm_svc_db.TripExpense, error) {

	return s.queries.CreateTripExpense(ctx, dm_svc_db.CreateTripExpenseParams{
		TripID:     int32(tripExpense.TripID),
		ExpensesID: int32(tripExpense.ExpenseID),
		Cost:       tripExpense.Cost, // float64 now matches generated type
		ReceiptUrl: &tripExpense.ImageURL,
		Remark:     &tripExpense.Remark,
	})
}

func (s *DriverService) RegisterAdvancePaymentService(ctx context.Context, transaction domain.Transaction, userID string) (dm_svc_db.Transaction, error) {
	userUUID, _ := uuid.Parse(userID)
	tripID := int32(transaction.TripID)
	txRecord, err := s.queries.CreateTransaction(ctx, dm_svc_db.CreateTransactionParams{
		TripID:        &tripID,
		Amount:        transaction.Amount,
		Category:      dm_svc_db.TransactionCategoryINCOMEDEPOSIT,
		PaymentMethod: transaction.Method,
		HandledBy: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		}, // The missing piece
		HandlerName: &transaction.ReceiverName, // You should fetch real name
		Notes:       &transaction.Notes,
	})
	if err != nil {
		return dm_svc_db.Transaction{}, err
	}

	paidAmountDecimal := decimal.NewFromFloat(transaction.Amount)

	err = s.queries.UpdateTripPaymentState(ctx, dm_svc_db.UpdateTripPaymentStateParams{
		ID: tripID,
		PaidAmount: decimal.NullDecimal{
			Decimal: paidAmountDecimal,
			Valid:   true,
		},
	})
	if err != nil {
		return dm_svc_db.Transaction{}, fmt.Errorf("failed to update trip status: %v", err)
	}
	return txRecord, nil
}

func (s *DriverService) RegisterFinalPaymentService(ctx context.Context, transaction domain.Transaction, userID string) (dm_svc_db.Transaction, error) {
	userUUID, _ := uuid.Parse(userID)
	tripID := int32(transaction.TripID)
	txRecord, err := s.queries.CreateTransaction(ctx, dm_svc_db.CreateTransactionParams{
		TripID:        &tripID,
		Amount:        transaction.Amount,
		Category:      dm_svc_db.TransactionCategoryINCOMEFINAL,
		PaymentMethod: transaction.Method,
		HandledBy: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		HandlerName: &transaction.ReceiverName,
		Notes:       &transaction.Notes,
	})
	if err != nil {
		return dm_svc_db.Transaction{}, err
	}

	paidAmountDecimal := decimal.NewFromFloat(transaction.Amount)

	err = s.queries.UpdateTripPaymentState(ctx, dm_svc_db.UpdateTripPaymentStateParams{
		ID: tripID,
		PaidAmount: decimal.NullDecimal{
			Decimal: paidAmountDecimal,
			Valid:   true,
		},
	})
	if err != nil {
		return dm_svc_db.Transaction{}, fmt.Errorf("failed to update trip status: %v", err)
	}

	return txRecord, nil
}

func (s *DriverService) RegisterMaintenanceService(ctx context.Context, maintenance domain.MaintenanceAddition) (dm_svc_db.MaintenanceLog, error) {
	var maintenanceType dm_svc_db.MaintenanceType
	switch maintenance.Type {
	case "CORRECTIVE":
		maintenanceType = dm_svc_db.MaintenanceTypeCORRECTIVE
	case "PREVENTIVE":
		maintenanceType = dm_svc_db.MaintenanceTypePREVENTIVE
	default:
		return dm_svc_db.MaintenanceLog{}, fmt.Errorf("invalid maintenance type: %s", maintenance.Type)
	}
	maintenanceInfo, err := s.queries.CreateMaintenance(ctx, dm_svc_db.CreateMaintenanceParams{
		CarID:           maintenance.CarID,
		MaintenanceType: maintenanceType,
		Cost:            maintenance.Cost,
		MaintenanceName: maintenance.Name,
		Description:     maintenance.Description,
	})
	if err != nil {
		return dm_svc_db.MaintenanceLog{}, fmt.Errorf("failed to create maintenance record: %v", err)
	}

	_, err = s.UpdateCarStatus(ctx, maintenance.CarID, domain.Maintenance)
	if err != nil {
		return dm_svc_db.MaintenanceLog{}, fmt.Errorf("failed to update car status: %v", err)
	}

	txRecord, err := s.queries.CreateTransaction(ctx, dm_svc_db.CreateTransactionParams{
		Category: dm_svc_db.TransactionCategoryEXPENSEMAINTENANCE,
		Amount:   maintenance.Cost,
	})
	if err != nil {
		return dm_svc_db.MaintenanceLog{}, fmt.Errorf("failed to create transaction record: %v", err)
	}
	s.logger.Info("Transaction created successfully", "transaction_id", txRecord.ID)

	return maintenanceInfo, nil
}

func toFloat64(v interface{}) (float64, error) {
	switch value := v.(type) {
	case pgtype.Numeric:
		f, err := value.Float64Value()
		if err != nil {
			return 0, err
		}
		return f.Float64, nil
	case decimal.NullDecimal:
		if value.Valid {
			return value.Decimal.InexactFloat64(), nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("invalid type assertion: %T", v)
	}
}

func (s *DriverService) GetFinancialSummaryService(ctx context.Context) (domain.FinancialSummary, error) {
	summary, err := s.queries.GetFinancialSummary(ctx)
	if err != nil {
		return domain.FinancialSummary{}, err
	}
	s.logger.Info("Financial summary result", "summary", summary)

	totalRevenue, err := toFloat64(summary.TotalRevenue)
	if err != nil {
		return domain.FinancialSummary{}, err
	}

	maintenanceCost, err := toFloat64(summary.MaintenanceCost)
	if err != nil {
		return domain.FinancialSummary{}, err
	}

	driverPayroll, err := toFloat64(summary.DriverPayroll)
	if err != nil {
		return domain.FinancialSummary{}, err
	}

	tripOverheads, err := toFloat64(summary.TripOverheads)
	if err != nil {
		return domain.FinancialSummary{}, err
	}

	totalExpenses := maintenanceCost + driverPayroll + tripOverheads
	netProfit := totalRevenue - totalExpenses
	profitMargin := 0.0
	if totalRevenue > 0 {
		profitMargin = (netProfit / totalRevenue) * 100
	}

	return domain.FinancialSummary{
		Revenue: totalRevenue,
		Expenses: domain.ExpenseSummary{
			Maintenance:   maintenanceCost,
			DriverPayroll: driverPayroll,
			TripOverheads: tripOverheads,
		},
		NetProfit:    netProfit,
		ProfitMargin: profitMargin,
	}, nil
}

func (s *DriverService) GetTripProfitabilityService(ctx context.Context, tripID int32) (domain.TripProfitability, error) {
	profitability, err := s.queries.GetTripProfitability(ctx, tripID)
	if err != nil {
		return domain.TripProfitability{}, err
	}

	totalExpenses, err := toFloat64(profitability.TotalExpenses)
	if err != nil {
		return domain.TripProfitability{}, err
	}

	totalRevenue, err := toFloat64(profitability.TotalRevenue)
	if err != nil {
		return domain.TripProfitability{}, err
	}
	netProfit := totalRevenue - totalExpenses

	return domain.TripProfitability{
		TripID:        tripID,
		TotalRevenue:  totalRevenue,
		TotalExpenses: totalExpenses,
		NetProfit:     netProfit,
	}, nil
}

func (s *DriverService) RegisterDriverPaymentService(ctx context.Context, payment domain.DriverPayment, userID string) (dm_svc_db.DriverLedger, error) {
	driverID := int32(payment.DriverID)
	txRecord, err := s.queries.DriverLedger(ctx, dm_svc_db.DriverLedgerParams{
		DriverID:      driverID,
		EntryType:     "sometype",
		RelatedTripID: nil,
		Amount:        payment.Amount,
		Notes:         &payment.Notes,
	})
	if err != nil {
		return dm_svc_db.DriverLedger{}, err
	}
	return txRecord, nil
}

func (s *DriverService) RegisterTripOverheadService(ctx context.Context, overhead domain.TripOverhead, userID string) (dm_svc_db.Transaction, error) {
	userUUID, _ := uuid.Parse(userID)
	tripID := int32(overhead.TripID)
	txRecord, err := s.queries.CreateTransaction(ctx, dm_svc_db.CreateTransactionParams{
		TripID:        &tripID,
		Amount:        overhead.Amount,
		Category:      dm_svc_db.TransactionCategory("EXPENSE_TRIP"),
		PaymentMethod: overhead.Method,
		HandledBy: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		HandlerName: &overhead.ReceiverName,
		Notes:       &overhead.Notes,
	})
	if err != nil {
		return dm_svc_db.Transaction{}, err
	}
	return txRecord, nil
}
