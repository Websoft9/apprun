package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/ent"
	"apprun/ent/enttest"
	"apprun/internal/jwt"
	"apprun/internal/rbac"
	"apprun/modules/auth/handler"
	"apprun/modules/auth/service"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

// createAuthContext creates a context with JWT user ID
func createAuthContext(userID int64) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, jwt.UserIDKey, userID)
}

// setupPermissionTestHandler creates a permission test handler with all dependencies
func setupPermissionTestHandler(t *testing.T) (*handler.PermissionHandler, *ent.Client, *ent.User, *ent.Project) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	ctx := context.Background()

	// Create test user
	user, err := client.User.Create().
		SetUsername("testuser_perm").
		SetEmail("test_perm@example.com").
		SetPasswordHash("hashed_password").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test project
	project, err := client.Project.Create().
		SetName("Test Project Permission").
		SetDescription("Test Description").
		SetOwnerID(user.ID).
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create owner membership
	_, err = client.ProjectMember.Create().
		SetProjectID(project.ID).
		SetUserID(user.ID).
		SetRole("owner").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create project member: %v", err)
	}

	// Initialize RBAC (in-memory mode for testing)
	rbac.InitEnforcer(rbac.Config{})

	// Add role to RBAC enforcer
	enforcer := rbac.GetEnforcer()
	enforcer.AddGroupingPolicy(
		rbac.FormatUserKey(user.ID),
		"owner", // Role name matches policy file
		rbac.FormatDomain(project.ID),
	)

	// Setup handler
	permissionService := service.NewPermissionService()
	permissionHandler := handler.NewPermissionHandler(permissionService)

	return permissionHandler, client, user, project
}

func TestGetMyPermissions_Success(t *testing.T) {
	permHandler, client, user, project := setupPermissionTestHandler(t)
	defer client.Close()

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/permissions/my", project.ID), nil)
	req = req.WithContext(createAuthContext(user.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	permHandler.GetMyPermissions(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=false. Error: %s", resp.Error)
	}

	// Check that permissions are returned
	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Response data is not a map")
	}

	permissions, ok := respData["permissions"].([]interface{})
	if !ok {
		t.Fatalf("Response permissions is not an array")
	}

	// Owner should have permissions
	if len(permissions) == 0 {
		t.Errorf("Expected permissions for owner, got none")
	}
}

func TestCheckPermission_Success(t *testing.T) {
	permHandler, client, user, project := setupPermissionTestHandler(t)
	defer client.Close()

	// Prepare request - check if owner can read project
	reqBody := handler.CheckPermissionRequest{
		Resource: "project",
		Action:   "read",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/permissions/check", project.ID), bytes.NewReader(bodyBytes))
	req = req.WithContext(createAuthContext(user.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	permHandler.CheckPermission(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=false. Error: %s", resp.Error)
	}

	// Check permission result
	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Response data is not a map")
	}

	allowed, ok := respData["allowed"].(bool)
	if !ok {
		t.Fatalf("Response allowed is not a boolean")
	}

	// Owner should be allowed to read project
	if !allowed {
		t.Errorf("Expected allowed=true for owner reading project, got false")
	}
}

func TestCheckPermission_InvalidRequest(t *testing.T) {
	permHandler, client, user, project := setupPermissionTestHandler(t)
	defer client.Close()

	// Prepare request with missing fields
	reqBody := handler.CheckPermissionRequest{
		Resource: "",
		Action:   "read",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/permissions/check", project.ID), bytes.NewReader(bodyBytes))
	req = req.WithContext(createAuthContext(user.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	permHandler.CheckPermission(w, req)

	// Check response - should fail with 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success=false, got success=true")
	}
}

func TestCheckPermission_NoAuth(t *testing.T) {
	permHandler, client, _, project := setupPermissionTestHandler(t)
	defer client.Close()

	// Prepare request without user authentication
	reqBody := handler.CheckPermissionRequest{
		Resource: "project",
		Action:   "read",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/permissions/check", project.ID), bytes.NewReader(bodyBytes))
	// No authentication context added

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	permHandler.CheckPermission(w, req)

	// Check response - should fail with 400 (missing token is a bad request)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success=false, got success=true")
	}
}
