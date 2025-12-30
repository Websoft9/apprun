package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/pkg/response"
)

func TestDemoHandler_Success(t *testing.T) {
	handler := NewDemoHandler()
	req := httptest.NewRequest(http.MethodGet, "/demo/success", nil)
	w := httptest.NewRecorder()

	handler.Success(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestDemoHandler_Create(t *testing.T) {
	handler := NewDemoHandler()
	req := httptest.NewRequest(http.MethodPost, "/demo/create", nil)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/api/v1/demo/123" {
		t.Errorf("Expected Location header '/api/v1/demo/123', got '%s'", location)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Code != 201 {
		t.Errorf("Expected code 201, got %d", resp.Code)
	}
}

func TestDemoHandler_Delete(t *testing.T) {
	handler := NewDemoHandler()
	req := httptest.NewRequest(http.MethodDelete, "/demo/delete", nil)
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	if w.Body.Len() != 0 {
		t.Error("Expected empty body for NoContent response")
	}
}

func TestDemoHandler_List(t *testing.T) {
	handler := NewDemoHandler()
	req := httptest.NewRequest(http.MethodGet, "/demo/list", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if _, exists := dataMap["items"]; !exists {
		t.Error("Expected data to contain 'items' field")
	}

	if _, exists := dataMap["pagination"]; !exists {
		t.Error("Expected data to contain 'pagination' field")
	}
}

func TestDemoHandler_Error(t *testing.T) {
	handler := NewDemoHandler()
	req := httptest.NewRequest(http.MethodGet, "/demo/error", nil)
	w := httptest.NewRecorder()

	handler.Error(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Error("Expected success to be false")
	}

	if resp.Error == nil {
		t.Fatal("Expected error info to be present")
	}

	if resp.Error.Code != response.ErrCodeNotFound {
		t.Errorf("Expected error code %s, got %s", response.ErrCodeNotFound, resp.Error.Code)
	}
}

func TestDemoHandler_Validate(t *testing.T) {
	handler := NewDemoHandler()
	req := httptest.NewRequest(http.MethodPost, "/demo/validate", nil)
	w := httptest.NewRecorder()

	handler.Validate(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Error("Expected success to be false")
	}

	if resp.Error == nil {
		t.Fatal("Expected error info to be present")
	}

	if resp.Error.Code != response.ErrCodeInvalidParam {
		t.Errorf("Expected error code %s, got %s", response.ErrCodeInvalidParam, resp.Error.Code)
	}

	details, ok := resp.Error.Details.(map[string]interface{})
	if !ok {
		t.Fatal("Expected error details to be a map")
	}

	if details["field"] != "email" {
		t.Errorf("Expected field 'email', got '%v'", details["field"])
	}
}
