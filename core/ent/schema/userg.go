// filepath: /data/cdl/apprun/core/ent/schema/users.go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Users struct {
	ent.Schema
}

func (Users) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("email").Unique(),
	}
}

func (Users) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("servers", Servers.Type), // 移除.Required()
	}
}
