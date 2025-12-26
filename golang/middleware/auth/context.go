package auth

import "context"

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// userIDKey is the context key for storing the user ID
	userIDKey contextKey = "user_id"
)

// WithUserID returns a new context with the user ID set
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts the user ID from the context
// Returns empty string if not found
func UserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// MustUserIDFromContext extracts the user ID from the context
// Panics if not found (use after auth middleware has run)
func MustUserIDFromContext(ctx context.Context) string {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		panic("user ID not found in context - auth middleware not configured?")
	}
	return userID
}
