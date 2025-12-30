// core/pkg/response/response.go
package response

import (
	"context"
	"encoding/json"
	"net/http"

	"apprun/pkg/logger"

	"github.com/go-chi/chi/v5/middleware"
)

var log logger.Logger

func init() {
	// Initialize with NopLogger by default
	// In production, this should be configured externally via SetLogger
	log = &logger.NopLogger{}
}

// SetLogger allows external configuration of the logger
func SetLogger(l logger.Logger) {
	if l != nil {
		log = l
	}
}

type Response struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type PaginationInfo struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

type ListData struct {
	Items      interface{}     `json:"items"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// getRequestID extracts request ID from context if available
func getRequestID(ctx context.Context) string {
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		return reqID
	}
	return ""
}

func Success(w http.ResponseWriter, data interface{}) {
	SuccessWithRequest(w, nil, data)
}

func SuccessWithRequest(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := Response{
		Success: true,
		Code:    200,
		Data:    data,
	}
	if r != nil {
		resp.RequestID = getRequestID(r.Context())
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode success response", logger.Field{Key: "error", Value: err})
	}
}

func Error(w http.ResponseWriter, code int, errCode, message string) {
	ErrorWithRequest(w, nil, code, errCode, message)
}

func ErrorWithRequest(w http.ResponseWriter, r *http.Request, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := Response{
		Success: false,
		Code:    code,
		Error: &ErrorInfo{
			Code:    errCode,
			Message: message,
		},
	}
	if r != nil {
		resp.RequestID = getRequestID(r.Context())
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode error response",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "status_code", Value: code},
			logger.Field{Key: "error_code", Value: errCode})
	}
}

func List(w http.ResponseWriter, items interface{}, pagination *PaginationInfo) {
	ListWithRequest(w, nil, items, pagination)
}

func ListWithRequest(w http.ResponseWriter, r *http.Request, items interface{}, pagination *PaginationInfo) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	listData := ListData{
		Items:      items,
		Pagination: pagination,
	}

	resp := Response{
		Success: true,
		Code:    200,
		Data:    listData,
	}
	if r != nil {
		resp.RequestID = getRequestID(r.Context())
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode list response", logger.Field{Key: "error", Value: err})
	}
}

func Created(w http.ResponseWriter, data interface{}, location string) {
	CreatedWithRequest(w, nil, data, location)
}

func CreatedWithRequest(w http.ResponseWriter, r *http.Request, data interface{}, location string) {
	w.Header().Set("Content-Type", "application/json")
	if location != "" {
		w.Header().Set("Location", location)
	}
	w.WriteHeader(http.StatusCreated)

	resp := Response{
		Success: true,
		Code:    201,
		Data:    data,
	}
	if r != nil {
		resp.RequestID = getRequestID(r.Context())
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode created response",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "location", Value: location})
	}
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func ValidationError(w http.ResponseWriter, field, message string) {
	ValidationErrorWithRequest(w, nil, field, message)
}

func ValidationErrorWithRequest(w http.ResponseWriter, r *http.Request, field, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)

	var details interface{}
	if field != "" {
		details = map[string]string{"field": field}
	}

	resp := Response{
		Success: false,
		Code:    422,
		Error: &ErrorInfo{
			Code:    ErrCodeInvalidParam,
			Message: message,
			Details: details,
		},
	}
	if r != nil {
		resp.RequestID = getRequestID(r.Context())
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode validation error response",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "field", Value: field})
	}
}
