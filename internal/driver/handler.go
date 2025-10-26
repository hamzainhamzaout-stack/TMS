package driver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Robotech-Org/driver-management/internal/data/dm_svc_db"
	"github.com/Robotech-Org/driver-management/internal/domain"
	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/go-chi/chi/v5"
)

type DriverHandler struct {
	logger  *slog.Logger
	service DriverService
}

func NewDriverHandler(logger *slog.Logger, service DriverService) *DriverHandler {
	return &DriverHandler{
		logger:  logger,
		service: service,
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
	var driverData domain.Driver
	if err := json.NewDecoder(r.Body).Decode(&driverData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	driver, err := h.service.UpdateDriverStatus(r.Context(), &driverData)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusInternalServerError, w, nil)
		return
	}
	data := map[string]string{
		"driver": driver.FirstName,
	}
	responses.SuccessJSONResponse("updated successfully", data, http.StatusOK, w)
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
	if err := h.service.ChangeCarDriver(r.Context(), &changeDriver); err != nil {
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

func (h *DriverHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var tripData dm_svc_db.CreateTripParams
	if err := json.NewDecoder(r.Body).Decode(&tripData); err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}
	if err := h.service.CreateTrip(r.Context(), &tripData); err != nil {
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
