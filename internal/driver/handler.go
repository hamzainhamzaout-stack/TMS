package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Robotech-Org/driver-management/internal/auth"
	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/domain"
	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/greatcloak/decimal"
	"github.com/jackc/pgx/v5/pgtype"
)

type DriverHandler struct {
	logger     *slog.Logger
	service    DriverService
	cloudinary *cloudinary.Cloudinary
}

func NewDriverHandler(logger *slog.Logger, service DriverService, cloudinary *cloudinary.Cloudinary) *DriverHandler {
	return &DriverHandler{
		logger:     logger,
		service:    service,
		cloudinary: cloudinary,
	}
}

func (h *DriverHandler) CreateDriver(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating driver")
	var driverData domain.Driver
	if err := json.NewDecoder(r.Body).Decode(&driverData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	driver, err := h.service.CreateDriver(r.Context(), &driverData)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]string{
		"driver": driver.FirstName,
	}
	responses.SuccessJSONResponse("created successfully", data, http.StatusCreated, w)
}

func (h *DriverHandler) ListDrivers(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Listing drivers")
	drivers, err := h.service.ListDrivers(r.Context())
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]interface{}{
		"drivers": drivers,
	}
	responses.SuccessJSONResponse("list successfully", data, http.StatusOK, w)
}

func (h *DriverHandler) UpdateDriver(w http.ResponseWriter, r *http.Request) {
	var req domain.Driver
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	updated, err := h.service.UpdateDriverStatus(r.Context(), int32(id), &req)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	data := map[string]string{"driver": updated.FirstName}
	responses.SuccessJSONResponse("updated successfully", data, http.StatusOK, w)
}

func (h *DriverHandler) GetDriverDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	driver, err := h.service.GetDriverDetail(r.Context(), int32(id))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("driver detail", driver, http.StatusOK, w)
}

func (h *DriverHandler) DeleteDriver(w http.ResponseWriter, r *http.Request) {
	driverID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	err = h.service.DeleteDriver(r.Context(), int32(driverID))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("deleted successfully", nil, http.StatusOK, w)
}

func (h *DriverHandler) GetDriverLedger(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting driver ledger")
	driverID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	ledger, err := h.service.GetDriverLedger(r.Context(), int32(driverID))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]interface{}{
		"ledger": ledger,
	}
	responses.SuccessJSONResponse("ledger retrieved successfully", data, http.StatusOK, w)
}

func (h *DriverHandler) AddPayroll(w http.ResponseWriter, r *http.Request) {
	driverID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	var req domain.DriverLedger
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	err = h.service.AddPayrollService(r.Context(), int32(driverID), &req)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("added successfully", nil, http.StatusOK, w)
}

func (h *DriverHandler) ListCars(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Listing cars")
	cars, err := h.service.ListCars(r.Context())
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]interface{}{
		"cars": cars,
	}
	responses.SuccessJSONResponse("list successfully", data, http.StatusOK, w)
}

func (h *DriverHandler) CreateCar(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating cars")
	var carDetail dm_svc_db.Car
	if err := json.NewDecoder(r.Body).Decode(&carDetail); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	car, err := h.service.CreateCar(r.Context(), &carDetail)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]interface{}{
		"car": car,
	}
	responses.SuccessJSONResponse("created successfully", data, http.StatusCreated, w)
}

func (h *DriverHandler) UpdateCar(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Updating cars")
	var carDetail dm_svc_db.Car
	if err := json.NewDecoder(r.Body).Decode(&carDetail); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	carID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	carDetail.ID = int32(carID)
	car, err := h.service.UpdateCar(r.Context(), &carDetail)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]interface{}{
		"car": car,
	}
	responses.SuccessJSONResponse("updated successfully", data, http.StatusOK, w)
}

func (h *DriverHandler) UpdateCarStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Updating cars status")
	carID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	var carStatus domain.CarStatus
	if err := json.NewDecoder(r.Body).Decode(&carStatus); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	car, err := h.service.UpdateCarStatus(r.Context(), int32(carID), carStatus)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]interface{}{
		"car": car,
	}
	responses.SuccessJSONResponse("updated successfully", data, http.StatusOK, w)
}

func (h *DriverHandler) DeleteCar(w http.ResponseWriter, r *http.Request) {
	carID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	err = h.service.DeleteCar(r.Context(), int32(carID))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("deleted successfully", nil, http.StatusOK, w)
}

func (h *DriverHandler) ChangeCarDriver(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Changing car driver")
	var changeDriver dm_svc_db.ChangeCarDriverParams
	if err := json.NewDecoder(r.Body).Decode(&changeDriver); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	carID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	if changeDriver.CurrentDriverID == nil {
		responses.ErrorJSONResponse(errors.New("current_driver_id is required"), http.StatusBadRequest, w, nil)
		return
	}
	if err := h.service.ChangeCarDriver(r.Context(), *changeDriver.CurrentDriverID, int32(carID)); err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("updated successfully", nil, http.StatusOK, w)
}

func (h *DriverHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	trips, err := h.service.ListTrips(r.Context())
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("list successfully", trips, http.StatusOK, w)
}

func NewNullDecimalFromFloat(f float64) decimal.NullDecimal {
	return decimal.NullDecimal{
		Decimal: decimal.NewFromFloat(f),
		Valid:   true,
	}
}

type MyTime time.Time

func (t *MyTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	parsed, err := time.Parse("2006-01-02T15:04", s)
	if err != nil {
		return err
	}
	*t = MyTime(parsed)
	return nil
}
func (h *DriverHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		responses.ErrorJSONResponse(errors.New("authentication required"), http.StatusUnauthorized, w, nil)
		return
	}
	var req struct {
		DriverID      int32   `json:"driver_id"`
		TotalCost     float64 `json:"total_cost"`
		Origin        string  `json:"origin"`
		Destination   string  `json:"destination"`
		DepartureTime MyTime  `json:"departure_time"`
		ArrivalTime   MyTime  `json:"arrival_time"`
		Items         string  `json:"items"`
		TotalQuote    float64 `json:"total_quote"`
		ItemWeight    float64 `json:"item_weight"`
		CoDriverName  string  `json:"co_driver_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	approvedBy, err := uuid.Parse(userID)
	if err != nil {
		responses.ErrorJSONResponse(errors.New("invalid user ID in token"), http.StatusInternalServerError, w, nil)
		return
	}

	tripData := dm_svc_db.CreateTripParams{
		DriverID:    req.DriverID,
		TotalCost:   req.TotalCost,
		ApprovedBy:  approvedBy,
		Origin:      &req.Origin,
		Destination: &req.Destination,
		DepartureTime: pgtype.Timestamp{
			Time:  time.Time(req.DepartureTime),
			Valid: true,
		},
		ArrivalTime: pgtype.Timestamp{
			Time:  time.Time(req.ArrivalTime),
			Valid: true,
		},
		Items:        &req.Items,
		TotalQuote:   NewNullDecimalFromFloat(req.TotalCost),
		ItemWeight:   NewNullDecimalFromFloat(req.ItemWeight),
		CoDriverName: &req.CoDriverName,
	}
	if err := h.service.CreateTrip(r.Context(), tripData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("created successfully", nil, http.StatusOK, w)
}

func (h *DriverHandler) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	var tripData dm_svc_db.UpdateTripParams
	if err := json.NewDecoder(r.Body).Decode(&tripData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	trip, err := h.service.UpdateTrip(r.Context(), &tripData)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("updated successfully", trip, http.StatusOK, w)
}

func (h *DriverHandler) DeleteTrip(w http.ResponseWriter, r *http.Request) {
	tripID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	if err := h.service.DeleteTrip(r.Context(), int32(tripID)); err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("deleted successfully", nil, http.StatusOK, w)
}

func (h *DriverHandler) ChangeTripDriver(w http.ResponseWriter, r *http.Request) {
	var tripData dm_svc_db.UpdateTripDriverParams
	if err := json.NewDecoder(r.Body).Decode(&tripData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	trip, err := h.service.ChangeTripDriver(r.Context(), &tripData)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("changed successfully", trip, http.StatusOK, w)
}

func (h *DriverHandler) TripDetail(w http.ResponseWriter, r *http.Request) {

	tripID, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {

		h.logger.Error("Failed to parse trip ID", "error", err)

		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)

		return

	}

	trip, err := h.service.GetTripDetailService(r.Context(), int32(tripID))

	if err != nil {

		h.logger.Error("Failed to get trip detail", "error", err)

		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)

		return

	}

	responses.SuccessJSONResponse("Trip detail retrieved successfully", trip, http.StatusOK, w)

}

// func (h *DriverHandler) CreateTripExpenses(w http.ResponseWriter, r *http.Request) {
// 	h.logger.Info("creating trips and expenses")

// 	tripID, err := strconv.Atoi(chi.URLParam(r, "trip-id"))
// 	if err != nil {
// 		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
// 		return
// 	}

// 	var createTripExpense TripExpense
// }

func (h *DriverHandler) CreateTripExpenseHandler(w http.ResponseWriter, r *http.Request) {

	h.logger.Info("Uploading receipt")

	tripID, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	file, header, err := r.FormFile("receipt")
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	defer file.Close()

	h.logger.Info("File uploaded successfully", "filename", header.Filename, "size", header.Size)

	uploadResult, err := h.cloudinary.Upload.Upload(r.Context(), file, uploader.UploadParams{})
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	receiptURL := uploadResult.SecureURL
	// if err := h.service.UploadReceiptService(r.Context(), int32(tripID), receiptURL); err != nil {
	// 	responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
	// 	return
	// }

	expenseIDStr := r.FormValue("expense_id")
	if expenseIDStr == "" {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	costStr := r.FormValue("cost")
	if costStr == "" {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	cost, err := strconv.ParseFloat(costStr, 64)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	expenseID, err := strconv.Atoi(expenseIDStr)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	expenseTrip := domain.TripExpense{
		TripID:    tripID,
		ExpenseID: expenseID,
		Cost:      cost,
		ImageURL:  receiptURL,
	}
	res, err := h.service.CreateTripExpenseService(r.Context(), expenseTrip)
	responses.SuccessJSONResponse("receipt uploaded successfully", res.TripID, http.StatusOK, w)
}

func (h *DriverHandler) GetExpenseTypes(w http.ResponseWriter, r *http.Request) {

	h.logger.Info("Getting expense types")

	expenseTypes, err := h.service.GetExpenseTypesService(r.Context())
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("expense types retrieved successfully", expenseTypes, http.StatusOK, w)
}

type ExpenseType struct {
	Name string `json:"name"`
}

func (h *DriverHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating expense")
	var expenseData ExpenseType
	if err := json.NewDecoder(r.Body).Decode(&expenseData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	expense, err := h.service.CreateExpenseService(r.Context(), expenseData.Name)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	data := map[string]interface{}{
		"id": expense.ID,
	}

	responses.SuccessJSONResponse("expense created successfully", data, http.StatusCreated, w)
}

func (h *DriverHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Updating expense")
	var expenseData ExpenseType
	if err := json.NewDecoder(r.Body).Decode(&expenseData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	expenseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	expense, err := h.service.UpdateExpenseService(r.Context(), int32(expenseID), expenseData.Name)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	data := map[string]interface{}{
		"id": expense.ID,
	}

	responses.SuccessJSONResponse("expense updated successfully", data, http.StatusOK, w)
}

func (h *DriverHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Deleting expense")
	expenseID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	if err := h.service.DeleteExpenseService(r.Context(), int32(expenseID)); err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("expense deleted successfully", nil, http.StatusOK, w)
}

func (h *DriverHandler) GetStats(w http.ResponseWriter, r *http.Request) {

	h.logger.Info("Getting stats")

	stats, err := h.service.GetStatsService(r.Context())
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("stats retrieved successfully", stats, http.StatusOK, w)
}

func (h *DriverHandler) RegisterAdvancePayment(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Registering advance payments")

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		responses.ErrorJSONResponse(errors.New("authentication required"), http.StatusUnauthorized, w, nil)
		return
	}

	var req domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	transaction, err := h.service.RegisterAdvancePaymentService(r.Context(), req, userID)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("transaction created successfully", transaction, http.StatusOK, w)
}

func (h *DriverHandler) RegisterFinalPayment(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Registering final payments")

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		responses.ErrorJSONResponse(fmt.Errorf("unauthorized action"), http.StatusUnauthorized, w, nil)
		return
	}

	var req domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	transaction, err := h.service.RegisterFinalPaymentService(r.Context(), req, userID)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("transaction created successfully", transaction, http.StatusOK, w)
}

func (h *DriverHandler) RegisterMaintenance(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Registering maintenance")

	var req domain.MaintenanceAddition
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	maintenance, err := h.service.RegisterMaintenanceService(r.Context(), req)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("maintenance created successfully", maintenance, http.StatusOK, w)
}

func (h *DriverHandler) GetFinancialSummary(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting financial summary")

	summary, err := h.service.GetFinancialSummaryService(r.Context())
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("financial summary retrieved successfully", summary, http.StatusOK, w)
}

func (h *DriverHandler) GetTripProfitability(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting trip profitability")

	tripID, err := strconv.Atoi(chi.URLParam(r, "tripId"))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	profitability, err := h.service.GetTripProfitabilityService(r.Context(), int32(tripID))
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("trip profitability retrieved successfully", profitability, http.StatusOK, w)
}

func (h *DriverHandler) RegisterDriverPayment(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Registering driver payment")

	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		responses.ErrorJSONResponse(fmt.Errorf("unauthorized action"), http.StatusUnauthorized, w, nil)
		return
	}

	var req domain.DriverPayment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	payment, err := h.service.RegisterDriverPaymentService(r.Context(), req, userID)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("driver payment created successfully", payment, http.StatusOK, w)
}

func (h *DriverHandler) RegisterTripOverhead(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Registering trip overhead")

	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		responses.ErrorJSONResponse(fmt.Errorf("unauthorized action"), http.StatusUnauthorized, w, nil)
		return
	}

	var req domain.TripOverhead
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	overhead, err := h.service.RegisterTripOverheadService(r.Context(), req, userID)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("trip overhead created successfully", overhead, http.StatusOK, w)
}

func (h *DriverHandler) RegisterReturnPayment(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Registering return payment")

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		responses.ErrorJSONResponse(fmt.Errorf("unauthorized action"), http.StatusUnauthorized, w, nil)
		return
	}

	var req domain.ReturnPayment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	payment, err := h.service.RegisterReturnPaymentService(r.Context(), req, userID)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("return payment created successfully", payment, http.StatusOK, w)
}
