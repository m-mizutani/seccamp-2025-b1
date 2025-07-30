package transformer

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"

	"github.com/m-mizutani/seccamp-2025-b1/terraform/lambda/importer/client"
)

type JSONLTransformer struct{}

func NewJSONLTransformer() *JSONLTransformer {
	return &JSONLTransformer{}
}

// ToJSONL converts log entries to JSON Lines format
func (t *JSONLTransformer) ToJSONL(logs []client.LogEntry) ([]byte, error) {
	if len(logs) == 0 {
		return []byte{}, nil
	}

	var buffer bytes.Buffer
	
	for _, log := range logs {
		jsonData, err := json.Marshal(log)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal log entry: %w", err)
		}
		
		buffer.Write(jsonData)
		buffer.WriteByte('\n')
	}

	return buffer.Bytes(), nil
}

// Compress compresses data using gzip
func (t *JSONLTransformer) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	
	if _, err := gzipWriter.Write(data); err != nil {
		gzipWriter.Close()
		return nil, fmt.Errorf("failed to write gzip data: %w", err)
	}
	
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return compressed.Bytes(), nil
}

// Transform converts logs to compressed JSONL format
func (t *JSONLTransformer) Transform(logs []client.LogEntry) ([]byte, error) {
	jsonlData, err := t.ToJSONL(logs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSONL: %w", err)
	}

	compressedData, err := t.Compress(jsonlData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	return compressedData, nil
}