package main

import (
	"syscall"
	"testing"

	"context"
	"os"
	"os/signal"
	"time"

	storage "github.com/justEngineer/go-metrics-service/internal"
	client "github.com/justEngineer/go-metrics-service/internal/http/client"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	MetricStorage := storage.New()
	config := client.Parse()
	ClientHandler := client.New(MetricStorage, &config)
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
