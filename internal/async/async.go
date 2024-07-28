// Package async предоставляет утилитные функции и структуры, используемые для многопоточности.
package async

// Semaphore структура семафора
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

// Acquire отправляет пустую структуру в канал
func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

// Release получает пустую структуру из канала
func (s *Semaphore) Release() {
	<-s.semaCh
}
