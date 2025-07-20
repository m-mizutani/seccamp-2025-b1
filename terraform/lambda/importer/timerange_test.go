package main

import (
	"testing"
	"time"
)

func TestCalculateTimeRange(t *testing.T) {
	now := time.Date(2024, 8, 12, 10, 5, 0, 0, time.UTC)
	bufferMinutes := 2

	timeRange := CalculateTimeRange(now, bufferMinutes)

	expectedStart := time.Date(2024, 8, 12, 9, 58, 0, 0, time.UTC) // 7 minutes ago
	expectedEnd := time.Date(2024, 8, 12, 10, 5, 0, 0, time.UTC)  // now

	if !timeRange.StartTime.Equal(expectedStart) {
		t.Errorf("Expected start time %v, got %v", expectedStart, timeRange.StartTime)
	}

	if !timeRange.EndTime.Equal(expectedEnd) {
		t.Errorf("Expected end time %v, got %v", expectedEnd, timeRange.EndTime)
	}

	if timeRange.Duration() != 7*time.Minute {
		t.Errorf("Expected duration 7 minutes, got %v", timeRange.Duration())
	}
}

func TestTimeRangeIsValid(t *testing.T) {
	now := time.Date(2024, 8, 12, 10, 5, 0, 0, time.UTC)
	
	// Valid range
	validRange := TimeRange{
		StartTime: now.Add(-5 * time.Minute),
		EndTime:   now,
	}
	if !validRange.IsValid() {
		t.Error("Expected valid range to be valid")
	}

	// Invalid range (start after end)
	invalidRange := TimeRange{
		StartTime: now,
		EndTime:   now.Add(-5 * time.Minute),
	}
	if invalidRange.IsValid() {
		t.Error("Expected invalid range to be invalid")
	}

	// Equal times (invalid)
	equalRange := TimeRange{
		StartTime: now,
		EndTime:   now,
	}
	if equalRange.IsValid() {
		t.Error("Expected equal times to be invalid")
	}
}

func TestTimeRangeString(t *testing.T) {
	start := time.Date(2024, 8, 12, 9, 58, 0, 0, time.UTC)
	end := time.Date(2024, 8, 12, 10, 5, 0, 0, time.UTC)
	
	timeRange := TimeRange{
		StartTime: start,
		EndTime:   end,
	}

	expected := "2024-08-12T09:58:00Z to 2024-08-12T10:05:00Z"
	if timeRange.String() != expected {
		t.Errorf("Expected string %s, got %s", expected, timeRange.String())
	}
}