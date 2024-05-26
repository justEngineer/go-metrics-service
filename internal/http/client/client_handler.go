package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"context"
	"fmt"

	"compress/gzip"

	async "github.com/justEngineer/go-metrics-service/internal/async"
	logger "github.com/justEngineer/go-metrics-service/internal/logger"
	model "github.com/justEngineer/go-metrics-service/internal/models"
	security "github.com/justEngineer/go-metrics-service/internal/security"
	storage "github.com/justEngineer/go-metrics-service/internal/storage"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

type Handler struct {
	storage   *storage.MemStorage
	config    *ClientConfig
	appLogger *logger.Logger
}

func New(metricsService *storage.MemStorage, config *ClientConfig, appLogger *logger.Logger) *Handler {
	return &Handler{metricsService, config, appLogger}
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
			h.GetAdditionalMetrics()
			h.storage.Mutex.Unlock()
		}
	}
}

func (h *Handler) sendRequest(metric []model.Metrics, url *string, client *http.Client, limiter *async.Semaphore) error {
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
	if h.config.SHA256Key != "" {
		signedBody, err := security.AddSign(body, h.config.SHA256Key)
		if err != nil {
			return fmt.Errorf("error while adding SHA256 sign: %w", err)
		}
		request.Header.Set(security.HashHeader, hex.EncodeToString(signedBody))
	}
	request.Close = true
	if limiter != nil {
		limiter.Acquire()
		defer limiter.Release()
	}
	response, err := client.Do(request)
	if err != nil {
		h.appLogger.Log.Info("request sending is failed", zap.String("error", err.Error()))
		return err
	}
	response.Body.Close()
	return nil
}

func (h *Handler) SendMetrics(ctx context.Context, client *http.Client, limiter *async.Semaphore) {
	sendTicker := time.NewTicker(time.Duration(h.config.reportInterval) * time.Second)
	defer sendTicker.Stop()
	url := "http://" + h.config.endpoint + "/updates/"
	for {
		select {
		case <-ctx.Done():
			return
		case <-sendTicker.C:
			h.storage.Mutex.Lock()
			var metricsBatch []model.Metrics
			for id, value := range h.storage.Gauge {
				metric := model.Metrics{
					ID:    id,
					MType: "gauge",
					Value: &value,
				}
				metricsBatch = append(metricsBatch, metric)
			}
			for id, value := range h.storage.Counter {
				metric := model.Metrics{
					ID:    id,
					MType: "counter",
					Delta: &value,
				}
				metricsBatch = append(metricsBatch, metric)
			}
			err := h.sendRequest(metricsBatch, &url, client, limiter)
			if err != nil {
				h.appLogger.Log.Info("request sending is failed", zap.String("error", err.Error()))
			}
			h.storage.Mutex.Unlock()
		}
	}
}

func (h *Handler) GetAdditionalMetrics() {
	cpuPercents, err := cpu.Percent(0, true)
	if err != nil {
		h.appLogger.Log.Info("getting CPU percentage failed", zap.String("error", err.Error()))
		return
	}
	h.storage.Gauge["CPUtilization1"] = float64(cpuPercents[1])

	v, err := mem.SwapMemory()
	if err != nil {
		h.appLogger.Log.Info("getting swap memory metrics failed", zap.String("error", err.Error()))
		return
	}
	h.storage.Gauge["TotalMemory"] = float64(v.Total / (1024 * 1024))
	h.storage.Gauge["FreeMemory"] = float64(v.Free / (1024 * 1024))
}
