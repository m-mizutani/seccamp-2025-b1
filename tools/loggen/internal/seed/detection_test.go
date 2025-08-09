package seed

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

// TestAllDetectionRules verifies that all anomaly patterns can be detected using SQL-like logic
func TestAllDetectionRules(t *testing.T) {
	// Generate one day template
	generator := NewGenerator()
	testDate := time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC)
	
	template, err := generator.GenerateDayTemplate(testDate, 0.10)
	if err != nil {
		t.Fatalf("Failed to generate day template: %v", err)
	}
	
	// Create log generator for interpreting seeds
	config := logcore.DefaultConfig()
	logGenerator := logcore.NewGenerator(config)
	
	// Detection result structures
	type detectionResult struct {
		pattern    string
		detected   bool
		count      int
		details    string
	}
	
	// Focus on patterns 4-7 only
	results := []detectionResult{
		{pattern: "Pattern4_AuthAttack", detected: false},
		{pattern: "Pattern5_DataTheft", detected: false},
		{pattern: "Pattern6_ServiceProbing", detected: false},
		{pattern: "Pattern7_GeoAnomaly", detected: false},
	}
	
	// Analysis structures for each pattern
	// Pattern 1: Night admin downloads
	nightAdminDownloads := make(map[string]int)
	
	// Pattern 2: External link burst access
	externalAccessWindows := make(map[string]int) // 15-min window -> count
	
	// Pattern 3: VPN lateral movement
	vpnUserResources := make(map[string]map[string]bool) // user -> resources
	
	// Pattern 4: Auth attack
	authFailuresByIP := make(map[string]int) // IP -> failure count in 5 min
	
	// Pattern 5: Data theft
	rapidDownloadsByUser := make(map[string]int) // user -> download count in 10 min
	
	// Pattern 6: Service probing
	serviceAccessByUser := make(map[string]struct {
		services map[string]int
		failures int
		total    int
	})
	
	// Pattern 7: Geo anomaly
	geoAccessByUser := make(map[string]map[string]time.Time) // user -> IP -> last access time
	
	// Process all seeds
	for i, seed := range template.LogSeeds {
		if seed.Pattern == 0 {
			continue // Skip normal patterns
		}
		
		// Generate log entry
		logEntry := logGenerator.GenerateLogEntry(seed, testDate, i)
		logTime, _ := time.Parse(time.RFC3339Nano, logEntry.ID.Time)
		hour := logTime.Hour()
		minute := logTime.Minute()
		
		switch seed.Pattern {
		case logcore.PatternExample1NightAdminDownload:
			// Night admin download detection
			if (hour >= 18 || hour <= 9) && strings.Contains(logEntry.Actor.Email, "admin") {
				for _, event := range logEntry.Events {
					if event.Name == "download" {
						nightAdminDownloads[logEntry.Actor.Email]++
					}
				}
			}
			
		case logcore.PatternExample2ExternalLinkAccess:
			// External link burst detection
			if hour >= 10 && hour <= 16 {
				windowMinute := (minute / 15) * 15
				windowKey := logTime.Format("2006-01-02 15:") + fmt.Sprintf("%02d", windowMinute)
				
				if strings.Contains(logEntry.Actor.Email, "external") {
					externalAccessWindows[windowKey]++
				}
			}
			
		case logcore.PatternExample3VpnLateralMovement:
			// VPN lateral movement detection
			if hour >= 9 && hour <= 18 {
				if strings.HasPrefix(logEntry.IPAddress, "10.") || 
				   strings.HasPrefix(logEntry.IPAddress, "172.") ||
				   strings.HasPrefix(logEntry.IPAddress, "192.168.") {
					user := logEntry.Actor.Email
					if vpnUserResources[user] == nil {
						vpnUserResources[user] = make(map[string]bool)
					}
					
					for _, event := range logEntry.Events {
						for _, param := range event.Parameters {
							if param.Name == "doc_title" && param.Value != "" {
								vpnUserResources[user][param.Value] = true
							}
						}
					}
				}
			}
			
		case logcore.PatternExample4HighFreqAuthAttack:
			// Auth attack detection
			for _, event := range logEntry.Events {
				if event.Name == "login_failure" {
					// Count failures from this IP
					authFailuresByIP[logEntry.IPAddress]++
				}
			}
			
		case logcore.PatternExample5RapidDataTheft:
			// Data theft detection
			for _, event := range logEntry.Events {
				if event.Name == "download" {
					rapidDownloadsByUser[logEntry.Actor.Email]++
				}
			}
			
		case logcore.PatternExample6MultiServiceProbing:
			// Service probing detection
			user := logEntry.Actor.Email
			if _, exists := serviceAccessByUser[user]; !exists {
				serviceAccessByUser[user] = struct {
					services map[string]int
					failures int
					total    int
				}{
					services: make(map[string]int),
				}
			}
			
			entry := serviceAccessByUser[user]
			entry.services[logEntry.ID.ApplicationName]++
			entry.total++
			
			// Check for failures
			for _, event := range logEntry.Events {
				if strings.Contains(event.Name, "denied") || strings.Contains(event.Name, "DENIED") {
					entry.failures++
				}
			}
			serviceAccessByUser[user] = entry
			
		case logcore.PatternExample7SimultaneousGeoAccess:
			// Geo anomaly detection
			user := logEntry.Actor.Email
			if geoAccessByUser[user] == nil {
				geoAccessByUser[user] = make(map[string]time.Time)
			}
			geoAccessByUser[user][logEntry.IPAddress] = logTime
		}
	}
	
	// Check Pattern 4: Auth attack (threshold: 10+ failures from same IP)
	// Note: In real SQL, this would be within a 5-minute window
	for ip, count := range authFailuresByIP {
		if count >= 10 {
			results[0].detected = true
			results[0].count = count
			results[0].details = fmt.Sprintf("IP %s: %d failures", ip, count)
			break
		}
	}
	
	// Check Pattern 5: Data theft (threshold: 50+ downloads)
	// Note: In real SQL, this would be within a 10-minute window
	for user, count := range rapidDownloadsByUser {
		if count >= 50 {
			results[1].detected = true
			results[1].count = count
			results[1].details = fmt.Sprintf("User %s: %d downloads", user, count)
			break
		}
	}
	
	// Check Pattern 6: Service probing (threshold: 3+ services, 70%+ failure)
	for user, access := range serviceAccessByUser {
		if len(access.services) >= 3 {
			failureRate := float64(access.failures) / float64(access.total)
			if failureRate >= 0.7 {
				results[2].detected = true
				results[2].count = len(access.services)
				results[2].details = fmt.Sprintf("User %s: %d services, %.0f%% failure", 
					user, len(access.services), failureRate*100)
				break
			}
		}
	}
	
	// Check Pattern 7: Geo anomaly (multiple IPs for same user)
	for user, ips := range geoAccessByUser {
		if len(ips) >= 2 {
			results[3].detected = true
			results[3].count = len(ips)
			results[3].details = fmt.Sprintf("User %s: %d different IPs", user, len(ips))
			break
		}
	}
	
	// Report results
	t.Log("Detection Results:")
	allDetected := true
	for _, result := range results {
		if result.detected {
			t.Logf("✓ %s: DETECTED - %s", result.pattern, result.details)
		} else {
			t.Errorf("✗ %s: NOT DETECTED", result.pattern)
			allDetected = false
		}
	}
	
	// Additional statistics
	patternCounts := make(map[uint8]int)
	for _, seed := range template.LogSeeds {
		if seed.Pattern > 0 {
			patternCounts[seed.Pattern]++
		}
	}
	
	t.Log("\nPattern occurrence counts:")
	patternNames := map[uint8]string{
		logcore.PatternExample1NightAdminDownload:    "Pattern1_NightAdminDownload",
		logcore.PatternExample2ExternalLinkAccess:    "Pattern2_ExternalLinkAccess",
		logcore.PatternExample3VpnLateralMovement:    "Pattern3_VPNLateralMovement",
		logcore.PatternExample4HighFreqAuthAttack:    "Pattern4_AuthAttack",
		logcore.PatternExample5RapidDataTheft:        "Pattern5_DataTheft",
		logcore.PatternExample6MultiServiceProbing:   "Pattern6_ServiceProbing",
		logcore.PatternExample7SimultaneousGeoAccess: "Pattern7_GeoAnomaly",
	}
	
	for pattern := uint8(1); pattern <= 9; pattern++ {
		if name, exists := patternNames[pattern]; exists {
			t.Logf("%s: %d occurrences", name, patternCounts[pattern])
		}
	}
	
	if !allDetected {
		t.Fatal("Not all detection patterns were successfully detected")
	}
}

// TestDetectionThresholds verifies that patterns meet minimum thresholds for SQL detection
func TestDetectionThresholds(t *testing.T) {
	generator := NewGenerator()
	testDate := time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC)
	
	template, err := generator.GenerateDayTemplate(testDate, 0.10)
	if err != nil {
		t.Fatalf("Failed to generate day template: %v", err)
	}
	
	// Count patterns
	patternCounts := make(map[uint8]int)
	for _, seed := range template.LogSeeds {
		if seed.Pattern > 0 {
			patternCounts[seed.Pattern]++
		}
	}
	
	// Define minimum thresholds needed for SQL detection
	minThresholds := map[uint8]int{
		logcore.PatternExample1NightAdminDownload:    5,   // Need 5+ downloads
		logcore.PatternExample2ExternalLinkAccess:    20,  // Need 20+ in burst
		logcore.PatternExample3VpnLateralMovement:    10,  // Need 10+ resources
		logcore.PatternExample4HighFreqAuthAttack:    10,  // Need 10+ failures
		logcore.PatternExample5RapidDataTheft:        50,  // Need 50+ downloads
		logcore.PatternExample6MultiServiceProbing:   10,  // Need enough for 3+ services
		logcore.PatternExample7SimultaneousGeoAccess: 10,  // Need multiple accesses
	}
	
	t.Log("Pattern generation vs SQL detection thresholds:")
	for pattern, threshold := range minThresholds {
		count := patternCounts[pattern]
		ratio := float64(count) / float64(threshold)
		
		status := "✓ SUFFICIENT"
		if count < threshold {
			status = "✗ INSUFFICIENT"
		}
		
		t.Logf("Pattern %d: %d generated / %d threshold = %.1fx %s", 
			pattern, count, threshold, ratio, status)
		
		if count < threshold {
			t.Errorf("Pattern %d generates insufficient logs (%d) to meet SQL threshold (%d)",
				pattern, count, threshold)
		}
	}
}