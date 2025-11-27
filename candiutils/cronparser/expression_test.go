package cronexpr

import (
	"testing"
	"time"
)

func TestHourlyIntervalCron(t *testing.T) {
	sched := MustParse("@hourly")
	start := time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC)
	next := sched.Next(start)
	if !next.After(start) {
		t.Errorf("Next did not return a future time for @hourly: got %v, want after %v", next, start)
	}
	// Should be exactly one hour later
	expected := time.Date(2025, 11, 27, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next for @hourly from %v: got %v, want %v", start, next, expected)
	}
}

func TestHourlyAtMinuteZero(t *testing.T) {
	sched := MustParse("0 * * * *")

	// Test from exact match time - should return next hour
	start := time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC)
	next := sched.Next(start)
	expected := time.Date(2025, 11, 27, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next for '0 * * * *' from %v: got %v, want %v", start, next, expected)
	}

	// Test from middle of hour - should return next hour
	start = time.Date(2025, 11, 27, 10, 30, 0, 0, time.UTC)
	next = sched.Next(start)
	expected = time.Date(2025, 11, 27, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next for '0 * * * *' from %v: got %v, want %v", start, next, expected)
	}

	// Test from one second before the hour
	// After adding 1 second, we're at 11:00:00 which matches the cron,
	// so Next should return the following occurrence at 12:00:00
	start = time.Date(2025, 11, 27, 10, 59, 59, 0, time.UTC)
	next = sched.Next(start)
	expected = time.Date(2025, 11, 27, 12, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next for '0 * * * *' from %v: got %v, want %v", start, next, expected)
	}

	// Test from just after the hour mark
	start = time.Date(2025, 11, 27, 10, 0, 1, 0, time.UTC)
	next = sched.Next(start)
	expected = time.Date(2025, 11, 27, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next for '0 * * * *' from %v: got %v, want %v", start, next, expected)
	}
}

func TestEveryTwoHours(t *testing.T) {
	sched := MustParse("0 */2 * * *")

	start := time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC)
	next := sched.Next(start)
	expected := time.Date(2025, 11, 27, 12, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next for '0 */2 * * *' from %v: got %v, want %v", start, next, expected)
	}
}

func TestEveryMinute(t *testing.T) {
	sched := MustParse("* * * * *")

	start := time.Date(2025, 11, 27, 10, 30, 0, 0, time.UTC)
	next := sched.Next(start)
	expected := time.Date(2025, 11, 27, 10, 31, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next for '* * * * *' from %v: got %v, want %v", start, next, expected)
	}
}
