// core/pkg/response/response.go
package response

import (
	"context"
	"encoding/json"
	stderrors "errors" // stdlib for errors.As
	"net/http"

	"apprun/pkg/errors"
	"apprun/pkg/errors/httpmap"
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

// AppError handles errors.AppError and maps to HTTP response
func AppError(w http.ResponseWriter, err error) {
	AppErrorWithRequest(w, nil, err)
}

// AppErrorWithRequest handles errors.AppError with request context
func AppErrorWithRequest(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		Success(w, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get HTTP status code from error
	statusCode := httpmap.ToHTTPStatus(err)
	w.WriteHeader(statusCode)

	// Build response
	resp := Response{
		Success: false,
		Code:    statusCode,
	}

	// Extract error details
	var appErr *errors.AppError
	if stderrors.As(err, &appErr) {
		resp.Error = &ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
		}

		// Add context as details if present
		if len(appErr.Context) > 0 {
			resp.Error.Details = appErr.Context
		}
	} else {
		// Fallback for non-AppError
		resp.Error = &ErrorInfo{
			Code:    errors.ErrCodeInternalError,
			Message: err.Error(),
		}
	}

	if r != nil {
		resp.RequestID = getRequestID(r.Context())
	}

	// Log system errors
	if errors.IsSystem(err) {
		log.Error("system error in response",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "status_code", Value: statusCode})
	}

	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		log.Error("failed to encode error response",
			logger.Field{Key: "error", Value: encErr},
			logger.Field{Key: "original_error", Value: err.Error()})
	}
}

// Error provides backward compatibility with old error code format
//
// Deprecated: Use AppError instead
func Error(w http.ResponseWriter, code int, errCode, message string) {
	ErrorWithRequest(w, nil, code, errCode, message)
}

// ErrorWithRequest provides backward compatibility
//
// Deprecated: Use AppErrorWithRequest instead
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

// ValidationError provides backward compatibility
//
// Deprecated: Use AppError with errors.New(errors.ErrCodeInvalidParam, ...) instead
func ValidationError(w http.ResponseWriter, field, message string) {
	ValidationErrorWithRequest(w, nil, field, message)
}

// ValidationErrorWithRequest provides backward compatibility
//
// Deprecated: Use AppErrorWithRequest instead
func ValidationErrorWithRequest(w http.ResponseWriter, r *http.Request, field, message string) {
	// Create AppError for validation
	err := errors.New(errors.ErrCodeInvalidParam, message)
	if field != "" {
		_ = err.WithContext("field", field)
	}

	AppErrorWithRequest(w, r, err)
}
