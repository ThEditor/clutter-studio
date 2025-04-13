package routes

import (
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/go-chi/chi/v5"
)

func UsersRouter(s *common.Server) http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello Users!"))
	})

	return r
}
