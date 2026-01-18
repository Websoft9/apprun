package rbac

import (
	"fmt"
	"strconv"
)

// FormatUserKey formats a user ID as a Casbin subject key
// Format: u:<userID>
func FormatUserKey(userID int64) string {
	return fmt.Sprintf("u:%d", userID)
}

// FormatRole formats a project role as a Casbin role name
// Format: p:<projectID>:<role> (e.g., "p:123:owner")
func FormatRole(projectID int64, role string) string {
	return fmt.Sprintf("p:%d:%s", projectID, role)
}

// FormatDomain formats a project ID as a Casbin domain
// Returns "platform" for projectID=0, otherwise "p:<projectID>"
func FormatDomain(projectID int64) string {
	if projectID == 0 {
		return PlatformDomain
	}
	return fmt.Sprintf("p:%d", projectID)
}

// ParseUserKey extracts the user ID from a formatted user key
// Input: "u:123" -> Output: 123, nil
func ParseUserKey(userKey string) (int64, error) {
	var userID int64
	_, err := fmt.Sscanf(userKey, "u:%d", &userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user key format: %s", userKey)
	}
	return userID, nil
}

// ParseRole extracts the project ID and role from a formatted role
// Input: "p:123:owner" -> Output: 123, "owner", nil
func ParseRole(roleStr string) (int64, string, error) {
	var projectID int64
	var role string
	_, err := fmt.Sscanf(roleStr, "p:%d:%s", &projectID, &role)
	if err != nil {
		return 0, "", fmt.Errorf("invalid role format: %s", roleStr)
	}
	return projectID, role, nil
}

// ParseDomain extracts the project ID from a formatted domain
// Input: "p:123" -> Output: 123, nil
// Input: "platform" -> Output: 0, nil
func ParseDomain(domain string) (int64, error) {
	if domain == PlatformDomain {
		return 0, nil
	}

	var projectID int64
	_, err := fmt.Sscanf(domain, "p:%d", &projectID)
	if err != nil {
		return 0, fmt.Errorf("invalid domain format: %s", domain)
	}
	return projectID, nil
}

// IsValidProjectRole checks if a role is a valid project-level role
func IsValidProjectRole(role string) bool {
	switch role {
	case RoleProjectOwner, RoleProjectAdmin, RoleProjectMember, RoleProjectViewer:
		return true
	default:
		return false
	}
}

// IsValidPlatformRole checks if a role is a valid platform-level role
func IsValidPlatformRole(role string) bool {
	switch role {
	case RolePlatformAdmin, RolePlatformUser:
		return true
	default:
		return false
	}
}

// ConvertToInt64 safely converts interface{} to int64
func ConvertToInt64(val interface{}) int64 {
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}
