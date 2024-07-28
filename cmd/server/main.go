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

	database "github.com/justEngineer/go-metrics-service/internal/database"
	filedump "github.com/justEngineer/go-metrics-service/internal/filestorage"
	config "github.com/justEngineer/go-metrics-service/internal/http/server/config"
	server "github.com/justEngineer/go-metrics-service/internal/http/server/handlers"
	routing "github.com/justEngineer/go-metrics-service/internal/http/server/routing"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
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

	router := chi.NewRouter()
	routing.SetMiddlewares(router, appLogger, &cfg.SHA256Key)
	routing.SetRequestRouting(router, ServerHandler)

	port := strings.Split(cfg.Endpoint, ":")
	log.Fatal(http.ListenAndServe(":"+port[1], router))

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel
}
