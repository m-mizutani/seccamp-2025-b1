package model

import (
	"time"
)

// RawLog represents the schema of raw log data from JSONL files
type RawLog struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Success   bool      `json:"success"`
	Remote    string    `json:"remote"`
}

// ParquetLog represents the log data for Parquet output (Security Lake compatible)
type ParquetLog struct {
	EventTime    int64       `parquet:"name=event_time, type=INT64"`
	ActivityID   int32       `parquet:"name=activity_id, type=INT32"`
	ActivityName string      `parquet:"name=activity_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	Actor        Actor       `parquet:"name=actor"`
	API          API         `parquet:"name=api"`
	SrcEndpoint  SrcEndpoint `parquet:"name=src_endpoint"`
	StatusID     int32       `parquet:"name=status_id, type=INT32"`
	StatusDetail string      `parquet:"name=status_detail, type=BYTE_ARRAY, convertedtype=UTF8"`
	Time         int64       `parquet:"name=time, type=INT64"`
	TypeUID      int64       `parquet:"name=type_uid, type=INT64"`
	CategoryUID  int32       `parquet:"name=category_uid, type=INT32"`
	ClassUID     int32       `parquet:"name=class_uid, type=INT32"`
	Severity     string      `parquet:"name=severity, type=BYTE_ARRAY, convertedtype=UTF8"`
	Metadata     Metadata    `parquet:"name=metadata"`
}

type Actor struct {
	User User `parquet:"name=user"`
}

type User struct {
	Name string `parquet:"name=name, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type API struct {
	Operation string  `parquet:"name=operation, type=BYTE_ARRAY, convertedtype=UTF8"`
	Request   Request `parquet:"name=request"`
}

type Request struct {
	UID string `parquet:"name=uid, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type SrcEndpoint struct {
	IP string `parquet:"name=ip, type=BYTE_ARRAY, convertedtype=UTF8"`
}

type Metadata struct {
	ProductName    string `parquet:"name=product_name, type=BYTE_ARRAY, convertedtype=UTF8"`
	ProductVersion string `parquet:"name=product_version, type=BYTE_ARRAY, convertedtype=UTF8"`
	Version        string `parquet:"name=version, type=BYTE_ARRAY, convertedtype=UTF8"`
}

// ConvertToParquet converts RawLog to ParquetLog
func (r *RawLog) ConvertToParquet() *ParquetLog {
	var activityID int32
	var statusID int32

	// Map action to activity_id
	switch r.Action {
	case "read":
		activityID = 2 // Read
	case "write":
		activityID = 3 // Write
	case "login":
		activityID = 1 // Logon
	default:
		activityID = 99 // Other
	}

	// Map success to status_id
	if r.Success {
		statusID = 1 // Success
	} else {
		statusID = 2 // Failure
	}

	return &ParquetLog{
		EventTime:    r.Timestamp.UnixMilli(),
		ActivityID:   activityID,
		ActivityName: r.Action,
		Actor: Actor{
			User: User{
				Name: r.User,
			},
		},
		API: API{
			Operation: r.Action,
			Request: Request{
				UID: r.Target,
			},
		},
		SrcEndpoint: SrcEndpoint{
			IP: r.Remote,
		},
		StatusID:     statusID,
		StatusDetail: map[bool]string{true: "Success", false: "Failure"}[r.Success],
		Time:         r.Timestamp.UnixMilli(),
		TypeUID:      300200, // User Activity
		CategoryUID:  3,      // Identity & Access Management
		ClassUID:     3002,   // User Activity
		Severity:     "Informational",
		Metadata: Metadata{
			ProductName:    "SecCamp2025-Service",
			ProductVersion: "1.0",
			Version:        "1.1.0",
		},
	}
}
