package jwt

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	value, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}

	return value, true
}
