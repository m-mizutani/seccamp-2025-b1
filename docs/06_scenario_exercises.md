# シナリオベース検知ルール実装演習

## 🎯 実践シナリオ演習（3つから選択）

参加者は以下の3つのシナリオから1つを選択して、検知ルールを実装します。

### シナリオ選択ガイド

| シナリオ | 難易度 | 主要技術 | 想定業界 |
|----------|--------|----------|----------|
| **ネットワーク異常通信** | ⭐⭐☆ | VPC Flow Logs分析 | 製造業、金融 |
| **DNS異常クエリ** | ⭐⭐⭐ | 正規表現、統計分析 | IT企業、SaaS |
| **特権エスカレーション** | ⭐☆☆ | CloudTrail分析 | 全業界 |

## 🌐 シナリオ1：ネットワーク異常通信検知（VPC Flow Logs活用）

### 背景とビジネスケース
```
🏭 製造業での実例
課題: 工場内IoTデバイスからの異常な外部通信
リスク: 産業制御システムへの侵入、設計図データの窃取
要件: 内部から外部への大量データ転送の早期検知
```

### 検知対象
- **大量データ転送**: 短時間での大容量通信
- **異常な通信先**: 通常とは異なる外部IPアドレス
- **非業務時間の通信**: 夜間・休日の大量通信
- **異常なポート使用**: 一般的でないポートでの通信

### 実装SQLクエリ（完全版）
```sql
-- ネットワーク異常通信検知クエリ
WITH 
-- 1. 時間別の正常ベースライン（過去7日間）
baseline AS (
    SELECT 
        date_format(from_unixtime(time/1000), '%H') as hour,
        AVG(connection_info.bytes) as avg_bytes,
        STDDEV(connection_info.bytes) as stddev_bytes,
        PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY connection_info.bytes) as p95_bytes
    FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
    WHERE dt BETWEEN DATE_FORMAT(DATE_SUB(CURRENT_DATE, INTERVAL 7 DAY), '%Y-%m-%d') 
                 AND DATE_FORMAT(DATE_SUB(CURRENT_DATE, INTERVAL 1 DAY), '%Y-%m-%d')
        AND src_endpoint.ip LIKE '10.%'  -- 内部ネットワーク
        AND dst_endpoint.ip NOT LIKE '10.%'  -- 外部向け通信
        AND disposition = 'Allowed'
    GROUP BY date_format(from_unixtime(time/1000), '%H')
),

-- 2. 当日の通信集計（送信元IP・時間別）
current_traffic AS (
    SELECT 
        src_endpoint.ip as source_ip,
        dst_endpoint.ip as destination_ip,
        dst_endpoint.port as destination_port,
        connection_info.protocol_name as protocol,
        date_format(from_unixtime(time/1000), '%H') as hour,
        date_format(from_unixtime(time/1000), '%Y-%m-%d %H:%i') as time_window,
        SUM(connection_info.bytes) as total_bytes,
        COUNT(*) as connection_count,
        COUNT(DISTINCT dst_endpoint.ip) as unique_destinations,
        MIN(time) as first_connection,
        MAX(time) as last_connection
    FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
    WHERE dt = '2024-08-12'  -- 分析対象日
        AND src_endpoint.ip LIKE '10.%'  -- 内部ネットワーク
        AND dst_endpoint.ip NOT LIKE '10.%'  -- 外部向け通信
        AND disposition = 'Allowed'
        AND connection_info.bytes > 1000000  -- 1MB以上の通信のみ
    GROUP BY 
        src_endpoint.ip, 
        dst_endpoint.ip, 
        dst_endpoint.port,
        connection_info.protocol_name,
        date_format(from_unixtime(time/1000), '%H'),
        date_format(from_unixtime(time/1000), '%Y-%m-%d %H:%i')
),

-- 3. 異常判定
anomaly_detection AS (
    SELECT 
        ct.*,
        bl.avg_bytes,
        bl.stddev_bytes,
        bl.p95_bytes,
        CASE 
            WHEN bl.avg_bytes IS NULL THEN 'NO_BASELINE'
            WHEN ct.total_bytes > (bl.avg_bytes + (3 * bl.stddev_bytes)) THEN 'STATISTICAL_ANOMALY'
            WHEN ct.total_bytes > bl.p95_bytes * 10 THEN 'EXTREME_VOLUME'
            WHEN ct.total_bytes > 1000000000 THEN 'HIGH_VOLUME'  -- 1GB以上
            ELSE 'NORMAL'
        END as anomaly_type,
        ct.total_bytes / NULLIF(bl.avg_bytes, 0) as volume_ratio,
        -- 業務時間外判定（平日9-18時以外）
        CASE 
            WHEN CAST(ct.hour AS INTEGER) < 9 OR CAST(ct.hour AS INTEGER) >= 18 THEN 'OFF_HOURS'
            WHEN DAYOFWEEK(CURRENT_DATE) IN (1, 7) THEN 'WEEKEND'  -- 日曜日=1, 土曜日=7
            ELSE 'BUSINESS_HOURS'
        END as time_classification
    FROM current_traffic ct
    LEFT JOIN baseline bl ON ct.hour = bl.hour
),

-- 4. リスクスコア計算
risk_scoring AS (
    SELECT 
        *,
        (CASE anomaly_type
            WHEN 'EXTREME_VOLUME' THEN 100
            WHEN 'STATISTICAL_ANOMALY' THEN 80
            WHEN 'HIGH_VOLUME' THEN 60
            WHEN 'NO_BASELINE' THEN 40
            ELSE 0
        END) +
        (CASE time_classification
            WHEN 'OFF_HOURS' THEN 30
            WHEN 'WEEKEND' THEN 20
            ELSE 0
        END) +
        (CASE 
            WHEN unique_destinations > 10 THEN 20  -- 多数の宛先
            WHEN destination_port NOT IN (80, 443, 22, 21) THEN 15  -- 一般的でないポート
            ELSE 0
        END) as risk_score
    FROM anomaly_detection
    WHERE anomaly_type != 'NORMAL'
)

-- 5. 最終結果（アラート対象）
SELECT 
    source_ip,
    destination_ip,
    destination_port,
    protocol,
    time_window,
    total_bytes,
    connection_count,
    unique_destinations,
    anomaly_type,
    time_classification,
    risk_score,
    CASE 
        WHEN risk_score >= 80 THEN 'HIGH'
        WHEN risk_score >= 50 THEN 'MEDIUM'
        ELSE 'LOW'
    END as severity,
    from_unixtime(first_connection/1000) as first_connection_time,
    from_unixtime(last_connection/1000) as last_connection_time,
    volume_ratio
FROM risk_scoring
WHERE risk_score >= 50  -- Medium以上のリスクのみアラート
ORDER BY risk_score DESC, total_bytes DESC
LIMIT 100;
```

### Go実装例
```go
// internal/detector/network.go
package detector

import (
    "context"
    "fmt"
    "time"

    "github.com/aws/aws-sdk-go-v2/service/athena"
    "seccamp-detection/internal/alert"
)

func DetectNetworkAnomalies(ctx context.Context, client *athena.Client) ([]alert.SecurityAlert, error) {
    query := `
    -- 上記のSQLクエリ
    `
    
    results, err := executeAthenaQuery(ctx, client, query)
    if err != nil {
        return nil, fmt.Errorf("failed to execute network anomaly query: %w", err)
    }

    var alerts []alert.SecurityAlert
    for _, row := range results.ResultSet.Rows[1:] { // Skip header
        if len(row.Data) < 16 {
            continue
        }

        sourceIP := getStringValue(row.Data[0])
        destIP := getStringValue(row.Data[1])
        totalBytes := getStringValue(row.Data[5])
        severity := getStringValue(row.Data[12])
        riskScore := getStringValue(row.Data[11])

        alert := alert.SecurityAlert{
            Severity:    severity,
            Title:       "Network Anomaly Detected",
            Description: fmt.Sprintf("Large data transfer detected from %s to %s (%s bytes)", 
                        sourceIP, destIP, totalBytes),
            SourceData:  "VPC Flow Logs",
            Timestamp:   time.Now().Format(time.RFC3339),
            Details: map[string]interface{}{
                "source_ip":      sourceIP,
                "destination_ip": destIP,
                "total_bytes":    totalBytes,
                "risk_score":     riskScore,
            },
        }
        alerts = append(alerts, alert)
    }

    return alerts, nil
}
```

## 🔍 シナリオ2：DNS異常クエリ検知（DNS Logs活用）

### 背景とビジネスケース
```
💼 SaaS企業での実例
課題: マルウェア感染による外部C&Cサーバーとの通信
リスク: 顧客データの窃取、サービス停止
要件: DGA（ドメイン生成アルゴリズム）疑いドメインの検知
```

### 検知対象
- **DGA疑いドメイン**: 機械生成された長いランダム文字列
- **短命ドメイン**: TTL値が異常に短い
- **DNS Tunneling**: 異常に長いクエリ・大量リクエスト
- **既知の悪性ドメイン**: 脅威インテリジェンスとの照合

### 実装SQLクエリ（完全版）
```sql
-- DNS異常クエリ検知クエリ
WITH 
-- 1. ドメインエントロピー計算（ランダム性評価）
domain_entropy AS (
    SELECT 
        query.hostname as domain,
        LENGTH(query.hostname) as domain_length,
        -- 簡易エントロピー計算（文字種の多様性）
        LENGTH(query.hostname) - LENGTH(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
            UPPER(query.hostname), 
            'A', ''), 'B', ''), 'C', ''), 'D', ''), 'E', ''), 'F', ''), 'G', ''), 'H', ''), 'I', ''), 'J', '')) as char_diversity,
        -- 数字の比率
        (LENGTH(query.hostname) - LENGTH(REGEXP_REPLACE(query.hostname, '[0-9]', ''))) * 1.0 / LENGTH(query.hostname) as digit_ratio,
        -- 子音の連続性（読みやすさの逆指標）
        CASE WHEN query.hostname RLIKE '[bcdfghjklmnpqrstvwxyz]{4,}' THEN 1 ELSE 0 END as has_consonant_clusters
    FROM amazon_security_lake_table_us_east_1_route53_2_0
    WHERE dt = '2024-08-12'
        AND query.hostname IS NOT NULL
        AND LENGTH(query.hostname) >= 8
    GROUP BY query.hostname
),

-- 2. DNS クエリ統計
dns_stats AS (
    SELECT 
        query.hostname as domain,
        COUNT(*) as query_count,
        COUNT(DISTINCT src_endpoint.ip) as unique_clients,
        COUNT(DISTINCT query.type) as query_types_count,
        array_agg(DISTINCT query.type) as query_types,
        AVG(CAST(answers[1].ttl AS BIGINT)) as avg_ttl,
        MIN(CAST(answers[1].ttl AS BIGINT)) as min_ttl,
        MAX(CAST(answers[1].ttl AS BIGINT)) as max_ttl,
        -- 時間分散（短時間での集中アクセス検知）
        (MAX(time) - MIN(time)) / 1000 / 60 as time_span_minutes,
        MIN(time) as first_query,
        MAX(time) as last_query
    FROM amazon_security_lake_table_us_east_1_route53_2_0
    WHERE dt = '2024-08-12'
        AND query.hostname IS NOT NULL
        AND response_code = 'NOERROR'
    GROUP BY query.hostname
),

-- 3. DGA判定スコア計算
dga_scoring AS (
    SELECT 
        de.domain,
        de.domain_length,
        de.char_diversity,
        de.digit_ratio,
        de.has_consonant_clusters,
        ds.query_count,
        ds.unique_clients,
        ds.avg_ttl,
        ds.min_ttl,
        ds.time_span_minutes,
        -- DGAスコア計算（各要素に重み付け）
        (CASE 
            WHEN de.domain_length > 30 THEN 30
            WHEN de.domain_length > 20 THEN 20
            WHEN de.domain_length > 15 THEN 10
            ELSE 0
        END) +
        (CASE WHEN de.char_diversity > de.domain_length * 0.6 THEN 25 ELSE 0 END) +
        (CASE WHEN de.digit_ratio > 0.3 THEN 20 ELSE 0 END) +
        (CASE WHEN de.has_consonant_clusters = 1 THEN 15 ELSE 0 END) +
        (CASE 
            WHEN ds.query_count <= 5 THEN 20  -- 低頻度（DGAの特徴）
            WHEN ds.query_count <= 10 THEN 10
            ELSE 0
        END) +
        (CASE WHEN ds.unique_clients <= 2 THEN 15 ELSE 0 END) +
        (CASE 
            WHEN ds.min_ttl < 300 THEN 20  -- 5分未満の短いTTL
            WHEN ds.min_ttl < 900 THEN 10  -- 15分未満
            ELSE 0
        END) as dga_score,
        from_unixtime(ds.first_query/1000) as first_query_time,
        from_unixtime(ds.last_query/1000) as last_query_time
    FROM domain_entropy de
    JOIN dns_stats ds ON de.domain = ds.domain
),

-- 4. 既知脅威パターンとの照合
threat_patterns AS (
    SELECT 
        domain,
        CASE 
            -- 既知のDGAファミリーパターン
            WHEN domain RLIKE '^[a-z]{12,16}\.(com|net|org)$' THEN 'CONFICKER_LIKE'
            WHEN domain RLIKE '^[a-z0-9]{8,12}\.(tk|ml|ga|cf)$' THEN 'GOOTKIT_LIKE'
            WHEN domain RLIKE '^[0-9a-f]{32}\.' THEN 'MD5_LIKE'
            -- 疑わしいTLD
            WHEN domain RLIKE '\.(tk|ml|ga|cf|pw|cc)$' THEN 'SUSPICIOUS_TLD'
            -- IP直接指定
            WHEN domain RLIKE '^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$' THEN 'DIRECT_IP'
            ELSE 'UNKNOWN'
        END as threat_pattern,
        dga_score
    FROM dga_scoring
    WHERE dga_score >= 30  -- 一定スコア以上のみ
),

-- 5. 最終異常判定
final_anomalies AS (
    SELECT 
        ds.*,
        tp.threat_pattern,
        CASE 
            WHEN ds.dga_score >= 80 THEN 'HIGH'
            WHEN ds.dga_score >= 50 THEN 'MEDIUM'
            WHEN tp.threat_pattern != 'UNKNOWN' THEN 'MEDIUM'
            ELSE 'LOW'
        END as severity
    FROM dga_scoring ds
    LEFT JOIN threat_patterns tp ON ds.domain = tp.domain
    WHERE ds.dga_score >= 30 OR tp.threat_pattern != 'UNKNOWN'
)

-- 6. 結果出力
SELECT 
    domain,
    domain_length,
    query_count,
    unique_clients,
    dga_score,
    threat_pattern,
    severity,
    avg_ttl,
    time_span_minutes,
    first_query_time,
    last_query_time,
    -- アラートメッセージ生成
    CASE threat_pattern
        WHEN 'CONFICKER_LIKE' THEN 'Potential Conficker family DGA domain detected'
        WHEN 'GOOTKIT_LIKE' THEN 'Potential Gootkit family DGA domain detected'
        WHEN 'MD5_LIKE' THEN 'MD5-like domain pattern detected'
        WHEN 'SUSPICIOUS_TLD' THEN 'Domain using suspicious TLD detected'
        WHEN 'DIRECT_IP' THEN 'Direct IP address query detected'
        ELSE 'Machine-generated domain pattern detected'
    END as alert_message
FROM final_anomalies
ORDER BY dga_score DESC, query_count ASC
LIMIT 50;
```

## 👑 シナリオ3：特権エスカレーション検知（CloudTrail活用）

### 背景とビジネスケース
```
🏦 金融機関での実例
課題: 内部者による段階的な権限昇格
リスク: 機密データへの不正アクセス、金融取引の改ざん
要件: 短時間での複数権限操作の検知
```

### 検知対象
- **権限付与操作**: IAMポリシーのアタッチ・作成
- **ロール取得**: AssumeRole操作の異常パターン
- **管理者権限**: 管理者権限に関連する操作
- **時系列パターン**: 段階的権限昇格の検知

### 実装SQLクエリ（完全版）
```sql
-- 特権エスカレーション検知クエリ
WITH 
-- 1. 権限関連操作の抽出
privilege_operations AS (
    SELECT 
        time,
        actor.user.name as user_name,
        actor.user.type as user_type,
        api.operation as operation,
        api.service.name as service,
        src_endpoint.ip as source_ip,
        http_request.user_agent as user_agent,
        CASE api.operation
            WHEN 'AttachUserPolicy' THEN 90
            WHEN 'AttachRolePolicy' THEN 90
            WHEN 'PutUserPolicy' THEN 85
            WHEN 'PutRolePolicy' THEN 85
            WHEN 'CreateRole' THEN 80
            WHEN 'AssumeRole' THEN 70
            WHEN 'CreateUser' THEN 65
            WHEN 'AddUserToGroup' THEN 60
            WHEN 'CreateAccessKey' THEN 55
            WHEN 'UpdateAssumeRolePolicy' THEN 85
            WHEN 'DetachUserPolicy' THEN 40  -- 権限削除（リスクスコア低）
            WHEN 'DetachRolePolicy' THEN 40
            ELSE 0
        END as risk_weight,
        -- 管理者権限関連操作判定
        CASE 
            WHEN UPPER(CAST(api.request AS VARCHAR)) LIKE '%ADMIN%' THEN 1
            WHEN UPPER(CAST(api.request AS VARCHAR)) LIKE '%POWERUSER%' THEN 1
            WHEN UPPER(CAST(api.request AS VARCHAR)) LIKE '%FULLACCESS%' THEN 1
            ELSE 0
        END as is_admin_operation
    FROM amazon_security_lake_table_us_east_1_cloud_trail_mgmt_2_0
    WHERE dt = '2024-08-12'
        AND api.operation IN (
            'AttachUserPolicy', 'AttachRolePolicy', 
            'PutUserPolicy', 'PutRolePolicy',
            'AssumeRole', 'CreateRole', 'CreateUser',
            'AddUserToGroup', 'CreateAccessKey',
            'UpdateAssumeRolePolicy', 'DetachUserPolicy', 'DetachRolePolicy'
        )
        AND actor.user.name IS NOT NULL
        AND actor.user.name != 'root'  -- ルートユーザー除外
),

-- 2. ユーザー別操作集計（時間窓）
user_activity_windows AS (
    SELECT 
        user_name,
        user_type,
        source_ip,
        user_agent,
        -- 30分間隔での集計
        FLOOR(time / (30 * 60 * 1000)) * (30 * 60 * 1000) as time_window,
        COUNT(*) as operation_count,
        COUNT(DISTINCT operation) as unique_operations,
        SUM(risk_weight) as total_risk_score,
        SUM(is_admin_operation) as admin_operations_count,
        array_agg(operation ORDER BY time) as operations_sequence,
        MIN(time) as window_start,
        MAX(time) as window_end,
        (MAX(time) - MIN(time)) / 1000 / 60 as duration_minutes
    FROM privilege_operations
    GROUP BY 
        user_name, user_type, source_ip, user_agent,
        FLOOR(time / (30 * 60 * 1000))
),

-- 3. 異常パターンの検知
escalation_patterns AS (
    SELECT 
        *,
        -- 権限昇格パターンの判定
        CASE 
            WHEN operation_count >= 5 AND duration_minutes <= 30 THEN 'RAPID_ESCALATION'
            WHEN admin_operations_count >= 2 THEN 'ADMIN_PRIVILEGE_ATTEMPT'
            WHEN unique_operations >= 4 THEN 'DIVERSE_PRIVILEGE_OPS'
            WHEN total_risk_score >= 200 THEN 'HIGH_RISK_OPERATIONS'
            ELSE 'NORMAL'
        END as escalation_pattern,
        -- 時系列分析（段階的昇格の検知）
        CASE 
            WHEN array_join(operations_sequence, ',') LIKE '%CreateUser%AddUserToGroup%AttachUserPolicy%' THEN 'SYSTEMATIC_ESCALATION'
            WHEN array_join(operations_sequence, ',') LIKE '%CreateRole%AssumeRole%' THEN 'ROLE_BASED_ESCALATION'
            ELSE 'OTHER'
        END as escalation_method
    FROM user_activity_windows
    WHERE operation_count >= 3  -- 3回以上の操作
),

-- 4. 重大度とリスクスコア計算
risk_assessment AS (
    SELECT 
        *,
        (CASE escalation_pattern
            WHEN 'RAPID_ESCALATION' THEN 100
            WHEN 'ADMIN_PRIVILEGE_ATTEMPT' THEN 90
            WHEN 'DIVERSE_PRIVILEGE_OPS' THEN 80
            WHEN 'HIGH_RISK_OPERATIONS' THEN 70
            ELSE 0
        END) +
        (CASE escalation_method
            WHEN 'SYSTEMATIC_ESCALATION' THEN 50
            WHEN 'ROLE_BASED_ESCALATION' THEN 30
            ELSE 0
        END) +
        (CASE 
            WHEN duration_minutes <= 10 THEN 30  -- 10分以内の集中操作
            WHEN duration_minutes <= 30 THEN 20
            ELSE 0
        END) +
        (CASE user_type
            WHEN 'IAMUser' THEN 20  -- 通常ユーザーのほうがリスク高
            ELSE 0
        END) as final_risk_score
    FROM escalation_patterns
    WHERE escalation_pattern != 'NORMAL'
),

-- 5. 業務時間・地理的異常の考慮
contextual_analysis AS (
    SELECT 
        ra.*,
        -- 業務時間判定
        CASE 
            WHEN CAST(date_format(from_unixtime(window_start/1000), '%H') AS INTEGER) < 9 
                 OR CAST(date_format(from_unixtime(window_start/1000), '%H') AS INTEGER) >= 18 THEN 'OFF_HOURS'
            WHEN DAYOFWEEK(from_unixtime(window_start/1000)) IN (1, 7) THEN 'WEEKEND'
            ELSE 'BUSINESS_HOURS'
        END as time_context,
        -- 地理的異常の簡易判定（IPアドレス範囲）
        CASE 
            WHEN source_ip NOT LIKE '10.%' AND source_ip NOT LIKE '172.16.%' AND source_ip NOT LIKE '192.168.%' THEN 'EXTERNAL_IP'
            ELSE 'INTERNAL_IP'
        END as ip_context
    FROM risk_assessment
),

-- 6. 最終判定
final_assessment AS (
    SELECT 
        ca.*,
        ca.final_risk_score +
        (CASE time_context
            WHEN 'OFF_HOURS' THEN 25
            WHEN 'WEEKEND' THEN 15
            ELSE 0
        END) +
        (CASE ip_context
            WHEN 'EXTERNAL_IP' THEN 30
            ELSE 0
        END) as adjusted_risk_score,
        CASE 
            WHEN ca.final_risk_score >= 100 THEN 'HIGH'
            WHEN ca.final_risk_score >= 70 THEN 'MEDIUM'
            ELSE 'LOW'
        END as base_severity
    FROM contextual_analysis ca
)

-- 7. 最終結果
SELECT 
    user_name,
    user_type,
    source_ip,
    from_unixtime(window_start/1000) as escalation_start_time,
    from_unixtime(window_end/1000) as escalation_end_time,
    duration_minutes,
    operation_count,
    unique_operations,
    admin_operations_count,
    operations_sequence,
    escalation_pattern,
    escalation_method,
    time_context,
    ip_context,
    adjusted_risk_score,
    CASE 
        WHEN adjusted_risk_score >= 120 THEN 'HIGH'
        WHEN adjusted_risk_score >= 80 THEN 'MEDIUM'
        ELSE 'LOW'
    END as final_severity,
    -- アラートメッセージ
    CONCAT(
        'Potential privilege escalation detected for user: ', user_name,
        ' (Pattern: ', escalation_pattern, 
        ', Method: ', escalation_method,
        ', Operations: ', CAST(operation_count AS VARCHAR), ')'
    ) as alert_message
FROM final_assessment
WHERE adjusted_risk_score >= 60  -- Medium以上のリスクのみ
ORDER BY adjusted_risk_score DESC, duration_minutes ASC
LIMIT 30;
```

## 🔄 実装フロー（各シナリオ共通）

### 1. SQLクエリ作成
```bash
# sql/ディレクトリにクエリファイル作成
# シナリオ選択に応じて
touch sql/network_anomaly.sql     # シナリオ1
touch sql/dns_anomaly.sql         # シナリオ2  
touch sql/privilege_escalation.sql # シナリオ3
```

### 2. ローカルテスト
```bash
# Athenaコンソールまたは AWS CLIでクエリテスト
aws athena start-query-execution \
  --query-string "$(cat sql/network_anomaly.sql)" \
  --work-group security-lake-workgroup
```

### 3. Lambda統合
```go
// internal/detector/network.go に実装
func DetectNetworkAnomalies(ctx context.Context, client *athena.Client) ([]alert.SecurityAlert, error) {
    sqlBytes, err := os.ReadFile("sql/network_anomaly.sql")
    if err != nil {
        return nil, err
    }
    
    results, err := executeQuery(ctx, client, string(sqlBytes))
    // ... 結果解析とアラート生成
}
```

### 4. CI/CDデプロイ
```bash
git add .
git commit -m "Add network anomaly detection logic"
git push origin main

# GitHub Actions自動実行確認
gh run watch
```

### 5. 動作確認
```bash
# Lambda手動実行
aws lambda invoke \
  --function-name security-detection-lambda \
  --payload '{"scenario": "network"}' \
  response.json

# Slackチャンネルでアラート確認
```

## 🎓 評価ポイント

### 技術的完成度（40点）
- **SQLクエリの正確性**: 構文エラーなし、期待する結果取得
- **パフォーマンス**: パーティション活用、適切なインデックス使用
- **Go実装品質**: エラーハンドリング、ログ出力

### 実用性（30点）
- **誤検知率**: 正常操作を異常として検知しない
- **検知精度**: 実際の脅威を適切に検知
- **しきい値設定**: 統計的根拠に基づく適切な設定

### 運用考慮（20点）
- **スケーラビリティ**: 大量データ処理への対応
- **保守性**: コードの可読性、設定の変更しやすさ
- **監視**: ログ出力、メトリクス取得

### セキュリティ視点（10点）
- **脅威モデル理解**: 実際の攻撃手法への理解
- **ビジネス影響**: 検知する脅威のビジネスインパクト理解
- **継続改善**: 新しい脅威への対応方針

## 💡 実装のコツ

### 1. 段階的開発
```
Phase 1: 基本的なしきい値ベース検知
Phase 2: 統計的異常検知の追加
Phase 3: 時系列分析、パターンマッチング
Phase 4: 機械学習ベースの高度化
```

### 2. 誤検知削減
```sql
-- ホワイトリスト活用
WHERE source_ip NOT IN ('10.0.1.100', '10.0.1.101')  -- 管理サーバー除外
  AND user_name NOT LIKE '%service%'  -- サービスアカウント除外
```

### 3. パフォーマンス最適化
```sql
-- 適切なパーティション絞り込み
WHERE dt = '2024-08-12'  -- 必須
  AND hour BETWEEN '09' AND '18'  -- 業務時間のみ
```

---

**これで実装演習の準備完了です！選択したシナリオで検知ルールを実装してみましょう！** 🚀 