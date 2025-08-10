package logcore

import (
	"fmt"
	"math/rand"
	"time"
)

// ログエントリジェネレータ
type Generator struct {
	config           *Config
	anomalyGenerator *AnomalyGenerator
}

// 新しいジェネレータを作成
func NewGenerator(config *Config) *Generator {
	return &Generator{
		config:           config,
		anomalyGenerator: NewAnomalyGenerator(),
	}
}

// シードからログエントリを生成
func (g *Generator) GenerateLogEntry(seed LogSeed, baseDate time.Time, sequenceInSecond int) *GoogleWorkspaceLogEntry {
	// 1. タイムスタンプ計算
	timestamp := baseDate.Add(time.Duration(seed.Timestamp) * time.Second)

	// 2. シードベースの疑似乱数生成器
	rng := rand.New(rand.NewSource(int64(seed.Seed) + int64(sequenceInSecond)))

	// 3. ユーザー・リソース解決
	user := g.config.Users[seed.UserIndex%uint8(len(g.config.Users))]
	resource := g.config.Resources[seed.ResourceIdx%uint8(len(g.config.Resources))]

	// 4. 基本ログエントリ作成
	logEntry := &GoogleWorkspaceLogEntry{
		Kind: "audit#activity",
		ID: LogID{
			Time:            timestamp.Format(time.RFC3339Nano),
			UniqueQualifier: g.generateUniqueQualifier(rng),
			ApplicationName: g.determineApplicationName(seed.EventType),
			CustomerID:      g.config.CustomerID,
		},
		Actor: Actor{
			CallerType: "USER",
			Email:      user.Email,
			ProfileID:  user.ProfileID,
		},
		OwnerDomain: g.config.OwnerDomain,
		IPAddress:   g.generateIPAddress(user, timestamp),
		Events:      []Event{},
	}

	// 5. イベント種別による生成分岐
	switch seed.EventType {
	case EventTypeDriveAccess:
		logEntry.Events = append(logEntry.Events, g.generateDriveAccessEvent(user, resource, rng))
	case EventTypeLogin:
		logEntry.Events = append(logEntry.Events, g.generateLoginEvent(user, timestamp, rng))
	case EventTypeAdmin:
		logEntry.Events = append(logEntry.Events, g.generateAdminEvent(user, resource, rng))
	case EventTypeCalendar:
		logEntry.Events = append(logEntry.Events, g.generateCalendarEvent(user, resource, rng))
	case EventTypeGmail:
		logEntry.Events = append(logEntry.Events, g.generateGmailEvent(user, resource, rng))
	default:
		logEntry.Events = append(logEntry.Events, g.generateDriveAccessEvent(user, resource, rng))
	}

	// 6. 異常パターンの適用
	if seed.Pattern > 0 {
		return g.anomalyGenerator.ApplyAnomalyPattern(logEntry, seed.Pattern, rng)
	}

	return logEntry
}

// ユニーク修飾子を生成
func (g *Generator) generateUniqueQualifier(rng *rand.Rand) string {
	return fmt.Sprintf("%d", rng.Int63())
}

// アプリケーション名を決定
func (g *Generator) determineApplicationName(eventType uint8) string {
	switch eventType {
	case EventTypeDriveAccess:
		return "drive"
	case EventTypeLogin:
		return "login"
	case EventTypeAdmin:
		return "admin"
	case EventTypeCalendar:
		return "calendar"
	case EventTypeGmail:
		return "gmail"
	default:
		return "drive"
	}
}

// IPアドレスを生成
func (g *Generator) generateIPAddress(user User, timestamp time.Time) string {
	// 外部ユーザーは別のIPレンジを使用
	if user.Role == "external" {
		// 外部ユーザーは203.0.113.x（ドキュメント用のIPレンジ）
		// ただし .99 は攻撃者用に予約されているので避ける
		hash := 0
		for _, c := range user.Email {
			hash = hash*31 + int(c)
		}
		baseIP := (hash % 70) + 1  // 1-70の範囲
		
		// 外部ユーザーも2-3のIPを使い分け
		ipVariation := g.selectIPVariation(timestamp, hash)
		finalIP := baseIP + (ipVariation * 20)
		
		// 99を避ける（70 + 2*20 = 110が最大値なので98を上限に）
		if finalIP > 98 {
			finalIP = 98
		}
		return fmt.Sprintf("203.0.113.%d", finalIP)
	}
	
	// 内部ユーザーのIPアドレス生成
	// ユーザーごとに一貫したIPアドレスを生成（ユーザーのメールアドレスをハッシュ化）
	hash := 0
	for _, c := range user.Email {
		hash = hash*31 + int(c)
	}
	
	// 時間帯と曜日に基づいて2-3のIPアドレスを使い分け
	ipVariation := g.selectIPVariation(timestamp, hash)
	
	// 各ユーザーに固有の2-3個のIPアドレスを割り当て
	// ユーザーのハッシュ値に基づいて異なるIPレンジから選択
	return g.generateUserSpecificIP(hash, ipVariation)
}

// ユーザー固有のIPアドレスを生成
func (g *Generator) generateUserSpecificIP(userHash int, variation int) string {
	// ユーザーごとに異なるが固定のIPを生成
	// 最後のオクテットはユーザーハッシュから決定
	lastOctet := (userHash % 200) + 20  // 20-219の範囲
	
	// オフィスIPは常に同じ（固定）
	officeIP := fmt.Sprintf("210.160.34.%d", lastOctet)
	
	// variation に基づいて異なるIPレンジを使用
	switch variation {
	case 0:
		// オフィス（固定IP）
		return officeIP
	case 1:
		// モバイル/カフェ（様々なISP）
		// ランダム性を高めるため、時間要素も含める
		ispChoice := userHash % 4
		randomOffset := (userHash * variation) % 50
		switch ispChoice {
		case 0:
			return fmt.Sprintf("126.204.%d.%d", (randomOffset%50)+100, (lastOctet+randomOffset)%256)  // docomo
		case 1:
			return fmt.Sprintf("110.163.%d.%d", (randomOffset%50)+150, (lastOctet+randomOffset)%256)  // au
		case 2:
			return fmt.Sprintf("101.142.%d.%d", (randomOffset%50)+100, (lastOctet+randomOffset)%256)  // softbank
		default:
			return fmt.Sprintf("114.156.%d.%d", (randomOffset%50)+50, (lastOctet+randomOffset)%256)   // その他
		}
	default:
		// 自宅（様々な一般ISP）
		ispChoice := (userHash + 1) % 4  // モバイルとは異なるISPを選択
		randomOffset := (userHash * (variation + 2)) % 50
		switch ispChoice {
		case 0:
			return fmt.Sprintf("118.103.%d.%d", (randomOffset%50)+100, (lastOctet+randomOffset)%256)  // OCN
		case 1:
			return fmt.Sprintf("122.208.%d.%d", (randomOffset%50)+50, (lastOctet+randomOffset)%256)   // BIGLOBE
		case 2:
			return fmt.Sprintf("125.198.%d.%d", (randomOffset%50)+100, (lastOctet+randomOffset)%256)  // So-net
		default:
			return fmt.Sprintf("133.200.%d.%d", (randomOffset%50)+50, (lastOctet+randomOffset)%256)   // その他
		}
	}
}

// IPバリエーションを選択（0, 1, 2のいずれか）
func (g *Generator) selectIPVariation(timestamp time.Time, userHash int) int {
	hour := timestamp.Hour()
	weekday := timestamp.Weekday()
	
	// 平日と週末で異なるパターン
	if weekday == time.Saturday || weekday == time.Sunday {
		// 週末は主に自宅IP（variation 2）
		if hour >= 10 && hour <= 16 {
			// 昼間は外出先（variation 1）
			return 1
		}
		return 2
	}
	
	// 平日
	if hour >= 9 && hour <= 12 {
		// 午前中はオフィス（variation 0）
		return 0
	} else if hour >= 13 && hour <= 17 {
		// 午後もオフィス（variation 0）
		return 0
	} else if hour >= 18 && hour <= 22 {
		// 夜は自宅（variation 2）
		return 2
	} else {
		// 深夜・早朝は自宅またはモバイル
		if userHash % 3 == 0 {
			return 1 // モバイル/カフェ
		}
		return 2 // 自宅
	}
}

// Drive アクセスイベントを生成
func (g *Generator) generateDriveAccessEvent(user User, resource Resource, rng *rand.Rand) Event {
	return Event{
		Type: "access",
		Name: "view",
		Parameters: []Parameter{
			{Name: "doc_id", Value: resource.ID},
			{Name: "doc_title", Value: resource.Name},
			{Name: "doc_type", Value: resource.Type},
			{Name: "owner", Value: user.Email},
			{Name: "visibility", Value: resource.Visibility},
			{Name: "primary_event", BoolValue: true},
		},
	}
}

// ログインイベントを生成
func (g *Generator) generateLoginEvent(user User, timestamp time.Time, rng *rand.Rand) Event {
	return Event{
		Type: "login",
		Name: "login_success",
		Parameters: []Parameter{
			{Name: "login_type", Value: "google_password"},
			{Name: "login_challenge_method", MultiStrValue: []string{"password"}},
		},
	}
}

// 管理イベントを生成
func (g *Generator) generateAdminEvent(user User, resource Resource, rng *rand.Rand) Event {
	return Event{
		Type: "USER_SETTINGS",
		Name: "CREATE_USER",
		Parameters: []Parameter{
			{Name: "USER_EMAIL", Value: user.Email},
			{Name: "DOMAIN_NAME", Value: g.config.OwnerDomain},
		},
	}
}

// カレンダーイベントを生成
func (g *Generator) generateCalendarEvent(user User, resource Resource, rng *rand.Rand) Event {
	return Event{
		Type: "event_change",
		Name: "create_event",
		Parameters: []Parameter{
			{Name: "calendar_id", Value: "primary"},
			{Name: "event_id", Value: g.generateEventID(rng)},
			{Name: "api_kind", Value: "web"},
		},
	}
}

// Gmailイベントを生成
func (g *Generator) generateGmailEvent(user User, resource Resource, rng *rand.Rand) Event {
	return Event{
		Type: "mail_action",
		Name: "send_message",
		Parameters: []Parameter{
			{Name: "message_id", Value: g.generateMessageID(rng)},
			{Name: "recipient", Value: g.generateRecipientEmail(rng)},
			{Name: "size_bytes", Value: fmt.Sprintf("%d", rng.Intn(50000)+1000)},
			{Name: "is_encrypted", BoolValue: rng.Float32() < 0.3},
		},
	}
}

// ドキュメントIDを生成
func (g *Generator) generateDocumentID(rng *rand.Rand) string {
	return fmt.Sprintf("1%s", g.generateRandomString(rng, 43))
}

// イベントIDを生成
func (g *Generator) generateEventID(rng *rand.Rand) string {
	return g.generateRandomString(rng, 12)
}

// メッセージIDを生成
func (g *Generator) generateMessageID(rng *rand.Rand) string {
	return fmt.Sprintf("<%s@mail.gmail.com>", g.generateRandomString(rng, 16))
}

// 受信者メールアドレスを生成
func (g *Generator) generateRecipientEmail(rng *rand.Rand) string {
	// 内部ユーザーまたは外部ユーザーを選択
	if rng.Float32() < 0.8 {
		// 80%は内部ユーザー
		return g.config.Users[rng.Intn(len(g.config.Users))].Email
	}
	// 20%は外部ユーザー
	domains := []string{"example.com", "test.co.jp", "partner.org"}
	return fmt.Sprintf("user%d@%s", rng.Intn(100), domains[rng.Intn(len(domains))])
}

// ランダム文字列を生成
func (g *Generator) generateRandomString(rng *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rng.Intn(len(chars))]
	}
	return string(result)
}

// 時間範囲内のシードを抽出
func ExtractSeedsInRange(template *DayTemplate, startTime, endTime time.Time) []LogSeed {
	// 日付の開始時刻を取得
	dayStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())

	// 秒単位での開始・終了位置計算
	startSecond := int64(startTime.Sub(dayStart).Seconds())
	endSecond := int64(endTime.Sub(dayStart).Seconds())

	var seeds []LogSeed
	for _, seed := range template.LogSeeds {
		if seed.Timestamp >= startSecond && seed.Timestamp < endSecond {
			seeds = append(seeds, seed)
		}
	}

	return seeds
}
