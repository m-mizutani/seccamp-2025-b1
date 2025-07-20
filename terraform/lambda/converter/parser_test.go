package main

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGoogleWorkspaceJSONL_ValidData(t *testing.T) {
	jsonlData := `{"kind":"audit#activity","id":{"time":"2024-08-12T10:15:30.123456Z","uniqueQualifier":"358068855354","applicationName":"drive","customerId":"C03az79cb"},"actor":{"callerType":"USER","email":"user@muhai-academy.com","profileId":"114511147312345678901"},"ownerDomain":"muhai-academy.com","ipAddress":"203.0.113.255","events":[{"type":"access","name":"view","parameters":[{"name":"doc_id","value":"1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"}]}]}
{"kind":"audit#activity","id":{"time":"2024-08-12T22:30:15.789Z","uniqueQualifier":"358068855355","applicationName":"drive","customerId":"C03az79cb"},"actor":{"callerType":"USER","email":"admin@muhai-academy.com","profileId":"114511147312345678902"},"ownerDomain":"muhai-academy.com","ipAddress":"198.51.100.42","events":[{"type":"access","name":"download","parameters":[{"name":"doc_id","value":"1A2B3C4D5E6F7G8H9I0J"}]}]}`

	var gwLogs []GoogleWorkspaceLog
	scanner := bufio.NewScanner(strings.NewReader(jsonlData))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var gwLog GoogleWorkspaceLog
		err := json.Unmarshal([]byte(line), &gwLog)
		require.NoError(t, err, "Line %d should parse successfully", lineNum)
		gwLogs = append(gwLogs, gwLog)
	}

	require.NoError(t, scanner.Err())
	assert.Len(t, gwLogs, 2)

	// Verify first log
	assert.Equal(t, "audit#activity", gwLogs[0].Kind)
	assert.Equal(t, "2024-08-12T10:15:30.123456Z", gwLogs[0].ID.Time)
	assert.Equal(t, "drive", gwLogs[0].ID.ApplicationName)
	assert.Equal(t, "user@muhai-academy.com", gwLogs[0].Actor.Email)
	assert.Equal(t, "203.0.113.255", gwLogs[0].IPAddress)
	assert.Len(t, gwLogs[0].Events, 1)
	assert.Equal(t, "view", gwLogs[0].Events[0].Name)

	// Verify second log
	assert.Equal(t, "audit#activity", gwLogs[1].Kind)
	assert.Equal(t, "2024-08-12T22:30:15.789Z", gwLogs[1].ID.Time)
	assert.Equal(t, "drive", gwLogs[1].ID.ApplicationName)
	assert.Equal(t, "admin@muhai-academy.com", gwLogs[1].Actor.Email)
	assert.Equal(t, "198.51.100.42", gwLogs[1].IPAddress)
	assert.Len(t, gwLogs[1].Events, 1)
	assert.Equal(t, "download", gwLogs[1].Events[0].Name)
}

func TestParseGoogleWorkspaceJSONL_InvalidData(t *testing.T) {
	testCases := []struct {
		name     string
		jsonLine string
		wantErr  bool
	}{
		{
			name:     "invalid json",
			jsonLine: `{"kind":"audit#activity","id":invalid}`,
			wantErr:  true,
		},
		{
			name:     "missing required field",
			jsonLine: `{"actor":{"email":"user@example.com"}}`,
			wantErr:  false, // json.Unmarshal doesn't fail on missing fields, just sets zero values
		},
		{
			name:     "invalid nested structure",
			jsonLine: `{"kind":"audit#activity","id":"not_an_object"}`,
			wantErr:  true,
		},
		{
			name:     "invalid array type",
			jsonLine: `{"kind":"audit#activity","events":"not_an_array"}`,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gwLog GoogleWorkspaceLog
			err := json.Unmarshal([]byte(tc.jsonLine), &gwLog)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseGoogleWorkspaceJSONL_EmptyLines(t *testing.T) {
	jsonlData := `{"kind":"audit#activity","id":{"time":"2024-08-12T10:15:30.123456Z","uniqueQualifier":"358068855354","applicationName":"drive","customerId":"C03az79cb"},"actor":{"callerType":"USER","email":"user@muhai-academy.com","profileId":"114511147312345678901"},"ownerDomain":"muhai-academy.com","ipAddress":"203.0.113.255","events":[{"type":"access","name":"view"}]}

{"kind":"audit#activity","id":{"time":"2024-08-12T22:30:15.789Z","uniqueQualifier":"358068855355","applicationName":"drive","customerId":"C03az79cb"},"actor":{"callerType":"USER","email":"admin@muhai-academy.com","profileId":"114511147312345678902"},"ownerDomain":"muhai-academy.com","ipAddress":"198.51.100.42","events":[{"type":"access","name":"download"}]}
	
`

	var gwLogs []GoogleWorkspaceLog
	scanner := bufio.NewScanner(strings.NewReader(jsonlData))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var gwLog GoogleWorkspaceLog
		err := json.Unmarshal([]byte(line), &gwLog)
		require.NoError(t, err)
		gwLogs = append(gwLogs, gwLog)
	}

	require.NoError(t, scanner.Err())
	assert.Len(t, gwLogs, 2, "Should parse 2 logs ignoring empty lines")
}
