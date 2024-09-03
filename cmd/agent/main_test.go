package main

import (
	"log"
	"syscall"
	"testing"

	"context"
	"os"
	"os/signal"
	"time"

	cfg "github.com/justEngineer/go-metrics-service/internal/handlers/client"
	client "github.com/justEngineer/go-metrics-service/internal/handlers/client/http"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	MetricStorage := storage.New()
	config := cfg.Parse()
	appLogger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("Logger wasn't initialized due to %s", err)
	}
	ClientHandler := client.New(MetricStorage, &config, appLogger)
	ctx, cancel := context.WithCancel(context.Background())
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	defer func() {
		signal.Stop(signalChannel)
		cancel()
	}()
	go ClientHandler.GetMetrics(ctx)
	time.Sleep(time.Second * 3)
	MetricStorage.Mutex.RLock()
	defer MetricStorage.Mutex.RUnlock()
	assert.Equal(t, int64(1), MetricStorage.Counter["PollCount"], "Количество запросов метрик совпадает с ожидаемым")
}
