package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"vn-admin-api/internal/cache"
	"vn-admin-api/internal/database"
	"vn-admin-api/internal/logger"
)

type Handler struct {
	repo  *database.Repository
	log   *logger.Logger
	cache cache.Cache // Interface - can be MemoryCache or RedisCache
}

func NewHandler(repo *database.Repository, log *logger.Logger) *Handler {
	return &Handler{
		repo:  repo,
		log:   log,
		cache: cache.NewMemoryCache(5 * time.Minute), // Default: in-memory
	}
}

// NewHandlerWithCache allows injecting custom cache (e.g., Redis)
func NewHandlerWithCache(repo *database.Repository, log *logger.Logger, c cache.Cache) *Handler {
	return &Handler{
		repo:  repo,
		log:   log,
		cache: c,
	}
}

// Response helpers - Standard format

func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"data": data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

func (h *Handler) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := map[string]string{
		"error": message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("Failed to encode error response", "error", err)
	}
}

// Health Check Handlers

// HealthCheck returns basic health status (liveness probe)
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ReadyCheck checks if the service is ready (readiness probe)
func (h *Handler) ReadyCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DB().Ping(); err != nil {
		h.log.Error("Readiness check failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

// API Handlers

// GetProvinces handles GET /api/v1/provinces
func (h *Handler) GetProvinces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check cache first
	if cached, ok := h.cache.GetProvinces(ctx); ok {
		h.respondSuccess(w, cached)
		return
	}

	provinces, err := h.repo.GetProvinces(ctx)
	if err != nil {
		h.log.Error("Failed to get provinces", "error", err)
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Store in cache (ignore error for cache)
	_ = h.cache.SetProvinces(ctx, provinces)

	h.respondSuccess(w, provinces)
}

// GetUnitsByProvince handles GET /api/v1/provinces/{id}/units
func (h *Handler) GetUnitsByProvince(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid Province ID")
		return
	}

	ctx := r.Context()

	// Check cache first
	if cached, ok := h.cache.GetUnits(ctx, id); ok {
		h.respondSuccess(w, cached)
		return
	}

	units, err := h.repo.GetUnitsByProvince(ctx, id)
	if err != nil {
		h.log.Error("Failed to get units", "province_id", id, "error", err)
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Store in cache
	_ = h.cache.SetUnits(ctx, id, units)

	h.respondSuccess(w, units)
}

// Search handles GET /api/v1/search?q=...
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		h.respondError(w, http.StatusBadRequest, "Query too short (min 2 chars)")
		return
	}

	ctx := r.Context()
	units, err := h.repo.SearchAdminUnits(ctx, query)
	if err != nil {
		h.log.Error("Failed to search units", "query", query, "error", err)
		h.respondError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	h.respondSuccess(w, units)
}
