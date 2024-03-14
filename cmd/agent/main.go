package main

import (
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type MemStorage struct {
	// указаны некоторые поля структуры
	gauge   map[string]float64
	counter map[string]int64
}

var MetricStorage MemStorage

func getMetrics() {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	MetricStorage.gauge["Alloc"] = float64(m.Alloc)
	MetricStorage.gauge["BuckHashSys"] = float64(m.BuckHashSys)
	MetricStorage.gauge["Frees"] = float64(m.Frees)
	MetricStorage.gauge["GCCPUFraction"] = float64(m.GCCPUFraction)
	MetricStorage.gauge["GCSys"] = float64(m.GCSys)
	MetricStorage.gauge["HeapAlloc"] = float64(m.HeapAlloc)
	MetricStorage.gauge["HeapIdle"] = float64(m.HeapIdle)
	MetricStorage.gauge["HeapInuse"] = float64(m.HeapInuse)
	MetricStorage.gauge["HeapObjects"] = float64(m.HeapObjects)
	MetricStorage.gauge["HeapReleased"] = float64(m.HeapReleased)
	MetricStorage.gauge["HeapSys"] = float64(m.HeapSys)
	MetricStorage.gauge["LastGC"] = float64(m.LastGC)
	MetricStorage.gauge["Lookups"] = float64(m.Lookups)
	MetricStorage.gauge["MCacheInuse"] = float64(m.MCacheInuse)
	MetricStorage.gauge["MCacheSys"] = float64(m.MCacheSys)
	MetricStorage.gauge["MSpanSys"] = float64(m.MSpanSys)
	MetricStorage.gauge["Mallocs"] = float64(m.Mallocs)
	MetricStorage.gauge["NextGC"] = float64(m.NextGC)
	MetricStorage.gauge["NumForcedGC"] = float64(m.NumForcedGC)
	MetricStorage.gauge["NumGC"] = float64(m.NumGC)
	MetricStorage.gauge["OtherSys"] = float64(m.OtherSys)
	MetricStorage.gauge["PauseTotalNs"] = float64(m.PauseTotalNs)
	MetricStorage.gauge["StackInuse"] = float64(m.StackInuse)
	MetricStorage.gauge["StackSys"] = float64(m.StackSys)
	MetricStorage.gauge["Sys"] = float64(m.Sys)
	MetricStorage.gauge["TotalAlloc"] = float64(m.TotalAlloc)
	MetricStorage.gauge["RandomValue"] = float64(rand.Float64() * 100)

	MetricStorage.counter["PollCount"] += 1
}

func sendMetrics(config *ClientConfig) {
	client := &http.Client{}
	for k, v := range MetricStorage.gauge {
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
	for k, v := range MetricStorage.counter {
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

func onTimer(sendTimeoutSec *int, pollTimeoutSec *int, config *ClientConfig) {
	*sendTimeoutSec += 1
	*pollTimeoutSec += 1
	if *pollTimeoutSec == int(config.pollInterval) {
		getMetrics()
		*pollTimeoutSec = 0
	}
	if *sendTimeoutSec == int(config.reportInterval) {
		sendMetrics(config)
		*sendTimeoutSec = 0
	}
}

func main() {
	config := GetClientConfig()

	MetricStorage.gauge = make(map[string]float64)
	MetricStorage.counter = make(map[string]int64)
	sendMetricsTimeout := 0
	pollMetricsTimeout := 0
	for {
		onTimer(&sendMetricsTimeout, &pollMetricsTimeout, &config)
		time.Sleep(time.Second)
	}
}
