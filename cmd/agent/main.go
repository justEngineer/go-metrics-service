package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	async "github.com/justEngineer/go-metrics-service/internal/async"
	client "github.com/justEngineer/go-metrics-service/internal/http/client"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
)

func main() {
	MetricStorage := storage.New()
	config := client.Parse()
	appLogger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}
	ClientHandler := client.New(MetricStorage, &config, appLogger)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ClientHandler.GetMetrics(ctx)
		stop()
	}()
	wg.Add(1)
	requestLimiter := async.NewSemaphore(int(config.RateLimit))
	go func() {
		defer wg.Done()
		client := http.Client{}
		ClientHandler.SendMetrics(ctx, &client, requestLimiter)
		stop()
	}()
	wg.Wait()
}
