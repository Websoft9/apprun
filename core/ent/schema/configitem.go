package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Configitem holds the schema definition for the Configitem entity.
type Configitem struct {
	ent.Schema
}

// Fields of the Configitem.
func (Configitem) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			Unique().
			NotEmpty().
			Comment("配置项的键，如 poc.enabled"),
		field.String("value").
			Comment("配置项的值（JSON字符串）"),
		field.Bool("is_dynamic").
			Default(false).
			Comment("是否为动态配置（db:true）"),
	}
}

// Edges of the Configitem.
func (Configitem) Edges() []ent.Edge {
	return nil
}
