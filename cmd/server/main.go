package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	filedump "github.com/justEngineer/go-metrics-service/internal/filestorage"
	compression "github.com/justEngineer/go-metrics-service/internal/gzip"
	server "github.com/justEngineer/go-metrics-service/internal/http/server"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
)

func main() {
	config := server.Parse()
	MetricStorage := storage.New()
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	appLogger, err := logger.New(config.LogLevel)
	filedump.New(MetricStorage, &config, ctx, appLogger)
	if err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}
	ServerHandler := server.New(MetricStorage, &config, appLogger)
	r := chi.NewRouter()
	r.Use(appLogger.RequestLogger)
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

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel
}
