package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	compression "github.com/justEngineer/go-metrics-service/internal/gzip"
	server "github.com/justEngineer/go-metrics-service/internal/http/server"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
)

func main() {
	config := server.Parse()
	MetricStorage := storage.New()
	ServerHandler := server.New(MetricStorage, &config)
	r := chi.NewRouter()

	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)
	// gzipMiddleware := middleware.NewCompressor(gzip.BestCompression)
	// r.Use(gzipMiddleware.Handler)
	r.Use(compression.GzipMiddleware)
	r.Post("/update/{type}/{name}/{value}", ServerHandler.UpdateMetric)
	r.Get("/value/{type}/{name}", ServerHandler.GetMetric)
	r.Get("/", ServerHandler.MainPage)
	r.Post("/update/", ServerHandler.UpdateMetricFromJSON)
	r.Post("/value/", ServerHandler.GetMetricAsJSON)

	port := strings.Split(config.Endpoint, ":")
	log.Fatal(http.ListenAndServe(":"+port[1], r))

}
