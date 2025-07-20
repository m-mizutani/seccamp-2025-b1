package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type AuditlogClient struct {
	httpClient *http.Client
	baseURL    string
}

type LogEntry struct {
	ID        LogID                  `json:"id"`
	Timestamp string                 `json:"timestamp"`
	User      LogUser               `json:"user"`
	Event     LogEvent              `json:"event"`
	Resource  LogResource           `json:"resource"`
	Metadata  LogMetadata           `json:"metadata"`
	Result    LogResult             `json:"result"`
}

type LogID struct {
	Time string `json:"time"`
}

type LogUser struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type LogEvent struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
}

type LogResource struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

type LogMetadata struct {
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent"`
	Location  LogLocation `json:"location"`
}

type LogLocation struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

type LogResult struct {
	Success      bool   `json:"success"`
	DeniedReason *string `json:"denied_reason"`
}

type LogResponse struct {
	Date     string           `json:"date"`
	Metadata ResponseMetadata `json:"metadata"`
	Logs     []LogEntry       `json:"logs"`
}

type ResponseMetadata struct {
	Total     int       `json:"total"`
	Offset    int       `json:"offset"`
	Limit     int       `json:"limit"`
	Generated time.Time `json:"generated"`
}

func NewAuditlogClient(baseURL string, timeout time.Duration) *AuditlogClient {
	return &AuditlogClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

func (c *AuditlogClient) FetchLogs(ctx context.Context, startTime, endTime time.Time, offset, limit int) (*LogResponse, error) {
	u, err := url.Parse(c.baseURL + "/logs")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("startTime", startTime.Format(time.RFC3339))
	q.Set("endTime", endTime.Format(time.RFC3339))
	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	var logResponse LogResponse
	if err := json.NewDecoder(resp.Body).Decode(&logResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &logResponse, nil
}

func (c *AuditlogClient) FetchAllLogs(ctx context.Context, startTime, endTime time.Time) ([]LogEntry, error) {
	var allLogs []LogEntry
	offset := 0
	limit := 100

	for {
		resp, err := c.FetchLogs(ctx, startTime, endTime, offset, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch logs at offset %d: %w", offset, err)
		}

		allLogs = append(allLogs, resp.Logs...)

		// Check if we've fetched all logs
		if offset+len(resp.Logs) >= resp.Metadata.Total {
			break
		}

		offset += limit
	}

	return allLogs, nil
}