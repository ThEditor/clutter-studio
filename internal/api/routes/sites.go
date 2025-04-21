package routes

import (
	"encoding/json"
	"net/http"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/api/middlewares"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/ThEditor/clutter-studio/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateRequest struct {
	SiteUrl string `json:"site_url" validate:"required,url"`
}

type AnalyticsRequest struct {
	From string `json:"from" validate:"YYYYMMDDdate"`
	To   string `json:"to" validate:"YYYYMMDDdate"`
}

type AnalyticsResponse struct {
	TopPages       []storage.PageStats     `json:"top_pages"`
	DeviceStats    []storage.DeviceStats   `json:"device_stats"`
	PageViews      int                     `json:"page_views"`
	TopReferrers   []storage.ReferrerStats `json:"top_referrers"`
	UniqueVisitors int                     `json:"unique_visitors"`
	VisitorGraph   []storage.VisitorStats  `json:"visitor_graph"`
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
		var req AnalyticsRequest
		req.From = r.URL.Query().Get("from")
		req.To = r.URL.Query().Get("to")

		if err := common.Validate.Struct(req); err != nil {
			http.Error(w, "Invalid query parameters: "+err.Error(), http.StatusBadRequest)
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

		topPages, err := s.ClickHouse.GetTopPages(site.ID, 10)

		if err != nil || topPages == nil {
			http.Error(w, "Couldn't find analytics data for site", http.StatusNotFound)
			return
		}

		deviceStats, err := s.ClickHouse.GetDeviceStats(site.ID)

		if err != nil || deviceStats == nil {
			http.Error(w, "Couldn't find analytics data for site", http.StatusNotFound)
			return
		}

		pageViews, err := s.ClickHouse.GetPageViews(site.ID)

		if err != nil {
			http.Error(w, "Couldn't find analytics data for site", http.StatusNotFound)
			return
		}

		topReferrers, err := s.ClickHouse.GetTopReferrers(site.ID, 10)

		if err != nil || topReferrers == nil {
			http.Error(w, "Couldn't find analytics data for site", http.StatusNotFound)
			return
		}

		uniqueVisitors, err := s.ClickHouse.GetUniqueVisitors(site.ID)

		if err != nil {
			http.Error(w, "Couldn't find analytics data for site", http.StatusNotFound)
			return
		}

		visitorGraph, err := s.ClickHouse.GetVisitorGraph(site.ID, req.From, req.To)

		if err != nil || visitorGraph == nil {
			http.Error(w, "Couldn't find analytics data for site", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(&AnalyticsResponse{
			TopPages:       topPages,
			DeviceStats:    deviceStats,
			PageViews:      pageViews,
			TopReferrers:   topReferrers,
			UniqueVisitors: uniqueVisitors,
			VisitorGraph:   visitorGraph,
		})
	})

	return r
}
