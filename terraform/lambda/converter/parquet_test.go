package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/parquet-go/parquet-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRawLogToParquetLog_Conversion(t *testing.T) {
	timestamp, _ := time.Parse(time.RFC3339, "2023-12-01T10:00:00Z")

	rawLog := RawLog{
		ID:        "log001",
		Timestamp: timestamp,
		User:      "alice",
		Action:    "login",
		Target:    "",
		Success:   true,
		Remote:    "192.168.1.100",
	}

	parquetLog := rawLog.ToParquetLog()

	assert.Equal(t, rawLog.ID, parquetLog.ID)
	assert.Equal(t, rawLog.Timestamp.UnixMilli(), parquetLog.Timestamp)
	assert.Equal(t, rawLog.User, parquetLog.User)
	assert.Equal(t, rawLog.Action, parquetLog.Action)
	assert.Equal(t, rawLog.Target, parquetLog.Target)
	assert.Equal(t, rawLog.Success, parquetLog.Success)
	assert.Equal(t, rawLog.Remote, parquetLog.Remote)
}

func TestGenerateParquetFile_ValidData(t *testing.T) {
	handler := &Handler{}

	timestamp1, _ := time.Parse(time.RFC3339, "2023-12-01T10:00:00Z")
	timestamp2, _ := time.Parse(time.RFC3339, "2023-12-01T10:05:00Z")

	logs := []ParquetLog{
		{
			ID:        "log001",
			Timestamp: timestamp1.UnixMilli(),
			User:      "alice",
			Action:    "login",
			Target:    "",
			Success:   true,
			Remote:    "192.168.1.100",
		},
		{
			ID:        "log002",
			Timestamp: timestamp2.UnixMilli(),
			User:      "bob",
			Action:    "read",
			Target:    "document1.txt",
			Success:   true,
			Remote:    "192.168.1.101",
		},
	}

	data, err := handler.generateParquetFile(logs)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Verify we can read back the data
	reader := parquet.NewGenericReader[ParquetLog](bytes.NewReader(data))
	defer reader.Close()

	readLogs := make([]ParquetLog, len(logs))
	n, err := reader.Read(readLogs)
	require.NoError(t, err)
	assert.Equal(t, len(logs), n)

	for i, expected := range logs {
		assert.Equal(t, expected.ID, readLogs[i].ID)
		assert.Equal(t, expected.Timestamp, readLogs[i].Timestamp)
		assert.Equal(t, expected.User, readLogs[i].User)
		assert.Equal(t, expected.Action, readLogs[i].Action)
		assert.Equal(t, expected.Target, readLogs[i].Target)
		assert.Equal(t, expected.Success, readLogs[i].Success)
		assert.Equal(t, expected.Remote, readLogs[i].Remote)
	}
}

func TestGenerateParquetFile_EmptyData(t *testing.T) {
	handler := &Handler{}

	logs := []ParquetLog{}

	data, err := handler.generateParquetFile(logs)
	require.NoError(t, err)
	require.NotEmpty(t, data) // Even empty parquet files have metadata

	// Verify we can read back the empty data
	reader := parquet.NewGenericReader[ParquetLog](bytes.NewReader(data))
	defer reader.Close()

	readLogs := make([]ParquetLog, 10) // Allocate buffer
	n, err := reader.Read(readLogs)
	if err != nil && err.Error() != "EOF" {
		require.NoError(t, err) // Only fail if it's not EOF
	}
	assert.Equal(t, 0, n)
}

func TestGenerateParquetFile_SchemaValidation(t *testing.T) {
	handler := &Handler{}

	// Test with maximum field lengths and edge cases
	timestamp := time.Now().Unix()
	logs := []ParquetLog{
		{
			ID:        "very-long-id-that-tests-string-handling-in-parquet-format",
			Timestamp: timestamp * 1000, // Convert to milliseconds for consistency
			User:      "user_with_special_chars@domain.com",
			Action:    "read",
			Target:    "documents/subfolder/very-long-filename-with-spaces and special chars.txt",
			Success:   false,
			Remote:    "255.255.255.255",
		},
	}

	data, err := handler.generateParquetFile(logs)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Verify schema integrity by reading back
	reader := parquet.NewGenericReader[ParquetLog](bytes.NewReader(data))
	defer reader.Close()

	readLogs := make([]ParquetLog, 1)
	n, err := reader.Read(readLogs)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, logs[0], readLogs[0])
}
