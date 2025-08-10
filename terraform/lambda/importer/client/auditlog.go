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
	Kind        string    `json:"kind"`
	ID          LogID     `json:"id"`
	Actor       LogActor  `json:"actor"`
	OwnerDomain string    `json:"ownerDomain"`
	IPAddress   string    `json:"ipAddress"`
	Events      []LogEvent `json:"events"`
}

type LogID struct {
	Time             string `json:"time"`
	UniqueQualifier  string `json:"uniqueQualifier"`
	ApplicationName  string `json:"applicationName"`
	CustomerID       string `json:"customerId"`
}

type LogActor struct {
	CallerType string `json:"callerType"`
	Email      string `json:"email"`
	ProfileID  string `json:"profileId"`
}

type LogEvent struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Parameters []LogParameter `json:"parameters,omitempty"`
}

type LogParameter struct {
	Name         string      `json:"name"`
	Value        interface{} `json:"value"`
	IntValue     *int64      `json:"intValue,omitempty"`
	BoolValue    *bool       `json:"boolValue,omitempty"`
	MultiValue   []string    `json:"multiValue,omitempty"`
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
	u, err := url.Parse(c.baseURL)
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
	limit := 5000

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