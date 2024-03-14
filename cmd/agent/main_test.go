package main

import (
	"testing"

	storage "github.com/justEngineer/go-metrics-service/internal"
	client "github.com/justEngineer/go-metrics-service/internal/http/client"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	MetricStorage := storage.New()
	ClientHandler := client.New(MetricStorage, nil)
	ClientHandler.GetMetrics(MetricStorage)
	assert.Equal(t, int64(1), MetricStorage.Counter["PollCount"], "Количество запросов метрик совпадает с ожидаемым")
}
