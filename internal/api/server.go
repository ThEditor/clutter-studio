package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/api/routes"
	"github.com/ThEditor/clutter-studio/internal/log"
	"github.com/ThEditor/clutter-studio/internal/mailer"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/ThEditor/clutter-studio/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func Start(ctx context.Context, address string, port int, repo *repository.Queries, clickhouse *storage.ClickHouseStorage, mailer *mailer.Mailer) {
	s := &common.Server{
		Ctx:        ctx,
		Repo:       repo,
		ClickHouse: clickhouse,
		Mailer:     mailer,
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:6789", "http://127.0.0.1:6789", "https://clutter.phy0.in"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Logger)
	r.Use(httprate.LimitByRealIP(100, time.Minute))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.Mount("/auth", routes.AuthRouter(s))
	r.Mount("/users", routes.UsersRouter(s))
	r.Mount("/sites", routes.SitesRouter(s))

	log.Info("API server listening on " + address + ":" + strconv.Itoa(port))
	err := http.ListenAndServe(address+":"+strconv.Itoa(port), r)
	if err != nil {
		log.Info("Server failed to start: " + err.Error())
	}
}
