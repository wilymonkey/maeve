package main

type Semaphore struct {
	permits chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		permits: make(chan struct{}, n),
	}
}

// Acquire takes a permit (blocks if none available)
func (s *Semaphore) Acquire() {
	s.permits <- struct{}{}
}

// Release returns a permit
func (s *Semaphore) Release() {
	<-s.permits
}
