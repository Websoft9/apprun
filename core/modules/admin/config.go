// Package admin provides platform-level user management functionality.
package admin

// ============================================================================
// Module Constants (Roles)
// ============================================================================

// User roles
const (
	RolePlatformAdmin = "platform_admin"
	RolePlatformUser  = "platform_user"
	RoleSystemUser    = "system_user" // Reserved for future use
)

// ValidRoles contains all valid user roles
var ValidRoles = []string{RolePlatformAdmin, RolePlatformUser}

// IsValidRole checks if a role string is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}
