package main

import (
	"fmt"
	"strings"
	"time"
)

// ConvertToOCSF converts Google Workspace log to OCSF Web Resources Activity format
func ConvertToOCSF(log *GoogleWorkspaceLog, region string, accountID string) (*OCSFWebResourceActivity, error) {
	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, log.ID.Time)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	// Get first event (Google Workspace logs can have multiple events)
	var firstEvent struct {
		Type   string `json:"type"`
		Name   string `json:"name"`
		Action string `json:"action"`
	}
	if len(log.Events) > 0 {
		firstEvent.Type = log.Events[0].Type
		firstEvent.Name = log.Events[0].Name
		firstEvent.Action = log.ID.ApplicationName // Use application name as action
	}

	// Determine activity_id based on event type and name
	activityID := mapActivityIDFromEvent(firstEvent)

	// Determine severity_id
	severityID := mapSeverityIDFromEvent(firstEvent)

	// Determine status_id (assume success since Google Workspace logs successful events)
	statusID := 1 // Success

	// Determine user type (admin or regular user)
	userTypeID := mapUserTypeIDFromEvent(firstEvent)

	ocsf := &OCSFWebResourceActivity{
		// Basic classification
		CategoryUID: 6,    // Application Activity
		ClassUID:    6001, // Web Resources Activity
		TypeUID:     600100 + activityID,
		ActivityID:  activityID,
		SeverityID:  severityID,
		Time:        timestamp.UnixMilli(),
		StatusID:    statusID,

		// Partitioning fields
		Region:    region,
		AccountID: accountID,
		EventHour: timestamp.Format("2006-01-02-15"),
	}

	// Actor information
	ocsf.Actor.User.UID = log.Actor.ProfileID
	if ocsf.Actor.User.UID == "" {
		ocsf.Actor.User.UID = log.Actor.Email // Fallback to email if ProfileID not available
	}
	ocsf.Actor.User.EmailAddr = log.Actor.Email
	ocsf.Actor.User.Domain = log.OwnerDomain
	ocsf.Actor.User.TypeID = userTypeID

	// Session information - generate UID from email and timestamp
	ocsf.Actor.Session.UID = fmt.Sprintf("%s_%d", log.Actor.Email, timestamp.Unix())
	ocsf.Actor.Session.CreatedTime = timestamp.Add(-1 * time.Hour).UnixMilli() // Estimate session start

	// App information
	ocsf.Actor.AppName = "Google Workspace"
	ocsf.Actor.AppUID = log.ID.ApplicationName

	// API information
	ocsf.API.Service.Name = mapServiceNameFromEvent(firstEvent)
	ocsf.API.Service.Version = "v3"
	ocsf.API.Operation = firstEvent.Name
	if ocsf.API.Operation == "" {
		ocsf.API.Operation = firstEvent.Type
	}
	ocsf.API.Request.UID = fmt.Sprintf("req_%d", timestamp.Unix())
	ocsf.API.Response.Code = getResponseCode(statusID)
	ocsf.API.Response.Message = getResponseMessage(statusID)

	// Cloud information
	ocsf.Cloud.Provider = "Google Cloud"
	ocsf.Cloud.Account.UID = log.ID.CustomerID
	ocsf.Cloud.Account.Name = strings.Split(log.OwnerDomain, ".")[0]
	ocsf.Cloud.Org.Name = log.OwnerDomain
	ocsf.Cloud.Org.UID = log.ID.CustomerID
	ocsf.Cloud.Region = "asia-northeast1" // Default to Tokyo region

	// Source endpoint
	ocsf.SrcEndpoint.IP = log.IPAddress
	// Location information is not available in the original Google Workspace log structure

	// Web resources - extract from event parameters
	ocsf.WebResources = extractWebResourcesFromEventParameters(log.Events)

	// Store original log data in observables for easier analysis
	observables := []struct {
		Name  string `parquet:"name"`
		Type  string `parquet:"type"`
		Value string `parquet:"value"`
	}{
		{Name: "kind", Type: "original", Value: log.Kind},
		{Name: "unique_qualifier", Type: "original", Value: log.ID.UniqueQualifier},
		{Name: "application_name", Type: "original", Value: log.ID.ApplicationName},
		{Name: "customer_id", Type: "original", Value: log.ID.CustomerID},
		{Name: "caller_type", Type: "original", Value: log.Actor.CallerType},
		{Name: "actor_email", Type: "original", Value: log.Actor.Email},
		{Name: "actor_profile_id", Type: "original", Value: log.Actor.ProfileID},
		{Name: "owner_domain", Type: "original", Value: log.OwnerDomain},
		{Name: "ip_address", Type: "original", Value: log.IPAddress},
	}

	// Add event information
	for i, event := range log.Events {
		observables = append(observables,
			struct {
				Name  string `parquet:"name"`
				Type  string `parquet:"type"`
				Value string `parquet:"value"`
			}{Name: fmt.Sprintf("event_%d_type", i), Type: "original", Value: event.Type},
			struct {
				Name  string `parquet:"name"`
				Type  string `parquet:"type"`
				Value string `parquet:"value"`
			}{Name: fmt.Sprintf("event_%d_name", i), Type: "original", Value: event.Name},
		)
	}

	// Filter out empty values
	ocsf.Observables = []struct {
		Name  string `parquet:"name"`
		Type  string `parquet:"type"`
		Value string `parquet:"value"`
	}{}

	for _, obs := range observables {
		if obs.Value != "" {
			ocsf.Observables = append(ocsf.Observables, obs)
		}
	}

	// Metadata
	ocsf.Metadata.UID = log.ID.UniqueQualifier                                           // Store uniqueQualifier as UID
	ocsf.Metadata.CorrelationUID = fmt.Sprintf("gw_%s_%s", log.Actor.Email, log.ID.Time) // Unique ID for log deduplication
	ocsf.Metadata.OriginalTime = log.ID.Time
	ocsf.Metadata.Processed = time.Now().UnixMilli()
	ocsf.Metadata.ProductName = "Google Workspace"
	ocsf.Metadata.Version = "1.0.0"

	// Add original log information to labels
	labels := []string{}
	if log.Kind != "" {
		labels = append(labels, fmt.Sprintf("kind:%s", log.Kind))
	}
	if log.ID.ApplicationName != "" {
		labels = append(labels, fmt.Sprintf("application:%s", log.ID.ApplicationName))
	}
	if log.Actor.CallerType != "" {
		labels = append(labels, fmt.Sprintf("caller_type:%s", log.Actor.CallerType))
	}
	if log.OwnerDomain != "" {
		labels = append(labels, fmt.Sprintf("domain:%s", log.OwnerDomain))
	}
	if len(log.Events) > 0 {
		labels = append(labels, fmt.Sprintf("event_type:%s", log.Events[0].Type))
		if log.Events[0].Name != "" {
			labels = append(labels, fmt.Sprintf("event_name:%s", log.Events[0].Name))
		}
	}
	ocsf.Metadata.Labels = labels

	return ocsf, nil
}

// extractWebResourcesFromEventParameters extracts file/resource information from event parameters
func extractWebResourcesFromEventParameters(events []struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Parameters []struct {
		Name       string      `json:"name"`
		Value      interface{} `json:"value"`
		IntValue   *int64      `json:"intValue,omitempty"`
		BoolValue  *bool       `json:"boolValue,omitempty"`
		MultiValue []string    `json:"multiValue,omitempty"`
	} `json:"parameters,omitempty"`
}) []struct {
	Name      string `parquet:"name,optional"`
	UID       string `parquet:"uid,optional"`
	Type      string `parquet:"type,optional"`
	URLString string `parquet:"url_string,optional"`
	Data      struct {
		Classification string `parquet:"classification,optional"`
	} `parquet:"data,optional"`
} {
	var resources []struct {
		Name      string `parquet:"name,optional"`
		UID       string `parquet:"uid,optional"`
		Type      string `parquet:"type,optional"`
		URLString string `parquet:"url_string,optional"`
		Data      struct {
			Classification string `parquet:"classification,optional"`
		} `parquet:"data,optional"`
	}

	// Extract document/resource information from event parameters
	for _, event := range events {
		for _, param := range event.Parameters {
			if param.Name == "doc_id" || param.Name == "document_id" || param.Name == "file_id" {
				if docID, ok := param.Value.(string); ok && docID != "" {
					webResource := struct {
						Name      string `parquet:"name,optional"`
						UID       string `parquet:"uid,optional"`
						Type      string `parquet:"type,optional"`
						URLString string `parquet:"url_string,optional"`
						Data      struct {
							Classification string `parquet:"classification,optional"`
						} `parquet:"data,optional"`
					}{
						UID:  docID,
						Type: "document",
					}

					// Build URL based on document ID
					webResource.URLString = fmt.Sprintf("https://docs.google.com/document/d/%s", docID)
					webResource.Data.Classification = "internal"

					resources = append(resources, webResource)
				}
			}
		}
	}

	return resources
}

// mapActivityIDFromEvent maps Google Workspace events to OCSF activity_id
func mapActivityIDFromEvent(event struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
}) int {
	eventType := event.Type
	eventName := event.Name
	eventAction := event.Action

	// Use name first, then action, then type for mapping
	eventToMap := eventName
	if eventToMap == "" {
		eventToMap = eventAction
	}
	if eventToMap == "" {
		eventToMap = eventType
	}

	// Map based on common patterns
	eventToMapLower := strings.ToLower(eventToMap)

	if strings.Contains(eventToMapLower, "create") || strings.Contains(eventToMapLower, "add") || strings.Contains(eventToMapLower, "new") {
		return 1 // Create
	}
	if strings.Contains(eventToMapLower, "view") || strings.Contains(eventToMapLower, "read") || strings.Contains(eventToMapLower, "access") || strings.Contains(eventToMapLower, "login") {
		return 2 // Read
	}
	if strings.Contains(eventToMapLower, "edit") || strings.Contains(eventToMapLower, "update") || strings.Contains(eventToMapLower, "modify") || strings.Contains(eventToMapLower, "change") {
		return 3 // Update
	}
	if strings.Contains(eventToMapLower, "delete") || strings.Contains(eventToMapLower, "remove") || strings.Contains(eventToMapLower, "trash") {
		return 4 // Delete
	}
	if strings.Contains(eventToMapLower, "download") || strings.Contains(eventToMapLower, "export") || strings.Contains(eventToMapLower, "print") {
		return 7 // Export
	}
	if strings.Contains(eventToMapLower, "upload") || strings.Contains(eventToMapLower, "import") {
		return 6 // Import
	}
	if strings.Contains(eventToMapLower, "share") || strings.Contains(eventToMapLower, "permission") {
		return 8 // Share
	}

	return 99 // Other
}

// mapSeverityIDFromEvent determines severity based on event type
func mapSeverityIDFromEvent(event struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
}) int {
	eventType := event.Type
	eventName := event.Name
	eventAction := event.Action

	// Use name first, then action, then type for mapping
	eventToMap := eventName
	if eventToMap == "" {
		eventToMap = eventAction
	}
	if eventToMap == "" {
		eventToMap = eventType
	}

	eventToMapLower := strings.ToLower(eventToMap)

	// High severity events
	if strings.Contains(eventToMapLower, "security") || strings.Contains(eventToMapLower, "admin") || strings.Contains(eventToMapLower, "suspend") {
		return 4 // High
	}

	// Medium severity events
	if strings.Contains(eventToMapLower, "delete") || strings.Contains(eventToMapLower, "remove") || strings.Contains(eventToMapLower, "password") {
		return 3 // Medium
	}

	// Low severity events
	if strings.Contains(eventToMapLower, "failure") || strings.Contains(eventToMapLower, "denied") || strings.Contains(eventToMapLower, "share") || strings.Contains(eventToMapLower, "error") {
		return 2 // Low
	}

	return 1 // Informational
}

// mapUserTypeIDFromEvent determines if user is admin or regular user
func mapUserTypeIDFromEvent(event struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
}) int {
	eventType := event.Type
	eventName := event.Name
	eventAction := event.Action

	// Check for admin-related patterns
	eventToCheck := strings.ToLower(eventType + " " + eventName + " " + eventAction)

	if strings.Contains(eventToCheck, "admin") || strings.Contains(eventToCheck, "settings") ||
		strings.Contains(eventToCheck, "manage") || strings.Contains(eventToCheck, "configure") {
		return 2 // Admin
	}

	return 1 // Regular user
}

// mapServiceNameFromEvent maps event type to service name
func mapServiceNameFromEvent(event struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
}) string {
	eventType := strings.ToLower(event.Type)
	eventName := strings.ToLower(event.Name)
	eventAction := strings.ToLower(event.Action)

	if strings.Contains(eventType, "login") || strings.Contains(eventName, "login") || strings.Contains(eventAction, "login") {
		return "Google Identity"
	}
	if strings.Contains(eventType, "drive") || strings.Contains(eventName, "drive") || strings.Contains(eventAction, "drive") {
		return "Google Drive API"
	}
	if strings.Contains(eventType, "admin") || strings.Contains(eventName, "admin") || strings.Contains(eventAction, "admin") {
		return "Google Admin API"
	}
	if strings.Contains(eventType, "calendar") || strings.Contains(eventName, "calendar") || strings.Contains(eventAction, "calendar") {
		return "Google Calendar API"
	}

	return "Google Workspace API"
}

// getResponseCode returns HTTP response code based on status
func getResponseCode(statusID int) int {
	if statusID == 1 {
		return 200 // Success
	}
	return 403 // Forbidden
}

// getResponseMessage returns response message based on status
func getResponseMessage(statusID int) string {
	if statusID == 1 {
		return "Success"
	}
	return "Access Denied"
}
