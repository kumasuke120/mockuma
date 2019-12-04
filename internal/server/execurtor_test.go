package server

import (
	"testing"
	"time"

	"github.com/kumasuke120/mockuma/internal/mckmaps"
)

func TestWaitBeforeReturns(t *testing.T) {
	i1 := &mckmaps.Interval{Min: 100, Max: 100}
	i1Start := time.Now()
	waitBeforeReturns(i1)
	i1Elapsed := time.Since(i1Start)
	if i1Elapsed.Milliseconds() < 100 {
		t.Error("i1: error")
	}

	i2 := &mckmaps.Interval{Min: 10, Max: 30}
	i2Start := time.Now()
	for i := 0; i < 10; i++ {
		waitBeforeReturns(i2)
	}
	i2Elapsed := time.Since(i2Start)
	if i2Elapsed.Milliseconds() < 100 || i2Elapsed.Milliseconds() > 350 {
		t.Error("i2: error")
	}
}
