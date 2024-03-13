package storage

type MemStorage struct {
	// указаны некоторые поля структуры
	Gauge   map[string]float64
	Counter map[string]int64
}
