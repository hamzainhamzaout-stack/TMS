package auth

import (
	"context"
)

type contextKey string

const UserContextKey = contextKey("user")
const RoleContextKey = contextKey("role")

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserContextKey).(string)
	return userID, ok
}

func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleContextKey).(string)
	return role, ok
}
