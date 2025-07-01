package main

// Alert represents the alert notification format for SNS
type Alert struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Attrs       map[string]interface{} `json:"attrs"`
}

// QueryResult represents a row returned from Athena query
type QueryResult struct {
	Data map[string]string `json:"data"`
}

// QueryDefinition represents a SQL query to execute
type QueryDefinition struct {
	Name        string
	Description string
	SQL         string
}
