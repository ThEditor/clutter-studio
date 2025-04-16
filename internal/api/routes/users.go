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
	r.Use(middlewares.AuthMiddleware)

	r.Post("/me", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := s.Repo.FindPlayerByID(s.Ctx, claims.UserID)
		if err != nil {
			http.Error(w, "Cannot find user", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"id":         user.ID.String(),
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt.String(),
			"updated_at": user.UpdatedAt.String(),
		})
	})

	return r
}
