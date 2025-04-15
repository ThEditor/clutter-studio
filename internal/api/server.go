package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/api/routes"
	"github.com/ThEditor/clutter-studio/internal/log"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func Start(ctx context.Context, address string, port int, repo *repository.Queries) {
	s := &common.Server{
		Ctx:  ctx,
		Repo: repo,
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.Mount("/auth", routes.AuthRouter(s))
	r.Mount("/users", routes.UsersRouter(s))

	log.Info("API server listening on " + address + ":" + strconv.Itoa(port))
	err := http.ListenAndServe(address+":"+strconv.Itoa(port), r)
	if err != nil {
		log.Info("Server failed to start: " + err.Error())
	}
}
