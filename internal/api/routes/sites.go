package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/api/middlewares"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/go-chi/chi/v5"
)

type CreateRequest struct {
	SiteUrl string `json:"site_url" validate:"required,url"`
}

func SitesRouter(s *common.Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middlewares.AuthMiddleware)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := common.Validate.Struct(req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		userId := claims.UserID

		_, err := s.Repo.FindSiteByUserIDAndURL(s.Ctx, repository.FindSiteByUserIDAndURLParams{
			UserID:  userId,
			SiteUrl: req.SiteUrl,
		})

		if err == nil {
			http.Error(w, "Site already exists for this user", http.StatusConflict)
			return
		}

		site, err := s.Repo.CreateSite(s.Ctx, repository.CreateSiteParams{
			UserID:  userId,
			SiteUrl: req.SiteUrl,
		})

		if err != nil {
			http.Error(w, "Couldn't create site", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Site " + site.SiteUrl + " added successfully!",
		})
	})

	return r
}
