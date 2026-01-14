package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Comment("Internal database primary key for foreign key relationships"),
		field.String("uuid").
			MaxLen(36).
			Unique().
			Immutable().
			DefaultFunc(func() string {
				return uuid.New().String()
			}).
			Comment("External UUID for API usage in REST paths"),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			MaxLen(500).
			Optional(),
		field.Int64("owner_id").
			Comment("Project owner user_id, foreign key to users.id"),
		field.Int8("status").
			Default(1).
			Comment("Status: 0-disabled, 1-enabled, 2-archived"),
		field.Time("created_at").
			Immutable().
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("owned_projects").
			Field("owner_id").
			Unique().
			Required(),
		edge.To("members", ProjectMember.Type),
	}
}

// Indexes of the Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uuid").Unique(),
		index.Fields("owner_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
