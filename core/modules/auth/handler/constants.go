package handler

import "apprun/internal/rbac"

// ValidProjectRoles defines all valid project roles
var ValidProjectRoles = []string{
	rbac.RoleProjectOwner,
	rbac.RoleProjectAdmin,
	rbac.RoleProjectMember,
	rbac.RoleProjectViewer,
}

// IsValidProjectRole checks if a role is valid
func IsValidProjectRole(role string) bool {
	for _, validRole := range ValidProjectRoles {
		if role == validRole {
			return true
		}
	}
	return false
}
