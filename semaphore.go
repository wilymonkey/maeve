package main

import "runtime"

type Semaphore struct {
	permits chan struct{}
}

// Creates a Semaphore with permits that scale with cpu number.
func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		permits: make(chan struct{}, runtime.NumCPU()*n),
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
