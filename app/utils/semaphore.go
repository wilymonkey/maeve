package utils

import "runtime"

type Semaphore struct {
	permits chan struct{}
}

// Creates a Semaphore with n permits that scale with cpu number.
func NewScalingSema(n int) *Semaphore {
	return New(runtime.NumCPU() * n)
}

// Creates a Semaphore with n permits.
func New(n int) *Semaphore {
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

func (s *Semaphore) Close() {
	close(s.permits)
}
