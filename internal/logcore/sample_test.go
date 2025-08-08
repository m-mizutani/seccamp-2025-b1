package logcore

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestShowSampleResources(t *testing.T) {
	resources := generateResources()
	
	fmt.Println("\n=== Sample Resources Generated ===")
	fmt.Println(strings.Repeat("=", 100))
	
	// カテゴリ別にサンプルを表示
	categories := map[string][]Resource{
		"教材": []Resource{},
		"事務": []Resource{},
		"経理": []Resource{},
		"人事": []Resource{},
		"プロジェクト": []Resource{},
		"学習者管理": []Resource{},
		"経営戦略": []Resource{},
		"マーケティング": []Resource{},
		"技術開発": []Resource{},
		"法務": []Resource{},
	}
	
	// カテゴリ別に分類
	for _, r := range resources {
		parts := strings.Split(r.Name, "/")
		if len(parts) > 0 {
			category := parts[0]
			if list, exists := categories[category]; exists {
				categories[category] = append(list, r)
			}
		}
	}
	
	// 各カテゴリから最大3個ずつサンプルを表示
	for cat, list := range categories {
		if len(list) > 0 {
			fmt.Printf("\n【%s】\n", cat)
			
			// ランダムに並び替え
			rand.Shuffle(len(list), func(i, j int) {
				list[i], list[j] = list[j], list[i]
			})
			
			count := 3
			if len(list) < count {
				count = len(list)
			}
			
			for i := 0; i < count; i++ {
				r := list[i]
				fmt.Printf("  - %s (Type: %s, Visibility: %s)\n", r.Name, r.Type, r.Visibility)
			}
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 100))
}