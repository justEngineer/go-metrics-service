// Package routing предоставляет обработчики для HTTP запросов.
package routing

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	compression "github.com/justEngineer/go-metrics-service/internal/gzip"
	server "github.com/justEngineer/go-metrics-service/internal/http/server/handlers"
	profiler "github.com/justEngineer/go-metrics-service/internal/http/server/profiler"
	"github.com/justEngineer/go-metrics-service/internal/logger"
	"github.com/justEngineer/go-metrics-service/internal/security"
)

// SetMiddlewares добавляет промежуточные обработчики запросов.
func SetMiddlewares(router *chi.Mux, appLogger *logger.Logger, SHA256Key *string) {
	router.Use(appLogger.RequestLogger)
	router.Use(middleware.Recoverer)
	router.Use(compression.GzipMiddleware)
	if *SHA256Key != "" {
		router.Use(security.New(*SHA256Key))
	}
}

// SetRequestRouting добавляет обработчики для HTTP запросов.
func SetRequestRouting(router *chi.Mux, ServerHandler *server.Handler) {
	router.Mount("/debug", profiler.Profiler())
	router.Post("/update/{type}/{name}/{value}", ServerHandler.UpdateMetric)
	router.Get("/value/{type}/{name}", ServerHandler.GetMetric)
	router.Get("/", ServerHandler.MainPage)
	router.Post("/update/", ServerHandler.UpdateMetricFromJSON)
	router.Post("/updates/", server.TimeoutMiddleware(time.Second, ServerHandler.UpdateMetricsFromBatch))
	router.Post("/value/", ServerHandler.GetMetricAsJSON)
	router.Get("/ping", ServerHandler.CheckDBConnection)
}
