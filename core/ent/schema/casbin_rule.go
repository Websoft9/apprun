package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CasbinRule holds the schema definition for the CasbinRule entity.
// This is used for database-backed Casbin policy storage (optional).
type CasbinRule struct {
	ent.Schema
}

// Fields of the CasbinRule.
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("ptype").
			MaxLen(100).
			Comment("Policy type: p (policy) or g (grouping)"),
		field.String("v0").
			MaxLen(100).
			Optional().
			Comment("Subject (user/role)"),
		field.String("v1").
			MaxLen(100).
			Optional().
			Comment("Domain (project_id) or Role for grouping"),
		field.String("v2").
			MaxLen(100).
			Optional().
			Comment("Object (resource) for policy"),
		field.String("v3").
			MaxLen(100).
			Optional().
			Comment("Action for policy"),
		field.String("v4").
			MaxLen(100).
			Optional(),
		field.String("v5").
			MaxLen(100).
			Optional(),
	}
}

// Indexes of the CasbinRule.
func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ptype"),
		index.Fields("v0"),
		index.Fields("v1"),
	}
}
