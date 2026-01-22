package rbac

import (
	"context"
	"fmt"
	"strings"

	"apprun/ent"
	"apprun/ent/casbinrule"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// EntAdapter is a Casbin adapter implementation using Ent ORM.
// This adapter implements the persist.Adapter interface for Casbin v2,
// allowing policies to be stored in the database via the casbin_rule table.
type EntAdapter struct {
	client *ent.Client
}

// NewEntAdapter creates a new EntAdapter with the given Ent client.
func NewEntAdapter(client *ent.Client) (*EntAdapter, error) {
	if client == nil {
		return nil, fmt.Errorf("ent client cannot be nil")
	}
	return &EntAdapter{client: client}, nil
}

// LoadPolicy loads all policy rules from the storage.
// Implements persist.Adapter interface.
func (a *EntAdapter) LoadPolicy(m model.Model) error {
	ctx := context.Background()

	rules, err := a.client.CasbinRule.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("failed to load casbin rules: %w", err)
	}

	for _, rule := range rules {
		loadPolicyLine(rule, m)
	}

	return nil
}

// SavePolicy saves all policy rules to the storage.
// Implements persist.Adapter interface.
func (a *EntAdapter) SavePolicy(m model.Model) error {
	ctx := context.Background()

	// Start transaction
	tx, err := a.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Delete all existing rules
	_, err = tx.CasbinRule.Delete().Exec(ctx)
	if err != nil {
		_ = tx.Rollback() //nolint:errcheck // Rollback error is not critical here
		return fmt.Errorf("failed to clear existing rules: %w", err)
	}

	// Save policy rules (p, p2, ...)
	for ptype, ast := range m["p"] {
		for _, rule := range ast.Policy {
			if err := savePolicyLine(ctx, tx, ptype, rule); err != nil {
				_ = tx.Rollback() //nolint:errcheck // Rollback error is not critical here
				return err
			}
		}
	}

	// Save grouping rules (g, g2, ...)
	for ptype, ast := range m["g"] {
		for _, rule := range ast.Policy {
			if err := savePolicyLine(ctx, tx, ptype, rule); err != nil {
				_ = tx.Rollback() //nolint:errcheck // Rollback error is not critical here
				return err
			}
		}
	}

	return tx.Commit()
}

// AddPolicy adds a policy rule to the storage.
// Implements persist.Adapter interface.
func (a *EntAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()
	return savePolicyLine(ctx, a.client, ptype, rule)
}

// RemovePolicy removes a policy rule from the storage.
// Implements persist.Adapter interface.
func (a *EntAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()

	query := a.client.CasbinRule.Delete().Where(casbinrule.PtypeEQ(ptype))

	// Build where clause based on rule length
	if len(rule) > 0 {
		query = query.Where(casbinrule.V0EQ(rule[0]))
	}
	if len(rule) > 1 {
		query = query.Where(casbinrule.V1EQ(rule[1]))
	}
	if len(rule) > 2 {
		query = query.Where(casbinrule.V2EQ(rule[2]))
	}
	if len(rule) > 3 {
		query = query.Where(casbinrule.V3EQ(rule[3]))
	}
	if len(rule) > 4 {
		query = query.Where(casbinrule.V4EQ(rule[4]))
	}
	if len(rule) > 5 {
		query = query.Where(casbinrule.V5EQ(rule[5]))
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// Implements persist.Adapter interface.
func (a *EntAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	ctx := context.Background()

	query := a.client.CasbinRule.Delete().Where(casbinrule.PtypeEQ(ptype))

	// Build where clause based on fieldIndex and fieldValues
	for i, value := range fieldValues {
		if value == "" {
			continue
		}
		switch fieldIndex + i {
		case 0:
			query = query.Where(casbinrule.V0EQ(value))
		case 1:
			query = query.Where(casbinrule.V1EQ(value))
		case 2:
			query = query.Where(casbinrule.V2EQ(value))
		case 3:
			query = query.Where(casbinrule.V3EQ(value))
		case 4:
			query = query.Where(casbinrule.V4EQ(value))
		case 5:
			query = query.Where(casbinrule.V5EQ(value))
		}
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove filtered policy: %w", err)
	}

	return nil
}

// Helper functions

// loadPolicyLine loads a single policy rule into the Casbin model.
func loadPolicyLine(rule *ent.CasbinRule, m model.Model) {
	lineText := rule.Ptype

	if rule.V0 != "" {
		lineText += ", " + rule.V0
	}
	if rule.V1 != "" {
		lineText += ", " + rule.V1
	}
	if rule.V2 != "" {
		lineText += ", " + rule.V2
	}
	if rule.V3 != "" {
		lineText += ", " + rule.V3
	}
	if rule.V4 != "" {
		lineText += ", " + rule.V4
	}
	if rule.V5 != "" {
		lineText += ", " + rule.V5
	}

	_ = persist.LoadPolicyLine(lineText, m) //nolint:errcheck // Error is not critical
}

// savePolicyLine saves a single policy rule to the database.
// Works with both *ent.Client and *ent.Tx via the CasbinRuleClient interface.
func savePolicyLine(ctx context.Context, client interface{}, ptype string, rule []string) error {
	var create *ent.CasbinRuleCreate

	switch c := client.(type) {
	case *ent.Client:
		create = c.CasbinRule.Create()
	case *ent.Tx:
		create = c.CasbinRule.Create()
	default:
		return fmt.Errorf("unsupported client type")
	}

	create.SetPtype(ptype)

	if len(rule) > 0 {
		create.SetV0(rule[0])
	}
	if len(rule) > 1 {
		create.SetV1(rule[1])
	}
	if len(rule) > 2 {
		create.SetV2(rule[2])
	}
	if len(rule) > 3 {
		create.SetV3(rule[3])
	}
	if len(rule) > 4 {
		create.SetV4(rule[4])
	}
	if len(rule) > 5 {
		create.SetV5(rule[5])
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to save policy line: %w", err)
	}

	return nil
}

// Verify interface implementation at compile time
var _ persist.Adapter = (*EntAdapter)(nil)

// policyToStringArray converts a CasbinRule to string array for internal use.
func policyToStringArray(rule *ent.CasbinRule) []string {
	result := make([]string, 0, 6)

	if rule.V0 != "" {
		result = append(result, rule.V0)
	}
	if rule.V1 != "" {
		result = append(result, rule.V1)
	}
	if rule.V2 != "" {
		result = append(result, rule.V2)
	}
	if rule.V3 != "" {
		result = append(result, rule.V3)
	}
	if rule.V4 != "" {
		result = append(result, rule.V4)
	}
	if rule.V5 != "" {
		result = append(result, rule.V5)
	}

	return result
}

// GetFilteredPolicy gets policy rules that match the filter.
// This is a helper method not required by the basic Adapter interface.
func (a *EntAdapter) GetFilteredPolicy(ptype string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	ctx := context.Background()

	query := a.client.CasbinRule.Query().Where(casbinrule.PtypeEQ(ptype))

	// Build where clause based on fieldIndex and fieldValues
	for i, value := range fieldValues {
		if value == "" {
			continue
		}
		switch fieldIndex + i {
		case 0:
			query = query.Where(casbinrule.V0EQ(value))
		case 1:
			query = query.Where(casbinrule.V1EQ(value))
		case 2:
			query = query.Where(casbinrule.V2EQ(value))
		case 3:
			query = query.Where(casbinrule.V3EQ(value))
		case 4:
			query = query.Where(casbinrule.V4EQ(value))
		case 5:
			query = query.Where(casbinrule.V5EQ(value))
		}
	}

	rules, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered policy: %w", err)
	}

	result := make([][]string, len(rules))
	for i, rule := range rules {
		result[i] = policyToStringArray(rule)
	}

	return result, nil
}

// GetAllPolicy gets all policy rules from the database.
// This is a helper method for debugging and inspection.
func (a *EntAdapter) GetAllPolicy() ([][]string, error) {
	ctx := context.Background()

	rules, err := a.client.CasbinRule.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all policies: %w", err)
	}

	result := make([][]string, len(rules))
	for i, rule := range rules {
		// Include ptype in the result
		line := []string{rule.Ptype}
		line = append(line, policyToStringArray(rule)...)
		result[i] = line
	}

	return result, nil
}

// parseCSVLine parses a CSV line into ptype and rule.
// Format: "p, role, resource, action" or "g, user, role, domain"
func parseCSVLine(line string) (ptype string, rule []string, ok bool) {
	parts := strings.Split(line, ",")
	if len(parts) < 2 {
		return "", nil, false
	}

	ptype = strings.TrimSpace(parts[0])
	rule = make([]string, len(parts)-1)
	for i := 1; i < len(parts); i++ {
		rule[i-1] = strings.TrimSpace(parts[i])
	}

	return ptype, rule, true
}
