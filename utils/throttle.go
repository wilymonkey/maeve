package utils

import (
	"time"
)

func Throttle[T any](in <-chan T, onEach func(T), emit func()) {
	go func() {
		tick := time.NewTicker(200 * time.Millisecond)
		defer tick.Stop()

		dirty := false
		for {
			select {
			case v, ok := <-in:
				if !ok {
					if dirty {
						emit()
					} // final flush
					return
				}
				onEach(v)
				dirty = true

			case <-tick.C:
				if dirty {
					emit()
					dirty = false
				}
			}
		}
	}()
}
