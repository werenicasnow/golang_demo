package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"database/sql"
	"demo/app-1/cache"
	"demo/app-1/config"
	"demo/app-1/domain"
	"demo/app-1/usecase"
)

type LinkHandler struct {
	usecase *usecase.LinkUsecase
	cache   *cache.Cache
	config  *config.Config
}

func NewLinkHandler(u *usecase.LinkUsecase, c *cache.Cache, cfg *config.Config) *LinkHandler {
	return &LinkHandler{
		usecase: u,
		cache:   c,
		config:  cfg,
	}
}

func (h *LinkHandler) GetLinks(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	offsetStr := r.URL.Query().Get("offset")
	offset, _ := strconv.Atoi(offsetStr)

	links, err := h.usecase.GetLinks(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func (h *LinkHandler) PostLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.CreateLinkRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortCode, err := h.usecase.CreateShortLink(req)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"short_code": shortCode}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *LinkHandler) GetLink(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		http.Error(w, "short_code is required", http.StatusBadRequest)
		return
	}

	originalURL, visits, err := h.usecase.GetLinkAndRedirect(shortCode)
	if err == sql.ErrNoRows {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"url":    originalURL,
		"visits": visits,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LinkHandler) DeleteLink(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		http.Error(w, "short_code is required", http.StatusBadRequest)
		return
	}

	deleted, err := h.usecase.DeleteLink(shortCode)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !deleted {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	cacheKey := "stats:" + string(shortCode)
	h.cache.Delete(cacheKey)
	w.WriteHeader(http.StatusNoContent)
}

func (h *LinkHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	if shortCode == "" {
		http.Error(w, "short_code is required", http.StatusBadRequest)
		return
	}

	cacheKey := "stats:" + shortCode
	if cachedData, found := h.cache.Get(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedData)
		return
	}

	response, err := h.usecase.GetLinkStats(shortCode)
	if err == sql.ErrNoRows {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.cache.Set(cacheKey, response, h.config.CacheTTL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}