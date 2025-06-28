# Security Lake基礎クエリとログ構造理解

## 📊 対象ログソースの構造理解

### VPC Flow Logs：ネットワーク通信の可視化

#### OCSF形式でのVPC Flow Logs構造
```json
{
  "metadata": {
    "version": "1.0.0",
    "product": {
      "name": "Amazon VPC Flow Logs",
      "vendor_name": "Amazon Web Services"
    }
  },
  "time": 1691836800000,
  "class_name": "Network Activity",
  "activity_id": 5,
  "severity_id": 1,
  "src_endpoint": {
    "ip": "10.0.1.100",
    "port": 443,
    "vpc_uid": "vpc-12345678"
  },
  "dst_endpoint": {
    "ip": "203.0.113.10",
    "port": 80
  },
  "connection_info": {
    "protocol_name": "TCP",
    "protocol_num": 6,
    "bytes": 2048,
    "packets": 15
  },
  "disposition": "Allowed",
  "disposition_id": 1,
  "traffic": {
    "bytes": 2048,
    "packets": 15
  }
}
```

#### キーフィールドの説明

| フィールド | 説明 | 分析での活用 |
|------------|------|--------------|
| `src_endpoint.ip` | 送信元IPアドレス | 内部ホストの特定、通信パターン分析 |
| `dst_endpoint.ip` | 宛先IPアドレス | 外部通信先の特定、脅威IOCとの照合 |
| `connection_info.bytes` | 転送データ量 | 大量データ転送の検知 |
| `disposition` | 通信の許可/拒否 | ブロックされた通信の分析 |
| `time` | 通信時刻 | 時間パターン分析 |

### DNS Logs：ドメイン解決パターンの分析

#### OCSF形式でのDNS Logs構造
```json
{
  "metadata": {
    "version": "1.0.0",
    "product": {
      "name": "Amazon Route 53 Resolver",
      "vendor_name": "Amazon Web Services"
    }
  },
  "time": 1691836800000,
  "class_name": "DNS Activity", 
  "activity_id": 2,
  "severity_id": 1,
  "query": {
    "hostname": "malicious-domain.example.com",
    "type": "A",
    "class": "IN"
  },
  "answers": [
    {
      "rdata": "192.0.2.100",
      "type": "A",
      "ttl": 300
    }
  ],
  "src_endpoint": {
    "ip": "10.0.1.50",
    "vpc_uid": "vpc-12345678"
  },
  "response_code": "NOERROR",
  "response_code_id": 0
}
```

#### DNS分析の重要ポイント

| 分析項目 | 検知内容 | 実装例 |
|----------|----------|--------|
| **DGA検知** | 機械生成ドメイン | ランダム文字列パターン、エントロピー分析 |
| **C&C通信** | 既知悪性ドメイン | 脅威インテリジェンスとの照合 |
| **データ漏洩** | DNS Tunneling | 異常に長いクエリ、高頻度クエリ |
| **フィッシング** | 類似ドメイン | ブランド名との類似度分析 |

### CloudTrail：API操作の監査

#### OCSF形式でのCloudTrail構造
```json
{
  "metadata": {
    "version": "1.0.0", 
    "product": {
      "name": "AWS CloudTrail",
      "vendor_name": "Amazon Web Services"
    }
  },
  "time": 1691836800000,
  "class_name": "API Activity",
  "activity_id": 3,
  "severity_id": 2,
  "api": {
    "operation": "AssumeRole",
    "service": {
      "name": "sts"
    },
    "request": {
      "uid": "arn:aws:iam::123456789012:role/PowerUser"
    }
  },
  "actor": {
    "user": {
      "type": "IAMUser",
      "name": "admin-user",
      "uid": "AIDACKCEVSQ6C2EXAMPLE"
    },
    "session": {
      "mfa": true,
      "issuer": "arn:aws:iam::123456789012:user/admin-user"
    }
  },
  "src_endpoint": {
    "ip": "203.0.113.50"
  },
  "http_request": {
    "user_agent": "aws-cli/2.0.0"
  }
}
```

#### CloudTrail分析の着眼点

| 監視項目 | リスクレベル | 検知例 |
|----------|--------------|--------|
| **特権操作** | 高 | AssumeRole、AttachUserPolicy |
| **リソース削除** | 高 | DeleteRole、DeleteBucket |
| **設定変更** | 中 | PutBucketPolicy、ModifyDBInstance |
| **大量操作** | 中 | 短時間での同一操作の繰り返し |

## 🔍 OCSF基礎クエリ練習

### 基本的なSELECT文

#### 1. データ確認クエリ
```sql
-- VPC Flow Logsの基本確認
SELECT 
    time,
    src_endpoint.ip as source_ip,
    dst_endpoint.ip as destination_ip,
    connection_info.bytes as bytes_transferred,
    disposition
FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
WHERE dt = '2024-08-12'
LIMIT 10;
```

#### 2. DNS Logsの基本確認
```sql
-- DNS クエリの確認
SELECT 
    time,
    src_endpoint.ip as client_ip,
    query.hostname as queried_domain,
    query.type as query_type,
    response_code
FROM amazon_security_lake_table_us_east_1_route53_2_0  
WHERE dt = '2024-08-12'
LIMIT 10;
```

#### 3. CloudTrail の基本確認
```sql
-- API操作の確認
SELECT 
    time,
    actor.user.name as user_name,
    api.operation as operation,
    api.service.name as service,
    src_endpoint.ip as source_ip
FROM amazon_security_lake_table_us_east_1_cloud_trail_mgmt_2_0
WHERE dt = '2024-08-12'
LIMIT 10;
```

### データ集計とフィルタリング

#### 通信量による集計
```sql
-- 時間帯別通信量の集計
SELECT 
    date_format(from_unixtime(time/1000), '%H') as hour,
    COUNT(*) as connection_count,
    SUM(connection_info.bytes) as total_bytes,
    AVG(connection_info.bytes) as avg_bytes
FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
WHERE dt = '2024-08-12'
    AND disposition = 'Allowed'
GROUP BY date_format(from_unixtime(time/1000), '%H')
ORDER BY hour;
```

#### 送信元IP別通信先集計
```sql
-- 送信元IPごとの外部通信先数
SELECT 
    src_endpoint.ip as source_ip,
    COUNT(DISTINCT dst_endpoint.ip) as unique_destinations,
    COUNT(*) as total_connections,
    SUM(connection_info.bytes) as total_bytes
FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
WHERE dt = '2024-08-12'
    AND src_endpoint.ip LIKE '10.%'  -- 内部IPアドレス
    AND dst_endpoint.ip NOT LIKE '10.%'  -- 外部IPアドレス
GROUP BY src_endpoint.ip
HAVING unique_destinations > 10  -- 10以上の宛先と通信
ORDER BY unique_destinations DESC;
```

#### DNS解決パターン分析
```sql
-- ドメイン別クエリ頻度
SELECT 
    query.hostname as domain,
    COUNT(*) as query_count,
    COUNT(DISTINCT src_endpoint.ip) as unique_clients,
    array_agg(DISTINCT query.type) as query_types
FROM amazon_security_lake_table_us_east_1_route53_2_0
WHERE dt = '2024-08-12'
GROUP BY query.hostname
HAVING query_count > 100  -- 高頻度クエリ
ORDER BY query_count DESC;
```

### 時間ベースの分析

#### 業務時間外のアクティビティ
```sql
-- 業務時間外のAPI操作（日本時間 22:00-06:00）
SELECT 
    actor.user.name as user_name,
    api.operation as operation,
    api.service.name as service,
    COUNT(*) as operation_count,
    min(time) as first_operation,
    max(time) as last_operation
FROM amazon_security_lake_table_us_east_1_cloud_trail_mgmt_2_0
WHERE dt = '2024-08-12'
    AND (
        date_format(from_unixtime(time/1000), '%H') >= '22' OR 
        date_format(from_unixtime(time/1000), '%H') < '06'
    )
    AND actor.user.name IS NOT NULL
GROUP BY actor.user.name, api.operation, api.service.name
ORDER BY operation_count DESC;
```

#### 異常な時間パターン
```sql
-- 通常と異なる時間帯での大量通信
WITH hourly_baseline AS (
    SELECT 
        date_format(from_unixtime(time/1000), '%H') as hour,
        AVG(connection_info.bytes) as avg_bytes,
        STDDEV(connection_info.bytes) as stddev_bytes
    FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
    WHERE dt BETWEEN '2024-08-05' AND '2024-08-11'  -- 過去1週間のベースライン
    GROUP BY date_format(from_unixtime(time/1000), '%H')
)
SELECT 
    v.src_endpoint.ip as source_ip,
    date_format(from_unixtime(v.time/1000), '%H') as hour,
    v.connection_info.bytes as current_bytes,
    b.avg_bytes + (2 * b.stddev_bytes) as threshold
FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0 v
JOIN hourly_baseline b ON date_format(from_unixtime(v.time/1000), '%H') = b.hour
WHERE v.dt = '2024-08-12'
    AND v.connection_info.bytes > (b.avg_bytes + (2 * b.stddev_bytes))  -- 2σ以上の異常
ORDER BY v.connection_info.bytes DESC;
```

## 📈 パフォーマンス最適化のテクニック

### パーティション活用
```sql
-- パーティション絞り込みによる高速化
SELECT src_endpoint.ip, COUNT(*)
FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
WHERE dt = '2024-08-12'  -- パーティション指定（必須）
    AND hour = '10'      -- さらに時間で絞り込み
    AND src_endpoint.ip LIKE '10.%'
GROUP BY src_endpoint.ip;
```

### LIMIT句の効果的な使用
```sql
-- 大量データから上位結果のみ取得
SELECT 
    dst_endpoint.ip,
    SUM(connection_info.bytes) as total_bytes
FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
WHERE dt = '2024-08-12'
GROUP BY dst_endpoint.ip
ORDER BY total_bytes DESC
LIMIT 100;  -- 上位100件のみ
```

### カラムストアの活用
```sql
-- 必要なカラムのみ選択（I/O削減）
SELECT 
    src_endpoint.ip,
    dst_endpoint.ip,
    connection_info.bytes
FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
WHERE dt = '2024-08-12'
-- 不要なカラム（metadata等）は選択しない
```

## 🎯 実習クエリテンプレート

### ネットワーク異常検知用クエリ
```sql
-- 大量データ転送検知
WITH transfer_stats AS (
    SELECT 
        src_endpoint.ip as source_ip,
        SUM(connection_info.bytes) as total_bytes,
        COUNT(*) as connection_count,
        COUNT(DISTINCT dst_endpoint.ip) as unique_destinations
    FROM amazon_security_lake_table_us_east_1_vpc_flow_2_0
    WHERE dt = '2024-08-12'
        AND src_endpoint.ip LIKE '10.%'  -- 内部ネットワーク
        AND connection_info.bytes > 1000000  -- 1MB以上の通信
    GROUP BY src_endpoint.ip
)
SELECT *
FROM transfer_stats
WHERE total_bytes > 1000000000  -- 1GB以上の転送
ORDER BY total_bytes DESC;
```

### DNS異常検知用クエリ
```sql
-- DGA疑いドメイン検知
SELECT 
    query.hostname as domain,
    LENGTH(query.hostname) as domain_length,
    COUNT(*) as query_count,
    COUNT(DISTINCT src_endpoint.ip) as unique_clients
FROM amazon_security_lake_table_us_east_1_route53_2_0
WHERE dt = '2024-08-12'
    AND LENGTH(query.hostname) > 20  -- 長いドメイン名
    AND query.hostname RLIKE '^[a-z0-9]{10,}\..*'  -- ランダム文字列パターン
GROUP BY query.hostname, LENGTH(query.hostname)
HAVING query_count < 5  -- 低頻度クエリ（DGAの特徴）
ORDER BY domain_length DESC;
```

### 特権エスカレーション検知用クエリ
```sql
-- 短時間での権限変更検知
SELECT 
    actor.user.name as user_name,
    COUNT(*) as privilege_operations,
    array_agg(api.operation) as operations,
    min(time) as first_operation,
    max(time) as last_operation,
    (max(time) - min(time))/1000/60 as duration_minutes
FROM amazon_security_lake_table_us_east_1_cloud_trail_mgmt_2_0
WHERE dt = '2024-08-12'
    AND api.operation IN (
        'AttachUserPolicy', 'AttachRolePolicy', 
        'PutUserPolicy', 'PutRolePolicy',
        'AssumeRole', 'CreateRole'
    )
GROUP BY actor.user.name
HAVING COUNT(*) >= 3  -- 3回以上の権限操作
    AND (max(time) - min(time))/1000/60 < 30  -- 30分以内
ORDER BY privilege_operations DESC;
```

## 💡 クエリ作成のベストプラクティス

### 1. 段階的な開発
```sql
-- Step 1: 基本クエリで動作確認
SELECT * FROM table_name WHERE dt = '2024-08-12' LIMIT 10;

-- Step 2: フィルタ条件追加
SELECT * FROM table_name 
WHERE dt = '2024-08-12' AND src_endpoint.ip LIKE '10.%';

-- Step 3: 集計処理追加
SELECT src_endpoint.ip, COUNT(*) 
FROM table_name 
WHERE dt = '2024-08-12' AND src_endpoint.ip LIKE '10.%'
GROUP BY src_endpoint.ip;
```

### 2. パフォーマンス考慮
- **パーティション**: 必ず`dt`で絞り込み
- **LIMIT**: 大量データ処理時は件数制限
- **インデックス**: 頻繁に使用するカラムを意識

### 3. 誤検知削減
- **ベースライン**: 正常パターンとの比較
- **しきい値**: 統計的根拠に基づく設定
- **除外条件**: 既知の正常パターンを除外

---

**次回**: 実際のシナリオベース検知ルール実装に進みます！ 