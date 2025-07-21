package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/parquet-go/parquet-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleWorkspaceLogToOCSF_Conversion(t *testing.T) {
	gwLog := &GoogleWorkspaceLog{
		Kind: "audit#activity",
		ID: struct {
			Time             string `json:"time"`
			UniqueQualifier  string `json:"uniqueQualifier"`
			ApplicationName  string `json:"applicationName"`
			CustomerID       string `json:"customerId"`
		}{
			Time:            "2024-08-12T10:15:30.123456Z",
			UniqueQualifier: "358068855354",
			ApplicationName: "drive",
			CustomerID:      "C03az79cb",
		},
		Actor: struct {
			CallerType string `json:"callerType"`
			Email      string `json:"email"`
			ProfileID  string `json:"profileId"`
		}{
			CallerType: "USER",
			Email:      "user@muhai-academy.com",
			ProfileID:  "114511147312345678901",
		},
		OwnerDomain: "muhai-academy.com",
		IPAddress:   "203.0.113.255",
		Events: []struct {
			Type       string `json:"type"`
			Name       string `json:"name"`
			Parameters []struct {
				Name           string   `json:"name"`
				Value          string   `json:"value,omitempty"`
				BoolValue      bool     `json:"boolValue,omitempty"`
				IntValue       int64    `json:"intValue,omitempty"`
				MultiStrValue  []string `json:"multiStrValue,omitempty"`
				MultiIntValue  []int64  `json:"multiIntValue,omitempty"`
			} `json:"parameters,omitempty"`
		}{
			{
				Type: "access",
				Name: "view",
				Parameters: []struct {
					Name           string   `json:"name"`
					Value          string   `json:"value,omitempty"`
					BoolValue      bool     `json:"boolValue,omitempty"`
					IntValue       int64    `json:"intValue,omitempty"`
					MultiStrValue  []string `json:"multiStrValue,omitempty"`
					MultiIntValue  []int64  `json:"multiIntValue,omitempty"`
				}{
					{Name: "doc_id", Value: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"},
					{Name: "doc_title", Value: "学習進捗データ"},
					{Name: "doc_type", Value: "spreadsheet"},
				},
			},
		},
	}

	ocsfLog, err := ConvertToOCSF(gwLog, "ap-northeast-1", "123456789012")
	require.NoError(t, err)

	assert.Equal(t, 6, ocsfLog.CategoryUID)
	assert.Equal(t, 6001, ocsfLog.ClassUID)
	assert.Equal(t, 2, ocsfLog.ActivityID) // Read for "view"
	assert.Equal(t, 1, ocsfLog.SeverityID) // Informational
	assert.Equal(t, 1, ocsfLog.StatusID)   // Success
	assert.Equal(t, "user@muhai-academy.com", ocsfLog.Actor.User.EmailAddr)
	assert.Equal(t, "203.0.113.255", ocsfLog.SrcEndpoint.IP)
	assert.Equal(t, "Google Drive API", ocsfLog.API.Service.Name)
	assert.Equal(t, "view", ocsfLog.API.Operation)
}

func TestGenerateOCSFParquetFile_ValidData(t *testing.T) {
	handler := &Handler{}

	timestamp1, _ := time.Parse(time.RFC3339, "2024-08-12T10:00:00Z")
	timestamp2, _ := time.Parse(time.RFC3339, "2024-08-12T10:05:00Z")

	logs := []OCSFWebResourceActivity{
		{
			CategoryUID: 6,
			ClassUID:    6001,
			TypeUID:     600102,
			ActivityID:  2,
			SeverityID:  1,
			StatusID:    1,
			Time:        timestamp1.UnixMilli(),
			Actor: struct {
				User struct {
					Name       string   `parquet:"name"`
					UID        string   `parquet:"uid"`
					EmailAddr  string   `parquet:"email_addr"`
					Domain     string   `parquet:"domain,optional"`
					TypeID     int      `parquet:"type_id"`
					Groups     []string `parquet:"groups,optional"`
				} `parquet:"user"`
				Session struct {
					UID         string `parquet:"uid"`
					CreatedTime int64  `parquet:"created_time,optional"`
					ExpTime     int64  `parquet:"exp_time,optional"`
				} `parquet:"session,optional"`
				AppName string `parquet:"app_name,optional"`
				AppUID  string `parquet:"app_uid,optional"`
			}{
				User: struct {
					Name       string   `parquet:"name"`
					UID        string   `parquet:"uid"`
					EmailAddr  string   `parquet:"email_addr"`
					Domain     string   `parquet:"domain,optional"`
					TypeID     int      `parquet:"type_id"`
					Groups     []string `parquet:"groups,optional"`
				}{
					EmailAddr: "alice@example.com",
					UID:       "user001",
					TypeID:    1,
				},
			},
			SrcEndpoint: struct {
				IP       string `parquet:"ip"`
				Hostname string `parquet:"hostname,optional"`
				Location struct {
					Country string `parquet:"country,optional"`
					Region  string `parquet:"src_region,optional"`
					City    string `parquet:"city,optional"`
				} `parquet:"location,optional"`
			}{
				IP: "192.168.1.100",
			},
			Region:    "ap-northeast-1",
			AccountID: "123456789012",
			EventHour: "2024-08-12-10",
		},
		{
			CategoryUID: 6,
			ClassUID:    6001,
			TypeUID:     600102,
			ActivityID:  2,
			SeverityID:  1,
			StatusID:    1,
			Time:        timestamp2.UnixMilli(),
			Actor: struct {
				User struct {
					Name       string   `parquet:"name"`
					UID        string   `parquet:"uid"`
					EmailAddr  string   `parquet:"email_addr"`
					Domain     string   `parquet:"domain,optional"`
					TypeID     int      `parquet:"type_id"`
					Groups     []string `parquet:"groups,optional"`
				} `parquet:"user"`
				Session struct {
					UID         string `parquet:"uid"`
					CreatedTime int64  `parquet:"created_time,optional"`
					ExpTime     int64  `parquet:"exp_time,optional"`
				} `parquet:"session,optional"`
				AppName string `parquet:"app_name,optional"`
				AppUID  string `parquet:"app_uid,optional"`
			}{
				User: struct {
					Name       string   `parquet:"name"`
					UID        string   `parquet:"uid"`
					EmailAddr  string   `parquet:"email_addr"`
					Domain     string   `parquet:"domain,optional"`
					TypeID     int      `parquet:"type_id"`
					Groups     []string `parquet:"groups,optional"`
				}{
					EmailAddr: "bob@example.com",
					UID:       "user002",
					TypeID:    1,
				},
			},
			SrcEndpoint: struct {
				IP       string `parquet:"ip"`
				Hostname string `parquet:"hostname,optional"`
				Location struct {
					Country string `parquet:"country,optional"`
					Region  string `parquet:"src_region,optional"`
					City    string `parquet:"city,optional"`
				} `parquet:"location,optional"`
			}{
				IP: "192.168.1.101",
			},
			Region:    "ap-northeast-1",
			AccountID: "123456789012",
			EventHour: "2024-08-12-10",
		},
	}

	data, err := handler.generateOCSFParquetFile(logs)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Verify we can read back the data
	reader := parquet.NewGenericReader[OCSFWebResourceActivity](bytes.NewReader(data))
	defer reader.Close()

	readLogs := make([]OCSFWebResourceActivity, len(logs))
	n, err := reader.Read(readLogs)
	require.NoError(t, err)
	assert.Equal(t, len(logs), n)

	for i, expected := range logs {
		assert.Equal(t, expected.CategoryUID, readLogs[i].CategoryUID)
		assert.Equal(t, expected.ClassUID, readLogs[i].ClassUID)
		assert.Equal(t, expected.Time, readLogs[i].Time)
		assert.Equal(t, expected.Actor.User.EmailAddr, readLogs[i].Actor.User.EmailAddr)
		assert.Equal(t, expected.SrcEndpoint.IP, readLogs[i].SrcEndpoint.IP)
	}
}

func TestGenerateOCSFParquetFile_EmptyData(t *testing.T) {
	handler := &Handler{}

	logs := []OCSFWebResourceActivity{}

	data, err := handler.generateOCSFParquetFile(logs)
	require.NoError(t, err)
	require.NotEmpty(t, data) // Even empty parquet files have metadata

	// Verify we can read back the empty data
	reader := parquet.NewGenericReader[OCSFWebResourceActivity](bytes.NewReader(data))
	defer reader.Close()

	readLogs := make([]OCSFWebResourceActivity, 10) // Allocate buffer
	n, err := reader.Read(readLogs)
	if err != nil && err.Error() != "EOF" {
		require.NoError(t, err) // Only fail if it's not EOF
	}
	assert.Equal(t, 0, n)
}

func TestGenerateOCSFParquetFile_SchemaValidation(t *testing.T) {
	handler := &Handler{}

	// Test with maximum field lengths and edge cases
	timestamp := time.Now()
	logs := []OCSFWebResourceActivity{
		{
			CategoryUID: 6,
			ClassUID:    6001,
			TypeUID:     600103,
			ActivityID:  3,
			SeverityID:  3,
			StatusID:    2,
			Time:        timestamp.UnixMilli(),
			Actor: struct {
				User struct {
					Name       string   `parquet:"name"`
					UID        string   `parquet:"uid"`
					EmailAddr  string   `parquet:"email_addr"`
					Domain     string   `parquet:"domain,optional"`
					TypeID     int      `parquet:"type_id"`
					Groups     []string `parquet:"groups,optional"`
				} `parquet:"user"`
				Session struct {
					UID         string `parquet:"uid"`
					CreatedTime int64  `parquet:"created_time,optional"`
					ExpTime     int64  `parquet:"exp_time,optional"`
				} `parquet:"session,optional"`
				AppName string `parquet:"app_name,optional"`
				AppUID  string `parquet:"app_uid,optional"`
			}{
				User: struct {
					Name       string   `parquet:"name"`
					UID        string   `parquet:"uid"`
					EmailAddr  string   `parquet:"email_addr"`
					Domain     string   `parquet:"domain,optional"`
					TypeID     int      `parquet:"type_id"`
					Groups     []string `parquet:"groups,optional"`
				}{
					EmailAddr: "user_with_special_chars@domain.com",
					UID:       "very-long-id-that-tests-string-handling-in-parquet-format",
					TypeID:    2,
					Domain:    "domain.com",
				},
			},
			SrcEndpoint: struct {
				IP       string `parquet:"ip"`
				Hostname string `parquet:"hostname,optional"`
				Location struct {
					Country string `parquet:"country,optional"`
					Region  string `parquet:"src_region,optional"`
					City    string `parquet:"city,optional"`
				} `parquet:"location,optional"`
			}{
				IP: "255.255.255.255",
			},
			WebResources: []struct {
				Name      string `parquet:"name,optional"`
				UID       string `parquet:"uid,optional"`
				Type      string `parquet:"type,optional"`
				URLString string `parquet:"url_string,optional"`
				Data      struct {
					Classification string `parquet:"classification,optional"`
				} `parquet:"data,optional"`
			}{
				{
					Name: "documents/subfolder/very-long-filename-with-spaces and special chars.txt",
					Type: "document",
				},
			},
			Region:    "ap-northeast-1",
			AccountID: "123456789012",
			EventHour: timestamp.Format("2006-01-02-15"),
		},
	}

	data, err := handler.generateOCSFParquetFile(logs)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Verify schema integrity by reading back
	reader := parquet.NewGenericReader[OCSFWebResourceActivity](bytes.NewReader(data))
	defer reader.Close()

	readLogs := make([]OCSFWebResourceActivity, 1)
	n, err := reader.Read(readLogs)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, logs[0].CategoryUID, readLogs[0].CategoryUID)
	assert.Equal(t, logs[0].ClassUID, readLogs[0].ClassUID)
	assert.Equal(t, logs[0].Actor.User.EmailAddr, readLogs[0].Actor.User.EmailAddr)
}
