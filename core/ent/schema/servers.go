// filepath: /data/cdl/apprun/core/ent/schema/servers.go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Servers struct {
	ent.Schema
}

func (Servers) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("ip").Unique(),
		// 移除user_id字段，让Ent自动创建外键
	}
}

func (Servers) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", Users.Type).Ref("servers").Unique().Required(), // 保持.Required()
	}
}
