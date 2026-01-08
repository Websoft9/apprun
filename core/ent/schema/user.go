package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		// External Identifier (Public UUID)
		field.UUID("uuid", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable().
			Comment("Public UUID for external API reference (immutable, globally unique)"),

		// Primary Key (Internal ID)
		field.Int64("id").
			Positive().
			Comment("Internal database primary key (not exposed in API)"),

		// Authentication Fields
		field.String("username").
			MaxLen(64).
			Unique().
			Optional().
			Comment("Username for login (optional, alphanumeric + underscore)"),
		field.String("email").
			MaxLen(255).
			NotEmpty().
			Unique().
			Comment("Email address for login (required, unique)"),
		field.String("password_hash").
			MaxLen(255).
			NotEmpty().
			Sensitive().
			Comment("Bcrypt password hash (cost=12)"),

		// Profile Fields
		field.String("nickname").
			MaxLen(64).
			Optional().
			Comment("Display name / nickname"),
		field.String("avatar").
			MaxLen(255).
			Optional().
			Comment("Avatar URL"),
		field.String("phone").
			MaxLen(20).
			Optional().
			Comment("Phone number"),
		field.Int8("gender").
			Default(0).
			Comment("Gender: 0-Unknown, 1-Male, 2-Female"),
		field.String("signature").
			MaxLen(255).
			Optional().
			Comment("User signature / bio"),

		// Account Status
		field.Int8("status").
			Default(1).
			Comment("Account status: 0-Disabled, 1-Active"),

		// Login History
		field.Time("last_login_at").
			Optional().
			Nillable().
			Comment("Last login timestamp"),
		field.String("last_login_ip").
			MaxLen(45).
			Optional().
			Comment("Last login IP address (supports IPv6)"),

		// Localization
		field.String("timezone").
			MaxLen(64).
			Default("UTC").
			Comment("User timezone (e.g., Asia/Shanghai)"),
		field.String("language").
			MaxLen(10).
			Default("zh-CN").
			Comment("Preferred language (e.g., zh-CN, en-US)"),

		// Audit Fields
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Account creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		// Unique indexes for authentication
		index.Fields("email").Unique(),
		index.Fields("username").Unique(),

		// Performance indexes
		index.Fields("status"),
		index.Fields("created_at"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("servers", Servers.Type),
	}
}
