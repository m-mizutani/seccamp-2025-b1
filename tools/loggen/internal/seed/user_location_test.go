package seed

import (
	"fmt"
	"testing"
	"time"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

// TestNormalUsersSingleCountryAccess tests that normal users only access from a single country
func TestNormalUsersSingleCountryAccess(t *testing.T) {
	// Generate test data
	generator := NewGenerator()
	targetDate := time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC)
	anomalyRatio := 0.15

	dayTemplate, err := generator.GenerateDayTemplate(targetDate, anomalyRatio)
	if err != nil {
		t.Fatalf("Failed to generate day template: %v", err)
	}

	// Create log generator
	config := logcore.DefaultConfig()
	logGenerator := logcore.NewGenerator(config)

	// Track IP addresses and locations by user
	userLocations := make(map[string]map[string]bool) // user -> country -> seen
	userIPs := make(map[string]map[string]bool)      // user -> IP -> seen

	// Process all log entries
	for i, seed := range dayTemplate.LogSeeds {
		logEntry := logGenerator.GenerateLogEntry(seed, targetDate, i)

		// Skip anomaly patterns (except Pattern 7 which we want to validate)
		if seed.Pattern != logcore.PatternNormal && seed.Pattern != logcore.PatternExample7SimultaneousGeoAccess {
			continue
		}

		user := logEntry.Actor.Email
		ip := logEntry.IPAddress

		// Initialize maps if needed
		if userIPs[user] == nil {
			userIPs[user] = make(map[string]bool)
			userLocations[user] = make(map[string]bool)
		}

		// Track IP
		userIPs[user][ip] = true

		// Determine country from IP
		country := getCountryFromIP(ip)
		userLocations[user][country] = true
	}

	// Verify requirements for each user
	for user, ips := range userIPs {
		countries := userLocations[user]
		ipCount := len(ips)
		countryCount := len(countries)

		// Special case: yamada.takeshi should have multiple countries (Pattern 7)
		if user == "yamada.takeshi@muhaijuku.com" {
			if countryCount < 2 {
				t.Errorf("User %s (Pattern 7) should have multiple countries, but has %d", user, countryCount)
			}
			continue
		}

		// Normal users should only have one country
		if countryCount > 1 {
			t.Errorf("Normal user %s has %d countries (expected 1): %v", user, countryCount, getKeys(countries))
			fmt.Printf("  IPs for %s: %v\n", user, getKeys(ips))
		}

		// Users should have 1-3 IPs
		if ipCount < 1 || ipCount > 3 {
			t.Errorf("User %s has %d IPs (expected 1-3): %v", user, ipCount, getKeys(ips))
		}
	}
}

// TestIPAddressVariation tests that non-office IPs have sufficient variation
func TestIPAddressVariation(t *testing.T) {
	generator := NewGenerator()
	targetDate := time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC)
	
	dayTemplate, err := generator.GenerateDayTemplate(targetDate, 0.0) // No anomalies
	if err != nil {
		t.Fatalf("Failed to generate day template: %v", err)
	}

	config := logcore.DefaultConfig()
	logGenerator := logcore.NewGenerator(config)

	// Track unique IP prefixes (first 3 octets)
	ipPrefixes := make(map[string]bool)
	officeIPs := 0
	totalIPs := 0

	for i, seed := range dayTemplate.LogSeeds {
		if i > 10000 { // Sample first 10000 logs
			break
		}

		logEntry := logGenerator.GenerateLogEntry(seed, targetDate, i)
		ip := logEntry.IPAddress

		totalIPs++

		// Check if office IP
		if len(ip) > 10 && ip[:10] == "210.160.34" {
			officeIPs++
		} else {
			// Extract prefix (first 3 octets)
			lastDot := 0
			dots := 0
			for j, c := range ip {
				if c == '.' {
					dots++
					if dots == 3 {
						lastDot = j
						break
					}
				}
			}
			if lastDot > 0 {
				prefix := ip[:lastDot]
				ipPrefixes[prefix] = true
			}
		}
	}

	// Verify sufficient variation in non-office IPs
	uniquePrefixes := len(ipPrefixes)
	if uniquePrefixes < 10 {
		t.Errorf("Insufficient IP prefix variation: only %d unique prefixes found", uniquePrefixes)
	}

	officeRatio := float64(officeIPs) / float64(totalIPs)
	if officeRatio < 0.1 || officeRatio > 0.5 {
		t.Errorf("Office IP ratio is %f (expected between 0.1 and 0.5)", officeRatio)
	}
}

// Helper function to determine country from IP based on converter logic
func getCountryFromIP(ip string) string {
	// Based on mapLocationFromIP in converter/convert.go
	if len(ip) < 7 {
		return "JP"
	}

	prefix := ip[:7]
	
	// US IPs
	if prefix == "198.51." {
		return "US"
	}
	
	// Japanese attack IPs (changed from Russian to avoid geo confusion)
	if len(ip) >= 10 && ip[:10] == "133.200.32" {
		return "JP"
	}
	
	// External partner IPs are all from Japan now
	if len(ip) >= 10 && ip[:10] == "203.0.113." {
		return "JP"
	}
	
	// Default to JP (all other IPs are Japanese)
	return "JP"
}

// Helper function to get keys from a map
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to check if user is external
func isExternalUser(email string) bool {
	return email == "external.partner@example.com" || email == "guest.user@external.org"
}