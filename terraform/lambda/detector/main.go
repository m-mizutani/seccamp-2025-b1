package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

//go:embed queries/suspicious_login.sql
var suspiciousLoginQuery string

//go:embed queries/mass_data_access.sql
var massDataAccessQuery string

//go:embed queries/failed_auth.sql
var failedAuthQuery string

type Handler struct {
	athenaClient   AthenaAPI
	snsClient      SNSAPI
	database       string
	resultsBucket  string
	alertsTopicArn string
	queries        []QueryDefinition
}

func NewHandler() (*Handler, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	database := os.Getenv("ATHENA_DATABASE")
	if database == "" {
		return nil, fmt.Errorf("ATHENA_DATABASE environment variable is required")
	}

	resultsBucket := os.Getenv("ATHENA_RESULTS_BUCKET")
	if resultsBucket == "" {
		return nil, fmt.Errorf("ATHENA_RESULTS_BUCKET environment variable is required")
	}

	alertsTopicArn := os.Getenv("ALERTS_SNS_TOPIC_ARN")
	if alertsTopicArn == "" {
		return nil, fmt.Errorf("ALERTS_SNS_TOPIC_ARN environment variable is required")
	}

	queries := []QueryDefinition{
		{
			Name:        "suspicious_login",
			Description: "同一IPアドレスからの複数回ログイン失敗を検知",
			SQL:         suspiciousLoginQuery,
		},
		{
			Name:        "mass_data_access",
			Description: "短時間での大量データアクセスを検知",
			SQL:         massDataAccessQuery,
		},
		{
			Name:        "failed_auth",
			Description: "複数IPからの継続的な認証失敗を検知",
			SQL:         failedAuthQuery,
		},
	}

	return &Handler{
		athenaClient:   athena.NewFromConfig(cfg),
		snsClient:      sns.NewFromConfig(cfg),
		database:       database,
		resultsBucket:  resultsBucket,
		alertsTopicArn: alertsTopicArn,
		queries:        queries,
	}, nil
}

func (h *Handler) HandleEvent(ctx context.Context, event events.CloudWatchEvent) error {
	log.Printf("Starting alert detection process")

	for _, query := range h.queries {
		if err := h.executeQuery(ctx, query); err != nil {
			log.Printf("Error executing query %s: %v", query.Name, err)
			// Continue with other queries even if one fails
			continue
		}
	}

	log.Printf("Alert detection process completed")
	return nil
}

func (h *Handler) executeQuery(ctx context.Context, query QueryDefinition) error {
	log.Printf("Executing query: %s", query.Name)

	// Start query execution
	resultLocation := fmt.Sprintf("s3://%s/query-results/%s/", h.resultsBucket, query.Name)

	input := &athena.StartQueryExecutionInput{
		QueryString: aws.String(query.SQL),
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: aws.String(h.database),
		},
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: aws.String(resultLocation),
		},
		WorkGroup: aws.String("primary"),
	}

	result, err := h.athenaClient.StartQueryExecution(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start query execution: %w", err)
	}

	queryExecutionId := *result.QueryExecutionId
	log.Printf("Query execution started with ID: %s", queryExecutionId)

	// Wait for query to complete
	if err := h.waitForQueryCompletion(ctx, queryExecutionId); err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	// Get query results
	results, err := h.getQueryResults(ctx, queryExecutionId)
	if err != nil {
		return fmt.Errorf("failed to get query results: %w", err)
	}

	// Process results and send alerts if needed
	if len(results) > 0 {
		if err := h.sendAlert(ctx, query, results); err != nil {
			return fmt.Errorf("failed to send alert: %w", err)
		}
	} else {
		log.Printf("No alerts generated for query: %s", query.Name)
	}

	return nil
}

func (h *Handler) waitForQueryCompletion(ctx context.Context, queryExecutionId string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		input := &athena.GetQueryExecutionInput{
			QueryExecutionId: aws.String(queryExecutionId),
		}

		result, err := h.athenaClient.GetQueryExecution(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to get query execution status: %w", err)
		}

		status := result.QueryExecution.Status.State
		log.Printf("Query status: %s", status)

		switch status {
		case types.QueryExecutionStateSucceeded:
			return nil
		case types.QueryExecutionStateFailed, types.QueryExecutionStateCancelled:
			reason := ""
			if result.QueryExecution.Status.StateChangeReason != nil {
				reason = *result.QueryExecution.Status.StateChangeReason
			}
			return fmt.Errorf("query execution %s: %s", status, reason)
		default:
			// Query is still running, wait before checking again
			time.Sleep(5 * time.Second)
		}
	}
}

func (h *Handler) getQueryResults(ctx context.Context, queryExecutionId string) ([]QueryResult, error) {
	input := &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionId),
		MaxResults:       aws.Int32(100), // Limit results for alert processing
	}

	result, err := h.athenaClient.GetQueryResults(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var results []QueryResult
	if len(result.ResultSet.Rows) <= 1 {
		// No data rows (only header or empty)
		return results, nil
	}

	// Extract column names from header row
	var columns []string
	if len(result.ResultSet.Rows) > 0 {
		headerRow := result.ResultSet.Rows[0]
		for _, col := range headerRow.Data {
			if col.VarCharValue != nil {
				columns = append(columns, *col.VarCharValue)
			}
		}
	}

	// Process data rows
	for i := 1; i < len(result.ResultSet.Rows); i++ {
		row := result.ResultSet.Rows[i]
		data := make(map[string]string)

		for j, col := range row.Data {
			if j < len(columns) && col.VarCharValue != nil {
				data[columns[j]] = *col.VarCharValue
			}
		}

		if len(data) > 0 {
			results = append(results, QueryResult{Data: data})
		}
	}

	return results, nil
}

func (h *Handler) sendAlert(ctx context.Context, query QueryDefinition, results []QueryResult) error {
	log.Printf("Sending alert for query: %s with %d results", query.Name, len(results))

	// Create alert based on query type
	alert := h.createAlert(query, results)

	// Convert alert to JSON
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	// Send SNS notification
	input := &sns.PublishInput{
		TopicArn: aws.String(h.alertsTopicArn),
		Message:  aws.String(string(alertJSON)),
		Subject:  aws.String(alert.Title),
	}

	_, err = h.snsClient.Publish(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish SNS message: %w", err)
	}

	log.Printf("Alert sent successfully for query: %s", query.Name)
	return nil
}

func (h *Handler) createAlert(query QueryDefinition, results []QueryResult) Alert {
	switch query.Name {
	case "suspicious_login":
		return h.createSuspiciousLoginAlert(results)
	case "mass_data_access":
		return h.createMassDataAccessAlert(results)
	case "failed_auth":
		return h.createFailedAuthAlert(results)
	default:
		return Alert{
			Title:       fmt.Sprintf("Security Alert: %s", query.Name),
			Description: fmt.Sprintf("Query %s detected %d suspicious activities", query.Name, len(results)),
			Attrs: map[string]interface{}{
				"query_name":   query.Name,
				"result_count": len(results),
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
			},
		}
	}
}

func (h *Handler) createSuspiciousLoginAlert(results []QueryResult) Alert {
	var ips []string
	var users []string

	for _, result := range results {
		if ip := result.Data["remote"]; ip != "" {
			ips = append(ips, ip)
		}
		if user := result.Data["user"]; user != "" {
			users = append(users, user)
		}
	}

	return Alert{
		Title:       "不審なログイン試行を検知",
		Description: fmt.Sprintf("過去1時間で%d個のIPアドレスから複数回のログイン失敗が発生しました。攻撃者による総当り攻撃の可能性があります。", len(results)),
		Attrs: map[string]interface{}{
			"alert_type":     "suspicious_login",
			"affected_ips":   ips,
			"targeted_users": users,
			"event_count":    len(results),
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func (h *Handler) createMassDataAccessAlert(results []QueryResult) Alert {
	var users []string
	var ips []string

	for _, result := range results {
		if user := result.Data["user"]; user != "" {
			users = append(users, user)
		}
		if ip := result.Data["remote"]; ip != "" {
			ips = append(ips, ip)
		}
	}

	return Alert{
		Title:       "大量データアクセスを検知",
		Description: fmt.Sprintf("過去1時間で%d人のユーザーが短時間に大量のデータアクセスを実行しました。データ流出の可能性があります。", len(results)),
		Attrs: map[string]interface{}{
			"alert_type":       "mass_data_access",
			"suspicious_users": users,
			"source_ips":       ips,
			"event_count":      len(results),
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func (h *Handler) createFailedAuthAlert(results []QueryResult) Alert {
	var users []string
	totalIPs := 0

	for _, result := range results {
		if user := result.Data["user"]; user != "" {
			users = append(users, user)
		}
		if ips := result.Data["unique_ips"]; ips != "" {
			// Parse unique IP count if needed
		}
	}

	return Alert{
		Title:       "継続的認証失敗を検知",
		Description: fmt.Sprintf("過去24時間で%d人のユーザーに対して複数のIPアドレスからの継続的な認証失敗が発生しました。パスワード攻撃の可能性があります。", len(results)),
		Attrs: map[string]interface{}{
			"alert_type":       "failed_auth",
			"targeted_users":   users,
			"total_unique_ips": totalIPs,
			"event_count":      len(results),
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func main() {
	handler, err := NewHandler()
	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(handler.HandleEvent)
}
