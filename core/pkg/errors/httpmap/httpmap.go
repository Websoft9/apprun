// Package httpmap provides HTTP status code mapping for AppError types.
//
// Example usage:
//
//	func (h *Handler) HandleError(w http.ResponseWriter, err error) {
//	    status := httpmap.ToHTTPStatus(err)
//	    w.WriteHeader(status)
//	    // Write error response...
//	}
//
// HTTP Status Code Mapping:
//   - VAL (Validation)    -> 400 Bad Request
//   - RES (Resource)      -> 404 Not Found (if contains "NOT_FOUND"), else 400
//   - AUTH (Authentication) -> 401 Unauthorized
//   - PERM (Permission)   -> 403 Forbidden
//   - BIZ (Business)      -> 422 Unprocessable Entity
//   - SYS (System)        -> 500 Internal Server Error
//   - Unknown/nil         -> 500 Internal Server Error

package httpmap

import (
	"net/http"
	"strings"

	"apprun/pkg/errors"
)

// ToHTTPStatus maps an AppError to the appropriate HTTP status code
func ToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Try to convert to AppError
	appErr, ok := err.(*errors.AppError)
	if !ok {
		// Not an AppError, return 500 as default
		return http.StatusInternalServerError
	}

	// Map based on category
	category := appErr.Category()
	switch category {
	case errors.CategoryValidation:
		return http.StatusBadRequest // 400

	case errors.CategoryResource:
		// Check for NOT_FOUND in code
		if strings.Contains(appErr.Code, "NOT_FOUND") {
			return http.StatusNotFound // 404
		}
		return http.StatusBadRequest // 400 for other resource errors

	case errors.CategoryAuth:
		return http.StatusUnauthorized // 401

	case errors.CategoryPermission:
		return http.StatusForbidden // 403

	case errors.CategoryBusiness:
		return http.StatusUnprocessableEntity // 422

	case errors.CategorySystem:
		return http.StatusInternalServerError // 500

	default:
		return http.StatusInternalServerError // 500
	}
}
