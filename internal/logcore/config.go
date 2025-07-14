package logcore

import (
	"fmt"
	"strings"
	"time"
)

// デフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		CustomerID:  "C03az79cb",
		OwnerDomain: "muhai-academy.com",
		BaseDate:    time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
		Users:       generateUsers(),
		Resources:   generateResources(),
	}
}

// ユーザーリストを生成（約90人）
func generateUsers() []User {
	users := []User{
		// 管理者・マネージャー
		{Email: "admin@muhai-academy.com", ProfileID: "114511147312345678901", Role: "admin"},
		{Email: "manager@muhai-academy.com", ProfileID: "114511147312345678902", Role: "manager"},
		{Email: "vice.manager@muhai-academy.com", ProfileID: "114511147312345678903", Role: "manager"},
		{Email: "it.admin@muhai-academy.com", ProfileID: "114511147312345678904", Role: "admin"},
		{Email: "security.admin@muhai-academy.com", ProfileID: "114511147312345678905", Role: "admin"},

		// 教師（50人）
		{Email: "teacher.math01@muhai-academy.com", ProfileID: "114511147312345678906", Role: "teacher"},
		{Email: "teacher.math02@muhai-academy.com", ProfileID: "114511147312345678907", Role: "teacher"},
		{Email: "teacher.math03@muhai-academy.com", ProfileID: "114511147312345678908", Role: "teacher"},
		{Email: "teacher.english01@muhai-academy.com", ProfileID: "114511147312345678909", Role: "teacher"},
		{Email: "teacher.english02@muhai-academy.com", ProfileID: "114511147312345678910", Role: "teacher"},
		{Email: "teacher.english03@muhai-academy.com", ProfileID: "114511147312345678911", Role: "teacher"},
		{Email: "teacher.science01@muhai-academy.com", ProfileID: "114511147312345678912", Role: "teacher"},
		{Email: "teacher.science02@muhai-academy.com", ProfileID: "114511147312345678913", Role: "teacher"},
		{Email: "teacher.history01@muhai-academy.com", ProfileID: "114511147312345678914", Role: "teacher"},
		{Email: "teacher.history02@muhai-academy.com", ProfileID: "114511147312345678915", Role: "teacher"},
		{Email: "teacher.japanese01@muhai-academy.com", ProfileID: "114511147312345678916", Role: "teacher"},
		{Email: "teacher.japanese02@muhai-academy.com", ProfileID: "114511147312345678917", Role: "teacher"},
		{Email: "teacher.pe01@muhai-academy.com", ProfileID: "114511147312345678918", Role: "teacher"},
		{Email: "teacher.pe02@muhai-academy.com", ProfileID: "114511147312345678919", Role: "teacher"},
		{Email: "teacher.art01@muhai-academy.com", ProfileID: "114511147312345678920", Role: "teacher"},
		{Email: "teacher.music01@muhai-academy.com", ProfileID: "114511147312345678921", Role: "teacher"},
		{Email: "teacher.computer01@muhai-academy.com", ProfileID: "114511147312345678922", Role: "teacher"},
		{Email: "teacher.computer02@muhai-academy.com", ProfileID: "114511147312345678923", Role: "teacher"},
		{Email: "teacher.economics01@muhai-academy.com", ProfileID: "114511147312345678924", Role: "teacher"},
		{Email: "teacher.psychology01@muhai-academy.com", ProfileID: "114511147312345678925", Role: "teacher"},
	}

	// 残りの教師を自動生成
	for i := 21; i <= 50; i++ {
		users = append(users, User{
			Email:     fmt.Sprintf("teacher%02d@muhai-academy.com", i),
			ProfileID: fmt.Sprintf("1145111473123456789%02d", 25+i),
			Role:      "teacher",
		})
	}

	// スタッフ（20人）
	staffRoles := []string{"staff", "counselor", "librarian", "nurse", "security", "maintenance"}
	for i := 1; i <= 20; i++ {
		role := staffRoles[(i-1)%len(staffRoles)]
		users = append(users, User{
			Email:     fmt.Sprintf("staff%02d@muhai-academy.com", i),
			ProfileID: fmt.Sprintf("11451114731234567%03d", 800+i),
			Role:      role,
		})
	}

	// 学生（10人）
	for i := 1; i <= 10; i++ {
		users = append(users, User{
			Email:     fmt.Sprintf("student%02d@muhai-academy.com", i),
			ProfileID: fmt.Sprintf("11451114731234567%03d", 900+i),
			Role:      "student",
		})
	}

	// 外部ユーザー（5人）
	externalUsers := []User{
		{Email: "external.instructor@partner-company.com", ProfileID: "114511147312345671001", Role: "external"},
		{Email: "consultant@consulting-firm.jp", ProfileID: "114511147312345671002", Role: "external"},
		{Email: "auditor@audit-firm.com", ProfileID: "114511147312345671003", Role: "external"},
		{Email: "contractor@tech-company.com", ProfileID: "114511147312345671004", Role: "external"},
		{Email: "guest.lecturer@university.ac.jp", ProfileID: "114511147312345671005", Role: "external"},
	}
	users = append(users, externalUsers...)

	return users
}

// リソースリストを生成（約200個）
func generateResources() []Resource {
	resources := []Resource{}
	idCounter := 1000

	// 教材カテゴリ（40個）
	subjects := []string{"数学", "英語", "国語", "理科", "社会", "体育", "音楽", "美術", "技術", "家庭科"}
	materials := []string{"教科書", "問題集", "解答集", "指導要領"}

	for _, subject := range subjects {
		for _, material := range materials {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("教材/%s/%s.pdf", subject, material),
				Type:       "document",
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "domain",
			})
			idCounter++
		}
	}

	// 成績・評価データ（30個）
	gradeTypes := []string{"中間テスト", "期末テスト", "小テスト", "宿題", "レポート", "出席"}
	for _, subject := range subjects[:5] { // 主要5科目
		for _, gradeType := range gradeTypes {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("成績/%s/%s結果.xlsx", subject, gradeType),
				Type:       "spreadsheet",
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "private",
			})
			idCounter++
		}
	}

	// 学習データ・AI関連（20個）
	aiDataTypes := []string{"学習進捗", "学習パターン", "理解度分析", "推奨コンテンツ"}
	for _, subject := range subjects[:5] {
		for _, dataType := range aiDataTypes {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("学習データ/%s/%s.json", subject, dataType),
				Type:       "file",
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "private",
			})
			idCounter++
		}
	}

	// 管理用ファイル（25個）
	adminCategories := []map[string][]string{
		{"財務": {"予算計画", "決算書", "監査報告", "経費明細", "給与データ"}},
		{"人事": {"職員名簿", "勤怠記録", "評価表", "研修記録", "契約書"}},
		{"学籍": {"入学者名簿", "卒業者名簿", "転校記録", "出席統計", "進路データ"}},
		{"施設": {"設備点検", "修繕記録", "安全点検", "清掃記録", "備品管理"}},
		{"システム": {"ユーザー権限", "バックアップ", "ログ設定", "監査設定", "セキュリティ設定"}},
	}

	for _, categoryMap := range adminCategories {
		for category, items := range categoryMap {
			for _, item := range items {
				resources = append(resources, Resource{
					Name:       fmt.Sprintf("管理/%s/%s.xlsx", category, item),
					Type:       "spreadsheet",
					ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
					Visibility: "private",
				})
				idCounter++
			}
		}
	}

	// 個人フォルダ（50個）
	for i := 1; i <= 50; i++ {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("個人/teacher%02d/", i),
			Type:       "folder",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 共有フォルダ・公開ファイル（35個）
	sharedItems := []map[string]string{
		{"共有/お知らせ/": "folder"},
		{"共有/行事予定/": "folder"},
		{"共有/委員会資料/": "folder"},
		{"共有/PTA資料/": "folder"},
		{"共有/研修資料/": "folder"},
		{"公開/学校案内.pdf": "document"},
		{"公開/入学要項.pdf": "document"},
		{"公開/年間行事.pdf": "document"},
		{"公開/学校沿革.pdf": "document"},
		{"公開/アクセス.pdf": "document"},
	}

	for _, itemMap := range sharedItems {
		for name, resourceType := range itemMap {
			visibility := "domain"
			if strings.HasPrefix(name, "公開") {
				visibility = "public_on_the_web"
			}

			resources = append(resources, Resource{
				Name:       name,
				Type:       resourceType,
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: visibility,
			})
			idCounter++
		}
	}

	// 残りを埋めるためのバリエーション追加
	for len(resources) < 200 {
		category := []string{"その他", "アーカイブ", "テンプレート", "サンプル"}[len(resources)%4]
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("%s/ファイル%03d.pdf", category, len(resources)),
			Type:       "document",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	return resources
}
