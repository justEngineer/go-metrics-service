package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	async "github.com/justEngineer/go-metrics-service/internal/async"
	"github.com/justEngineer/go-metrics-service/internal/encryption"
	cfg "github.com/justEngineer/go-metrics-service/internal/handlers/client"
	client "github.com/justEngineer/go-metrics-service/internal/handlers/client/http"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
)

func main() {
	MetricStorage := storage.New()
	config := cfg.Parse()
	appLogger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}

	ClientHandler := client.New(MetricStorage, &config, appLogger)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ClientHandler.GetMetrics(ctx)
	}()

	wg.Add(1)
	requestLimiter := async.NewSemaphore(int(config.RateLimit))
	go func() {
		defer wg.Done()
		client := http.Client{
			Transport: encryption.EncryptionMiddleware{
				Proxied:   http.DefaultTransport,
				PublicKey: config.PublicCryptoKey,
			}}
		ClientHandler.SendMetrics(ctx, &client, requestLimiter)
	}()

	<-signalChannel
	log.Println("Shutting down the agent...")

	stop()
	wg.Wait()
	log.Println("Agent stopped.")
}
