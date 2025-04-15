package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/go-chi/chi/v5"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func AuthRouter(s *common.Server) http.Handler {
	r := chi.NewRouter()

	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := common.Validate.Struct(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hashedPassword, err := common.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		user, err := s.Repo.Create(s.Ctx, repository.CreateParams{
			Username: req.Username,
			Email:    req.Email,
			Passhash: hashedPassword,
		})

		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		jwt, err := common.CreateJWT(user.ID, user.Email)

		if err != nil {
			http.Error(w, "Failed creating JWT", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message":      "Successfully created!",
			"access_token": jwt,
		})
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := common.Validate.Struct(req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		hashedPassword, err := common.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		user, err := s.Repo.FindPlayerByEmail(s.Ctx, req.Email)

		if err != nil || user.Passhash != hashedPassword {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		jwt, err := common.CreateJWT(user.ID, user.Email)

		if err != nil {
			http.Error(w, "Failed creating JWT", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message":      "Successfully logged in!",
			"access_token": jwt,
		})
	})

	return r
}
