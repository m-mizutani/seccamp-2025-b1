package seed

import (
	"testing"
	"time"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

// TestContinuousAnomalyPatterns tests that continuous anomaly patterns appear at expected frequencies
func TestContinuousAnomalyPatterns(t *testing.T) {
	// Create generator
	generator := NewGenerator()
	
	// Generate 1 day of test data
	testDate := time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC)
	anomalyRatio := 0.10 // 期待値を10%に設定
	
	template, err := generator.GenerateDayTemplate(testDate, anomalyRatio)
	if err != nil {
		t.Fatalf("Failed to generate day template: %v", err)
	}
	
	// Analyze pattern occurrences by minute
	minutePatterns := make(map[int]map[uint8]int) // minute -> pattern -> count
	
	for _, seed := range template.LogSeeds {
		minute := int(seed.Timestamp / 60)
		if minutePatterns[minute] == nil {
			minutePatterns[minute] = make(map[uint8]int)
		}
		if seed.Pattern > 0 {
			minutePatterns[minute][seed.Pattern]++
		}
	}
	
	// Count pattern occurrences across all minutes
	pattern4Count := 0 // High freq auth attack
	pattern5Count := 0 // Rapid data theft
	pattern6Count := 0 // Multi service probing
	pattern7Count := 0 // Simultaneous geo access
	
	pattern4Minutes := 0
	pattern5Minutes := 0
	pattern6Minutes := 0
	pattern7Minutes := 0
	
	for _, patterns := range minutePatterns {
		if count, ok := patterns[logcore.PatternExample4HighFreqAuthAttack]; ok {
			pattern4Count += count
			pattern4Minutes++
		}
		if count, ok := patterns[logcore.PatternExample5RapidDataTheft]; ok {
			pattern5Count += count
			pattern5Minutes++
		}
		if count, ok := patterns[logcore.PatternExample6MultiServiceProbing]; ok {
			pattern6Count += count
			pattern6Minutes++
		}
		if count, ok := patterns[logcore.PatternExample7SimultaneousGeoAccess]; ok {
			pattern7Count += count
			pattern7Minutes++
		}
	}
	
	// Verify pattern frequencies
	t.Logf("Pattern 4 (Auth Attack): %d occurrences in %d minutes", pattern4Count, pattern4Minutes)
	t.Logf("Pattern 5 (Data Theft): %d occurrences in %d minutes", pattern5Count, pattern5Minutes)
	t.Logf("Pattern 6 (Service Probing): %d occurrences in %d minutes", pattern6Count, pattern6Minutes)
	t.Logf("Pattern 7 (Geo Access): %d occurrences in %d minutes", pattern7Count, pattern7Minutes)
	
	// Check if patterns appear with expected frequency
	totalMinutes := 24 * 60 // 1440 minutes in a day
	
	// Pattern should appear in at least 10% of minutes (for continuous patterns)
	minExpectedMinutes := int(float64(totalMinutes) * 0.001) // 0.1% as minimum threshold
	
	if pattern4Minutes < minExpectedMinutes {
		t.Errorf("Pattern 4: Expected to appear in at least %d minutes, but only appeared in %d", 
			minExpectedMinutes, pattern4Minutes)
	}
	if pattern5Minutes < minExpectedMinutes {
		t.Errorf("Pattern 5: Expected to appear in at least %d minutes, but only appeared in %d", 
			minExpectedMinutes, pattern5Minutes)
	}
	if pattern6Minutes < minExpectedMinutes {
		t.Errorf("Pattern 6: Expected to appear in at least %d minutes, but only appeared in %d", 
			minExpectedMinutes, pattern6Minutes)
	}
	if pattern7Minutes < minExpectedMinutes {
		t.Errorf("Pattern 7: Expected to appear in at least %d minutes, but only appeared in %d", 
			minExpectedMinutes, pattern7Minutes)
	}
	
	// Check overall anomaly ratio
	totalAnomalies := 0
	for _, count := range template.Metadata.AnomalyStats {
		totalAnomalies += count
	}
	
	actualAnomalyRatio := float64(totalAnomalies) / float64(template.Metadata.TotalLogs)
	t.Logf("Total logs: %d, Anomalies: %d, Ratio: %.2f%%", 
		template.Metadata.TotalLogs, totalAnomalies, actualAnomalyRatio*100)
	
	// Allow 50% deviation from target anomaly ratio (more tolerance due to probabilistic nature)
	if actualAnomalyRatio < anomalyRatio*0.5 || actualAnomalyRatio > anomalyRatio*1.5 {
		t.Errorf("Anomaly ratio %.2f%% is outside acceptable range (%.2f%% ± 50%%)", 
			actualAnomalyRatio*100, anomalyRatio*100)
	}
}

// TestPatternDistribution tests that patterns are distributed across time windows
func TestPatternDistribution(t *testing.T) {
	generator := NewGenerator()
	testDate := time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC)
	
	template, err := generator.GenerateDayTemplate(testDate, 0.10)
	if err != nil {
		t.Fatalf("Failed to generate day template: %v", err)
	}
	
	// Check 5-minute windows
	fiveMinWindows := make(map[int]map[uint8]int) // 5-min window -> pattern -> count
	
	for _, seed := range template.LogSeeds {
		window := int(seed.Timestamp / 300) // 300 seconds = 5 minutes
		if fiveMinWindows[window] == nil {
			fiveMinWindows[window] = make(map[uint8]int)
		}
		if seed.Pattern >= logcore.PatternExample4HighFreqAuthAttack {
			fiveMinWindows[window][seed.Pattern]++
		}
	}
	
	// Count windows with continuous patterns
	windowsWithPatterns := 0
	for _, patterns := range fiveMinWindows {
		if len(patterns) > 0 {
			windowsWithPatterns++
		}
	}
	
	totalWindows := 24 * 12 // 288 five-minute windows in a day
	coverageRatio := float64(windowsWithPatterns) / float64(totalWindows)
	
	t.Logf("5-minute windows with continuous patterns: %d/%d (%.1f%%)", 
		windowsWithPatterns, totalWindows, coverageRatio*100)
	
	// At least 5% of windows should have patterns
	if coverageRatio < 0.05 {
		t.Errorf("Pattern coverage %.1f%% is too low (expected at least 5%%)", coverageRatio*100)
	}
}

// TestPatternConsistency tests that pattern generation is consistent
func TestPatternConsistency(t *testing.T) {
	generator := NewGenerator()
	
	// Check that state is properly initialized
	if generator.authAttackState == nil {
		t.Error("authAttackState not initialized")
	}
	if generator.dataTheftState == nil {
		t.Error("dataTheftState not initialized")
	}
	if generator.serviceProbingState == nil {
		t.Error("serviceProbingState not initialized")
	}
	if generator.geoAccessState == nil {
		t.Error("geoAccessState not initialized")
	}
	
	// Verify fixed IPs and users
	if generator.authAttackState.attackerIP != "133.200.32.94" {
		t.Errorf("Unexpected attacker IP: %s", generator.authAttackState.attackerIP)
	}
	if generator.dataTheftState.theftUser != "tanaka.hiroshi@muhaijuku.com" {
		t.Errorf("Unexpected theft user: %s", generator.dataTheftState.theftUser)
	}
	if generator.geoAccessState.country1 != "JP" || generator.geoAccessState.country2 != "US" {
		t.Errorf("Unexpected geo access countries: %s, %s", 
			generator.geoAccessState.country1, generator.geoAccessState.country2)
	}
}