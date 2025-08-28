package utils

import "sync"

func GoWait(f func()) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
	return &wg
}

func CollectChan[T any](c chan T) func() []T {
	var wg sync.WaitGroup
	var collector []T

	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range c {
			collector = append(collector, item)
		}
	}()

	return func() []T {
		wg.Wait()
		return collector
	}
}

type ThreadSafe[T any] struct {
	mu  sync.RWMutex
	val T
}

func (s *ThreadSafe[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.val
}

func (s *ThreadSafe[T]) Set(v T) {
	s.mu.Lock()
	s.val = v
	s.mu.Unlock()
}
