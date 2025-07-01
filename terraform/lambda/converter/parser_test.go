package main

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSONL_ValidData(t *testing.T) {
	jsonlData := `{"id":"log001","timestamp":"2023-12-01T10:00:00Z","user":"alice","action":"login","target":"","success":true,"remote":"192.168.1.100"}
{"id":"log002","timestamp":"2023-12-01T10:05:00Z","user":"bob","action":"read","target":"document1.txt","success":true,"remote":"192.168.1.101"}`

	var rawLogs []RawLog
	scanner := bufio.NewScanner(strings.NewReader(jsonlData))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var rawLog RawLog
		err := json.Unmarshal([]byte(line), &rawLog)
		require.NoError(t, err, "Line %d should parse successfully", lineNum)
		rawLogs = append(rawLogs, rawLog)
	}

	require.NoError(t, scanner.Err())
	assert.Len(t, rawLogs, 2)

	// Verify first log
	expectedTime1, _ := time.Parse(time.RFC3339, "2023-12-01T10:00:00Z")
	assert.Equal(t, "log001", rawLogs[0].ID)
	assert.Equal(t, expectedTime1, rawLogs[0].Timestamp)
	assert.Equal(t, "alice", rawLogs[0].User)
	assert.Equal(t, "login", rawLogs[0].Action)
	assert.Equal(t, "", rawLogs[0].Target)
	assert.Equal(t, true, rawLogs[0].Success)
	assert.Equal(t, "192.168.1.100", rawLogs[0].Remote)

	// Verify second log
	expectedTime2, _ := time.Parse(time.RFC3339, "2023-12-01T10:05:00Z")
	assert.Equal(t, "log002", rawLogs[1].ID)
	assert.Equal(t, expectedTime2, rawLogs[1].Timestamp)
	assert.Equal(t, "bob", rawLogs[1].User)
	assert.Equal(t, "read", rawLogs[1].Action)
	assert.Equal(t, "document1.txt", rawLogs[1].Target)
	assert.Equal(t, true, rawLogs[1].Success)
	assert.Equal(t, "192.168.1.101", rawLogs[1].Remote)
}

func TestParseJSONL_InvalidData(t *testing.T) {
	testCases := []struct {
		name     string
		jsonLine string
		wantErr  bool
	}{
		{
			name:     "invalid json",
			jsonLine: `{"id":"log001","timestamp":invalid}`,
			wantErr:  true,
		},
		{
			name:     "missing required field",
			jsonLine: `{"timestamp":"2023-12-01T10:00:00Z","user":"alice"}`,
			wantErr:  false, // json.Unmarshal doesn't fail on missing fields, just sets zero values
		},
		{
			name:     "invalid timestamp format",
			jsonLine: `{"id":"log001","timestamp":"invalid-time","user":"alice","action":"login","target":"","success":true,"remote":"192.168.1.100"}`,
			wantErr:  true,
		},
		{
			name:     "invalid boolean",
			jsonLine: `{"id":"log001","timestamp":"2023-12-01T10:00:00Z","user":"alice","action":"login","target":"","success":"not_bool","remote":"192.168.1.100"}`,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rawLog RawLog
			err := json.Unmarshal([]byte(tc.jsonLine), &rawLog)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseJSONL_EmptyLines(t *testing.T) {
	jsonlData := `{"id":"log001","timestamp":"2023-12-01T10:00:00Z","user":"alice","action":"login","target":"","success":true,"remote":"192.168.1.100"}

{"id":"log002","timestamp":"2023-12-01T10:05:00Z","user":"bob","action":"read","target":"document1.txt","success":true,"remote":"192.168.1.101"}
	
`

	var rawLogs []RawLog
	scanner := bufio.NewScanner(strings.NewReader(jsonlData))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var rawLog RawLog
		err := json.Unmarshal([]byte(line), &rawLog)
		require.NoError(t, err)
		rawLogs = append(rawLogs, rawLog)
	}

	require.NoError(t, scanner.Err())
	assert.Len(t, rawLogs, 2, "Should parse 2 logs ignoring empty lines")
}
