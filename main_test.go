package main_test

import (
	"testing"
	"time"
)

func preciseSleep(duration time.Duration) {
	start := time.Now()
	for time.Since(start) < duration {
		// Busy-wait
	}
}

func BenchmarkPreciseSleep(b *testing.B) {
	for b.Loop() {
		for range 1000 {
			preciseSleep(1000 * time.Nanosecond)
		}
	}

}
func BenchmarkTimeSleep(b *testing.B) {
	for b.Loop() {
		for range 1000 {
			time.Sleep(1000 * time.Nanosecond)
		}
	}

}
