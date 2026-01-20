package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AuditLog holds the schema definition for the AuditLog entity.
type AuditLog struct {
	ent.Schema
}

// Fields of the AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Unique().
			Immutable().
			Comment("Unique identifier for audit log entry"),
		field.Time("timestamp").
			Default(time.Now).
			Immutable().
			Comment("Timestamp when the action occurred"),
		field.UUID("operator_id", uuid.UUID{}).
			Optional().
			Comment("UUID of the user who performed the action"),
		field.String("action").
			NotEmpty().
			Comment("Action type (e.g., user.update_role, auth.login)"),
		field.String("target_id").
			Optional().
			Comment("ID of the target resource"),
		field.String("target_type").
			Optional().
			Comment("Type of the target resource (e.g., user, project, config)"),
		field.JSON("changes", map[string]interface{}{}).
			Optional().
			Comment("Detailed changes made (before/after values)"),
		field.String("ip_address").
			MaxLen(45).
			Optional().
			Comment("IP address of the requester (supports IPv6)"),
		field.String("user_agent").
			MaxLen(512).
			Optional().
			Comment("User agent string from HTTP request"),
		field.Int("status_code").
			Optional().
			Comment("HTTP response status code"),
		field.Int("response_time_ms").
			Optional().
			Comment("Response time in milliseconds"),
		field.String("method").
			MaxLen(10).
			Optional().
			Comment("HTTP method (GET, POST, PUT, DELETE, etc.)"),
		field.String("path").
			MaxLen(512).
			Optional().
			Comment("Request path"),
	}
}

// Indexes of the AuditLog.
func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		// Index for time-based queries
		index.Fields("timestamp"),
		// Index for user-based queries
		index.Fields("operator_id"),
		// Index for action type filtering
		index.Fields("action"),
		// Composite index for target resource queries
		index.Fields("target_type", "target_id"),
	}
}
