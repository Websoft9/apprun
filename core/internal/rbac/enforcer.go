package rbac

import (
	_ "embed"
	"fmt"
	"os"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

//go:embed model.conf
var modelContent string

//go:embed default_policy.csv
var defaultPolicyContent string

// Config holds Casbin enforcer configuration
type Config struct {
	PolicyPath  string // Optional: custom policy path, if empty uses embedded default_policy.csv
	UseDatabase bool   // MVP: false (use file), Production: true (use DB)
}

var (
	enforcer *casbin.Enforcer
	once     sync.Once
	initErr  error
)

// InitEnforcer initializes the global Casbin enforcer (singleton)
func InitEnforcer(cfg Config) error {
	once.Do(func() {
		// Create model from embedded content
		m, err := model.NewModelFromString(modelContent)
		if err != nil {
			initErr = fmt.Errorf("failed to parse Casbin model: %w", err)
			return
		}

		// Determine policy source
		var adapter persist.Adapter
		var tmpFileName string
		if cfg.PolicyPath != "" {
			// Use custom policy file path
			adapter = fileadapter.NewAdapter(cfg.PolicyPath)
		} else {
			// Use embedded default policy - create temp file
			tmpFile, tmpErr := os.CreateTemp("", "casbin_policy_*.csv")
			if tmpErr != nil {
				initErr = fmt.Errorf("failed to create temp policy file: %w", tmpErr)
				return
			}
			tmpFileName = tmpFile.Name()
			defer func() {
				if closeErr := tmpFile.Close(); closeErr != nil {
					fmt.Printf("Warning: failed to close temp file: %v\n", closeErr)
				}
				if tmpFileName != "" {
					if removeErr := os.Remove(tmpFileName); removeErr != nil {
						fmt.Printf("Warning: failed to remove temp file: %v\n", removeErr)
					}
				}
			}()

			if _, writeErr := tmpFile.WriteString(defaultPolicyContent); writeErr != nil {
				initErr = fmt.Errorf("failed to write policy content: %w", writeErr)
				return
			}

			adapter = fileadapter.NewAdapter(tmpFileName)
		}

		enforcer, err = casbin.NewEnforcer(m, adapter)
		if err != nil {
			initErr = fmt.Errorf("failed to create Casbin enforcer: %w", err)
			return
		}

		// Load policies
		if err := enforcer.LoadPolicy(); err != nil {
			initErr = fmt.Errorf("failed to load policies: %w", err)
			return
		}
	})

	if initErr != nil {
		return fmt.Errorf("enforcer initialization failed: %w", initErr)
	}
	return nil
}

// GetEnforcer returns the global enforcer instance
func GetEnforcer() *casbin.Enforcer {
	return enforcer
}

// CheckPermission checks if a user has permission for a resource action
// Returns (allowed, error)
func CheckPermission(userID, projectID int64, resource, action string) (bool, error) {
	if enforcer == nil {
		return false, fmt.Errorf("enforcer not initialized")
	}

	// Format subject: u:{userID}
	subject := FormatUserKey(userID)

	// Format domain
	domain := FormatDomain(projectID)

	// Enforce permission check
	allowed, err := enforcer.Enforce(subject, domain, resource, action)
	if err != nil {
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	return allowed, nil
}

// AddUserRole assigns a role to a user in a project
func AddUserRole(userID, projectID int64, role string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	subject := FormatUserKey(userID)
	domain := FormatDomain(projectID)

	_, err := enforcer.AddGroupingPolicy(subject, role, domain)
	if err != nil {
		return fmt.Errorf("failed to add user role: %w", err)
	}

	return nil
}

// RemoveUserRole removes a role from a user in a project
func RemoveUserRole(userID, projectID int64, role string) error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	subject := FormatUserKey(userID)
	domain := FormatDomain(projectID)

	_, err := enforcer.RemoveGroupingPolicy(subject, role, domain)
	if err != nil {
		return fmt.Errorf("failed to remove user role: %w", err)
	}

	return nil
}

// GetUserRoles returns all roles of a user in a project
func GetUserRoles(userID, projectID int64) ([]string, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("enforcer not initialized")
	}

	subject := FormatUserKey(userID)
	domain := FormatDomain(projectID)

	roles := enforcer.GetRolesForUserInDomain(subject, domain)

	return roles, nil
}

// ReloadPolicies explicitly reloads policies from source
// Call this after role changes or policy updates
func ReloadPolicies() error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to reload policies: %w", err)
	}

	return nil
}

// resetEnforcerForTest resets the enforcer for testing purposes
// This should ONLY be used in tests
func resetEnforcerForTest() {
	enforcer = nil
	once = sync.Once{}
	initErr = nil
}
