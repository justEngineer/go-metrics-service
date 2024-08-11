// Package async предоставляет утилитные функции и структуры, используемые для многопоточности.
package async

// Semaphore реализует простой семафор для ограничения количества одновременных запросов.
//
// Пример использования:
//
//	sem := NewSemaphore(10)
//	sem.Acquire()
//	defer sem.Release()
//	// Выполнение работы
type Semaphore struct {
	semaCh chan struct{}
}

// NewSemaphore создает семафор с буферизованным каналом емкостью maxReq
func NewSemaphore(maxReq int) *Semaphore {
	if maxReq > 0 {
		return &Semaphore{
			semaCh: make(chan struct{}, maxReq),
		}
	} else {
		return nil
	}
}

// Wait отправляет пустую структуру в канал
func (s *Semaphore) Wait() {
	s.semaCh <- struct{}{}
}

// Signal получает пустую структуру из канала
func (s *Semaphore) Signal() {
	<-s.semaCh
}
