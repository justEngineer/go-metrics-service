package server

import (
	"context"
	"log"
	"net/http"

	"time"

	database "github.com/justEngineer/go-metrics-service/internal/database"
	"github.com/justEngineer/go-metrics-service/internal/handlers/server/config"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	"github.com/justEngineer/go-metrics-service/internal/storage"

	"github.com/go-chi/chi/v5"
)

func Example() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r := chi.NewRouter()
	var cfg config.ServerConfig

	cfg.Endpoint = "localhost:8080"
	cfg.DatabaseDSN = "postgresql://localhost/dbname"
	dbConnecton, _ := database.NewConnection(ctx, &cfg)
	appLogger, _ := logger.New(cfg.LogLevel)
	MetricStorage := storage.New()
	ServerHandler := New(MetricStorage, &cfg, appLogger, dbConnecton)

	r.Post("/update/{type}/{name}/{value}", ServerHandler.UpdateMetric)
	r.Get("/value/{type}/{name}", ServerHandler.GetMetric)
	r.Get("/", ServerHandler.MainPage)
	r.Post("/update/", ServerHandler.UpdateMetricFromJSON)
	r.Post("/updates/", TimeoutMiddleware(time.Second, ServerHandler.UpdateMetricsFromBatch))
	r.Post("/value/", ServerHandler.GetMetricAsJSON)
	r.Get("/ping", ServerHandler.CheckDBConnection)

	log.Fatal(http.ListenAndServe(cfg.Endpoint, r))
	// server.Shutdown(context.Background())
}
