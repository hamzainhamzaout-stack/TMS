package auth

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/Robotech-Org/driver-management/utils/tokenGen"
)

type claimsKeyType struct{}

var ClaimsKey = claimsKeyType{}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("accessToken")
		if err != nil {
			responses.ErrorJSONResponse(err, http.StatusUnauthorized, w, nil)
			return
		}

		tokenString := cookie.Value
		claims, err := tokenGen.ParseToken(tokenString)
		if err != nil {
			responses.ErrorJSONResponse(err, http.StatusUnauthorized, w, nil)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleContextKey, claims.RoleName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole must run after AuthMiddleware. It rejects requests whose
// access token does not carry one of the allowed roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := GetRoleFromContext(r.Context())
			if !slices.Contains(roles, role) {
				responses.ErrorJSONResponse(errors.New("forbidden"), http.StatusForbidden, w, nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
