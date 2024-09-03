// Package routing предоставляет обработчики для HTTP запросов.
package routing

import (
	"crypto/rsa"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justEngineer/go-metrics-service/internal/encryption"
	compression "github.com/justEngineer/go-metrics-service/internal/gzip"
	"github.com/justEngineer/go-metrics-service/internal/handlers/server/config"
	server "github.com/justEngineer/go-metrics-service/internal/handlers/server/http"
	profiler "github.com/justEngineer/go-metrics-service/internal/handlers/server/http/profiler"
	"github.com/justEngineer/go-metrics-service/internal/logger"
	"github.com/justEngineer/go-metrics-service/internal/security"
)

func ServerStart(appLogger *logger.Logger, ServerHandler *server.Handler, cfg *config.ServerConfig) *http.Server {

	router := chi.NewRouter()
	SetMiddlewares(router, appLogger, &cfg.SHA256Key, cfg.PrivateCryptoKey)
	SetRequestRouting(router, ServerHandler, cfg.PrivateCryptoKey, &cfg.TrustedSubnet)

	endpoint := ":" + (strings.Split(cfg.Endpoint, ":"))[1]
	server := &http.Server{
		Addr:    endpoint,
		Handler: router,
	}

	log.Printf("Running server on endpoint: %s\n", cfg.Endpoint)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on endpoint: %s, error: %v\n", cfg.Endpoint, err)
	}
	return server
}

// SetMiddlewares добавляет промежуточные обработчики запросов.
func SetMiddlewares(router *chi.Mux, appLogger *logger.Logger, SHA256Key *string, cryptoKey *rsa.PrivateKey) {
	router.Use(appLogger.RequestLogger)
	router.Use(middleware.Recoverer)
	router.Use(compression.GzipMiddleware)
	if *SHA256Key != "" {
		router.Use(security.New(*SHA256Key))
	}
	router.Use(encryption.BodyDecrypt(cryptoKey))
}

// SetRequestRouting добавляет обработчики для HTTP запросов.
func SetRequestRouting(router *chi.Mux, ServerHandler *server.Handler, cryptoKey *rsa.PrivateKey, trustedSubnet *string) {
	router.Mount("/debug", profiler.Profiler())
	router.Post("/update/{type}/{name}/{value}", ServerHandler.UpdateMetric)
	router.Get("/value/{type}/{name}", ServerHandler.GetMetric)
	router.Get("/", ServerHandler.MainPage)
	router.Post("/update/", ServerHandler.UpdateMetricFromJSON)

	router.Route("/updates", func(r chi.Router) {
		r.Post("/",
			security.SubnetCheckerMiddleware(trustedSubnet)(
				encryption.DecryptMiddleware(cryptoKey)(
					server.TimeoutMiddleware(time.Second, ServerHandler.UpdateMetricsFromBatch),
				),
			))
	})

	router.Post("/updates/", server.TimeoutMiddleware(time.Second, ServerHandler.UpdateMetricsFromBatch))
	router.Post("/value/", ServerHandler.GetMetricAsJSON)
	router.Get("/ping", ServerHandler.CheckDBConnection)
}
