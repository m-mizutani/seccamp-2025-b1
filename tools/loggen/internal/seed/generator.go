package seed

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand"
	"time"

	"github.com/m-mizutani/seccamp-2025-b1/internal/logcore"
)

// シードジェネレータ
type Generator struct {
	config *logcore.Config
	// 常時発生型異常パターン用の状態管理
	authAttackState    *authAttackState
	dataTheftState     *dataTheftState
	serviceProbingState *serviceProbingState
	geoAccessState     *geoAccessState
}

// 常時発生型異常パターンの状態管理構造体
type authAttackState struct {
	attackerIP  string
	targetUsers []string
}

type dataTheftState struct {
	theftUser string
	theftIP   string
	fileIndex int
}

type serviceProbingState struct {
	probingUser  string
	serviceOrder []string
	currentIndex int
}

type geoAccessState struct {
	user       string
	country1   string
	country1IP string
	country2   string
	country2IP string
	lastCountry int // 0 or 1
}

// 新しいシードジェネレータを作成
func NewGenerator() *Generator {
	g := &Generator{
		config: logcore.DefaultConfig(),
	}
	g.initializeContinuousPatterns()
	return g
}

// 常時発生型異常パターンの初期化
func (g *Generator) initializeContinuousPatterns() {
	// Pattern 4: 高頻度認証攻撃
	g.authAttackState = &authAttackState{
		attackerIP: "133.200.32.94",
		targetUsers: []string{
			"admin@example.com",
			"user1@example.com",
			"user2@example.com",
			"support@example.com",
			"test@example.com",
		},
	}

	// Pattern 5: 超高速データ窃取
	g.dataTheftState = &dataTheftState{
		theftUser: "tanaka.hiroshi@muhaijuku.com",  // 乗っ取られた正規ユーザーアカウント
		theftIP:   "198.51.100.99",  // 海外からの不審なアクセス
		fileIndex: 0,
	}

	// Pattern 6: マルチサービス不正アクセス試行
	g.serviceProbingState = &serviceProbingState{
		probingUser: "sato.yuki@muhaijuku.com",  // マルウェアに感染したユーザーアカウント
		serviceOrder: []string{
			"Google Drive",
			"Google Calendar",
			"Google Gmail",
			"Google Admin",
		},
		currentIndex: 0,
	}

	// Pattern 7: 地理的同時アクセス
	g.geoAccessState = &geoAccessState{
		user:        "yamada.takeshi@muhaijuku.com",  // 出張中のユーザー（の認証情報が盗まれた）
		country1:    "JP",
		country1IP:  "192.0.2.10",
		country2:    "US",
		country2IP:  "198.51.100.20",
		lastCountry: 0,
	}
}

// 1日分のシードテンプレートを生成
func (g *Generator) GenerateDayTemplate(date time.Time, anomalyRatio float64) (*logcore.DayTemplate, error) {
	template := &logcore.DayTemplate{
		Date:     date.Format("2006-01-02"),
		LogSeeds: make([]logcore.LogSeed, 0, 864000), // 予想容量
		Metadata: logcore.SeedMeta{
			Generated:         time.Now(),
			LogCoreVersion:    logcore.LogCoreVersion,
			SeedFormatVersion: logcore.SeedFormatVersion,
			AnomalyStats:      make(map[string]int),
		},
	}

	// 1秒ごとにログシードを生成
	for second := 0; second < 86400; second++ {
		currentTime := date.Add(time.Duration(second) * time.Second)

		// この秒のログ件数決定（平均10件、時間帯による調整）
		logCount := g.generateLogCount(currentTime)

		for i := 0; i < logCount; i++ {
			seed := logcore.LogSeed{
				Timestamp: int64(second),
				Seed:      g.generateSecondSeed(currentTime, i),
			}

			// 異常パターンの配置判定（先に判定）
			seed.Pattern = g.determineAnomalyPattern(currentTime, i, anomalyRatio)

			// パターンに応じた活動パターン決定
			if seed.Pattern >= logcore.PatternExample4HighFreqAuthAttack {
				// 新しい常時発生型パターンの場合、特別な処理
				seed.EventType, seed.UserIndex, seed.ResourceIdx = g.selectContinuousPatternActivity(seed.Pattern, currentTime, i)
			} else {
				// 通常の時間帯・曜日による活動パターン決定
				seed.EventType, seed.UserIndex, seed.ResourceIdx = g.selectActivityPattern(currentTime, i)
			}

			// 正常パターンのログは50%の確率でスキップ（異常パターンは必ず追加）
			if seed.Pattern == logcore.PatternNormal && rand.Float64() < 0.5 {
				continue
			}

			template.LogSeeds = append(template.LogSeeds, seed)
		}
	}

	// 統計情報更新
	g.updateStats(template)

	return template, nil
}

// この秒のログ件数を決定
func (g *Generator) generateLogCount(t time.Time) int {
	hour := t.Hour()
	weekday := t.Weekday()

	// 時間帯・曜日による期待値調整
	expectedRate := g.getExpectedRate(hour, weekday)

	// ポワソン分布でログ件数決定
	return g.poissonRandom(expectedRate)
}

// 時間帯・曜日による期待ログレート計算
func (g *Generator) getExpectedRate(hour int, weekday time.Weekday) float64 {
	// 基本レート: 毎秒200件（さらに2倍に増加）
	baseRate := 200.0

	// 時間帯補正（最小値を引き上げ）
	timeMultiplier := map[int]float64{
		0: 0.3, 1: 0.2, 2: 0.2, 3: 0.2, 4: 0.2, 5: 0.3, // 深夜（最低40-60件/秒）
		6: 0.4, 7: 0.6, 8: 0.8, // 朝
		9: 1.2, 10: 1.5, 11: 1.8, 12: 1.0, // 午前〜昼
		13: 0.8, 14: 1.3, 15: 1.6, 16: 1.4, 17: 1.1, 18: 0.9, // 午後
		19: 0.6, 20: 0.5, 21: 0.4, 22: 0.3, 23: 0.3, // 夜
	}

	// 曜日補正
	weekdayMultiplier := map[time.Weekday]float64{
		time.Monday: 1.0, time.Tuesday: 1.1, time.Wednesday: 1.2,
		time.Thursday: 1.1, time.Friday: 0.9,
		time.Saturday: 0.5, time.Sunday: 0.4, // 週末も少し引き上げ
	}

	return baseRate * timeMultiplier[hour] * weekdayMultiplier[weekday]
}

// ポワソン分布による乱数生成
func (g *Generator) poissonRandom(lambda float64) int {
	if lambda == 0 {
		return 0
	}

	// 簡易ポワソン分布（小さいλの場合）
	if lambda < 10 {
		L := math.Exp(-lambda)
		k := 0
		p := 1.0

		for p > L {
			k++
			p *= rand.Float64()
		}
		return k - 1
	}

	// 大きいλの場合は正規近似
	mean := lambda
	stddev := math.Sqrt(lambda)
	return int(rand.NormFloat64()*stddev + mean + 0.5)
}

// 決定論的シード生成
func (g *Generator) generateSecondSeed(currentTime time.Time, sequence int) uint32 {
	h := sha256.New()
	h.Write([]byte(currentTime.Format(time.RFC3339)))
	h.Write([]byte{byte(sequence)})
	hash := h.Sum(nil)
	return binary.BigEndian.Uint32(hash[:4])
}

// 活動パターン選択
func (g *Generator) selectActivityPattern(currentTime time.Time, sequence int) (uint8, uint8, uint8) {
	// シーケンスベースのランダム生成器
	rng := rand.New(rand.NewSource(currentTime.Unix() + int64(sequence)))

	hour := currentTime.Hour()
	weekday := currentTime.Weekday()

	// イベントタイプ決定
	var eventType uint8
	switch {
	case hour >= 9 && hour <= 18 && isWeekday(weekday):
		// 業務時間内
		eventType = g.selectBusinessHoursEvent(rng)
	case hour >= 19 && hour <= 22:
		// 残業時間
		eventType = g.selectOvertimeEvent(rng)
	default:
		// その他の時間（夜間・早朝）も通常の業務時間と同じ比率を維持
		// ただし全体量は getExpectedRate で調整済み
		eventType = g.selectBusinessHoursEvent(rng)
	}

	// ユーザー選択（重み付き）
	userIndex := g.selectUser(rng, hour)

	// リソース選択
	resourceIndex := g.selectResource(rng, eventType)

	return eventType, userIndex, resourceIndex
}

// 業務時間のイベント選択
func (g *Generator) selectBusinessHoursEvent(rng *rand.Rand) uint8 {
	weights := map[uint8]int{
		logcore.EventTypeDriveAccess: 60,
		logcore.EventTypeGmail:       25,
		logcore.EventTypeCalendar:    9,
		logcore.EventTypeLogin:       5,
		logcore.EventTypeAdmin:       1,
	}

	return g.weightedSelect(rng, weights)
}

// 残業時間のイベント選択
func (g *Generator) selectOvertimeEvent(rng *rand.Rand) uint8 {
	weights := map[uint8]int{
		logcore.EventTypeDriveAccess: 70,
		logcore.EventTypeGmail:       20,
		logcore.EventTypeCalendar:    5,
		logcore.EventTypeLogin:       4,
		logcore.EventTypeAdmin:       1,
	}

	return g.weightedSelect(rng, weights)
}

// 重み付き選択
func (g *Generator) weightedSelect(rng *rand.Rand, weights map[uint8]int) uint8 {
	total := 0
	for _, weight := range weights {
		total += weight
	}

	r := rng.Intn(total)
	current := 0

	for value, weight := range weights {
		current += weight
		if r < current {
			return value
		}
	}

	// フォールバック
	for value := range weights {
		return value
	}
	return 1
}

// ユーザー選択
func (g *Generator) selectUser(rng *rand.Rand, hour int) uint8 {
	userCount := len(g.config.Users)

	// 時間帯によるユーザー活動確率調整
	if hour >= 9 && hour <= 18 {
		// 業務時間内は内部ユーザーが多い
		return uint8(rng.Intn(userCount - 2)) // 外部ユーザー除外
	}

	// その他の時間は全ユーザー
	return uint8(rng.Intn(userCount))
}

// リソース選択
func (g *Generator) selectResource(rng *rand.Rand, eventType uint8) uint8 {
	resourceCount := len(g.config.Resources)

	switch eventType {
	case logcore.EventTypeAdmin:
		// 管理者設定は管理リソース
		return uint8(rng.Intn(2) + 5) // 管理/ resources
	case logcore.EventTypeDriveAccess:
		// ドライブアクセスは主にファイル
		return uint8(rng.Intn(5)) // 教材/, 成績/
	default:
		return uint8(rng.Intn(resourceCount))
	}
}

// 異常パターン決定
func (g *Generator) determineAnomalyPattern(t time.Time, sequenceInSecond int, anomalyRatio float64) uint8 {
	hour := t.Hour()

	// 基本的な異常確率
	baseAnomalyProb := anomalyRatio

	// 新規パターン（常時発生）: 全体の2-3%程度
	if rand.Float64() < 0.025 {
		return g.selectContinuousAnomalyPattern(t, sequenceInSecond)
	}

	// 実例1: 夜間の管理者による大量学習データダウンロード
	if (hour >= 18 || hour <= 9) && rand.Float64() < baseAnomalyProb*0.2 {
		return logcore.PatternExample1NightAdminDownload
	}

	// 実例2: anyone with link設定ミスによる外部流出（まとまって発生）
	if hour >= 10 && hour <= 16 {
		// 15分間隔でバーストパターン（外部からの自動アクセス）
		if t.Minute()%15 < 3 && rand.Float64() < baseAnomalyProb*0.3 {
			return logcore.PatternExample2ExternalLinkAccess
		}
	}

	// 実例3: VPN脆弱性経由の水平移動攻撃（業務時間内の怪しい動作）
	if hour >= 9 && hour <= 18 && rand.Float64() < baseAnomalyProb*0.15 {
		return logcore.PatternExample3VpnLateralMovement
	}

	// 一般的な軽微異常
	if rand.Float64() < baseAnomalyProb*0.2 {
		return uint8(int(logcore.PatternTimeAnomaly) + rand.Intn(2)) // Pattern 4-5
	}

	return logcore.PatternNormal
}

// 常時発生型異常パターンの選択
func (g *Generator) selectContinuousAnomalyPattern(t time.Time, sequenceInSecond int) uint8 {
	second := t.Second()
	
	// Pattern 4: 高頻度認証攻撃（1分に3-5回）
	if sequenceInSecond < 5 && second%(60/4) == 0 {
		return logcore.PatternExample4HighFreqAuthAttack
	}
	
	// Pattern 5: 超高速データ窃取（1分に10-15回）
	if sequenceInSecond < 15 && second%(60/12) == 0 {
		return logcore.PatternExample5RapidDataTheft
	}
	
	// Pattern 6: マルチサービス不正アクセス（1分に3-5回）
	if sequenceInSecond < 5 && (second+30)%(60/4) == 0 {
		return logcore.PatternExample6MultiServiceProbing
	}
	
	// Pattern 7: 地理的同時アクセス（1分に4-6回、2カ国合計）
	if sequenceInSecond < 6 && (second+15)%(60/5) == 0 {
		return logcore.PatternExample7SimultaneousGeoAccess
	}
	
	// フォールバック
	patterns := []uint8{
		logcore.PatternExample4HighFreqAuthAttack,
		logcore.PatternExample5RapidDataTheft,
		logcore.PatternExample6MultiServiceProbing,
		logcore.PatternExample7SimultaneousGeoAccess,
	}
	return patterns[rand.Intn(len(patterns))]
}

// 統計情報更新
func (g *Generator) updateStats(template *logcore.DayTemplate) {
	template.Metadata.TotalLogs = len(template.LogSeeds)

	anomalyCount := 0
	for _, seed := range template.LogSeeds {
		if seed.Pattern > 0 {
			anomalyCount++
		}
	}

	template.Metadata.NormalRatio = float64(len(template.LogSeeds)-anomalyCount) / float64(len(template.LogSeeds))
	template.Metadata.AnomalyStats["total"] = anomalyCount
	template.Metadata.AnomalyStats["example1"] = g.countPattern(template.LogSeeds, logcore.PatternExample1NightAdminDownload)
	template.Metadata.AnomalyStats["example2"] = g.countPattern(template.LogSeeds, logcore.PatternExample2ExternalLinkAccess)
	template.Metadata.AnomalyStats["example3"] = g.countPattern(template.LogSeeds, logcore.PatternExample3VpnLateralMovement)
	template.Metadata.AnomalyStats["time_anomaly"] = g.countPattern(template.LogSeeds, logcore.PatternTimeAnomaly)
	template.Metadata.AnomalyStats["volume_anomaly"] = g.countPattern(template.LogSeeds, logcore.PatternVolumeAnomaly)
	// 新規パターンの統計
	template.Metadata.AnomalyStats["example4_auth"] = g.countPattern(template.LogSeeds, logcore.PatternExample4HighFreqAuthAttack)
	template.Metadata.AnomalyStats["example5_theft"] = g.countPattern(template.LogSeeds, logcore.PatternExample5RapidDataTheft)
	template.Metadata.AnomalyStats["example6_probe"] = g.countPattern(template.LogSeeds, logcore.PatternExample6MultiServiceProbing)
	template.Metadata.AnomalyStats["example7_geo"] = g.countPattern(template.LogSeeds, logcore.PatternExample7SimultaneousGeoAccess)
}

// パターン数カウント
func (g *Generator) countPattern(seeds []logcore.LogSeed, pattern uint8) int {
	count := 0
	for _, seed := range seeds {
		if seed.Pattern == pattern {
			count++
		}
	}
	return count
}

// 平日判定
func isWeekday(weekday time.Weekday) bool {
	return weekday >= time.Monday && weekday <= time.Friday
}

// 常時発生型パターン用の活動選択
func (g *Generator) selectContinuousPatternActivity(pattern uint8, t time.Time, sequence int) (uint8, uint8, uint8) {
	switch pattern {
	case logcore.PatternExample4HighFreqAuthAttack:
		return g.generatePattern4AuthAttack(t, sequence)
	case logcore.PatternExample5RapidDataTheft:
		return g.generatePattern5DataTheft(t, sequence)
	case logcore.PatternExample6MultiServiceProbing:
		return g.generatePattern6ServiceProbing(t, sequence)
	case logcore.PatternExample7SimultaneousGeoAccess:
		return g.generatePattern7GeoAccess(t, sequence)
	default:
		// フォールバック
		return g.selectActivityPattern(t, sequence)
	}
}

// Pattern 4: 高頻度認証攻撃
func (g *Generator) generatePattern4AuthAttack(t time.Time, sequence int) (uint8, uint8, uint8) {
	// 認証イベント
	eventType := logcore.EventTypeLogin
	
	// ランダムなターゲットユーザーを選択
	targetUserIndex := uint8(rand.Intn(len(g.authAttackState.targetUsers)))
	
	// リソースは認証なので0
	resourceIndex := uint8(0)
	
	return eventType, targetUserIndex, resourceIndex
}

// Pattern 5: 超高速データ窃取
func (g *Generator) generatePattern5DataTheft(t time.Time, sequence int) (uint8, uint8, uint8) {
	// ダウンロードイベント
	eventType := logcore.EventTypeDriveAccess
	
	// 特定の侵害されたユーザー
	// tanaka.hiroshi@muhaijuku.com is at index 2
	userIndex := uint8(2) // tanaka.kenji@muhai-academy.comを使いたいが、存在しないので代替
	
	// 異なるファイルを選択（インクリメント）
	g.dataTheftState.fileIndex++
	resourceIndex := uint8(g.dataTheftState.fileIndex % 100)
	
	return eventType, userIndex, resourceIndex
}

// Pattern 6: マルチサービス不正アクセス試行
func (g *Generator) generatePattern6ServiceProbing(t time.Time, sequence int) (uint8, uint8, uint8) {
	// サービスを順番に試行
	serviceTypes := []uint8{
		logcore.EventTypeDriveAccess,
		logcore.EventTypeCalendar,
		logcore.EventTypeGmail,
		logcore.EventTypeAdmin,
	}
	
	eventType := serviceTypes[g.serviceProbingState.currentIndex]
	g.serviceProbingState.currentIndex = (g.serviceProbingState.currentIndex + 1) % len(serviceTypes)
	
	// 特定の感染ユーザー
	userIndex := uint8(3) // sato.yuki@muhaijuku.com (index 3)
	
	// リソースはサービスに応じて選択
	resourceIndex := uint8(rand.Intn(10))
	
	return eventType, userIndex, resourceIndex
}

// Pattern 7: 地理的同時アクセス
func (g *Generator) generatePattern7GeoAccess(t time.Time, sequence int) (uint8, uint8, uint8) {
	// 通常の業務操作（ドライブアクセスなど）
	eventTypes := []uint8{
		logcore.EventTypeDriveAccess,
		logcore.EventTypeGmail,
		logcore.EventTypeCalendar,
	}
	eventType := eventTypes[rand.Intn(len(eventTypes))]
	
	// 特定のユーザー
	userIndex := uint8(0) // yamada.takeshi@muhaijuku.com (index 0)
	
	// 国を交互に切り替え
	g.geoAccessState.lastCountry = (g.geoAccessState.lastCountry + 1) % 2
	
	// リソースは通常のアクセス
	resourceIndex := uint8(rand.Intn(20))
	
	return eventType, userIndex, resourceIndex
}
