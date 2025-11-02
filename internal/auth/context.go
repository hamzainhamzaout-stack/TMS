package auth

import (
	"context"
	"log"
)

type contextKey string

const UserContextKey = contextKey("user")

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserContextKey).(string)
	log.Printf("UserID: %v | Ok: %v", userID, ok)
	return userID, ok
}
