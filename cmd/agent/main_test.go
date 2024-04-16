package main

import (
	"log"
	"syscall"
	"testing"

	"context"
	"os"
	"os/signal"
	"time"

	client "github.com/justEngineer/go-metrics-service/internal/http/client"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	MetricStorage := storage.New()
	config := client.Parse()
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
	MetricStorage.Mutex.Lock()
	defer MetricStorage.Mutex.Unlock()
	assert.Equal(t, int64(1), MetricStorage.Counter["PollCount"], "Количество запросов метрик совпадает с ожидаемым")
}
