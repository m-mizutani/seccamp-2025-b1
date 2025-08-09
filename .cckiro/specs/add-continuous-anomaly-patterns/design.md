# 設計: 継続的な異常パターンの追加

## 概要
loggenツールに4つの新しい常時発生型異常パターンを追加し、multiplierオプションを廃止して定量的なログ生成を実現する設計。

## アーキテクチャ設計

### 1. パターン定数の追加
`internal/logcore/types.go`に新しいパターン定数を追加：

```go
const (
    // 既存のパターン
    PatternNormal                     uint8 = 0
    PatternExample1NightAdminDownload uint8 = 1
    PatternExample2ExternalLinkAccess uint8 = 2
    PatternExample3VpnLateralMovement uint8 = 3
    PatternTimeAnomaly                uint8 = 4
    PatternVolumeAnomaly              uint8 = 5
    
    // 新規追加パターン
    PatternExample4HighFreqAuthAttack      uint8 = 6
    PatternExample5RapidDataTheft          uint8 = 7
    PatternExample6MultiServiceProbing     uint8 = 8
    PatternExample7SimultaneousGeoAccess   uint8 = 9
)
```

### 2. 異常パターン生成ロジック

#### 2.1 共通設計方針
- 各パターンは独立した関数として実装
- 時間帯に関係なく24時間発生
- 発生頻度は秒単位で制御
- 特定のユーザー/IPを事前に決定して一貫性を保つ

#### 2.2 各パターンの実装設計

**PatternExample4HighFreqAuthAttack**
```go
type authAttackState struct {
    attackerIP    string
    targetUsers   []string
    lastAttempt   time.Time
}

// 状態管理：同一IPから継続的な攻撃を表現
// 実装方法：
// - 攻撃者IPは固定（例：203.0.113.99）
// - 1分間に3-5回、ランダムなユーザーへ認証試行
// - 5%の確率で成功（その後すぐログアウト）
```

**PatternExample5RapidDataTheft**
```go
type dataTheftState struct {
    theftUser     string
    theftIP       string
    downloadedFiles map[string]bool
}

// 状態管理：同一ユーザーによる継続的ダウンロード
// 実装方法：
// - 特定ユーザー（例：compromised@example.com）
// - 特定IP（例：198.51.100.99）
// - 1分間に10-15件のダウンロード
// - ファイル名は動的生成、重複回避
```

**PatternExample6MultiServiceProbing**
```go
type serviceProbingState struct {
    probingUser   string
    serviceOrder  []string
    currentIndex  int
}

// 状態管理：各種サービスへの順次アクセス
// 実装方法：
// - 特定ユーザー（例：infected@example.com）
// - サービスリスト：Drive, Calendar, Gmail, Admin
// - 1分間に3-5回、異なるサービスへアクセス
// - 70%は権限エラー（status_id = 2）
```

**PatternExample7SimultaneousGeoAccess**
```go
type geoAccessState struct {
    user          string
    country1      string
    country1IP    string
    country2      string
    country2IP    string
}

// 状態管理：2カ国からの同時アクセス
// 実装方法：
// - 特定ユーザー（例：travel@example.com）
// - 国1：JP（192.0.2.10）、国2：US（198.51.100.20）
// - 各国から1分間に2-3回アクセス
// - 交互にアクセスイベントを生成
```

### 3. ログ生成頻度の調整

#### 3.1 基本レート計算の変更
```go
// 現在の実装を変更して、より安定した生成レートを実現
func (g *Generator) getExpectedRate(hour int, weekday time.Weekday) float64 {
    // 基本レート: 毎秒200件（さらに増加）
    baseRate := 200.0
    
    // 時間帯補正はそのまま維持
    // ただし、最小値を引き上げて常時一定量のログを確保
    timeMultiplier := map[int]float64{
        0: 0.3, 1: 0.2, 2: 0.2, // 深夜でも最低40-60件/秒
        // ... 既存の値を調整
    }
}
```

#### 3.2 異常パターンの配分
```go
func (g *Generator) determineAnomalyPattern(t time.Time, sequenceInSecond int, anomalyRatio float64) uint8 {
    // 全体の5-10%を異常パターンに割り当て
    // 新しいパターンは常時発生型なので時間帯判定なし
    
    r := rand.Float64()
    
    // 既存パターン（時間帯依存）: 2-3%
    // 新規パターン（常時発生）: 3-7%
    
    if r < 0.01 {
        return g.selectContinuousAnomalyPattern()
    }
    
    // 既存のパターン判定ロジック
}

func (g *Generator) selectContinuousAnomalyPattern() uint8 {
    patterns := []uint8{
        PatternExample4HighFreqAuthAttack,
        PatternExample5RapidDataTheft,
        PatternExample6MultiServiceProbing,
        PatternExample7SimultaneousGeoAccess,
    }
    return patterns[rand.Intn(len(patterns))]
}
```

### 4. Multiplierオプションの廃止

#### 4.1 コマンドライン引数の変更
```go
// cmd/generate.go から multiplier フラグを削除
// multiplySeeds 関数を削除
// 関連するロジックをすべて削除
```

#### 4.2 S3アップロード時のファイル名
```go
// 後方互換性のため、ファイル名は変更しない
// binary-compressed形式の場合は従来通り固定名
case "binary-compressed":
    filename = fmt.Sprintf("large-seed.bin.gz")  // 既存のまま維持
```

### 5. 統計情報の拡張

```go
// updateStats関数を拡張
template.Metadata.AnomalyStats["example4_auth"] = g.countPattern(template.LogSeeds, PatternExample4HighFreqAuthAttack)
template.Metadata.AnomalyStats["example5_theft"] = g.countPattern(template.LogSeeds, PatternExample5RapidDataTheft)
template.Metadata.AnomalyStats["example6_probe"] = g.countPattern(template.LogSeeds, PatternExample6MultiServiceProbing)
template.Metadata.AnomalyStats["example7_geo"] = g.countPattern(template.LogSeeds, PatternExample7SimultaneousGeoAccess)
```

## ドキュメント更新設計

### 1. docs/06_lambda_implementation_and_detection_rules.md
- 既存の3つの課題の後に4つの新規課題を追加
- 各課題は新しい異常パターンに対応
- フォーマットは既存課題と同一（ヒント3段階、回答例は隠す）

### 2. loggen使用方法ドキュメント
- multiplierオプションの削除を明記
- 新しい異常パターンの説明を追加
- 生成されるログ量の目安を更新（定量化）

## テスト設計

### 1. 単体テスト
- 各異常パターン生成関数のテスト
- 発生頻度の検証（1分、5分単位）
- 状態管理の一貫性確認

### 2. 統合テスト
- 24時間分のログ生成
- 異常パターンの分布確認（5-10%）
- 各パターンの特徴が正しく生成されているか確認

### 3. 異常パターン出現率テスト
```go
// TestContinuousAnomalyPatterns
// 生成されたseedファイルを読み込んで各異常パターンの出現を検証
func TestContinuousAnomalyPatterns(t *testing.T) {
    // 1. テスト用の1日分のseed生成
    // 2. 生成されたLogSeedを時間窓ごとに分析
    // 3. 各パターンの出現頻度を計測
    
    // 検証項目：
    // - Pattern4: 1分あたり3-5回の出現
    // - Pattern5: 1分あたり10-15回の出現
    // - Pattern6: 1分あたり3-5回の出現
    // - Pattern7: 1分あたり4-6回の出現（2カ国合計）
    // - 全体の異常率: 5-10%
}
```

## リスクと対策

### 1. パフォーマンスへの影響
- リスク：ログ生成量増加による処理時間増大
- 対策：並列処理の最適化、バッファリングの改善

### 2. 既存機能への影響
- リスク：既存パターンの動作変更
- 対策：既存コードは最小限の変更に留める

### 3. 検知難易度のバランス
- リスク：異常が見つけやすすぎる/難しすぎる
- 対策：発生頻度の細かい調整、テストでの検証