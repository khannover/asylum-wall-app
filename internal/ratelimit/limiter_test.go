package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllow(t *testing.T) {
	l := New()
	key := "test"

	for i := 0; i < 3; i++ {
		if !l.Allow(key, 3, time.Hour) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if l.Allow(key, 3, time.Hour) {
		t.Fatal("fourth request should be blocked")
	}
}