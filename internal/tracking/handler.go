package tracking

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Robotech-Org/driver-management/internal/auth"
	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	logger  *slog.Logger
	service *Service
}

func NewHandler(logger *slog.Logger, service *Service) *Handler {
	return &Handler{logger: logger, service: service}
}

// Ingest receives positions from Traccar Client. It is public: the device
// token in the payload is the credential.
func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	token, fixes, err := parseTraccarRequest(r)
	if err != nil {
		responses.ErrorJSONResponse(err, http.StatusBadRequest, w, nil)
		return
	}

	stored, err := h.service.Ingest(r.Context(), token, fixes)
	if err != nil {
		if errors.Is(err, ErrUnknownDevice) {
			responses.ErrorJSONResponse(err, http.StatusForbidden, w, nil)
			return
		}
		h.logger.Error("gps ingest failed", "error", err)
		responses.ErrorJSONResponse(errors.New("could not store location"), http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("ok", map[string]int{"received": len(fixes), "stored": stored}, http.StatusOK, w)
}

type deviceStatus struct {
	Enabled    bool       `json:"enabled"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

type enableResponse struct {
	deviceStatus
	DeviceID  string `json:"device_id"`
	ServerURL string `json:"server_url"`
}

func (h *Handler) EnableTracking(w http.ResponseWriter, r *http.Request) {
	driverID, ok := intParam(w, r, "id")
	if !ok {
		return
	}

	var createdBy pgtype.UUID
	if userID, ok := auth.GetUserIDFromContext(r.Context()); ok {
		if u, err := uuid.Parse(userID); err == nil {
			createdBy = pgtype.UUID{Bytes: u, Valid: true}
		}
	}

	token, device, err := h.service.EnableTracking(r.Context(), driverID, createdBy)
	if err != nil {
		if errors.Is(err, ErrDriverNotFound) {
			responses.ErrorJSONResponse(err, http.StatusNotFound, w, nil)
			return
		}
		h.logger.Error("enable tracking failed", "error", err)
		responses.ErrorJSONResponse(errors.New("could not enable tracking"), http.StatusInternalServerError, w, nil)
		return
	}

	responses.SuccessJSONResponse("tracking enabled", enableResponse{
		deviceStatus: deviceStatus{Enabled: true, CreatedAt: timePtr(device.CreatedAt)},
		DeviceID:     token,
		ServerURL:    trackingServerURL(r),
	}, http.StatusCreated, w)
}

func (h *Handler) DisableTracking(w http.ResponseWriter, r *http.Request) {
	driverID, ok := intParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.DisableTracking(r.Context(), driverID); err != nil {
		h.logger.Error("disable tracking failed", "error", err)
		responses.ErrorJSONResponse(errors.New("could not disable tracking"), http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("tracking disabled", deviceStatus{Enabled: false}, http.StatusOK, w)
}

func (h *Handler) TrackingStatus(w http.ResponseWriter, r *http.Request) {
	driverID, ok := intParam(w, r, "id")
	if !ok {
		return
	}
	device, err := h.service.TrackingStatus(r.Context(), driverID)
	if err != nil {
		h.logger.Error("tracking status failed", "error", err)
		responses.ErrorJSONResponse(errors.New("could not load tracking status"), http.StatusInternalServerError, w, nil)
		return
	}
	status := deviceStatus{Enabled: device != nil}
	if device != nil {
		status.CreatedAt = timePtr(device.CreatedAt)
		status.LastSeenAt = timePtr(device.LastSeenAt)
	}
	responses.SuccessJSONResponse("success", status, http.StatusOK, w)
}

func (h *Handler) LivePositions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.LivePositions(r.Context())
	if err != nil {
		h.logger.Error("live positions failed", "error", err)
		responses.ErrorJSONResponse(errors.New("could not load positions"), http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("success", rows, http.StatusOK, w)
}

// DriverLocations returns a driver's points between ?from= and ?to=
// (RFC 3339). Defaults to the last 24 hours.
func (h *Handler) DriverLocations(w http.ResponseWriter, r *http.Request) {
	driverID, ok := intParam(w, r, "id")
	if !ok {
		return
	}
	to := time.Now()
	from := to.Add(-24 * time.Hour)
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			responses.ErrorJSONResponse(errors.New("from must be RFC 3339"), http.StatusBadRequest, w, nil)
			return
		}
		from = t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			responses.ErrorJSONResponse(errors.New("to must be RFC 3339"), http.StatusBadRequest, w, nil)
			return
		}
		to = t
	}

	rows, err := h.service.DriverLocations(r.Context(), driverID, from, to)
	if err != nil {
		h.logger.Error("driver locations failed", "error", err)
		responses.ErrorJSONResponse(errors.New("could not load locations"), http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("success", rows, http.StatusOK, w)
}

func (h *Handler) TripLocations(w http.ResponseWriter, r *http.Request) {
	tripID, ok := intParam(w, r, "id")
	if !ok {
		return
	}
	rows, err := h.service.TripLocations(r.Context(), tripID)
	if err != nil {
		h.logger.Error("trip locations failed", "error", err)
		responses.ErrorJSONResponse(errors.New("could not load locations"), http.StatusInternalServerError, w, nil)
		return
	}
	responses.SuccessJSONResponse("success", rows, http.StatusOK, w)
}

// trackingServerURL is what the driver types into Traccar Client as the
// server URL. Set TRACKING_SERVER_URL to override the derived value.
func trackingServerURL(r *http.Request) string {
	if v := strings.TrimSpace(os.Getenv("TRACKING_SERVER_URL")); v != "" {
		return v
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/track"
}

func intParam(w http.ResponseWriter, r *http.Request, name string) (int32, bool) {
	n, err := strconv.ParseInt(chi.URLParam(r, name), 10, 32)
	if err != nil || n <= 0 {
		responses.ErrorJSONResponse(errors.New("invalid "+name), http.StatusBadRequest, w, nil)
		return 0, false
	}
	return int32(n), true
}

func timePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}
