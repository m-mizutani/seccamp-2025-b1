package main

import (
	"time"
)

type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

// CalculateTimeRange calculates the time range for log fetching
// Returns past 7 minutes (5 minutes interval + 2 minutes buffer)
func CalculateTimeRange(now time.Time, bufferMinutes int) TimeRange {
	endTime := now
	startTime := now.Add(-time.Duration(5+bufferMinutes) * time.Minute)

	return TimeRange{
		StartTime: startTime,
		EndTime:   endTime,
	}
}

// Duration returns the duration of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.EndTime.Sub(tr.StartTime)
}

// String returns a human-readable representation of the time range
func (tr TimeRange) String() string {
	return tr.StartTime.Format(time.RFC3339) + " to " + tr.EndTime.Format(time.RFC3339)
}

// IsValid checks if the time range is valid
func (tr TimeRange) IsValid() bool {
	return tr.StartTime.Before(tr.EndTime)
}