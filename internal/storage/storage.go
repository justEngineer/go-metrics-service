package storage

import (
	"sync"
)

type GaugeMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type CounterMetric struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type MetricsDump struct {
	Counters []CounterMetric `json:"counters"`
	Gauges   []GaugeMetric   `json:"gauges"`
}

type MemStorage struct {
	// указаны некоторые поля структуры
	Gauge   map[string]float64
	Counter map[string]int64
	Mutex   sync.Mutex
}

func New() *MemStorage {
	MetricStorage := &MemStorage{}
	MetricStorage.Gauge = make(map[string]float64)
	MetricStorage.Counter = make(map[string]int64)
	return MetricStorage
}

func (s *MemStorage) GetAllMetrics() MetricsDump {
	var gauges []GaugeMetric
	var counters []CounterMetric
	s.Mutex.Lock()
	for key, value := range s.Gauge {
		gauges = append(gauges, GaugeMetric{Name: key, Value: value})
	}
	for key, value := range s.Counter {
		counters = append(counters, CounterMetric{Name: key, Value: value})
	}
	s.Mutex.Unlock()
	return MetricsDump{
		Counters: counters,
		Gauges:   gauges,
	}
}
