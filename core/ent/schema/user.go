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
			Default("").
			Sensitive().
			Comment("Bcrypt password hash (cost=12, empty for system user)"),

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
		field.Text("bio").
			MaxLen(500).
			Optional().
			Comment("User biography / description"),

		// Account Status
		field.Int8("status").
			Default(1).
			Comment("Account status: 0-Disabled, 1-Active"),

		// Platform Role
		field.String("role").
			MaxLen(20).
			Default("platform_user").
			Comment("Platform role: platform_user, platform_admin"),
		field.Bool("is_system").
			Default(false).
			Comment("System user flag (e.g., System, cannot login)"),

		// Token Management
		field.Int("token_version").
			Default(0).
			Comment("Token version for revocation mechanism (incremented on disable/delete)"),

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
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Soft delete timestamp (NULL if not deleted)"),
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
		index.Fields("role"),
		index.Fields("is_system"),
		index.Fields("token_version"),
		index.Fields("deleted_at"),
		index.Fields("created_at"),

		// Composite indexes for soft delete queries
		index.Fields("email", "deleted_at"),

		// Composite indexes for admin user management queries
		// Search index for email and nickname
		index.Fields("email", "nickname"),
		// Filter by role and status
		index.Fields("role", "status"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("servers", Servers.Type),
		edge.To("owned_projects", Project.Type),
		edge.To("project_memberships", ProjectMember.Type),
	}
}
