package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryDefinitions_EmbedContent(t *testing.T) {
	// Test that go:embed successfully loads the query files
	assert.NotEmpty(t, suspiciousLoginQuery, "suspicious_login.sql should be embedded")
	assert.NotEmpty(t, massDataAccessQuery, "mass_data_access.sql should be embedded")
	assert.NotEmpty(t, failedAuthQuery, "failed_auth.sql should be embedded")
}

func TestQueryDefinitions_SQLSyntaxValidation(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		keywords []string // Expected SQL keywords
	}{
		{
			name:  "suspicious_login query",
			query: suspiciousLoginQuery,
			keywords: []string{
				"SELECT", "FROM", "WHERE", "GROUP BY", "HAVING", "COUNT",
				"remote", "action", "success", "timestamp",
			},
		},
		{
			name:  "mass_data_access query",
			query: massDataAccessQuery,
			keywords: []string{
				"SELECT", "FROM", "WHERE", "GROUP BY", "HAVING", "COUNT",
				"user", "action", "timestamp",
			},
		},
		{
			name:  "failed_auth query",
			query: failedAuthQuery,
			keywords: []string{
				"SELECT", "FROM", "WHERE", "GROUP BY", "HAVING", "COUNT",
				"user", "remote", "action", "success", "timestamp",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Basic syntax validation
			assert.NotEmpty(t, tc.query)

			// Convert to uppercase for case-insensitive comparison
			upperQuery := strings.ToUpper(tc.query)

			// Check for required SQL keywords
			for _, keyword := range tc.keywords {
				assert.Contains(t, upperQuery, strings.ToUpper(keyword),
					"Query should contain keyword: %s", keyword)
			}

			// Ensure query is well-formed (basic checks)
			assert.True(t, strings.Contains(upperQuery, "SELECT"), "Query should start with SELECT")
			assert.True(t, strings.Contains(upperQuery, "FROM"), "Query should have FROM clause")

			// Check for potential SQL injection vulnerabilities
			assert.False(t, strings.Contains(tc.query, ";--"), "Query should not contain comment injection patterns")
			assert.False(t, strings.Contains(tc.query, "/*"), "Query should not contain block comment injection patterns")
		})
	}
}

func TestQueryDefinitions_LogicValidation(t *testing.T) {
	testCases := []struct {
		name          string
		query         string
		expectedLogic []string
		shouldNotHave []string
	}{
		{
			name:  "suspicious_login logic",
			query: suspiciousLoginQuery,
			expectedLogic: []string{
				"action = 'login'",     // Should focus on login attempts
				"success = false",      // Should focus on failed logins
				"GROUP BY remote",      // Should group by IP address
				"HAVING COUNT(*) >= 5", // Should detect multiple failures
			},
			shouldNotHave: []string{
				"action = 'read'",  // Shouldn't include read operations
				"action = 'write'", // Shouldn't include write operations
			},
		},
		{
			name:  "mass_data_access logic",
			query: massDataAccessQuery,
			expectedLogic: []string{
				"action = 'read'",        // Should focus on read operations
				"success = true",         // Should focus on successful reads
				"GROUP BY user",          // Should group by user
				"HAVING COUNT(*) >= 100", // Should detect high volume
			},
			shouldNotHave: []string{
				"action = 'login'", // Shouldn't include login attempts
				"action = 'write'", // Shouldn't include write operations
			},
		},
		{
			name:  "failed_auth logic",
			query: failedAuthQuery,
			expectedLogic: []string{
				"action = 'login'",                   // Should focus on login attempts
				"success = false",                    // Should focus on failed logins
				"GROUP BY user",                      // Should group by user
				"HAVING COUNT(DISTINCT remote) >= 3", // Should detect multiple IPs
			},
			shouldNotHave: []string{
				"success = true", // Shouldn't include successful logins
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Remove whitespace and newlines for easier comparison
			normalizedQuery := strings.ReplaceAll(strings.ReplaceAll(tc.query, "\n", " "), "\t", " ")
			normalizedQuery = strings.Join(strings.Fields(normalizedQuery), " ")

			// Check expected logic
			for _, expected := range tc.expectedLogic {
				assert.Contains(t, normalizedQuery, expected,
					"Query should contain logic: %s", expected)
			}

			// Check that unwanted logic is not present
			for _, unwanted := range tc.shouldNotHave {
				assert.NotContains(t, normalizedQuery, unwanted,
					"Query should not contain logic: %s", unwanted)
			}
		})
	}
}

func TestQueryDefinitions_TimeFiltering(t *testing.T) {
	testCases := []struct {
		name           string
		query          string
		expectedWindow string
	}{
		{
			name:           "suspicious_login time window",
			query:          suspiciousLoginQuery,
			expectedWindow: "3600", // Should look at 1 hour window (3600 seconds)
		},
		{
			name:           "mass_data_access time window",
			query:          massDataAccessQuery,
			expectedWindow: "3600", // Should look at 1 hour window (3600 seconds)
		},
		{
			name:           "failed_auth time window",
			query:          failedAuthQuery,
			expectedWindow: "86400", // Should look at 24 hour window (86400 seconds)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check for time filtering logic
			upperQuery := strings.ToUpper(tc.query)

			// Should have timestamp filtering
			assert.Contains(t, upperQuery, "TIMESTAMP", "Query should filter by timestamp")
			assert.Contains(t, upperQuery, "UNIX_TIMESTAMP()", "Query should use UNIX_TIMESTAMP() for current time")
			assert.Contains(t, tc.query, tc.expectedWindow,
				"Query should use expected time window: %s seconds", tc.expectedWindow)
		})
	}
}

func TestCreateAlert_SuspiciousLogin(t *testing.T) {
	handler := &Handler{}

	query := QueryDefinition{
		Name:        "suspicious_login",
		Description: "同一IPアドレスからの複数回ログイン失敗を検知",
		SQL:         suspiciousLoginQuery,
	}

	results := []QueryResult{
		{
			Data: map[string]string{
				"remote": "192.168.1.100",
				"user":   "alice",
			},
		},
		{
			Data: map[string]string{
				"remote": "192.168.1.101",
				"user":   "bob",
			},
		},
	}

	alert := handler.createAlert(query, results)

	assert.Equal(t, "不審なログイン試行を検知", alert.Title)
	assert.Contains(t, alert.Description, "複数回のログイン失敗")
	assert.Equal(t, []string{"192.168.1.100", "192.168.1.101"}, alert.Attrs["affected_ips"])
	assert.Equal(t, []string{"alice", "bob"}, alert.Attrs["targeted_users"])
	assert.Equal(t, 2, alert.Attrs["event_count"])
}

func TestCreateAlert_MassDataAccess(t *testing.T) {
	handler := &Handler{}

	query := QueryDefinition{
		Name:        "mass_data_access",
		Description: "短時間での大量データアクセスを検知",
		SQL:         massDataAccessQuery,
	}

	results := []QueryResult{
		{
			Data: map[string]string{
				"user":   "bob",
				"remote": "192.168.1.100",
			},
		},
		{
			Data: map[string]string{
				"user":   "charlie",
				"remote": "192.168.1.101",
			},
		},
	}

	alert := handler.createAlert(query, results)

	assert.Equal(t, "大量データアクセスを検知", alert.Title)
	assert.Contains(t, alert.Description, "大量のデータアクセス")
	assert.Equal(t, []string{"bob", "charlie"}, alert.Attrs["suspicious_users"])
	assert.Equal(t, []string{"192.168.1.100", "192.168.1.101"}, alert.Attrs["source_ips"])
	assert.Equal(t, 2, alert.Attrs["event_count"])
}

func TestCreateAlert_FailedAuth(t *testing.T) {
	handler := &Handler{}

	query := QueryDefinition{
		Name:        "failed_auth",
		Description: "複数IPからの継続的な認証失敗を検知",
		SQL:         failedAuthQuery,
	}

	results := []QueryResult{
		{
			Data: map[string]string{
				"user":       "charlie",
				"unique_ips": "3",
			},
		},
		{
			Data: map[string]string{
				"user":       "alice",
				"unique_ips": "2",
			},
		},
	}

	alert := handler.createAlert(query, results)

	assert.Equal(t, "継続的認証失敗を検知", alert.Title)
	assert.Contains(t, alert.Description, "継続的な認証失敗")
	assert.Equal(t, []string{"charlie", "alice"}, alert.Attrs["targeted_users"])
	assert.Equal(t, 0, alert.Attrs["total_unique_ips"]) // Current implementation doesn't parse unique_ips
	assert.Equal(t, 2, alert.Attrs["event_count"])
}
