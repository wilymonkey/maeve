package utils

import (
	"time"
)

func Throttle[T any](tChan chan T, every func(T), emit func()) {
	go func() {
		throttleDuration := 200 * time.Millisecond

		var (
			timer   *time.Timer
			timerCh <-chan time.Time
			canEmit bool = true // Flag to control emission
		)

		for {
			select {
			case value := <-tChan:
				every(value)
				if canEmit {
					emit()
					canEmit = false
					timer = time.NewTimer(throttleDuration)
					timerCh = timer.C
				}

			case <-timerCh:
				canEmit = true
			}
		}
	}()
}
