package service

import (
	"context"

	"apprun/internal/rbac"
)

// PermissionService handles permission checking and querying
type PermissionService struct{}

// NewPermissionService creates a new permission service
func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

// PermissionInfo represents a permission for a resource
type PermissionInfo struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *PermissionService) CheckPermission(ctx context.Context, userID, projectID int64, resource, action string) (bool, error) {
	return rbac.CheckPermission(userID, projectID, resource, action)
}

// GetUserRoles retrieves all roles of a user in a project
func (s *PermissionService) GetUserRoles(userID, projectID int64) ([]string, error) {
	return rbac.GetUserRoles(userID, projectID)
}

// GetUserRoleInProject retrieves the user's role in a specific project
func (s *PermissionService) GetUserRoleInProject(ctx context.Context, userID, projectID int64) (string, error) {
	roles, err := rbac.GetUserRoles(userID, projectID)
	if err != nil {
		return "", err
	}
	if len(roles) == 0 {
		return "", nil
	}
	// Return the first role (users typically have one role per project)
	return roles[0], nil
}

// GetPermissionsForRole retrieves all permissions for a role in a project
func (s *PermissionService) GetPermissionsForRole(ctx context.Context, projectID int64, role string) (map[string][]string, error) {
	// Define all resources and actions
	resources := []string{
		rbac.ResourceConfig,
		rbac.ResourceData,
		rbac.ResourceStorage,
		rbac.ResourceFunction,
		rbac.ResourceWorkflow,
		rbac.ResourceMember,
		rbac.ResourceProject,
	}

	actions := []string{
		rbac.ActionCreate,
		rbac.ActionRead,
		rbac.ActionUpdate,
		rbac.ActionDelete,
		rbac.ActionExecute,
		rbac.ActionManage,
	}

	permissions := make(map[string][]string)

	// For each resource, check what actions the role can perform
	enforcer := rbac.GetEnforcer()
	domain := rbac.FormatDomain(projectID)
	roleKey := rbac.FormatRole(projectID, role)

	for _, resource := range resources {
		var allowedActions []string
		for _, action := range actions {
			// Check if the role has this permission
			ok, err := enforcer.Enforce(roleKey, domain, resource, action)
			if err != nil {
				continue
			}
			if ok {
				allowedActions = append(allowedActions, action)
			}
		}
		if len(allowedActions) > 0 {
			permissions[resource] = allowedActions
		}
	}

	return permissions, nil
}

// GetUserPermissions retrieves all permissions of a user in a project
// This iterates through all resources and actions to build a permission list
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID, projectID int64) ([]PermissionInfo, error) {
	// Define all resources and actions to check
	resources := []string{
		rbac.ResourceConfig,
		rbac.ResourceData,
		rbac.ResourceStorage,
		rbac.ResourceFunction,
		rbac.ResourceWorkflow,
		rbac.ResourceMember,
		rbac.ResourceProject,
	}

	actions := []string{
		rbac.ActionCreate,
		rbac.ActionRead,
		rbac.ActionUpdate,
		rbac.ActionDelete,
		rbac.ActionExecute,
		rbac.ActionManage,
	}

	var permissions []PermissionInfo

	// Check each resource
	for _, resource := range resources {
		var allowedActions []string

		// Check each action for this resource
		for _, action := range actions {
			allowed, err := rbac.CheckPermission(userID, projectID, resource, action)
			if err != nil {
				// Log error but continue
				continue
			}

			if allowed {
				allowedActions = append(allowedActions, action)
			}
		}

		// Only include resources with at least one allowed action
		if len(allowedActions) > 0 {
			permissions = append(permissions, PermissionInfo{
				Resource: resource,
				Actions:  allowedActions,
			})
		}
	}

	return permissions, nil
}
