package routes

import (
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/go-chi/chi/v5"
)

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func UsersRouter(s *common.Server) http.Handler {
	r := chi.NewRouter()

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello users!"))
	})

	return r
}
