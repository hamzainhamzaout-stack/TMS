package auth

import (
	"context"
	"net/http"

	"github.com/Robotech-Org/driver-management/utils/responses"
	"github.com/Robotech-Org/driver-management/utils/tokenGen"
)

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
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
