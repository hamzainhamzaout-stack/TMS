package responses

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Pagination struct {
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
	Total  int64 `json:"total"`
}

func GetPaginationParams(r *http.Request) (offset int, limit int) {
	defaultLimit := 10
	defaultOffset := 0

	limit = defaultLimit
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	offset = defaultOffset
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	return offset, limit
}

type PaginatedResponse struct {
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
	Count      int64       `json:"count"`
}

type SuccessResponse struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	StatusCode int         `json:"status_code"`
}

func (r *PaginatedResponse) SetData(data interface{}, total int64) {
	if data == nil {
		r.Data = []interface{}{}
	} else {
		r.Data = data
	}

	r.Count = total
}

func SuccessJSONResponse(message string, data interface{}, statusCode int, w http.ResponseWriter) {
	response := SuccessResponse{
		Message:    message,
		Data:       data,
		StatusCode: statusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func PaginatedJSONResponse(message string, data interface{}, total int64, statusCode int, w http.ResponseWriter) {
	response := PaginatedResponse{
		Message:    message,
		StatusCode: statusCode,
	}

	response.SetData(data, total)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

