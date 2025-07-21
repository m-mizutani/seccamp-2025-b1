package main

// GoogleWorkspaceLog represents the input log format from Google Workspace audit logs
type GoogleWorkspaceLog struct {
	Kind string `json:"kind"`
	ID   struct {
		Time             string `json:"time"`
		UniqueQualifier  string `json:"uniqueQualifier"`
		ApplicationName  string `json:"applicationName"`
		CustomerID       string `json:"customerId"`
	} `json:"id"`
	Actor struct {
		CallerType string `json:"callerType"`
		Email      string `json:"email"`
		ProfileID  string `json:"profileId"`
	} `json:"actor"`
	OwnerDomain string `json:"ownerDomain"`
	IPAddress   string `json:"ipAddress"`
	Events      []struct {
		Type       string `json:"type"`
		Name       string `json:"name"`
		Parameters []struct {
			Name           string   `json:"name"`
			Value          string   `json:"value,omitempty"`
			BoolValue      bool     `json:"boolValue,omitempty"`
			IntValue       int64    `json:"intValue,omitempty"`
			MultiStrValue  []string `json:"multiStrValue,omitempty"`
			MultiIntValue  []int64  `json:"multiIntValue,omitempty"`
		} `json:"parameters,omitempty"`
	} `json:"events"`
}

// OCSFWebResourceActivity represents the OCSF Web Resources Activity (Class ID: 6001) format
type OCSFWebResourceActivity struct {
	// Basic classification attributes (required)
	CategoryUID int    `parquet:"category_uid"`     // 6 (Application Activity)
	ClassUID    int    `parquet:"class_uid"`        // 6001 (Web Resources Activity)
	TypeUID     int    `parquet:"type_uid"`         // class_uid * 100 + activity_id
	ActivityID  int    `parquet:"activity_id"`      // 1=Create, 2=Read, 3=Update, 4=Delete, etc.
	SeverityID  int    `parquet:"severity_id"`      // 1=Informational, 2=Low, 3=Medium, 4=High
	Time        int64  `parquet:"time"`             // Unix timestamp in milliseconds
	StartTime   int64  `parquet:"start_time,optional"`
	EndTime     int64  `parquet:"end_time,optional"`
	StatusID    int    `parquet:"status_id"`        // 1=Success, 2=Failure
	Confidence  int    `parquet:"confidence,optional"`

	// Actor information
	Actor struct {
		User struct {
			Name       string   `parquet:"name"`
			UID        string   `parquet:"uid"`
			EmailAddr  string   `parquet:"email_addr"`
			Domain     string   `parquet:"domain,optional"`
			TypeID     int      `parquet:"type_id"`      // 1=User, 2=Admin
			Groups     []string `parquet:"groups,optional"`
		} `parquet:"user"`
		Session struct {
			UID         string `parquet:"uid"`
			CreatedTime int64  `parquet:"created_time,optional"`
			ExpTime     int64  `parquet:"exp_time,optional"`
		} `parquet:"session,optional"`
		AppName string `parquet:"app_name,optional"`
		AppUID  string `parquet:"app_uid,optional"`
	} `parquet:"actor"`

	// API information
	API struct {
		Service struct {
			Name    string `parquet:"name"`
			Version string `parquet:"version,optional"`
		} `parquet:"service"`
		Operation string `parquet:"operation"`
		Request struct {
			UID string `parquet:"uid"`
		} `parquet:"request"`
		Response struct {
			Code    int    `parquet:"code"`
			Message string `parquet:"message,optional"`
		} `parquet:"response,optional"`
	} `parquet:"api"`

	// Cloud environment
	Cloud struct {
		Provider string `parquet:"provider"`
		Account struct {
			UID  string `parquet:"uid"`
			Name string `parquet:"name,optional"`
		} `parquet:"account"`
		Org struct {
			Name string `parquet:"name"`
			UID  string `parquet:"uid,optional"`
		} `parquet:"org,optional"`
		Region string `parquet:"cloud_region,optional"`
	} `parquet:"cloud"`

	// Source endpoint
	SrcEndpoint struct {
		IP       string `parquet:"ip"`
		Hostname string `parquet:"hostname,optional"`
		Location struct {
			Country string `parquet:"country,optional"`
			Region  string `parquet:"src_region,optional"`
			City    string `parquet:"city,optional"`
		} `parquet:"location,optional"`
	} `parquet:"src_endpoint"`

	// Web resources
	WebResources []struct {
		Name      string `parquet:"name,optional"`
		UID       string `parquet:"uid,optional"`
		Type      string `parquet:"type,optional"`
		URLString string `parquet:"url_string,optional"`
		Data      struct {
			Classification string `parquet:"classification,optional"`
		} `parquet:"data,optional"`
	} `parquet:"web_resources,optional"`

	// Metadata
	Metadata struct {
		CorrelationUID string            `parquet:"correlation_uid,optional"`
		Labels         []string          `parquet:"labels,optional,list"`
		OriginalTime   string            `parquet:"original_time,optional"`
		Processed      int64             `parquet:"processed,optional"`
		ProductName    string            `parquet:"product_name,optional"`
		Version        string            `parquet:"version,optional"`
		Extension      map[string]string `parquet:"-"` // Maps are not supported in parquet-go, will need custom handling
	} `parquet:"metadata,optional"`

	// Observables
	Observables []struct {
		Name  string `parquet:"name"`
		Type  string `parquet:"type"`
		Value string `parquet:"value"`
	} `parquet:"observables,optional,list"`

	// Partitioning fields
	Region    string `parquet:"aws_region"`      // AWS region
	AccountID string `parquet:"account_id"` // AWS account ID
	EventHour string `parquet:"event_hour"` // YYYY-MM-DD-HH format
}
