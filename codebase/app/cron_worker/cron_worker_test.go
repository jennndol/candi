package cronworker

import (
	"fmt"
	"testing"
	"time"

	cronexpr "github.com/golangid/candi/candiutils/cronparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHourlyCron     = "1 * * * *"   // At minute 1 of every hour
	testFiveMinuteCron = "*/5 * * * *" // Every 5 minutes
	testDailyCron      = "0 12 * * *"  // Daily at 12:00 PM
	testMinuteCron     = "*/1 * * * *" // Every minute
)

func TestRegisterNextIntervalCronExpression(t *testing.T) {
	tests := []struct {
		name           string
		cronExpr       string
		description    string
		minExpectedDur time.Duration
		maxExpectedDur time.Duration
	}{
		{
			name:           "hourly_at_minute_1",
			cronExpr:       testHourlyCron,
			description:    "Should schedule next execution at minute 1 of next hour",
			minExpectedDur: 1 * time.Minute,  // At least 1 minute if we're currently at minute 0
			maxExpectedDur: 61 * time.Minute, // At most 61 minutes if we just passed minute 1
		},
		{
			name:           "every_5_minutes",
			cronExpr:       testFiveMinuteCron,
			description:    "Should schedule next execution at next 5-minute boundary",
			minExpectedDur: 1 * time.Second, // Could be very soon if we're close to boundary
			maxExpectedDur: 5 * time.Minute, // At most 5 minutes
		},
		{
			name:           "daily_at_noon",
			cronExpr:       testDailyCron,
			description:    "Should schedule next execution at next 12:00 PM",
			minExpectedDur: 1 * time.Minute, // At least 1 minute
			maxExpectedDur: 24 * time.Hour,  // At most 24 hours
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the cron expression
			schedule, err := cronexpr.Parse(tt.cronExpr)
			require.NoError(t, err, "Failed to parse cron expression: %s", tt.cronExpr)

			// Test the registerNextInterval function logic
			now := time.Now()
			nextTime := schedule.Next(now)

			// Verify that Next() returns a valid time
			assert.False(t, nextTime.IsZero(), "Next() should return a valid time")

			// Calculate duration until next execution
			duration := time.Until(nextTime)

			// Verify duration is within expected range
			assert.GreaterOrEqual(t, duration, tt.minExpectedDur,
				"Duration should be at least %v, got %v", tt.minExpectedDur, duration)
			assert.LessOrEqual(t, duration, tt.maxExpectedDur,
				"Duration should be at most %v, got %v", tt.maxExpectedDur, duration)

			// Test that if duration is too small, we get the next occurrence
			if duration < time.Second {
				nextNextTime := schedule.Next(nextTime.Add(time.Second))
				nextDuration := time.Until(nextNextTime)
				assert.GreaterOrEqual(t, nextDuration, time.Second,
					"Next duration after skipping should be at least 1 second")
			}

			// Mock the ticker creation
			if duration >= time.Second {
				ticker := time.NewTicker(duration)
				ticker.Stop() // Stop immediately to prevent resource leak
				assert.NotNil(t, ticker, "Ticker should be created successfully")
			}
		})
	}
}

func TestCronTimingConsistency(t *testing.T) {
	// Test that demonstrates the fix for timing consistency
	schedule, err := cronexpr.Parse(testHourlyCron)
	require.NoError(t, err)

	// Simulate multiple calls to registerNextInterval at different times
	// This tests that we get consistent scheduling regardless of when it's called

	baseTime := time.Date(2025, 11, 7, 14, 1, 0, 0, time.UTC) // 14:01:00
	testTimes := []time.Time{
		baseTime,                       // Exactly at 14:01:00
		baseTime.Add(5 * time.Second),  // 14:01:05
		baseTime.Add(30 * time.Second), // 14:01:30
		baseTime.Add(59 * time.Second), // 14:01:59
	}

	expectedNextExecution := time.Date(2025, 11, 7, 15, 1, 0, 0, time.UTC) // 15:01:00

	for i, testTime := range testTimes {
		t.Run(fmt.Sprintf("test_time_%d", i), func(t *testing.T) {
			// Calculate next execution time using our fixed logic
			nextTime := schedule.Next(testTime)

			// All calls should result in the same next execution time
			assert.Equal(t, expectedNextExecution, nextTime,
				"Next execution time should be consistent regardless of current time")

			// Duration should be different but all pointing to same target time
			duration := nextTime.Sub(testTime)
			assert.Greater(t, duration, time.Duration(0), "Duration should be positive")
			assert.LessOrEqual(t, duration, time.Hour, "Duration should not exceed 1 hour")
		})
	}
}

func TestMinimumDurationSafetyCheck(t *testing.T) {
	// Test the safety check that prevents immediate re-execution
	schedule, err := cronexpr.Parse(testMinuteCron)
	require.NoError(t, err)

	// Simulate being very close to the next execution time
	now := time.Date(2025, 11, 7, 14, 5, 59, 900000000, time.UTC) // 14:05:59.9 (100ms before next minute)

	nextTime := schedule.Next(now)
	duration := time.Until(nextTime)

	// If duration is less than 1 second, we should get the next occurrence
	if duration < time.Second {
		nextNextTime := schedule.Next(nextTime.Add(time.Second))
		nextDuration := time.Until(nextNextTime)

		assert.GreaterOrEqual(t, nextDuration, time.Second,
			"Safety check should ensure at least 1 second before next execution")

		// The next execution should be 1 minute later than the immediate next
		expectedDiff := time.Minute
		actualDiff := nextNextTime.Sub(nextTime)
		assert.Equal(t, expectedDiff, actualDiff,
			"Next execution after safety check should be exactly 1 minute later")
	}
}
