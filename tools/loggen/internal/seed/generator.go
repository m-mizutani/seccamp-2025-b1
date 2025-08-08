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
}

// 新しいシードジェネレータを作成
func NewGenerator() *Generator {
	return &Generator{
		config: logcore.DefaultConfig(),
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

			// 時間帯・曜日による活動パターン決定
			seed.EventType, seed.UserIndex, seed.ResourceIdx = g.selectActivityPattern(currentTime, i)

			// 異常パターンの配置判定
			seed.Pattern = g.determineAnomalyPattern(currentTime, i, anomalyRatio)

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
	// 基本レート: 毎秒10件
	baseRate := 10.0

	// 時間帯補正
	timeMultiplier := map[int]float64{
		0: 0.1, 1: 0.05, 2: 0.05, 3: 0.05, 4: 0.05, 5: 0.1, // 深夜
		6: 0.2, 7: 0.4, 8: 0.8, // 朝
		9: 1.2, 10: 1.5, 11: 1.8, 12: 1.0, // 午前〜昼
		13: 0.8, 14: 1.3, 15: 1.6, 16: 1.4, 17: 1.1, 18: 0.9, // 午後
		19: 0.6, 20: 0.4, 21: 0.3, 22: 0.2, 23: 0.15, // 夜
	}

	// 曜日補正
	weekdayMultiplier := map[time.Weekday]float64{
		time.Monday: 1.0, time.Tuesday: 1.1, time.Wednesday: 1.2,
		time.Thursday: 1.1, time.Friday: 0.9,
		time.Saturday: 0.3, time.Sunday: 0.2,
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

	// 実例1: 夜間の管理者による大量学習データダウンロード
	if (hour >= 18 || hour <= 9) && rand.Float64() < baseAnomalyProb*0.3 {
		return logcore.PatternExample1NightAdminDownload
	}

	// 実例2: anyone with link設定ミスによる外部流出（まとまって発生）
	if hour >= 10 && hour <= 16 {
		// 15分間隔でバーストパターン（外部からの自動アクセス）
		if t.Minute()%15 < 3 && rand.Float64() < baseAnomalyProb*0.5 {
			return logcore.PatternExample2ExternalLinkAccess
		}
	}

	// 実例3: VPN脆弱性経由の水平移動攻撃（業務時間内の怪しい動作）
	if hour >= 9 && hour <= 18 && rand.Float64() < baseAnomalyProb*0.2 {
		return logcore.PatternExample3VpnLateralMovement
	}

	// 一般的な軽微異常
	if rand.Float64() < baseAnomalyProb*0.3 {
		return uint8(int(logcore.PatternTimeAnomaly) + rand.Intn(2)) // Pattern 4-5
	}

	return logcore.PatternNormal
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
