package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"apprun/modules/auth/handler"
	"apprun/pkg/response"

	"github.com/go-chi/chi/v5"
)

// TestAddMember_DuplicateMember tests adding a member who already exists
func TestAddMember_DuplicateMember(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Create a new user
	newUser, err := client.User.Create().
		SetUsername("duplicate_user").
		SetEmail("duplicate@example.com").
		SetPasswordHash("hashed_password").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create new user: %v", err)
	}

	// Add member first time
	reqBody := handler.AddMemberRequest{
		UserID: newUser.ID,
		Role:   "member",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/members", project.ID), bytes.NewReader(bodyBytes))
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	memberHandler.AddMember(w, req)

	// First add should succeed
	if w.Code != http.StatusCreated {
		t.Errorf("First add: expected status 201, got %d", w.Code)
	}

	// Try to add the same member again
	bodyBytes2, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/members", project.ID), bytes.NewReader(bodyBytes2))
	req2 = req2.WithContext(createAuthContext(owner.ID))

	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))

	w2 := httptest.NewRecorder()
	memberHandler.AddMember(w2, req2)

	// Second add should fail with 409 Conflict
	if w2.Code != http.StatusConflict {
		t.Errorf("Duplicate add: expected status 409, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var resp response.Response
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Success {
		t.Errorf("Expected success=false for duplicate member, got success=true")
	}
}

// TestUpdateRole_Success tests successful role update
func TestUpdateRole_Success(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Create a new member
	newUser, err := client.User.Create().
		SetUsername("member_to_update").
		SetEmail("member_update@example.com").
		SetPasswordHash("hashed_password").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create new user: %v", err)
	}

	// Add as member
	member, err := client.ProjectMember.Create().
		SetProjectID(project.ID).
		SetUserID(newUser.ID).
		SetRole("member").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create member: %v", err)
	}

	// Update to admin
	reqBody := handler.UpdateRoleRequest{
		Role: "admin",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%d/members/%d", project.ID, member.ID), bytes.NewReader(bodyBytes))
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	rctx.URLParams.Add("member_id", fmt.Sprintf("%d", member.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	memberHandler.UpdateRole(w, req)

	// Should succeed
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

	// Verify role was updated in database
	updatedMember, err := client.ProjectMember.Get(ctx, member.ID)
	if err != nil {
		t.Fatalf("Failed to get updated member: %v", err)
	}

	if updatedMember.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", updatedMember.Role)
	}
}

// TestRemoveMember_Success tests successful member removal
func TestRemoveMember_Success(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Create a new member
	newUser, err := client.User.Create().
		SetUsername("member_to_remove").
		SetEmail("member_remove@example.com").
		SetPasswordHash("hashed_password").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create new user: %v", err)
	}

	// Add as member
	member, err := client.ProjectMember.Create().
		SetProjectID(project.ID).
		SetUserID(newUser.ID).
		SetRole("member").
		Save(ctx)
	if err != nil {
		t.Fatalf("Failed to create member: %v", err)
	}

	// Remove member
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%d/members/%d", project.ID, member.ID), nil)
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	rctx.URLParams.Add("member_id", fmt.Sprintf("%d", member.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	memberHandler.RemoveMember(w, req)

	// Should succeed with 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify member was removed from database
	_, err = client.ProjectMember.Get(ctx, member.ID)
	if err == nil {
		t.Errorf("Expected member to be deleted, but still exists")
	}
}

// TestListMembers_Pagination tests that pagination info is returned
func TestListMembers_Pagination(t *testing.T) {
	memberHandler, client, owner, project := setupTestHandler(t)
	defer client.Close()

	ctx := context.Background()

	// Add multiple members
	for i := 0; i < 5; i++ {
		newUser, err := client.User.Create().
			SetUsername(fmt.Sprintf("member_%d", i)).
			SetEmail(fmt.Sprintf("member_%d@example.com", i)).
			SetPasswordHash("hashed_password").
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}

		_, err = client.ProjectMember.Create().
			SetProjectID(project.ID).
			SetUserID(newUser.ID).
			SetRole("member").
			Save(ctx)
		if err != nil {
			t.Fatalf("Failed to create member %d: %v", i, err)
		}
	}

	// List members
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/members", project.ID), nil)
	req = req.WithContext(createAuthContext(owner.ID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("project_id", fmt.Sprintf("%d", project.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	memberHandler.ListMembers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	respData, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Response data is not a map")
	}

	// Check pagination info exists
	pagination, ok := respData["pagination"].(map[string]interface{})
	if !ok {
		t.Fatalf("Response pagination is not a map")
	}

	total, ok := pagination["total"].(float64)
	if !ok {
		t.Fatalf("Pagination total is not a number")
	}

	// Should have at least 6 members (1 owner + 5 added)
	if int(total) < 6 {
		t.Errorf("Expected at least 6 members, got %d", int(total))
	}
}
