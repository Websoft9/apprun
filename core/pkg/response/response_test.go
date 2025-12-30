package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccess(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantCode int
	}{
		{
			name:     "success with data",
			data:     map[string]string{"id": "123", "name": "Test"},
			wantCode: 200,
		},
		{
			name:     "success with nil data",
			data:     nil,
			wantCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Success(w, tt.data)

			if w.Code != tt.wantCode {
				t.Errorf("Success() status code = %v, want %v", w.Code, tt.wantCode)
			}

			var resp Response
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if !resp.Success {
				t.Error("Success() success should be true")
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Success() response code = %v, want %v", resp.Code, tt.wantCode)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		errCode     string
		message     string
		wantCode    int
		wantErrCode string
	}{
		{
			name:        "bad request error",
			code:        400,
			errCode:     "VAL_INVALID_PARAM_001",
			message:     "Invalid parameter",
			wantCode:    400,
			wantErrCode: "VAL_INVALID_PARAM_001",
		},
		{
			name:        "not found error",
			code:        404,
			errCode:     "RES_NOT_FOUND_001",
			message:     "Resource not found",
			wantCode:    404,
			wantErrCode: "RES_NOT_FOUND_001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Error(w, tt.code, tt.errCode, tt.message)

			if w.Code != tt.wantCode {
				t.Errorf("Error() status code = %v, want %v", w.Code, tt.wantCode)
			}

			var resp Response
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if resp.Success {
				t.Error("Error() success should be false")
			}

			if resp.Error == nil {
				t.Fatal("Error() should have error info")
			}

			if resp.Error.Code != tt.wantErrCode {
				t.Errorf("Error() error code = %v, want %v", resp.Error.Code, tt.wantErrCode)
			}

			if resp.Error.Message != tt.message {
				t.Errorf("Error() error message = %v, want %v", resp.Error.Message, tt.message)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name       string
		items      interface{}
		pagination *PaginationInfo
		wantCode   int
		wantTotal  int
		wantPage   int
	}{
		{
			name: "valid list with pagination",
			items: []map[string]string{
				{"id": "1", "name": "Item 1"},
				{"id": "2", "name": "Item 2"},
			},
			pagination: &PaginationInfo{
				Total:      100,
				Page:       1,
				PageSize:   10,
				TotalPages: 10,
			},
			wantCode:  200,
			wantTotal: 100,
			wantPage:  1,
		},
		{
			name:       "empty list",
			items:      []interface{}{},
			pagination: &PaginationInfo{Total: 0, Page: 1, PageSize: 10, TotalPages: 0},
			wantCode:   200,
			wantTotal:  0,
			wantPage:   1,
		},
		{
			name: "list without pagination",
			items: []map[string]string{
				{"id": "1", "name": "Item 1"},
			},
			pagination: nil,
			wantCode:   200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			List(w, tt.items, tt.pagination)

			if w.Code != tt.wantCode {
				t.Errorf("List() status code = %v, want %v", w.Code, tt.wantCode)
			}

			var resp Response
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if !resp.Success {
				t.Error("List() success should be true")
			}

			if resp.Code != tt.wantCode {
				t.Errorf("List() response code = %v, want %v", resp.Code, tt.wantCode)
			}

			// Verify data structure
			data, ok := resp.Data.(map[string]interface{})
			if !ok {
				t.Fatal("List() data should be a map")
			}

			if _, exists := data["items"]; !exists {
				t.Error("List() data should contain 'items' field")
			}

			if tt.pagination != nil {
				pagination, exists := data["pagination"]
				if !exists {
					t.Error("List() data should contain 'pagination' field")
				}
				pagMap, ok := pagination.(map[string]interface{})
				if !ok {
					t.Fatal("pagination should be a map")
				}
				if int(pagMap["total"].(float64)) != tt.wantTotal {
					t.Errorf("pagination total = %v, want %v", pagMap["total"], tt.wantTotal)
				}
				if int(pagMap["page"].(float64)) != tt.wantPage {
					t.Errorf("pagination page = %v, want %v", pagMap["page"], tt.wantPage)
				}
			}
		})
	}
}

func TestCreated(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		location     string
		wantCode     int
		wantLocation string
	}{
		{
			name:         "created with location",
			data:         map[string]string{"id": "123", "name": "New Resource"},
			location:     "/api/v1/projects/123",
			wantCode:     201,
			wantLocation: "/api/v1/projects/123",
		},
		{
			name:         "created without location",
			data:         map[string]string{"id": "456"},
			location:     "",
			wantCode:     201,
			wantLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Created(w, tt.data, tt.location)

			if w.Code != tt.wantCode {
				t.Errorf("Created() status code = %v, want %v", w.Code, tt.wantCode)
			}

			location := w.Header().Get("Location")
			if location != tt.wantLocation {
				t.Errorf("Created() Location header = %v, want %v", location, tt.wantLocation)
			}

			var resp Response
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if !resp.Success {
				t.Error("Created() success should be true")
			}

			if resp.Code != tt.wantCode {
				t.Errorf("Created() response code = %v, want %v", resp.Code, tt.wantCode)
			}
		})
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("NoContent() status code = %v, want %v", w.Code, http.StatusNoContent)
	}

	if w.Body.Len() != 0 {
		t.Errorf("NoContent() should have empty body, got %v bytes", w.Body.Len())
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		message   string
		wantCode  int
		wantField string
	}{
		{
			name:      "validation error with field",
			field:     "email",
			message:   "Email format is invalid",
			wantCode:  422,
			wantField: "email",
		},
		{
			name:      "validation error with empty field",
			field:     "",
			message:   "Validation failed",
			wantCode:  422,
			wantField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ValidationError(w, tt.field, tt.message)

			if w.Code != tt.wantCode {
				t.Errorf("ValidationError() status code = %v, want %v", w.Code, tt.wantCode)
			}

			var resp Response
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if resp.Success {
				t.Error("ValidationError() success should be false")
			}

			if resp.Error == nil {
				t.Fatal("ValidationError() should have error info")
			}

			if resp.Error.Code != "VAL_INVALID_PARAM_001" {
				t.Errorf("ValidationError() error code = %v, want VAL_INVALID_PARAM_001", resp.Error.Code)
			}

			if resp.Error.Message != tt.message {
				t.Errorf("ValidationError() error message = %v, want %v", resp.Error.Message, tt.message)
			}

			if tt.field != "" {
				details, ok := resp.Error.Details.(map[string]interface{})
				if !ok {
					t.Fatal("ValidationError() details should be a map")
				}
				if details["field"] != tt.wantField {
					t.Errorf("ValidationError() details field = %v, want %v", details["field"], tt.wantField)
				}
			}
		})
	}
}
