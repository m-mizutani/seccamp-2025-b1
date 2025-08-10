package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Credential represents the structure of a credential JSON file
type Credential struct {
	AccountID    string `json:"account_id"`
	ConsoleURL   string `json:"console_url"`
	Instructions string `json:"instructions"`
	Password     string `json:"password"`
	Username     string `json:"username"`
}

// TeamsData represents the structure of teams.json
type TeamsData struct {
	Teams map[string]string `json:"teams"`
}

func main() {
	// Get the project root directory (two levels up from tools/cred2md)
	projectRoot := filepath.Join("..", "..")
	credentialsDir := filepath.Join(projectRoot, "credentials")
	teamsFile := filepath.Join(projectRoot, "terraform", "teams.json")

	// Check if credentials directory exists
	if _, err := os.Stat(credentialsDir); os.IsNotExist(err) {
		log.Fatalf("Credentials directory not found: %s", credentialsDir)
	}

	// Read teams.json to get GitHub usernames
	var teamsData TeamsData
	teamsDataBytes, err := os.ReadFile(teamsFile)
	if err != nil {
		log.Printf("Warning: Could not read teams.json: %v", err)
	} else {
		if err := json.Unmarshal(teamsDataBytes, &teamsData); err != nil {
			log.Printf("Warning: Could not parse teams.json: %v", err)
		}
	}

	// Read all JSON files from the credentials directory
	files, err := filepath.Glob(filepath.Join(credentialsDir, "*.json"))
	if err != nil {
		log.Fatalf("Error reading credential files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No credential files found in", credentialsDir)
		return
	}

	// Parse all credential files
	var credentials []Credential
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", file, err)
			continue
		}

		var cred Credential
		if err := json.Unmarshal(data, &cred); err != nil {
			log.Printf("Error parsing JSON from %s: %v", file, err)
			continue
		}

		credentials = append(credentials, cred)
	}

	// Sort credentials by username for consistent output
	sort.Slice(credentials, func(i, j int) bool {
		return credentials[i].Username < credentials[j].Username
	})

	// Output Markdown table
	fmt.Println("# AWS Credentials")
	fmt.Println()
	fmt.Println("| IAM User name | GitHub User | Initial Password |")
	fmt.Println("|------|-------------|------------------|")

	for _, cred := range credentials {
		// Escape pipe characters in password if any
		password := strings.ReplaceAll(cred.Password, "|", "\\|")

		// Get GitHub username from teams data
		githubUser := teamsData.Teams[cred.Username]
		if githubUser == "" {
			githubUser = "-"
		}

		fmt.Printf("| %s | %s | `%s` |\n", cred.Username, githubUser, password)
	}

	// Add additional information if needed
	if len(credentials) > 0 && credentials[0].Instructions != "" {
		fmt.Println()
		fmt.Println("**Note:**", credentials[0].Instructions)
	}
	if len(credentials) > 0 && credentials[0].ConsoleURL != "" {
		fmt.Println()
		fmt.Println("**Console URL:**", credentials[0].ConsoleURL)
	}
}
