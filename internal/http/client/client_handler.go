package client

import (
	"math/rand"
	"net/http"
	"runtime"
	"strconv"

	storage "github.com/justEngineer/go-metrics-service/internal"
)

type Handler struct {
	storage        *storage.MemStorage
	config         *ClientConfig
	sendTimeoutSec int
	pollTimeoutSec int
}

func New(metricsService *storage.MemStorage, config *ClientConfig) *Handler {
	return &Handler{metricsService, config, 0, 0}
}

func (h *Handler) GetMetrics(MetricStorage *storage.MemStorage) {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	MetricStorage.Gauge["Alloc"] = float64(m.Alloc)
	MetricStorage.Gauge["BuckHashSys"] = float64(m.BuckHashSys)
	MetricStorage.Gauge["Frees"] = float64(m.Frees)
	MetricStorage.Gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	MetricStorage.Gauge["GCSys"] = float64(m.GCSys)
	MetricStorage.Gauge["HeapAlloc"] = float64(m.HeapAlloc)
	MetricStorage.Gauge["HeapIdle"] = float64(m.HeapIdle)
	MetricStorage.Gauge["HeapInuse"] = float64(m.HeapInuse)
	MetricStorage.Gauge["HeapObjects"] = float64(m.HeapObjects)
	MetricStorage.Gauge["HeapReleased"] = float64(m.HeapReleased)
	MetricStorage.Gauge["HeapSys"] = float64(m.HeapSys)
	MetricStorage.Gauge["LastGC"] = float64(m.LastGC)
	MetricStorage.Gauge["Lookups"] = float64(m.Lookups)
	MetricStorage.Gauge["MCacheInuse"] = float64(m.MCacheInuse)
	MetricStorage.Gauge["MCacheSys"] = float64(m.MCacheSys)
	MetricStorage.Gauge["MSpanSys"] = float64(m.MSpanSys)
	MetricStorage.Gauge["Mallocs"] = float64(m.Mallocs)
	MetricStorage.Gauge["NextGC"] = float64(m.NextGC)
	MetricStorage.Gauge["NumForcedGC"] = float64(m.NumForcedGC)
	MetricStorage.Gauge["NumGC"] = float64(m.NumGC)
	MetricStorage.Gauge["OtherSys"] = float64(m.OtherSys)
	MetricStorage.Gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	MetricStorage.Gauge["StackInuse"] = float64(m.StackInuse)
	MetricStorage.Gauge["StackSys"] = float64(m.StackSys)
	MetricStorage.Gauge["Sys"] = float64(m.Sys)
	MetricStorage.Gauge["TotalAlloc"] = float64(m.TotalAlloc)
	MetricStorage.Gauge["RandomValue"] = float64(rand.Float64() * 100)

	MetricStorage.Counter["PollCount"] += 1
}

func (h *Handler) sendMetrics(config *ClientConfig, MetricStorage *storage.MemStorage) {
	client := &http.Client{}
	for k, v := range MetricStorage.Gauge {
		url := "http://" + config.endpoint + "/update/gauge/" + k + "/" + strconv.FormatFloat(v, 'f', -1, 64)
		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			panic(err)
		}
		request.Header.Add("Content-Type", "text/plain")
		response, err := client.Do(request)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()
	}
	for k, v := range MetricStorage.Counter {
		url := config.endpoint + "update/counter/" + k + "/" + strconv.FormatInt(v, 10)
		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			panic(err)
		}
		request.Header.Add("Content-Type", "text/plain")
		response, err := client.Do(request)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()
	}
}

func (h *Handler) OnTimer() {
	h.sendTimeoutSec += 1
	h.pollTimeoutSec += 1
	if h.pollTimeoutSec == int(h.config.pollInterval) {
		h.GetMetrics(h.storage)
		h.pollTimeoutSec = 0
	}
	if h.sendTimeoutSec == int(h.config.reportInterval) {
		h.sendMetrics(h.config, h.storage)
		h.sendTimeoutSec = 0
	}
}
