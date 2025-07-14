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

// リソースリストを生成（約500個）
func generateResources() []Resource {
	resources := []Resource{}
	idCounter := 1000

	// 教材カテゴリ（120個）
	subjects := []string{"数学", "英語", "国語", "理科", "社会", "体育", "音楽", "美術", "技術", "家庭科", "情報", "商業"}
	materials := []string{"教科書", "問題集", "解答集", "指導要領", "副教材", "参考書", "ワークブック", "プリント", "動画教材", "テスト問題"}

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

	// 成績・評価データ（84個）
	gradeTypes := []string{"中間テスト", "期末テスト", "小テスト", "宿題", "レポート", "出席", "実技評価"}
	for _, subject := range subjects { // 全12科目
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

	// 学習データ・AI関連（60個）
	aiDataTypes := []string{"学習進捗", "学習パターン", "理解度分析", "推奨コンテンツ", "学習履歴"}
	for _, subject := range subjects { // 全12科目
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

	// 管理用ファイル（70個）
	adminCategories := []map[string][]string{
		{"財務": {"予算計画", "決算書", "監査報告", "経費明細", "給与データ", "税務申告", "補助金申請", "寄付管理", "支払明細", "収支報告"}},
		{"人事": {"職員名簿", "勤怠記録", "評価表", "研修記録", "契約書", "採用資料", "昇進記録", "退職手続", "健康診断", "労働契約"}},
		{"学籍": {"入学者名簿", "卒業者名簿", "転校記録", "出席統計", "進路データ", "奨学金", "保護者連絡", "学籍変更", "休学届", "成績証明"}},
		{"施設": {"設備点検", "修繕記録", "安全点検", "清掃記録", "備品管理", "工事記録", "保険関連", "法定点検", "環境測定", "防災計画"}},
		{"システム": {"ユーザー権限", "バックアップ", "ログ設定", "監査設定", "セキュリティ設定", "ライセンス", "障害記録", "更新履歴", "アクセス制御", "運用手順"}},
		{"教務": {"カリキュラム", "時間割", "教材管理", "試験管理", "行事計画", "授業記録", "評価基準", "進級判定", "補習計画", "教育実習"}},
		{"総務": {"会議録", "規程集", "通達文書", "外部連絡", "統計資料", "年報", "理事会資料", "認可申請", "法人登記", "印鑑管理"}},
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

	// 個人フォルダ（90個）
	for i := 1; i <= 90; i++ {
		var prefix string
		switch {
		case i <= 50:
			prefix = "teacher"
		case i <= 70:
			prefix = "staff"
		case i <= 80:
			prefix = "student"
		default:
			prefix = "external"
		}
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("個人/%s%02d/", prefix, i),
			Type:       "folder",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 共有フォルダ・公開ファイル（80個）
	sharedFolders := []string{
		"共有/お知らせ/", "共有/行事予定/", "共有/委員会資料/", "共有/PTA資料/", "共有/研修資料/",
		"共有/職員会議/", "共有/教科会議/", "共有/学年会議/", "共有/安全管理/", "共有/健康管理/",
		"共有/図書室/", "共有/保健室/", "共有/相談室/", "共有/実験室/", "共有/体育館/",
		"共有/音楽室/", "共有/美術室/", "共有/技術室/", "共有/家庭科室/", "共有/コンピュータ室/",
	}
	
	publicFiles := []string{
		"公開/学校案内.pdf", "公開/入学要項.pdf", "公開/年間行事.pdf", "公開/学校沿革.pdf", "公開/アクセス.pdf",
		"公開/教育方針.pdf", "公開/進路実績.pdf", "公開/部活動紹介.pdf", "公開/制服案内.pdf", "公開/学費案内.pdf",
		"公開/入試要項.pdf", "公開/説明会案内.pdf", "公開/奨学金案内.pdf", "公開/施設紹介.pdf", "公開/教員紹介.pdf",
		"公開/保護者の声.pdf", "公開/卒業生の声.pdf", "公開/Q&A.pdf", "公開/お問い合わせ.pdf", "公開/交通案内.pdf",
	}
	
	internalDocs := []string{
		"内部/規程集.pdf", "内部/職員ハンドブック.pdf", "内部/緊急時対応.pdf", "内部/個人情報保護.pdf", "内部/セキュリティ規程.pdf",
		"内部/服務規程.pdf", "内部/評価基準.pdf", "内部/授業指導要領.pdf", "内部/生活指導要領.pdf", "内部/進路指導要領.pdf",
		"内部/保護者対応.pdf", "内部/事故対応.pdf", "内部/災害対応.pdf", "内部/感染症対応.pdf", "内部/いじめ対応.pdf",
		"内部/特別支援.pdf", "内部/国際交流.pdf", "内部/地域連携.pdf", "内部/広報活動.pdf", "内部/募集活動.pdf",
	}
	
	// 全てを統合
	var sharedItems []map[string]string
	for _, folder := range sharedFolders {
		sharedItems = append(sharedItems, map[string]string{folder: "folder"})
	}
	for _, file := range publicFiles {
		sharedItems = append(sharedItems, map[string]string{file: "document"})
	}
	for _, file := range internalDocs {
		sharedItems = append(sharedItems, map[string]string{file: "document"})
	}

	for _, itemMap := range sharedItems {
		for name, resourceType := range itemMap {
			var visibility string
			if strings.HasPrefix(name, "公開") {
				visibility = "public_on_the_web"
			} else if strings.HasPrefix(name, "内部") {
				visibility = "private"
			} else {
				visibility = "domain"
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

	// 残りを埋めるためのバリエーション追加（500個まで）
	categories := []string{"アーカイブ", "テンプレート", "サンプル", "ドラフト", "バックアップ", "ログ", "統計", "レポート"}
	fileTypes := []string{"pdf", "xlsx", "docx", "pptx", "json", "csv"}
	
	for len(resources) < 500 {
		category := categories[len(resources)%len(categories)]
		fileType := fileTypes[len(resources)%len(fileTypes)]
		docType := "document"
		if fileType == "xlsx" || fileType == "csv" {
			docType = "spreadsheet"
		}
		
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("%s/ファイル%03d.%s", category, len(resources), fileType),
			Type:       docType,
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	return resources
}
