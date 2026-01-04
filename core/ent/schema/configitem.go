package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Configitem holds the schema definition for the Configitem entity.
type Configitem struct {
	ent.Schema
}

// Fields of the Configitem.
func (Configitem) Fields() []ent.Field {
	return []ent.Field{
		// 核心字段
		field.String("key").
			Unique().
			NotEmpty().
			Comment("配置项的键，如 app.name"),
		field.String("value").
			Comment("配置项的值（字符串）"),
		field.Bool("is_dynamic").
			Default(false).
			Comment("是否为动态配置（db:true）"),

		// 状态管理
		field.Enum("status").
			Values("active", "inactive").
			Default("active").
			Comment("配置项状态，支持软删除"),

		// 审计字段（自动时间戳）
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

// Indexes of the Configitem.
func (Configitem) Indexes() []ent.Index {
	return []ent.Index{
		// key 唯一索引（主查询字段）
		index.Fields("key").Unique(),
		// status 索引（支持按状态筛选）
		index.Fields("status"),
	}
}

// Edges of the Configitem.
func (Configitem) Edges() []ent.Edge {
	return nil
}
