# errors - Business Error Wrapper Framework

A structured error handling framework for apprun with error codes, context, and HTTP mapping.

## Quick Start

```go
import "apprun/pkg/errors"

// Create error
err := errors.New(errors.ErrCodeNotFound, "User not found")

// Add context
err.WithContext(errors.ContextKeyUserID, "user123")

// Check type
if errors.IsNotFound(err) {
    // Handle...
}

// Map to HTTP status
status := httpmap.ToHTTPStatus(err) // 404
```

## Error Code Format

`<MODULE>_<CATEGORY>_<NAME>_<NNN>`

Example: `AUTH_VAL_INVALID_EMAIL_001`

**Categories**: VAL (validation), RES (resource), AUTH (auth), PERM (permission), BIZ (business), SYS (system)

## Core Functions

```go
// Create
errors.New(code, message)
errors.Newf(code, format, args...)

// Wrap
errors.Wrap(err, code, message)
errors.Wrapf(err, code, format, args...)

// Check
errors.IsValidation(err)
errors.IsNotFound(err)
errors.IsAuth(err)
errors.IsSystem(err)

// HTTP mapping
httpmap.ToHTTPStatus(err)
```

## Complete Example

```go
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    
    // Validate
    if userID == "" {
        err := errors.New(errors.ErrCodeInvalidParam, "User ID required").
            WithContext(errors.ContextKeyRequestID, r.Context().Value("request_id"))
        h.writeError(w, err)
        return
    }
    
    // Query
    user, err := h.userRepo.GetByID(userID)
    if err != nil {
        appErr := errors.Wrap(err, errors.ErrCodeInternalError, "Failed to get user").
            WithContext(errors.ContextKeyUserID, userID)
        h.writeError(w, appErr)
        return
    }
    
    if user == nil {
        err := errors.New(errors.ErrCodeNotFound, "User not found").
            WithContext(errors.ContextKeyUserID, userID)
        h.writeError(w, err)
        return
    }
    
    json.NewEncoder(w).Encode(user)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
    status := httpmap.ToHTTPStatus(err)
    w.WriteHeader(status)
    
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error":   appErr.Code,
            "message": appErr.Message,
        })
    }
}
```

## HTTP Status Mapping

| Category | HTTP Status | Code |
|----------|-------------|------|
| VAL | 400 Bad Request | Validation errors |
| RES | 404 Not Found | Resource not found (if code contains "NOT_FOUND") |
| RES | 400 Bad Request | Other resource errors |
| AUTH | 401 Unauthorized | Authentication errors |
| PERM | 403 Forbidden | Permission errors |
| BIZ | 422 Unprocessable Entity | Business logic errors |
| SYS | 500 Internal Server Error | System errors |

## Package Structure

```
pkg/errors/
├── errors.go       # Core types and functions
├── codes.go        # Error code definitions (114+ codes, 13 modules)
├── httpmap/        # HTTP status mapping
│   └── httpmap.go
└── README.md
```

## Adding New Error Codes

Edit `codes.go`:

```go
const (
    ErrCodeYourNewError = "MODULE_CATEGORY_NAME_001"
)
```

Follow naming: `<MODULE>_<CATEGORY>_<NAME>_<NNN>`

## Features

- ✅ Structured error codes (114+ predefined)
- ✅ Error chaining (Go 1.13+ compatible)
- ✅ Type-safe context keys
- ✅ HTTP status mapping
- ✅ Helper functions (IsValidation, IsNotFound, etc.)
- ✅ 100% test coverage

## Documentation

See package godoc for detailed API documentation.
