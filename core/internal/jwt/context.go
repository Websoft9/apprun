package jwt

import "context"

// contextKey is a private type to avoid context key collisions
type contextKey struct{ name string }

var (
	// UserIDKey is the context key for user ID
	UserIDKey = &contextKey{"user_id"}
	// UsernameKey is the context key for username
	UsernameKey = &contextKey{"username"}
	// EmailKey is the context key for email
	EmailKey = &contextKey{"email"}
	// ProjectIDKey is the context key for project ID (for RBAC)
	ProjectIDKey = &contextKey{"project_id"}
)

// SetUserID injects user ID into context
func SetUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// SetUsername injects username into context
func SetUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, UsernameKey, username)
}

// SetEmail injects email into context
func SetEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, EmailKey, email)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) int64 {
	if userID, ok := ctx.Value(UserIDKey).(int64); ok {
		return userID
	}
	return 0
}

// GetUsername retrieves username from context
func GetUsername(ctx context.Context) string {
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		return username
	}
	return ""
}

// GetEmail retrieves email from context
func GetEmail(ctx context.Context) string {
	if email, ok := ctx.Value(EmailKey).(string); ok {
		return email
	}
	return ""
}

// SetProjectID injects project ID into context
func SetProjectID(ctx context.Context, projectID int64) context.Context {
	return context.WithValue(ctx, ProjectIDKey, projectID)
}

// GetProjectID retrieves project ID from context
func GetProjectID(ctx context.Context) int64 {
	if projectID, ok := ctx.Value(ProjectIDKey).(int64); ok {
		return projectID
	}
	return 0
}
