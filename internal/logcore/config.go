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
		OwnerDomain: "muhaijuku.com",
		BaseDate:    time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
		Users:       generateUsers(),
		Resources:   generateResources(),
	}
}

// ユーザーリストを生成（約90人）
func generateUsers() []User {
	users := []User{
		// 経営陣・管理職
		{Email: "yamada.takeshi@muhaijuku.com", ProfileID: "114511147312345678901", Role: "ceo"},
		{Email: "suzuki.keiko@muhaijuku.com", ProfileID: "114511147312345678902", Role: "cto"},
		{Email: "tanaka.hiroshi@muhaijuku.com", ProfileID: "114511147312345678903", Role: "cfo"},
		{Email: "sato.yuki@muhaijuku.com", ProfileID: "114511147312345678904", Role: "vp_product"},
		{Email: "watanabe.jun@muhaijuku.com", ProfileID: "114511147312345678905", Role: "vp_sales"},

		// 講師・インストラクター（50人）
		{Email: "kobayashi.akira@muhaijuku.com", ProfileID: "114511147312345678906", Role: "instructor"},
		{Email: "nakamura.yui@muhaijuku.com", ProfileID: "114511147312345678907", Role: "instructor"},
		{Email: "yamamoto.shinji@muhaijuku.com", ProfileID: "114511147312345678908", Role: "instructor"},
		{Email: "johnson.sarah@muhaijuku.com", ProfileID: "114511147312345678909", Role: "instructor"},
		{Email: "smith.michael@muhaijuku.com", ProfileID: "114511147312345678910", Role: "instructor"},
		{Email: "brown.jennifer@muhaijuku.com", ProfileID: "114511147312345678911", Role: "instructor"},
		{Email: "ito.masahiro@muhaijuku.com", ProfileID: "114511147312345678912", Role: "instructor"},
		{Email: "takahashi.emi@muhaijuku.com", ProfileID: "114511147312345678913", Role: "instructor"},
		{Email: "mori.kenichi@muhaijuku.com", ProfileID: "114511147312345678914", Role: "instructor"},
		{Email: "ishikawa.naomi@muhaijuku.com", ProfileID: "114511147312345678915", Role: "instructor"},
		{Email: "fujita.haruka@muhaijuku.com", ProfileID: "114511147312345678916", Role: "instructor"},
		{Email: "okamoto.satoshi@muhaijuku.com", ProfileID: "114511147312345678917", Role: "instructor"},
		{Email: "kimura.daichi@muhaijuku.com", ProfileID: "114511147312345678918", Role: "instructor"},
		{Email: "hayashi.yuka@muhaijuku.com", ProfileID: "114511147312345678919", Role: "instructor"},
		{Email: "saito.noriko@muhaijuku.com", ProfileID: "114511147312345678920", Role: "instructor"},
		{Email: "kawaguchi.ryo@muhaijuku.com", ProfileID: "114511147312345678921", Role: "instructor"},
		{Email: "nishimura.kenji@muhaijuku.com", ProfileID: "114511147312345678922", Role: "instructor"},
		{Email: "matsuda.ayumi@muhaijuku.com", ProfileID: "114511147312345678923", Role: "instructor"},
		{Email: "honda.takuya@muhaijuku.com", ProfileID: "114511147312345678924", Role: "instructor"},
		{Email: "ogawa.mika@muhaijuku.com", ProfileID: "114511147312345678925", Role: "instructor"},
	}

	// 残りのインストラクターを自動生成
	instructorNames := []string{
		"yoshida.tomohiro", "hasegawa.miyuki", "shimizu.kazuya", "inoue.sayuri", "kato.daisuke",
		"sakamoto.rina", "abe.shinichiro", "nakajima.megumi", "fujimoto.tatsuya", "ohara.kanako",
		"matsumoto.yuichi", "ikeda.mari", "hashimoto.kenta", "yamazaki.chihiro", "kuroda.noboru",
		"ueda.asuka", "morita.ryuji", "takeuchi.yumiko", "kaneko.hideki", "nagai.sachiko",
		"doi.masaki", "fukuda.eriko", "tsuchiya.koichi", "kamata.yukari", "higuchi.tetsuo",
		"miyamoto.nanami", "ozawa.jiro", "maeda.kaori", "sugiyama.akihiko", "nomura.reiko",
	}
	
	for i := 0; i < len(instructorNames); i++ {
		users = append(users, User{
			Email:     fmt.Sprintf("%s@muhaijuku.com", instructorNames[i]),
			ProfileID: fmt.Sprintf("1145111473123456789%02d", 26+i),
			Role:      "instructor",
		})
	}

	// スタッフ（20人）
	staffNames := []string{
		"yoshimura.takako", "aoki.hiroyuki", "yamashita.kumiko", "ishida.masaru", "ogura.chiaki",
		"kondo.yasuhiro", "sasaki.natsumi", "miura.shigeru", "fujikawa.yoko", "okada.kazuhiko",
		"harada.midori", "nishida.toshio", "maruyama.sanae", "imai.ryosuke", "takeda.mihoko",
		"yokoyama.kengo", "matsushita.aya", "ono.takahiro", "tamura.kyoko", "nakano.shingo",
	}
	
	staffRoles := []string{"hr", "finance", "marketing", "sales", "support", "engineer", "designer", "analyst"}
	for i := 0; i < len(staffNames); i++ {
		role := staffRoles[i%len(staffRoles)]
		users = append(users, User{
			Email:     fmt.Sprintf("%s@muhaijuku.com", staffNames[i]),
			ProfileID: fmt.Sprintf("11451114731234567%03d", 801+i),
			Role:      role,
		})
	}

	// 学習者（10人）
	learnerNames := []string{
		"sato.haruto", "suzuki.yua", "takahashi.sota", "tanaka.mio", "ito.ren",
		"watanabe.aoi", "yamamoto.haruki", "nakamura.sakura", "kobayashi.yuto", "kato.himari",
	}
	
	for i := 0; i < len(learnerNames); i++ {
		users = append(users, User{
			Email:     fmt.Sprintf("%s@muhaijuku.com", learnerNames[i]),
			ProfileID: fmt.Sprintf("11451114731234567%03d", 901+i),
			Role:      "learner",
		})
	}

	// 外部ユーザー（5人）
	externalUsers := []User{
		{Email: "taniguchi.keisuke@partner-company.com", ProfileID: "114511147312345671001", Role: "external"},
		{Email: "miyazaki.hiromi@consulting-firm.jp", ProfileID: "114511147312345671002", Role: "external"},
		{Email: "kimura.toshiyuki@audit-firm.com", ProfileID: "114511147312345671003", Role: "external"},
		{Email: "anderson.james@tech-company.com", ProfileID: "114511147312345671004", Role: "external"},
		{Email: "yamaguchi.professor@university.ac.jp", ProfileID: "114511147312345671005", Role: "external"},
	}
	users = append(users, externalUsers...)

	// 攻撃対象専用ユーザー（Pattern 6用）
	attackTargetUsers := []User{
		{Email: "takano.masaki@muhaijuku.com", ProfileID: "114511147312345671006", Role: "compromised"},
		{Email: "ishida.kaori@muhaijuku.com", ProfileID: "114511147312345671007", Role: "compromised"},
	}
	users = append(users, attackTargetUsers...)

	return users
}

// リソースリストを生成（700個以上）
func generateResources() []Resource {
	resources := []Resource{}
	idCounter := 1000

	// 1. 教材・コース関連（200個以上）
	// 無敗塾（基礎コース）
	basicCourses := []string{
		"プログラミング基礎/Python入門", "プログラミング基礎/JavaScript基礎", "プログラミング基礎/Web開発入門",
		"データ分析入門/統計学基礎", "データ分析入門/Excel活用", "データ分析入門/SQLマスター",
		"英語基礎/ビジネス英会話", "英語基礎/TOEIC対策", "英語基礎/技術英語",
		"数学基礎/微積分", "数学基礎/線形代数", "数学基礎/確率統計",
		"ビジネス基礎/経営戦略", "ビジネス基礎/マーケティング", "ビジネス基礎/会計入門",
	}
	
	courseTypes := []string{"01_講義動画.mp4", "02_演習問題.pdf", "03_解答解説.pdf", "04_参考資料.pdf", "05_スライド.pptx", "06_テキスト教材.pdf", "07_確認テスト.docx", "08_修了テスト.pdf", "09_補足資料.pdf"}
	
	for _, course := range basicCourses {
		for _, courseType := range courseTypes {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("教材/無敗塾/%s/%s", course, courseType),
				Type:       getResourceType(courseType),
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "domain",
			})
			idCounter++
		}
	}

	// 無敗ラーニング（法人向け）
	corporateCourses := []string{
		"リーダーシップ研修/基礎編", "リーダーシップ研修/実践編", "リーダーシップ研修/上級編",
		"プロジェクトマネジメント/PMP対策", "プロジェクトマネジメント/アジャイル入門", "プロジェクトマネジメント/リスク管理",
		"ビジネスコミュニケーション/プレゼン技術", "ビジネスコミュニケーション/交渉術", "ビジネスコミュニケーション/文書作成",
		"DX推進/デジタル戦略", "DX推進/業務効率化", "DX推進/データ活用",
		"AI活用/機械学習入門", "AI活用/ディープラーニング", "AI活用/ChatGPT活用術",
	}
	
	for _, course := range corporateCourses {
		for i := 1; i <= 5; i++ {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("教材/無敗ラーニング/%s/第%d回資料.pdf", course, i),
				Type:       "document",
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "domain",
			})
			idCounter++
		}
	}

	// 無敗大学（学位プログラム）
	degreePrograms := []string{
		"経営学修士/1年次春期", "経営学修士/1年次秋期", "経営学修士/2年次春期", "経営学修士/2年次秋期",
		"データサイエンス修士/基礎科目", "データサイエンス修士/専門科目", "データサイエンス修士/研究指導",
		"MBA/コア科目", "MBA/専門科目", "MBA/ケーススタディ",
		"情報工学/プログラミング", "情報工学/アルゴリズム", "情報工学/システム設計",
	}
	
	academicDocs := []string{"シラバス.pdf", "講義ノート.pdf", "課題レポート.docx", "中間試験.pdf", "期末試験.pdf", "研究計画書.docx", "論文ドラフト.docx"}
	
	for _, program := range degreePrograms {
		for _, doc := range academicDocs {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("教材/無敗大学/%s/%s", program, doc),
				Type:       getResourceType(doc),
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "private",
			})
			idCounter++
		}
	}

	// 2. 事務系ファイル（100個）
	contractTypes := []string{
		"法人契約/A社_基本契約書_202401.pdf", "法人契約/B社_サービス利用契約_202402.pdf", "法人契約/C社_年間契約_202403.pdf",
		"講師契約/講師契約書_山田太郎_202401.pdf", "講師契約/講師契約書_鈴木花子_202402.pdf", "講師契約/業務委託契約_田中_202403.pdf",
		"パートナー契約/販売代理店契約_202401.pdf", "パートナー契約/コンテンツ提供契約_202402.pdf", "パートナー契約/技術提携契約_202403.pdf",
	}
	
	for _, contract := range contractTypes {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("事務/契約書/%s", contract),
			Type:       "document",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 請求書・見積書
	for year := 2023; year <= 2024; year++ {
		for month := 1; month <= 12; month++ {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("事務/請求書/%d年%02d月/請求書一覧.xlsx", year, month),
				Type:       "spreadsheet",
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "private",
			})
			idCounter++
			
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("事務/見積書/%d年%02d月/見積書管理表.xlsx", year, month),
				Type:       "spreadsheet",
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "private",
			})
			idCounter++
		}
	}

	// 3. 経理・財務関連（80個）
	financeCategories := []string{
		"決算書/2023年度_決算報告書.pdf", "決算書/2023年度_貸借対照表.xlsx", "決算書/2023年度_損益計算書.xlsx",
		"月次決算/2024年01月_月次決算.xlsx", "月次決算/2024年02月_月次決算.xlsx", "月次決算/2024年03月_月次決算.xlsx",
		"予算計画/2024年度_予算計画書.xlsx", "予算計画/2024年度_部門別予算.xlsx", "予算計画/2024年度_投資計画.xlsx",
		"経費精算/2024年01月_経費精算書.xlsx", "経費精算/2024年02月_経費精算書.xlsx", "経費精算/2024年03月_経費精算書.xlsx",
		"税務申告/2023年度_法人税申告書.pdf", "税務申告/2023年度_消費税申告書.pdf", "税務申告/源泉徴収票_2023.xlsx",
		"監査資料/内部監査報告書_2023.pdf", "監査資料/外部監査報告書_2023.pdf", "監査資料/改善計画書_2024.docx",
	}
	
	for _, finance := range financeCategories {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("経理/%s", finance),
			Type:       getResourceType(finance),
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 売上分析レポート
	salesReports := []string{
		"コース別売上", "法人顧客別売上", "地域別売上", "講師別売上", "月次売上推移", "四半期売上分析", "年間売上サマリー",
	}
	
	for _, report := range salesReports {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("経理/売上分析/2024年_%s.xlsx", report),
			Type:       "spreadsheet",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 4. 進行管理・プロジェクト管理（60個）
	projects := []string{
		"新コース開発/AI基礎コース", "新コース開発/ブロックチェーン入門", "新コース開発/IoT実践",
		"システム開発/学習管理システム", "システム開発/決済システム", "システム開発/分析ダッシュボード",
		"マーケティング/春季キャンペーン", "マーケティング/法人営業強化", "マーケティング/SEO対策",
		"業務改善/カスタマーサポート改善", "業務改善/講師評価制度", "業務改善/コンテンツ品質向上",
	}
	
	projectDocs := []string{"プロジェクト計画書.docx", "WBS.xlsx", "進捗報告書.pptx", "課題管理表.xlsx", "リスク管理表.xlsx"}
	
	for _, project := range projects {
		for _, doc := range projectDocs {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("プロジェクト/%s/%s", project, doc),
				Type:       getResourceType(doc),
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "domain",
			})
			idCounter++
		}
	}

	// 5. 人事・労務関連（80個）
	hrCategories := []string{
		"採用/2024年_新卒採用", "採用/2024年_中途採用", "採用/2024年_インターン採用",
		"評価/2023年度_人事評価", "評価/2024年度_目標設定", "評価/360度評価",
		"勤怠/2024年01月_勤怠記録", "勤怠/2024年02月_勤怠記録", "勤怠/2024年03月_勤怠記録",
		"研修/新入社員研修", "研修/管理職研修", "研修/技術研修",
		"労務/就業規則", "労務/給与規程", "労務/福利厚生ガイド",
	}
	
	hrDocs := []string{"履歴書", "面接評価シート", "採用判定書", "雇用契約書", "評価シート"}
	
	for _, category := range hrCategories {
		for i := 1; i <= 3; i++ {
			docType := hrDocs[i%len(hrDocs)]
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("人事/%s/%s_%03d.docx", category, docType, i),
				Type:       "document",
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "private",
			})
			idCounter++
		}
	}

	// 6. 学習者管理・成績（100個）
	learnerData := []string{
		"学習進捗/2024年Q1_進捗レポート.xlsx", "学習進捗/個人別進捗_202401.csv", "学習進捗/コース完了率_202402.xlsx",
		"成績管理/プログラミング基礎_成績表.xlsx", "成績管理/データ分析入門_成績表.xlsx", "成績管理/ビジネス基礎_成績表.xlsx",
		"修了証明/修了証明書_発行リスト.xlsx", "修了証明/デジタル証明書_管理台帳.xlsx", "修了証明/認定証_テンプレート.docx",
		"学習分析/学習行動分析_202401.pdf", "学習分析/離脱率分析_202402.pdf", "学習分析/満足度調査_202403.xlsx",
		"受講履歴/アクセスログ_202401.csv", "受講履歴/視聴履歴_202402.csv", "受講履歴/課題提出状況_202403.xlsx",
	}
	
	for _, data := range learnerData {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("学習者管理/%s", data),
			Type:       getResourceType(data),
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 個人別学習データ
	learnerProfiles := []string{"sato_haruto", "suzuki_yua", "takahashi_sota", "tanaka_mio", "ito_ren"}
	learnerDocs := []string{"学習履歴.json", "成績表.xlsx", "進捗レポート.pdf", "フィードバック.docx", "学習計画.xlsx"}
	
	for _, profile := range learnerProfiles {
		for _, doc := range learnerDocs {
			resources = append(resources, Resource{
				Name:       fmt.Sprintf("学習者管理/個人データ/%s/%s", profile, doc),
				Type:       getResourceType(doc),
				ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
				Visibility: "private",
			})
			idCounter++
		}
	}

	// 7. 経営・戦略関連（50個）
	strategyDocs := []string{
		"事業計画/2024年度_事業計画書.pptx", "事業計画/中期経営計画_2024-2026.pptx", "事業計画/新規事業提案書.docx",
		"KPI管理/月次KPIレポート_202401.xlsx", "KPI管理/四半期KPIレビュー_2024Q1.pptx", "KPI管理/年間KPI達成状況.xlsx",
		"市場分析/競合分析レポート_2024.pdf", "市場分析/市場調査結果_EdTech.pdf", "市場分析/顧客セグメント分析.xlsx",
		"投資家向け/決算説明資料_2023Q4.pptx", "投資家向け/成長戦略説明書.pdf", "投資家向け/IR資料_最新.pptx",
		"取締役会/議事録_202401.docx", "取締役会/議事録_202402.docx", "取締役会/議事録_202403.docx",
	}
	
	for _, doc := range strategyDocs {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("経営戦略/%s", doc),
			Type:       getResourceType(doc),
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 8. マーケティング・営業（70個）
	marketingDocs := []string{
		"キャンペーン/春の新規獲得キャンペーン/企画書.pptx", "キャンペーン/春の新規獲得キャンペーン/予算計画.xlsx", "キャンペーン/春の新規獲得キャンペーン/効果測定.xlsx",
		"広告/Google広告_レポート_202401.xlsx", "広告/Facebook広告_分析_202402.xlsx", "広告/YouTube広告_実績_202403.xlsx",
		"SEO/キーワード分析_2024.xlsx", "SEO/コンテンツ計画_Q1.docx", "SEO/順位レポート_月次.xlsx",
		"営業資料/サービス紹介資料_最新版.pptx", "営業資料/料金プラン説明書.pdf", "営業資料/導入事例集.pdf",
		"顧客分析/顧客満足度調査_2024Q1.xlsx", "顧客分析/NPS分析レポート.pdf", "顧客分析/離脱要因分析.xlsx",
		"提案書/A社向け_提案書_202401.pptx", "提案書/B社向け_カスタマイズ提案.pptx", "提案書/C社向け_導入提案.pptx",
	}
	
	for _, doc := range marketingDocs {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("マーケティング/%s", doc),
			Type:       getResourceType(doc),
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "domain",
		})
		idCounter++
	}

	// 9. 技術・開発関連（60個）
	techDocs := []string{
		"設計書/システムアーキテクチャ図.pdf", "設計書/データベース設計書.xlsx", "設計書/API仕様書.yaml",
		"開発/ソースコード仕様.md", "開発/コーディング規約.pdf", "開発/テスト計画書.docx",
		"インフラ/AWS構成図.pdf", "インフラ/監視設定書.xlsx", "インフラ/バックアップ計画.docx",
		"セキュリティ/脆弱性診断報告書_2024Q1.pdf", "セキュリティ/ペネトレーションテスト結果.pdf", "セキュリティ/インシデント対応手順.docx",
		"運用/障害対応記録_202401.xlsx", "運用/メンテナンス計画_2024.xlsx", "運用/SLA管理表.xlsx",
		"技術調査/AI技術トレンド_2024.pdf", "技術調査/新技術評価レポート.docx", "技術調査/PoC実施結果.pptx",
	}
	
	for _, doc := range techDocs {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("技術開発/%s", doc),
			Type:       getResourceType(doc),
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "domain",
		})
		idCounter++
	}

	// 10. 法務・コンプライアンス（40個）
	legalDocs := []string{
		"規程/利用規約_最新版.pdf", "規程/プライバシーポリシー_20240401.pdf", "規程/特定商取引法表記.pdf",
		"コンプライアンス/個人情報保護方針.pdf", "コンプライアンス/情報セキュリティポリシー.pdf", "コンプライアンス/内部統制規程.pdf",
		"GDPR/GDPR対応チェックリスト.xlsx", "GDPR/データ処理契約書_テンプレート.docx", "GDPR/プライバシー影響評価.pdf",
		"認証/ISO27001_認証書.pdf", "認証/SOC2_監査報告書.pdf", "認証/プライバシーマーク_申請書.docx",
		"契約管理/契約書管理台帳.xlsx", "契約管理/契約更新スケジュール.xlsx", "契約管理/契約書テンプレート集.zip",
		"知的財産/商標登録証_無敗塾.pdf", "知的財産/著作権管理台帳.xlsx", "知的財産/特許出願書類.pdf",
	}
	
	for _, doc := range legalDocs {
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("法務/%s", doc),
			Type:       getResourceType(doc),
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 個人フォルダ（ユーザーごとに生成）
	allUsers := generateUsers()
	for _, user := range allUsers {
		emailPrefix := strings.Split(user.Email, "@")[0]
		resources = append(resources, Resource{
			Name:       fmt.Sprintf("個人/%s/", emailPrefix),
			Type:       "folder",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "private",
		})
		idCounter++
	}

	// 共有フォルダ
	sharedFolders := []string{
		"共有/お知らせ/", "共有/社内イベント/", "共有/研修資料/", "共有/業務マニュアル/", "共有/テンプレート/",
		"共有/会議資料/経営会議/", "共有/会議資料/部門会議/", "共有/会議資料/全社会議/",
		"共有/プロジェクト/", "共有/ナレッジベース/", "共有/ベストプラクティス/",
	}

	for _, folder := range sharedFolders {
		resources = append(resources, Resource{
			Name:       folder,
			Type:       "folder",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "domain",
		})
		idCounter++
	}

	// 公開ファイル
	publicFiles := []string{
		"公開/サービス案内.pdf", "公開/料金表_2024.pdf", "公開/導入事例集.pdf", "公開/お客様の声.pdf",
		"公開/よくある質問.pdf", "公開/利用ガイド.pdf", "公開/セミナー資料.pdf", "公開/ホワイトペーパー.pdf",
	}

	for _, file := range publicFiles {
		resources = append(resources, Resource{
			Name:       file,
			Type:       "document",
			ID:         fmt.Sprintf("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE%04d", idCounter),
			Visibility: "public_on_the_web",
		})
		idCounter++
	}

	return resources
}

// ファイル拡張子からリソースタイプを判定
func getResourceType(filename string) string {
	if strings.HasSuffix(filename, ".xlsx") || strings.HasSuffix(filename, ".csv") {
		return "spreadsheet"
	} else if strings.HasSuffix(filename, ".mp4") || strings.HasSuffix(filename, ".avi") || strings.HasSuffix(filename, ".mov") {
		return "video"
	} else if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".gif") {
		return "image"
	} else if strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".md") || strings.HasSuffix(filename, ".zip") {
		return "file"
	} else if strings.HasSuffix(filename, "/") {
		return "folder"
	}
	return "document"
}