package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	server := routing.ServerStart(appLogger, ServerHandler, &cfg)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signalChannel

	log.Println("Shutting down the server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped.")
}
