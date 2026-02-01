package api

import (
	"net/http"

	"vn-admin-api/internal/cache"
	"vn-admin-api/internal/database"
	"vn-admin-api/internal/logger"
)

func NewRouter(repo *database.Repository, log *logger.Logger) http.Handler {
	handler := NewHandler(repo, log)
	return buildRouter(handler, log)
}

func NewRouterWithCache(repo *database.Repository, log *logger.Logger, c cache.Cache) http.Handler {
	handler := NewHandlerWithCache(repo, log, c)
	return buildRouter(handler, log)
}

func buildRouter(handler *Handler, log *logger.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health Check Endpoints
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("GET /ready", handler.ReadyCheck)

	// API Routes
	mux.HandleFunc("GET /api/v1/provinces", handler.GetProvinces)
	mux.HandleFunc("GET /api/v1/provinces/{id}/units", handler.GetUnitsByProvince)
	mux.HandleFunc("GET /api/v1/search", handler.Search)

	// Middleware Chain
	return ChainMiddleware(mux,
		RecoveryMiddleware(log),
		LoggerMiddleware(log),
		CORSMiddleware(),
		RateLimitMiddleware(),
		GzipMiddleware(),
	)
}
