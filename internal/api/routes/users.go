package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/api/middlewares"
	"github.com/go-chi/chi/v5"
)

func UsersRouter(s *common.Server) http.Handler {
	r := chi.NewRouter()

	// endpoint for basic user info
	r.With(middlewares.AuthWithoutEmailVerifiedMiddleware).
		Get("/me", func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := s.Repo.FindUserByID(s.Ctx, claims.UserID)
			if err != nil {
				http.Error(w, "Cannot find user", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]any{
				"id":             user.ID,
				"username":       user.Username,
				"email":          user.Email,
				"email_verified": user.EmailVerified,
				"created_at":     user.CreatedAt,
				"updated_at":     user.UpdatedAt,
			})
		})

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware)
		// other endpoints
	})

	return r
}
