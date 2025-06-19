package main

import "runtime"

type Semaphore struct {
	permits chan struct{}
}

// Creates a Semaphore with n permits that scale with cpu number.
func NewScalingSemaphore(n int) *Semaphore {
	return NewSemaphore(runtime.NumCPU() * n)
}

// Creates a Semaphore with n permits.
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
