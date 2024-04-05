package client

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"context"

	"compress/gzip"

	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	model "github.com/justEngineer/go-metrics-service/internal/models"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
	"go.uber.org/zap"
)

type Handler struct {
	storage *storage.MemStorage
	config  *ClientConfig
}

func New(metricsService *storage.MemStorage, config *ClientConfig) *Handler {
	return &Handler{metricsService, config}
}

func (h *Handler) GetMetrics(ctx context.Context) {
	pollTicker := time.NewTicker(time.Duration(h.config.pollInterval) * time.Second)
	defer pollTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			m := &runtime.MemStats{}
			runtime.ReadMemStats(m)
			h.storage.Mutex.Lock()
			h.storage.Gauge["Alloc"] = float64(m.Alloc)
			h.storage.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
			h.storage.Gauge["Frees"] = float64(m.Frees)
			h.storage.Gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
			h.storage.Gauge["GCSys"] = float64(m.GCSys)
			h.storage.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
			h.storage.Gauge["HeapIdle"] = float64(m.HeapIdle)
			h.storage.Gauge["HeapInuse"] = float64(m.HeapInuse)
			h.storage.Gauge["HeapObjects"] = float64(m.HeapObjects)
			h.storage.Gauge["HeapReleased"] = float64(m.HeapReleased)
			h.storage.Gauge["HeapSys"] = float64(m.HeapSys)
			h.storage.Gauge["LastGC"] = float64(m.LastGC)
			h.storage.Gauge["Lookups"] = float64(m.Lookups)
			h.storage.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
			h.storage.Gauge["MCacheSys"] = float64(m.MCacheSys)
			h.storage.Gauge["MSpanInuse"] = float64(m.MSpanInuse)
			h.storage.Gauge["MSpanSys"] = float64(m.MSpanSys)
			h.storage.Gauge["Mallocs"] = float64(m.Mallocs)
			h.storage.Gauge["NextGC"] = float64(m.NextGC)
			h.storage.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
			h.storage.Gauge["NumGC"] = float64(m.NumGC)
			h.storage.Gauge["OtherSys"] = float64(m.OtherSys)
			h.storage.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
			h.storage.Gauge["StackInuse"] = float64(m.StackInuse)
			h.storage.Gauge["StackSys"] = float64(m.StackSys)
			h.storage.Gauge["Sys"] = float64(m.Sys)
			h.storage.Gauge["TotalAlloc"] = float64(m.TotalAlloc)
			h.storage.Gauge["RandomValue"] = float64(rand.Float64() * 100)

			h.storage.Counter["PollCount"] += 1
			h.storage.Mutex.Unlock()
			//time.Sleep(time.Second * time.Duration(h.config.pollInterval))
		}
	}
}

func sendRequest(metric model.Metrics, url *string, client *http.Client) error {
	body, err := json.Marshal(metric)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err = gzipWriter.Write(body)
	if err != nil {
		panic(err)
	}
	err = gzipWriter.Close()
	if err != nil {
		panic(err)
	}
	body = buf.Bytes()
	request, err := http.NewRequest(http.MethodPost, *url, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Set("Accept-Encoding", "gzip")
	request.Header.Set("Content-Encoding", "gzip")
	response, err := client.Do(request)
	if err != nil {
		logger.Log.Info("request sending is failed", zap.String("error", err.Error()))
		return err
	}
	response.Body.Close()
	return nil
}

func (h *Handler) SendMetrics(ctx context.Context) {
	sendTicker := time.NewTicker(time.Duration(h.config.reportInterval) * time.Second)
	defer sendTicker.Stop()

	client := http.Client{Timeout: time.Second * 1}
	url := "http://" + h.config.endpoint + "/update/"
	for {
		select {
		case <-ctx.Done():
			return
		case <-sendTicker.C:
			h.storage.Mutex.Lock()
			for id, value := range h.storage.Gauge {
				metric := model.Metrics{
					ID:    id,
					MType: "gauge",
					Value: &value,
				}
				if sendRequest(metric, &url, &client) != nil {
					break
				}
			}
			for id, value := range h.storage.Counter {
				metric := model.Metrics{
					ID:    id,
					MType: "counter",
					Delta: &value,
				}
				if sendRequest(metric, &url, &client) != nil {
					break
				}
			}
			h.storage.Mutex.Unlock()
		}
	}
}
