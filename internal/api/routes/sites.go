package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/api/middlewares"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
			"site_id": site.ID.String(),
			"message": "Site " + site.SiteUrl + " added successfully!",
		})
	})

	r.Get("/all", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sites, err := s.Repo.ListSitesByUserID(s.Ctx, claims.UserID)

		if err != nil {
			http.Error(w, "Couldn't fetch list of sites", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(sites)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		siteId, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		site, err := s.Repo.FindSiteByID(s.Ctx, siteId)

		if err != nil {
			http.Error(w, "Couldn't find site", http.StatusNotFound)
			return
		}

		if site.UserID != claims.UserID {
			http.Error(w, "You do not have access to this site", http.StatusForbidden)
			return
		}

		json.NewEncoder(w).Encode(site)
	})

	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		siteId, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		site, err := s.Repo.FindSiteByID(s.Ctx, siteId)

		if err != nil {
			http.Error(w, "Couldn't find site", http.StatusNotFound)
			return
		}

		if site.UserID != claims.UserID {
			http.Error(w, "You do not have access to this site", http.StatusForbidden)
			return
		}

		err = s.Repo.DeleteSite(s.Ctx, repository.DeleteSiteParams{
			ID:     siteId,
			UserID: claims.UserID,
		})

		if err != nil {
			http.Error(w, "Could not delete site", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Site " + site.SiteUrl + " successfully deleted!",
		})
	})

	r.Get("/{id}/analytics", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		siteId, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "Invalid UUID", http.StatusBadRequest)
			return
		}

		site, err := s.Repo.FindSiteByID(s.Ctx, siteId)

		if err != nil {
			http.Error(w, "Couldn't find site", http.StatusNotFound)
			return
		}

		if site.UserID != claims.UserID {
			http.Error(w, "You do not have access to this site", http.StatusForbidden)
			return
		}

		data, err := s.ClickHouse.GetSiteEventData(site.ID)

		if err != nil || data == nil {
			http.Error(w, "Couldn't find analytics data for site", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(data)
	})

	return r
}
