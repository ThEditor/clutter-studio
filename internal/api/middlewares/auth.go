package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func baseAuthMiddleware(next http.Handler, requireEmailVerification bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("accessToken")

		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "cookie not found", http.StatusBadRequest)
			default:
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return
		}

		tokenString := cookie.Value

		claims, err := common.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if requireEmailVerification && !claims.EmailVerified {
			http.Error(w, "Email not verified", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthWithoutEmailVerifiedMiddleware(next http.Handler) http.Handler {
	return baseAuthMiddleware(next, false)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return baseAuthMiddleware(next, true)
}
