package rbac

import "fmt"

// SavePolicies saves current policies to the adapter
// This is called when policies are modified programmatically
func SavePolicies() error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	if err := enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save policies: %w", err)
	}

	// Clear cache after policy changes
	ClearAllCache()

	return nil
}

// AddPlatformRole assigns a platform-level role to a user
func AddPlatformRole(userID int64, role string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	subject := fmt.Sprintf("user:%d", userID)

	// Use g2 for platform roles (no domain)
	_, err := enforcer.AddRoleForUser(subject, role)
	if err != nil {
		return fmt.Errorf("failed to add platform role: %w", err)
	}

	return SavePolicies()
}

// RemovePlatformRole removes a platform-level role from a user
func RemovePlatformRole(userID int64, role string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	subject := fmt.Sprintf("user:%d", userID)

	_, err := enforcer.DeleteRoleForUser(subject, role)
	if err != nil {
		return fmt.Errorf("failed to remove platform role: %w", err)
	}

	return SavePolicies()
}

// GetPlatformRoles returns all platform-level roles of a user
func GetPlatformRoles(userID int64) ([]string, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("enforcer not initialized")
	}

	subject := fmt.Sprintf("user:%d", userID)

	roles, err := enforcer.GetRolesForUser(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform roles: %w", err)
	}

	return roles, nil
}
