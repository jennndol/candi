package cronworker

import (
	"time"

	cronexpr "github.com/golangid/candi/candiutils/cronparser"
	"github.com/golangid/candi/codebase/factory/types"
)

// Job model
type Job struct {
	HandlerName  string              `json:"handler_name"`
	Interval     string              `json:"interval"`
	Handler      types.WorkerHandler `json:"-"`
	Params       string              `json:"params"`
	WorkerIndex  int                 `json:"worker_index"`
	ticker       *time.Ticker        `json:"-"`
	schedule     cronexpr.Schedule   `json:"-"`
	nextDuration *time.Duration      `json:"-"`
	nextTime     *time.Time          `json:"-"` // For cron expressions, track exact next run time
}

// calculateNextTime calculates the duration until the next job execution
// For cron expressions, it dynamically calculates based on the cron schedule
// For regular intervals, it uses the fixed duration
func (j *Job) calculateNextTime() time.Duration {
	if j.schedule != nil {
		// For cron expressions, calculate the exact next execution time
		now := time.Now()
		nextTime := j.schedule.Next(now)
		j.nextTime = &nextTime

		duration := time.Until(nextTime)
		// Ensure minimum duration to prevent immediate re-execution
		if duration < time.Second {
			nextTime = j.schedule.Next(nextTime.Add(time.Second))
			j.nextTime = &nextTime
			duration = time.Until(nextTime)
		}
		return duration
	}

	// For regular intervals, use nextDuration if available
	if j.nextDuration != nil {
		return *j.nextDuration
	}

	// This shouldn't happen for properly initialized jobs
	return time.Minute
}
