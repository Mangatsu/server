package utils

import (
	"errors"
	"os"
	"time"
)

// Clamp clamps the given value to the given range.
func Clamp(value, min, max int64) int64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// PeriodicTask loops the given function in separate thread between the given interval.
func PeriodicTask(d time.Duration, f func()) {
	go func() {
		for {
			f()
			time.Sleep(d)
		}
	}()
}

// PathExists checks if the given path exists.
func PathExists(pathTo string) bool {
	_, err := os.OpenFile(pathTo, os.O_RDONLY, 0444)
	return !errors.Is(err, os.ErrNotExist)
}
