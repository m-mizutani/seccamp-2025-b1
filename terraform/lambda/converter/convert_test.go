package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestConvertToOCSF_WebResourcesPopulated(t *testing.T) {
	// Read test data
	data, err := os.ReadFile("testdata/drive.jsonl")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		t.Fatal("No test data found")
	}

	// Test cases for different log types
	testCases := []struct {
		name                  string
		lineIndex             int
		expectWebResources    bool
		expectedResourceName  string
		expectedResourceID    string
		expectedResourceType  string
	}{
		{
			name:                 "Drive access log should have web resources",
			lineIndex:            0, // First line is a drive access log
			expectWebResources:   true,
			expectedResourceName: "教材/数学/教科書.pdf",
			expectedResourceID:   "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE1000",
			expectedResourceType: "document",
		},
		{
			name:               "Login log should not have web resources",
			lineIndex:          1, // Second line is a login log
			expectWebResources: false,
		},
		{
			name:                 "Another drive access log should have web resources",
			lineIndex:            2, // Third line is a drive access log
			expectWebResources:   true,
			expectedResourceName: "教材/数学/問題集.pdf",
			expectedResourceID:   "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE1001",
			expectedResourceType: "document",
		},
		{
			name:               "Calendar log should not have web resources",
			lineIndex:          4, // Fifth line is a calendar log
			expectWebResources: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the log line
			var log GoogleWorkspaceLog
			if err := json.Unmarshal([]byte(lines[tc.lineIndex]), &log); err != nil {
				t.Fatalf("Failed to parse log: %v", err)
			}

			// Convert to OCSF
			ocsf, err := ConvertToOCSF(&log, "ap-northeast-1", "123456789012")
			if err != nil {
				t.Fatalf("Failed to convert to OCSF: %v", err)
			}

			// Check web resources
			if tc.expectWebResources {
				if len(ocsf.WebResources) == 0 {
					t.Error("Expected web resources but got none")
				} else {
					// Verify the first web resource
					resource := ocsf.WebResources[0]
					
					if resource.Name != tc.expectedResourceName {
						t.Errorf("Expected resource name %q, got %q", tc.expectedResourceName, resource.Name)
					}
					
					if resource.UID != tc.expectedResourceID {
						t.Errorf("Expected resource ID %q, got %q", tc.expectedResourceID, resource.UID)
					}
					
					if resource.Type != tc.expectedResourceType {
						t.Errorf("Expected resource type %q, got %q", tc.expectedResourceType, resource.Type)
					}
					
					// Check URL construction
					expectedURL := "https://docs.google.com/document/d/" + tc.expectedResourceID
					if resource.URLString != expectedURL {
						t.Errorf("Expected URL %q, got %q", expectedURL, resource.URLString)
					}
					
					// Check classification
					if resource.Data.Classification != "internal" {
						t.Errorf("Expected classification 'internal', got %q", resource.Data.Classification)
					}
				}
			} else {
				if len(ocsf.WebResources) > 0 {
					t.Errorf("Expected no web resources but got %d", len(ocsf.WebResources))
				}
			}
		})
	}
}

func TestExtractWebResourcesFromEventParameters(t *testing.T) {
	// Test the extraction function directly
	events := []struct {
		Type       string `json:"type"`
		Name       string `json:"name"`
		Parameters []struct {
			Name       string      `json:"name"`
			Value      interface{} `json:"value"`
			IntValue   *int64      `json:"intValue,omitempty"`
			BoolValue  *bool       `json:"boolValue,omitempty"`
			MultiValue []string    `json:"multiValue,omitempty"`
		} `json:"parameters,omitempty"`
	}{
		{
			Type: "access",
			Name: "view",
			Parameters: []struct {
				Name       string      `json:"name"`
				Value      interface{} `json:"value"`
				IntValue   *int64      `json:"intValue,omitempty"`
				BoolValue  *bool       `json:"boolValue,omitempty"`
				MultiValue []string    `json:"multiValue,omitempty"`
			}{
				{Name: "doc_id", Value: "1234567890"},
				{Name: "doc_title", Value: "Test Document.pdf"},
				{Name: "doc_type", Value: "document"},
			},
		},
	}

	resources := extractWebResourcesFromEventParameters(events)

	if len(resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(resources))
	}

	resource := resources[0]
	if resource.Name != "Test Document.pdf" {
		t.Errorf("Expected name 'Test Document.pdf', got %q", resource.Name)
	}
	if resource.UID != "1234567890" {
		t.Errorf("Expected UID '1234567890', got %q", resource.UID)
	}
	if resource.Type != "document" {
		t.Errorf("Expected type 'document', got %q", resource.Type)
	}
}

func TestWebResourceURLGeneration(t *testing.T) {
	// Test URL generation for different document types
	testCases := []struct {
		docType     string
		docID       string
		expectedURL string
	}{
		{
			docType:     "spreadsheet",
			docID:       "abc123",
			expectedURL: "https://docs.google.com/spreadsheets/d/abc123",
		},
		{
			docType:     "presentation",
			docID:       "def456",
			expectedURL: "https://docs.google.com/presentation/d/def456",
		},
		{
			docType:     "folder",
			docID:       "ghi789",
			expectedURL: "https://drive.google.com/drive/folders/ghi789",
		},
		{
			docType:     "document",
			docID:       "jkl012",
			expectedURL: "https://docs.google.com/document/d/jkl012",
		},
		{
			docType:     "unknown",
			docID:       "mno345",
			expectedURL: "https://docs.google.com/document/d/mno345", // Default to document
		},
	}

	for _, tc := range testCases {
		t.Run(tc.docType, func(t *testing.T) {
			events := []struct {
				Type       string `json:"type"`
				Name       string `json:"name"`
				Parameters []struct {
					Name       string      `json:"name"`
					Value      interface{} `json:"value"`
					IntValue   *int64      `json:"intValue,omitempty"`
					BoolValue  *bool       `json:"boolValue,omitempty"`
					MultiValue []string    `json:"multiValue,omitempty"`
				} `json:"parameters,omitempty"`
			}{
				{
					Type: "access",
					Name: "view",
					Parameters: []struct {
						Name       string      `json:"name"`
						Value      interface{} `json:"value"`
						IntValue   *int64      `json:"intValue,omitempty"`
						BoolValue  *bool       `json:"boolValue,omitempty"`
						MultiValue []string    `json:"multiValue,omitempty"`
					}{
						{Name: "doc_id", Value: tc.docID},
						{Name: "doc_title", Value: "Test Document"},
						{Name: "doc_type", Value: tc.docType},
					},
				},
			}

			resources := extractWebResourcesFromEventParameters(events)
			if len(resources) != 1 {
				t.Fatalf("Expected 1 resource, got %d", len(resources))
			}

			if resources[0].URLString != tc.expectedURL {
				t.Errorf("Expected URL %q, got %q", tc.expectedURL, resources[0].URLString)
			}
		})
	}
}