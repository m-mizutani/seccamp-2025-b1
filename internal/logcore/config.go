package logcore

import "time"

// デフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		CustomerID:  "C03az79cb",
		OwnerDomain: "muhai-academy.com",
		BaseDate:    time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
		Users: []User{
			{Email: "admin@muhai-academy.com", ProfileID: "114511147312345678901", Role: "admin"},
			{Email: "manager@muhai-academy.com", ProfileID: "114511147312345678902", Role: "manager"},
			{Email: "teacher1@muhai-academy.com", ProfileID: "114511147312345678903", Role: "teacher"},
			{Email: "teacher2@muhai-academy.com", ProfileID: "114511147312345678904", Role: "teacher"},
			{Email: "teacher3@muhai-academy.com", ProfileID: "114511147312345678905", Role: "teacher"},
			{Email: "staff1@muhai-academy.com", ProfileID: "114511147312345678906", Role: "staff"},
			{Email: "staff2@muhai-academy.com", ProfileID: "114511147312345678907", Role: "staff"},
			{Email: "external.instructor@partner-company.com", ProfileID: "114511147312345678908", Role: "external"},
			{Email: "consultant@consulting-firm.jp", ProfileID: "114511147312345678909", Role: "external"},
		},
		Resources: []Resource{
			{Name: "教材/数学/教科書.pdf", Type: "document", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", Visibility: "private"},
			{Name: "教材/英語/問題集.pdf", Type: "document", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upmt", Visibility: "private"},
			{Name: "成績/数学テスト結果.xlsx", Type: "spreadsheet", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upmu", Visibility: "private"},
			{Name: "成績/英語テスト結果.xlsx", Type: "spreadsheet", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upmv", Visibility: "private"},
			{Name: "学習データ/学生レポート.zip", Type: "file", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upmw", Visibility: "private"},
			{Name: "管理/財務報告書/", Type: "folder", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upmx", Visibility: "private"},
			{Name: "管理/ユーザー権限設定", Type: "settings", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upmy", Visibility: "private"},
			{Name: "個人/教師フォルダ/", Type: "folder", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upmz", Visibility: "private"},
			{Name: "共有/お知らせ/", Type: "folder", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upn0", Visibility: "domain"},
			{Name: "公開/学校案内.pdf", Type: "document", ID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upn1", Visibility: "public_on_the_web"},
		},
	}
}