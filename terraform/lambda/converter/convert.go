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

	// Determine activity_id based on event type and name
	activityID := mapActivityID(log.ID.ApplicationName, log.Events)
	
	// Determine severity_id
	severityID := mapSeverityID(log.ID.ApplicationName, log.Events)
	
	// Determine status_id
	statusID := mapStatusID(log.Events)
	
	// Determine user type (admin or regular user)
	userTypeID := mapUserTypeID(log.ID.ApplicationName, log.Events)

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
	ocsf.Actor.User.Name = log.Actor.Email
	ocsf.Actor.User.UID = log.Actor.ProfileID
	ocsf.Actor.User.EmailAddr = log.Actor.Email
	ocsf.Actor.User.Domain = log.OwnerDomain
	ocsf.Actor.User.TypeID = userTypeID
	
	// Session information
	ocsf.Actor.Session.UID = log.ID.UniqueQualifier
	ocsf.Actor.Session.CreatedTime = timestamp.Add(-1 * time.Hour).UnixMilli() // Estimate session start
	
	// App information
	ocsf.Actor.AppName = "Google Workspace"
	ocsf.Actor.AppUID = log.ID.ApplicationName

	// API information
	ocsf.API.Service.Name = mapServiceName(log.ID.ApplicationName)
	ocsf.API.Service.Version = "v3"
	if len(log.Events) > 0 {
		ocsf.API.Operation = log.Events[0].Name
	}
	ocsf.API.Request.UID = log.ID.UniqueQualifier
	ocsf.API.Response.Code = getResponseCode(statusID)
	ocsf.API.Response.Message = getResponseMessage(statusID)

	// Cloud information
	ocsf.Cloud.Provider = "Google Cloud"
	ocsf.Cloud.Account.UID = log.ID.CustomerID
	ocsf.Cloud.Account.Name = strings.Split(log.OwnerDomain, ".")[0]
	ocsf.Cloud.Org.Name = log.OwnerDomain
	ocsf.Cloud.Org.UID = strings.Split(log.OwnerDomain, ".")[0]
	ocsf.Cloud.Region = "asia-northeast1" // Default to Tokyo region

	// Source endpoint
	ocsf.SrcEndpoint.IP = log.IPAddress
	
	// Web resources (if applicable)
	ocsf.WebResources = extractWebResources(log.Events)

	// Metadata
	ocsf.Metadata.OriginalTime = log.ID.Time
	ocsf.Metadata.Processed = time.Now().UnixMilli()
	ocsf.Metadata.ProductName = "Google Workspace"
	ocsf.Metadata.Version = "1.0.0"

	return ocsf, nil
}

// mapActivityID maps Google Workspace events to OCSF activity_id
func mapActivityID(appName string, events []struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Parameters []struct {
		Name          string   `json:"name"`
		Value         string   `json:"value,omitempty"`
		BoolValue     bool     `json:"boolValue,omitempty"`
		IntValue      int64    `json:"intValue,omitempty"`
		MultiStrValue []string `json:"multiStrValue,omitempty"`
		MultiIntValue []int64  `json:"multiIntValue,omitempty"`
	} `json:"parameters,omitempty"`
}) int {
	if len(events) == 0 {
		return 99 // Other
	}

	eventType := events[0].Type
	eventName := events[0].Name

	// Map based on schema.md
	switch appName {
	case "login":
		switch eventName {
		case "login_success", "login_failure", "login_challenge":
			return 2 // Read
		case "logout":
			return 99 // Other
		case "suspicious_login":
			return 2 // Read
		}
	case "drive":
		switch eventName {
		case "create":
			return 1 // Create
		case "view", "preview", "access_denied":
			return 2 // Read
		case "edit", "move", "rename":
			return 3 // Update
		case "trash", "delete":
			return 4 // Delete
		case "download", "print":
			return 7 // Export
		case "upload":
			return 6 // Import
		case "share", "unshare":
			return 8 // Share
		}
	case "admin":
		switch eventType {
		case "USER_SETTINGS":
			switch eventName {
			case "CREATE_USER":
				return 1 // Create
			case "DELETE_USER":
				return 4 // Delete
			default:
				return 3 // Update
			}
		case "GROUP_SETTINGS":
			switch eventName {
			case "CREATE_GROUP":
				return 1 // Create
			case "DELETE_GROUP":
				return 4 // Delete
			default:
				return 3 // Update
			}
		default:
			return 3 // Update for most admin operations
		}
	case "calendar":
		switch eventName {
		case "create_event":
			return 1 // Create
		case "view_event":
			return 2 // Read
		case "edit_event", "invite_respond":
			return 3 // Update
		case "delete_event":
			return 4 // Delete
		case "share_calendar":
			return 8 // Share
		}
	}

	return 99 // Other
}

// mapSeverityID determines severity based on event type
func mapSeverityID(appName string, events []struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Parameters []struct {
		Name          string   `json:"name"`
		Value         string   `json:"value,omitempty"`
		BoolValue     bool     `json:"boolValue,omitempty"`
		IntValue      int64    `json:"intValue,omitempty"`
		MultiStrValue []string `json:"multiStrValue,omitempty"`
		MultiIntValue []int64  `json:"multiIntValue,omitempty"`
	} `json:"parameters,omitempty"`
}) int {
	if len(events) == 0 {
		return 1 // Informational
	}

	eventType := events[0].Type
	eventName := events[0].Name

	// Based on schema.md severity rules
	switch appName {
	case "login":
		switch eventName {
		case "login_failure":
			return 2 // Low
		case "suspicious_login":
			return 3 // Medium
		default:
			return 1 // Informational
		}
	case "drive":
		switch eventName {
		case "share":
			return 2 // Low
		case "delete":
			return 2 // Low
		case "access_denied":
			return 2 // Low
		default:
			return 1 // Informational
		}
	case "admin":
		switch eventType {
		case "USER_SETTINGS":
			switch eventName {
			case "DELETE_USER":
				return 3 // Medium
			case "SUSPEND_USER":
				return 3 // Medium
			case "CHANGE_USER_PASSWORD":
				return 3 // Medium
			default:
				return 2 // Low
			}
		case "SECURITY_SETTINGS":
			return 4 // High
		case "DOMAIN_SETTINGS":
			return 3 // Medium
		default:
			return 2 // Low
		}
	case "calendar":
		if eventName == "share_calendar" {
			return 2 // Low
		}
		return 1 // Informational
	}

	return 1 // Informational
}

// mapStatusID determines if the operation succeeded or failed
func mapStatusID(events []struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Parameters []struct {
		Name          string   `json:"name"`
		Value         string   `json:"value,omitempty"`
		BoolValue     bool     `json:"boolValue,omitempty"`
		IntValue      int64    `json:"intValue,omitempty"`
		MultiStrValue []string `json:"multiStrValue,omitempty"`
		MultiIntValue []int64  `json:"multiIntValue,omitempty"`
	} `json:"parameters,omitempty"`
}) int {
	if len(events) == 0 {
		return 1 // Success
	}

	eventName := events[0].Name
	
	// Check for failure indicators
	if strings.Contains(eventName, "failure") || 
	   strings.Contains(eventName, "denied") || 
	   strings.Contains(eventName, "error") {
		return 2 // Failure
	}

	return 1 // Success
}

// mapUserTypeID determines if user is admin or regular user
func mapUserTypeID(appName string, events []struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Parameters []struct {
		Name          string   `json:"name"`
		Value         string   `json:"value,omitempty"`
		BoolValue     bool     `json:"boolValue,omitempty"`
		IntValue      int64    `json:"intValue,omitempty"`
		MultiStrValue []string `json:"multiStrValue,omitempty"`
		MultiIntValue []int64  `json:"multiIntValue,omitempty"`
	} `json:"parameters,omitempty"`
}) int {
	if appName == "admin" {
		return 2 // Admin
	}

	// Check for admin-related event types
	if len(events) > 0 {
		eventType := events[0].Type
		if strings.Contains(eventType, "_SETTINGS") {
			return 2 // Admin
		}
	}

	return 1 // Regular user
}

// mapServiceName maps application name to service name
func mapServiceName(appName string) string {
	switch appName {
	case "login":
		return "Google Identity"
	case "drive":
		return "Google Drive API"
	case "admin":
		return "Google Admin API"
	case "calendar":
		return "Google Calendar API"
	default:
		return "Google Workspace API"
	}
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

// extractWebResources extracts file/resource information from events
func extractWebResources(events []struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Parameters []struct {
		Name          string   `json:"name"`
		Value         string   `json:"value,omitempty"`
		BoolValue     bool     `json:"boolValue,omitempty"`
		IntValue      int64    `json:"intValue,omitempty"`
		MultiStrValue []string `json:"multiStrValue,omitempty"`
		MultiIntValue []int64  `json:"multiIntValue,omitempty"`
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

	for _, event := range events {
		var docID, docTitle, docType, visibility string
		
		for _, param := range event.Parameters {
			switch param.Name {
			case "doc_id":
				docID = param.Value
			case "doc_title":
				docTitle = param.Value
			case "doc_type":
				docType = param.Value
			case "visibility":
				visibility = param.Value
			}
		}

		if docID != "" || docTitle != "" {
			resource := struct {
				Name      string `parquet:"name,optional"`
				UID       string `parquet:"uid,optional"`
				Type      string `parquet:"type,optional"`
				URLString string `parquet:"url_string,optional"`
				Data      struct {
					Classification string `parquet:"classification,optional"`
				} `parquet:"data,optional"`
			}{
				Name: docTitle,
				UID:  docID,
				Type: docType,
			}

			// Build URL based on doc type
			if docID != "" && docType != "" {
				switch docType {
				case "spreadsheet":
					resource.URLString = fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", docID)
				case "document":
					resource.URLString = fmt.Sprintf("https://docs.google.com/document/d/%s", docID)
				case "presentation":
					resource.URLString = fmt.Sprintf("https://docs.google.com/presentation/d/%s", docID)
				default:
					resource.URLString = fmt.Sprintf("https://drive.google.com/file/d/%s", docID)
				}
			}

			// Set classification based on visibility
			if visibility != "" {
				switch visibility {
				case "private":
					resource.Data.Classification = "confidential"
				case "people_with_link":
					resource.Data.Classification = "internal"
				case "public":
					resource.Data.Classification = "public"
				default:
					resource.Data.Classification = "internal"
				}
			}

			resources = append(resources, resource)
		}
	}

	return resources
}