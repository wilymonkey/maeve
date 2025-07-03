package utils

import (
	"time"
)

func Throttle[T any](tChan chan T, every func(T), emit func()) {
	go func() {
		throttleDuration := 200 * time.Millisecond

		var (
			timer     *time.Timer
			timerChan <-chan time.Time
			canEmit   = true
		)

		for {
			select {
			case value, ok := <-tChan:
				if ok {
					every(value)
					if canEmit {
						emit()
						canEmit = false
						timer = time.NewTimer(throttleDuration)
						timerChan = timer.C
					}
				} else {
					emit()
					return
				}

			case <-timerChan:
				canEmit = true
			}
		}
	}()
}
