package main

import (
	"time"
)

// RawLog represents the input log format from JSONL files
type RawLog struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"` // read, write, login
	Target    string    `json:"target"` // empty for login
	Success   bool      `json:"success"`
	Remote    string    `json:"remote"` // IPv4 address
}

// ParquetLog represents the output format for Security Lake
type ParquetLog struct {
	ID        string `parquet:"id,optional"`
	Timestamp int64  `parquet:"timestamp,optional"` // Unix timestamp in milliseconds
	User      string `parquet:"user,optional"`
	Action    string `parquet:"action,optional"`
	Target    string `parquet:"target,optional"`
	Success   bool   `parquet:"success,optional"`
	Remote    string `parquet:"remote,optional"`
	Year      int    `parquet:"year,optional"`  // For partitioning
	Month     int    `parquet:"month,optional"` // For partitioning
	Day       int    `parquet:"day,optional"`   // For partitioning
	Hour      int    `parquet:"hour,optional"`  // For partitioning
}

// ToParquetLog converts RawLog to ParquetLog
func (r *RawLog) ToParquetLog() ParquetLog {
	return ParquetLog{
		ID:        r.ID,
		Timestamp: r.Timestamp.UnixMilli(),
		User:      r.User,
		Action:    r.Action,
		Target:    r.Target,
		Success:   r.Success,
		Remote:    r.Remote,
		Year:      r.Timestamp.Year(),
		Month:     int(r.Timestamp.Month()),
		Day:       r.Timestamp.Day(),
		Hour:      r.Timestamp.Hour(),
	}
}
