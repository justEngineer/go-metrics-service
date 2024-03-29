package storage

type MemStorage struct {
	// указаны некоторые поля структуры
	Gauge   map[string]float64
	Counter map[string]int64
}

func New() *MemStorage {
	MetricStorage := &MemStorage{}
	MetricStorage.Gauge = make(map[string]float64)
	MetricStorage.Counter = make(map[string]int64)
	return MetricStorage
}
