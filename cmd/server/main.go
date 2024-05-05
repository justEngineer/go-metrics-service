package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	database "github.com/justEngineer/go-metrics-service/internal/database"
	filedump "github.com/justEngineer/go-metrics-service/internal/filestorage"
	compression "github.com/justEngineer/go-metrics-service/internal/gzip"
	config "github.com/justEngineer/go-metrics-service/internal/http/server/config"
	server "github.com/justEngineer/go-metrics-service/internal/http/server/handlers"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	security "github.com/justEngineer/go-metrics-service/internal/security"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
)

func main() {
	cfg := config.Parse()
	MetricStorage := storage.New()
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	dbConnecton, err := database.NewConnection(ctx, &cfg)
	if err != nil {
		log.Printf("Database connection failed %s", err)
	} else {
		defer dbConnecton.Connections.Close()
	}
	appLogger, err := logger.New(cfg.LogLevel)
	filedump.New(MetricStorage, &cfg, ctx, appLogger)
	if err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}
	ServerHandler := server.New(MetricStorage, &cfg, appLogger, dbConnecton)
	r := chi.NewRouter()
	r.Use(appLogger.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(compression.GzipMiddleware)
	if cfg.SHA256Key != "" {
		r.Use(security.New(cfg.SHA256Key))
	}
	r.Post("/update/{type}/{name}/{value}", ServerHandler.UpdateMetric)
	r.Get("/value/{type}/{name}", ServerHandler.GetMetric)
	r.Get("/", ServerHandler.MainPage)
	r.Post("/update/", ServerHandler.UpdateMetricFromJSON)
	r.Post("/updates/", server.TimeoutMiddleware(time.Second, ServerHandler.UpdateMetricsFromBatch))
	r.Post("/value/", ServerHandler.GetMetricAsJSON)
	r.Get("/ping", ServerHandler.CheckDBConnection)

	port := strings.Split(cfg.Endpoint, ":")
	log.Fatal(http.ListenAndServe(":"+port[1], r))

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel
}
