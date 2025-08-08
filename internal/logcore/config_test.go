package logcore

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateResources(t *testing.T) {
	resources := generateResources()
	
	// リソース数の確認
	if len(resources) < 700 {
		t.Errorf("Expected at least 700 resources, got %d", len(resources))
	}
	
	fmt.Printf("Total resources generated: %d\n", len(resources))
	
	// カテゴリ別の集計
	categories := make(map[string]int)
	for _, r := range resources {
		parts := strings.Split(r.Name, "/")
		if len(parts) > 0 {
			categories[parts[0]]++
		}
	}
	
	fmt.Println("\nResource categories:")
	for cat, count := range categories {
		fmt.Printf("  %s: %d\n", cat, count)
	}
	
	// 必須カテゴリの確認
	requiredCategories := []string{
		"教材", "事務", "経理", "プロジェクト", "人事", 
		"学習者管理", "経営戦略", "マーケティング", "技術開発", "法務",
		"個人", "共有", "公開",
	}
	
	for _, cat := range requiredCategories {
		if _, exists := categories[cat]; !exists {
			t.Errorf("Required category '%s' not found", cat)
		}
	}
	
	// ファイルタイプの多様性確認
	fileTypes := make(map[string]int)
	for _, r := range resources {
		if strings.Contains(r.Name, ".") {
			parts := strings.Split(r.Name, ".")
			ext := parts[len(parts)-1]
			fileTypes[ext]++
		}
	}
	
	fmt.Println("\nFile types:")
	for ext, count := range fileTypes {
		fmt.Printf("  .%s: %d\n", ext, count)
	}
	
	// 期待されるファイルタイプの確認
	expectedTypes := []string{"pdf", "xlsx", "docx", "pptx", "json", "csv", "mp4", "yaml", "md"}
	for _, ext := range expectedTypes {
		if _, exists := fileTypes[ext]; !exists {
			t.Errorf("Expected file type '.%s' not found", ext)
		}
	}
	
	// 無敗塾特有のコンテンツ確認
	muhaijukuContent := 0
	for _, r := range resources {
		if strings.Contains(r.Name, "無敗") {
			muhaijukuContent++
		}
	}
	
	if muhaijukuContent < 50 {
		t.Errorf("Expected at least 50 Muhaijuku-specific content items, got %d", muhaijukuContent)
	}
	
	fmt.Printf("\nMuhaijuku-specific content: %d items\n", muhaijukuContent)
}

func TestGenerateUsers(t *testing.T) {
	users := generateUsers()
	
	// ユーザー数の確認
	if len(users) < 90 {
		t.Errorf("Expected at least 90 users, got %d", len(users))
	}
	
	fmt.Printf("\nTotal users generated: %d\n", len(users))
	
	// ロール別の集計
	roles := make(map[string]int)
	for _, u := range users {
		roles[u.Role]++
	}
	
	fmt.Println("\nUser roles:")
	for role, count := range roles {
		fmt.Printf("  %s: %d\n", role, count)
	}
	
	// ドメインの確認
	muhaijukuDomain := 0
	for _, u := range users {
		if strings.Contains(u.Email, "@muhaijuku.com") {
			muhaijukuDomain++
		}
	}
	
	if muhaijukuDomain < 85 {
		t.Errorf("Expected at least 85 users with @muhaijuku.com domain, got %d", muhaijukuDomain)
	}
	
	fmt.Printf("\nUsers with @muhaijuku.com domain: %d\n", muhaijukuDomain)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	// ドメインの確認
	if config.OwnerDomain != "muhaijuku.com" {
		t.Errorf("Expected OwnerDomain to be 'muhaijuku.com', got '%s'", config.OwnerDomain)
	}
	
	fmt.Printf("\nDefault config domain: %s\n", config.OwnerDomain)
}