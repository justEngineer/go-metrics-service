package storage

import (
	"context"
	"errors"
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

func (s *MemStorage) GetGaugeMetric(ctx context.Context, id string) (float64, error) {
	val, ok := s.Gauge[id]
	if !ok {
		return 0.0, errors.New("gauge metric is not found")
	} else {
		return val, nil
	}
}

func (s *MemStorage) GetCounterMetric(ctx context.Context, id string) (int64, error) {
	val, ok := s.Counter[id]
	if !ok {
		return 0.0, errors.New("counter metric is not found")
	} else {
		return val, nil
	}
}

func (s *MemStorage) SetGaugeMetric(ctx context.Context, id string, value float64) error {
	s.Gauge[id] = value
	return nil
}

func (s *MemStorage) SetCounterMetric(ctx context.Context, id string, value int64) error {
	s.Counter[id] += value
	return nil
}

func (s *MemStorage) SetMetricsBatch(ctx context.Context, gaugesBatch []GaugeMetric, countersBatch []CounterMetric) error {
	for _, gaugeMetric := range gaugesBatch {
		if err := s.SetGaugeMetric(ctx, gaugeMetric.Name, gaugeMetric.Value); err != nil {
			return err
		}
	}
	for _, counterMetric := range countersBatch {
		if err := s.SetCounterMetric(ctx, counterMetric.Name, counterMetric.Value); err != nil {
			return err
		}
	}
	return nil
}
