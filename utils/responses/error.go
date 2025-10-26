package responses

import (
	"encoding/json"
	"net/http"

	validator "github.com/Robotech-Org/driver-management/utils/validators"
)

type ErrorResponse struct {
	ErrorMessage string                      `json:"message"`
	Errors       *validator.ValidationErrors `json:"errors,omitempty"`
	StatusCode   int                         `json:"status_code"`
}

func (e ErrorResponse) Error() string {
	return e.ErrorMessage
}

func ErrorJSONResponse(err error, statusCode int, w http.ResponseWriter, valErrs *validator.ValidationErrors) {
	response := ErrorResponse{
		ErrorMessage: err.Error(),
		StatusCode:   statusCode,
		Errors:       valErrs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
