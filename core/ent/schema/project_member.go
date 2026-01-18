package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProjectMember holds the schema definition for the ProjectMember entity.
type ProjectMember struct {
	ent.Schema
}

// Fields of the ProjectMember.
func (ProjectMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique(),
		field.Int64("project_id").
			Comment("Foreign key to projects.id"),
		field.Int64("user_id").
			Comment("Foreign key to users.id"),
		field.String("role").
			MaxLen(20).
			NotEmpty().
			Comment("Role: owner, admin, member, viewer"),
		field.Time("joined_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the ProjectMember.
func (ProjectMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("members").
			Field("project_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("project_memberships").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes of the ProjectMember.
func (ProjectMember) Indexes() []ent.Index {
	return []ent.Index{
		// Ensure one user has only one role in one project
		index.Fields("project_id", "user_id").Unique(),
		index.Fields("user_id"),
		index.Fields("role"),
	}
}
