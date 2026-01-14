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
	"apprun/ent/projectmember"
	"apprun/internal/rbac"
	"apprun/modules/auth/handler"
	"apprun/modules/auth/repository"
	"apprun/modules/auth/service"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *ent.Client {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	return client
}

// setupTestHandler creates a test handler with all dependencies
func setupTestHandler(t *testing.T) (*handler.ProjectMemberHandler, *ent.Client, *ent.User, *ent.Project) {
	client := setupTestDB(t)
	ctx := context.Background()

	// Create test user
	user, err := client.User.Create().
		SetUsername("testuser").
		SetEmail("test@example.com").
		SetPasswordHash("hashed_password").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test project
	project, err := client.Project.Create().
		SetName("Test Project").
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

	// Initialize RBAC
	rbac.InitEnforcer(rbac.Config{UseDatabase: false})

	// Add owner role to RBAC enforcer
	enforcer := rbac.GetEnforcer()
	enforcer.AddGroupingPolicy(
		rbac.FormatUserKey(user.ID),
		"owner", // Role name matches policy file
		rbac.FormatDomain(project.ID),
	)
	enforcer.SavePolicy()

	// Setup handler
	projectRepo := repository.NewProjectRepository(client)
	memberRepo := repository.NewProjectMemberRepository(client)
	memberService := service.NewProjectMemberService(memberRepo, projectRepo)
	memberHandler := handler.NewProjectMemberHandler(memberService)

	return memberHandler, client, user, project
}

func TestAddMember_Success(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Create a new user to add as member
	newUser, err := client.User.Create().
		SetUsername("newmember").
		SetEmail("newmember@example.com").
		SetPasswordHash("hashed_password").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create new user: %v", err)
	}

	// Prepare request
	reqBody := handler.AddMemberRequest{
		UserID: newUser.ID,
		Role:   "member",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/members", project.ID), bytes.NewReader(bodyBytes))
	req = req.WithContext(createAuthContext(owner.ID))

	// Add URL parameters using chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Execute handler
	memberHandler.AddMember(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=false. Error: %s", resp.Error)
	}
}

func TestAddMember_InvalidRole(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Create a new user
	newUser, err := client.User.Create().
		SetUsername("newmember2").
		SetEmail("newmember2@example.com").
		SetPasswordHash("hashed_password").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create new user: %v", err)
	}

	// Prepare request with invalid role
	reqBody := handler.AddMemberRequest{
		UserID: newUser.ID,
		Role:   "invalid_role",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/members", project.ID), bytes.NewReader(bodyBytes))
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	memberHandler.AddMember(w, req)

	// Check response - should fail with 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success=false, got success=true")
	}
}

func TestListMembers_Success(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/members", project.ID), nil)
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	memberHandler.ListMembers(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success=true, got success=false")
	}

	// Check that at least the owner is in the list
	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Response data is not a map")
	}

	items, ok := respData["items"].([]interface{})
	if !ok {
		t.Fatalf("Response items is not an array")
	}

	if len(items) < 1 {
		t.Errorf("Expected at least 1 member, got %d", len(items))
	}
}

func TestUpdateRole_CannotModifyOwner(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Get the owner's member record
	ownerMember, err := client.ProjectMember.Query().
		Where(
			projectmember.ProjectIDEQ(project.ID),
			projectmember.UserIDEQ(owner.ID),
		).
		Only(ctx)
	if err != nil {
		t.Fatalf("Failed to get owner member: %v", err)
	}

	// Try to update owner role
	reqBody := handler.UpdateRoleRequest{
		Role: "admin",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%d/members/%d", project.ID, ownerMember.ID), bytes.NewReader(bodyBytes))
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	rctx.URLParams.Add("member_id", fmt.Sprintf("%d", ownerMember.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	memberHandler.UpdateRole(w, req)

	// Check response - should fail with 422 (business rule violation)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success=false, got success=true")
	}
}

func TestRemoveMember_CannotRemoveOwner(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Get the owner's member record
	ownerMember, err := client.ProjectMember.Query().
		Where(
			projectmember.ProjectIDEQ(project.ID),
			projectmember.UserIDEQ(owner.ID),
		).
		Only(ctx)
	if err != nil {
		t.Fatalf("Failed to get owner member: %v", err)
	}

	// Try to remove owner
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%d/members/%d", project.ID, ownerMember.ID), nil)
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	rctx.URLParams.Add("member_id", fmt.Sprintf("%d", ownerMember.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	memberHandler.RemoveMember(w, req)

	// Check response - should fail with 422 (business rule violation)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success=false, got success=true")
	}
}
