package transformer

import (
	"compress/gzip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/m-mizutani/seccamp-2025-b1/terraform/lambda/importer/client"
)

func TestToJSONL(t *testing.T) {
	transformer := NewJSONLTransformer()

	// Test with empty logs
	emptyResult, err := transformer.ToJSONL([]client.LogEntry{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(emptyResult) != 0 {
		t.Error("Expected empty result for empty logs")
	}

	// Test with sample logs
	logs := []client.LogEntry{
		{
			ID: client.LogID{Time: "2024-08-12T10:05:00Z"},
			Timestamp: "2024-08-12T10:05:00Z",
			User: client.LogUser{
				Email:  "test@example.com",
				Name:   "Test User",
				Domain: "example.com",
			},
			Event: client.LogEvent{
				Type:   "drive",
				Name:   "access",
				Action: "view",
			},
		},
		{
			ID: client.LogID{Time: "2024-08-12T10:06:00Z"},
			Timestamp: "2024-08-12T10:06:00Z",
			User: client.LogUser{
				Email:  "admin@example.com",
				Name:   "Admin User",
				Domain: "example.com",
			},
			Event: client.LogEvent{
				Type:   "admin",
				Name:   "user_create",
				Action: "create",
			},
		},
	}

	result, err := transformer.ToJSONL(logs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(string(result), "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var logEntry client.LogEntry
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestCompress(t *testing.T) {
	transformer := NewJSONLTransformer()

	// Test with empty data
	emptyResult, err := transformer.Compress([]byte{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(emptyResult) != 0 {
		t.Error("Expected empty result for empty data")
	}

	// Test with sample data
	testData := []byte("test data\nline 2\nline 3")
	compressed, err := transformer.Compress(testData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify compression worked
	if len(compressed) == 0 {
		t.Error("Expected non-empty compressed data")
	}

	// Verify we can decompress
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	var decompressed bytes.Buffer
	if _, err := decompressed.ReadFrom(reader); err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	if !bytes.Equal(testData, decompressed.Bytes()) {
		t.Error("Decompressed data doesn't match original")
	}
}

func TestTransform(t *testing.T) {
	transformer := NewJSONLTransformer()

	logs := []client.LogEntry{
		{
			ID: client.LogID{Time: "2024-08-12T10:05:00Z"},
			Timestamp: "2024-08-12T10:05:00Z",
			User: client.LogUser{
				Email:  "test@example.com",
				Name:   "Test User",
				Domain: "example.com",
			},
		},
	}

	result, err := transformer.Transform(logs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}

	// Verify it's compressed by trying to decompress
	reader, err := gzip.NewReader(bytes.NewReader(result))
	if err != nil {
		t.Fatalf("Result is not valid gzip: %v", err)
	}
	defer reader.Close()

	var decompressed bytes.Buffer
	if _, err := decompressed.ReadFrom(reader); err != nil {
		t.Fatalf("Failed to decompress result: %v", err)
	}

	// Verify decompressed data is valid JSONL
	lines := strings.Split(strings.TrimSuffix(decompressed.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	var logEntry client.LogEntry
	if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
		t.Errorf("Decompressed line is not valid JSON: %v", err)
	}
}