package rbac

import (
	_ "embed"
	"fmt"
	"sync"

	"apprun/ent"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

//go:embed model.conf
var modelContent string

//go:embed default_policy.csv
var defaultPolicyContent string

// Config holds Casbin enforcer configuration
type Config struct {
	PolicyPath string      // Optional: custom policy path (deprecated, use EntClient instead)
	EntClient  *ent.Client // Optional: Ent client for database adapter
}

var (
	enforcer     *casbin.Enforcer
	entAdapter   *EntAdapter // Store adapter for direct policy operations
	once         sync.Once
	initErr      error
	usingDBStore bool // Track if using database storage
)

// InitEnforcer initializes the global Casbin enforcer (singleton)
// If cfg.EntClient is provided, uses database storage (recommended for production)
// Otherwise, uses embedded default policy (for testing/development)
func InitEnforcer(cfg Config) error {
	once.Do(func() {
		// Create model from embedded content
		m, err := model.NewModelFromString(modelContent)
		if err != nil {
			initErr = fmt.Errorf("failed to parse Casbin model: %w", err)
			return
		}

		var adapter persist.Adapter

		if cfg.EntClient != nil {
			// Use database adapter (production mode)
			entAdapter, err = NewEntAdapter(cfg.EntClient)
			if err != nil {
				initErr = fmt.Errorf("failed to create ent adapter: %w", err)
				return
			}
			adapter = entAdapter
			usingDBStore = true
		} else {
			// Fallback: use in-memory model without persistent adapter
			// This is for testing or when database is not available
			// Policies will be loaded from embedded default_policy.csv into memory
			adapter = nil
			usingDBStore = false
		}

		// Create enforcer
		if adapter != nil {
			enforcer, err = casbin.NewEnforcer(m, adapter)
		} else {
			// No adapter - create enforcer with model only
			enforcer, err = casbin.NewEnforcer(m)
		}
		if err != nil {
			initErr = fmt.Errorf("failed to create Casbin enforcer: %w", err)
			return
		}

		// Load policies
		if adapter != nil {
			// Load from database
			if err := enforcer.LoadPolicy(); err != nil {
				initErr = fmt.Errorf("failed to load policies from database: %w", err)
				return
			}
		} else {
			// Load embedded default policies into memory
			if err := loadEmbeddedPolicies(enforcer); err != nil {
				initErr = fmt.Errorf("failed to load embedded policies: %w", err)
				return
			}
		}
	})

	if initErr != nil {
		return fmt.Errorf("enforcer initialization failed: %w", initErr)
	}
	return nil
}

// loadEmbeddedPolicies loads policies from the embedded default_policy.csv into the enforcer.
// This is used when no database adapter is available (testing/development fallback).
func loadEmbeddedPolicies(e *casbin.Enforcer) error {
	lines := splitLines(defaultPolicyContent)

	for _, line := range lines {
		line = trimLine(line)
		if line == "" || line[0] == '#' {
			continue // Skip empty lines and comments
		}

		ptype, rule, ok := parseCSVLine(line)
		if !ok {
			continue
		}

		var err error
		switch ptype {
		case "p":
			_, err = e.AddPolicy(rule)
		case "g":
			_, err = e.AddGroupingPolicy(rule)
		case "g2":
			_, err = e.AddNamedGroupingPolicy("g2", rule)
		}
		if err != nil {
			return fmt.Errorf("failed to add policy %v: %w", rule, err)
		}
	}

	return nil
}

// splitLines splits content by newlines
func splitLines(content string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			result = append(result, content[start:i])
			start = i + 1
		}
	}
	if start < len(content) {
		result = append(result, content[start:])
	}
	return result
}

// trimLine trims whitespace from a line
func trimLine(line string) string {
	start := 0
	end := len(line)
	for start < end && (line[start] == ' ' || line[start] == '\t' || line[start] == '\r') {
		start++
	}
	for end > start && (line[end-1] == ' ' || line[end-1] == '\t' || line[end-1] == '\r') {
		end--
	}
	return line[start:end]
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

// SavePolicyIfNeeded saves policies to the adapter if using database storage.
// When using in-memory mode (no adapter), this is a no-op and returns nil.
// This should be called after adding/removing policies.
func SavePolicyIfNeeded() error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	// Only save if using database storage (has adapter)
	if !usingDBStore {
		return nil // In-memory mode, no need to persist
	}

	return enforcer.SavePolicy()
}

// resetEnforcerForTest resets the enforcer for testing purposes
// This should ONLY be used in tests
func resetEnforcerForTest() {
	enforcer = nil
	entAdapter = nil
	once = sync.Once{}
	initErr = nil
	usingDBStore = false
}

// IsUsingDatabaseStorage returns true if the enforcer is using database storage
func IsUsingDatabaseStorage() bool {
	return usingDBStore
}

// GetEntAdapter returns the Ent adapter instance (for direct policy operations)
// Returns nil if not using database storage
func GetEntAdapter() *EntAdapter {
	return entAdapter
}

// SeedDefaultPolicies imports default policies from embedded CSV into the database.
// This function is idempotent - it only adds policies that don't already exist.
// Should be called during application bootstrap when using database storage.
func SeedDefaultPolicies() error {
	if enforcer == nil {
		return fmt.Errorf("enforcer not initialized")
	}

	if !usingDBStore {
		// When not using DB storage, policies are already loaded from embedded CSV
		return nil
	}

	lines := splitLines(defaultPolicyContent)
	addedCount := 0
	skippedCount := 0

	for _, line := range lines {
		line = trimLine(line)
		if line == "" || line[0] == '#' {
			continue // Skip empty lines and comments
		}

		ptype, rule, ok := parseCSVLine(line)
		if !ok {
			continue
		}

		var hasPolicy bool
		var err error

		switch ptype {
		case "p":
			hasPolicy = enforcer.HasPolicy(rule)
			if !hasPolicy {
				_, err = enforcer.AddPolicy(rule)
				if err == nil {
					addedCount++
				}
			} else {
				skippedCount++
			}
		case "g":
			hasPolicy = enforcer.HasGroupingPolicy(rule)
			if !hasPolicy {
				_, err = enforcer.AddGroupingPolicy(rule)
				if err == nil {
					addedCount++
				}
			} else {
				skippedCount++
			}
		case "g2":
			hasPolicy = enforcer.HasNamedGroupingPolicy("g2", rule)
			if !hasPolicy {
				_, err = enforcer.AddNamedGroupingPolicy("g2", rule)
				if err == nil {
					addedCount++
				}
			} else {
				skippedCount++
			}
		}

		if err != nil {
			return fmt.Errorf("failed to add policy %s %v: %w", ptype, rule, err)
		}
	}

	// Policies are automatically persisted by the adapter
	// No need to call SavePolicy() as AddPolicy/AddGroupingPolicy already persist

	return nil
}

// GetPolicySummary returns a summary of loaded policies for logging/debugging
func GetPolicySummary() (pCount, gCount, g2Count int) {
	if enforcer == nil {
		return 0, 0, 0
	}

	policies := enforcer.GetPolicy()
	groupingPolicies := enforcer.GetGroupingPolicy()
	g2Policies := enforcer.GetNamedGroupingPolicy("g2")

	return len(policies), len(groupingPolicies), len(g2Policies)
}
