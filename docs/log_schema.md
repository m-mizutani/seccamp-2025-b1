# Security Lake ログスキーマ・リファレンス（分析者向け）

このドキュメントでは、AWS Athena で Security Lake のデータを分析する際に使用するフィールドについて説明します。

## 概要

Security Lake には Google Workspace のログが OCSF (Open Cybersecurity Schema Framework) 形式で保存されています。OCSF は異なるセキュリティ製品のログを標準化するフレームワークで、統一的な分析を可能にします。

## 完全フィールド一覧

### 基本分類フィールド

| カラム名 | 型 | 説明 | 主な値 |
|---------|---|------|--------|
| `category_uid` | int | カテゴリ識別子 | `6`=Application Activity |
| `class_uid` | int | クラス識別子 | `6001`=Web Resources Activity |
| `type_uid` | int | タイプ識別子（class_uid × 100 + activity_id） | `600101`=Create<br>`600102`=Read<br>`600103`=Update<br>`600104`=Delete<br>`600107`=Export<br>`600108`=Share |
| `activity_id` | int | アクティビティの種類 | `1`=Create（作成）<br>`2`=Read（読み取り）<br>`3`=Update（更新）<br>`4`=Delete（削除）<br>`7`=Export（ダウンロード）<br>`8`=Share（共有） |
| `severity_id` | int | イベントの重要度 | `1`=Informational（情報）<br>`2`=Low（低）<br>`3`=Medium（中）<br>`4`=High（高） |
| `status_id` | int | 操作の成功/失敗 | `1`=Success（成功）<br>`2`=Failure（失敗） |
| `confidence` | int | イベントの信頼度（オプション） | `0`～`100` |

### 時刻フィールド

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `time` | bigint | イベント発生時刻（Unix ミリ秒） | `1640995200000` |
| `start_time` | bigint | イベント開始時刻（オプション） | Unix ミリ秒 |
| `end_time` | bigint | イベント終了時刻（オプション） | Unix ミリ秒 |

### アクター（操作実行者）情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `actor.user.name` | string | ユーザー名 | `田中太郎` |
| `actor.user.uid` | string | ユーザー一意識別子 | `114511147312345678901` |
| `actor.user.email_addr` | string | ユーザーのメールアドレス | `user@example.com` |
| `actor.user.domain` | string | ユーザーのドメイン | `example.com` |
| `actor.user.type_id` | int | ユーザーの種別 | `1`=一般ユーザー<br>`2`=管理者 |
| `actor.user.groups` | array<string> | ユーザーが所属するグループ（オプション） | `["sales", "marketing"]` |
| `actor.session.uid` | string | セッション識別子（オプション） | セッションID |
| `actor.session.created_time` | bigint | セッション作成時刻（オプション） | Unix ミリ秒 |
| `actor.session.exp_time` | bigint | セッション有効期限（オプション） | Unix ミリ秒 |
| `actor.app_name` | string | アプリケーション名（オプション） | `Google Chrome` |
| `actor.app_uid` | string | アプリケーション識別子（オプション） | アプリID |

### API 情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `api.service.name` | string | 使用されたサービス | `Google Identity`<br>`Google Drive API`<br>`Google Admin API` |
| `api.service.version` | string | サービスバージョン（オプション） | `v1`, `v2` |
| `api.operation` | string | 実行された操作 | `login`, `download`, `share`, etc. |
| `api.request.uid` | string | リクエスト識別子 | リクエストID |
| `api.response.code` | int | HTTP レスポンスコード（オプション） | `200`=成功<br>`403`=アクセス拒否<br>`404`=見つからない<br>`500`=サーバーエラー |
| `api.response.message` | string | レスポンスメッセージ（オプション） | `OK`, `Forbidden` |

### クラウド環境情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `cloud.provider` | string | クラウドプロバイダー | `Google Workspace` |
| `cloud.account.uid` | string | アカウント識別子 | `C03az79cb` |
| `cloud.account.name` | string | アカウント名（オプション） | `example-company` |
| `cloud.org.name` | string | 組織名（オプション） | `Example Corp` |
| `cloud.org.uid` | string | 組織識別子（オプション） | 組織ID |
| `cloud.cloud_region` | string | クラウドリージョン（オプション） | `asia-northeast1` |

### アクセス元情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `src_endpoint.ip` | string | アクセス元 IP アドレス | `192.0.2.1` |
| `src_endpoint.hostname` | string | アクセス元ホスト名（オプション） | `workstation-01.example.com` |
| `src_endpoint.location.country` | string | アクセス元国（オプション） | `JP`, `US` |
| `src_endpoint.location.src_region` | string | アクセス元地域（オプション） | `Tokyo`, `California` |
| `src_endpoint.location.city` | string | アクセス元都市（オプション） | `千代田区`, `San Francisco` |

### Web リソース情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `web_resources` | array | アクセスされたリソースの配列 | 複数のリソース情報 |
| `web_resources[n].name` | string | リソース名（オプション） | `重要資料.xlsx` |
| `web_resources[n].uid` | string | リソース識別子（オプション） | ドキュメントID |
| `web_resources[n].type` | string | リソースタイプ（オプション） | `document`, `spreadsheet`, `folder` |
| `web_resources[n].url_string` | string | リソースURL（オプション） | `https://docs.google.com/...` |
| `web_resources[n].data.classification` | string | データ分類（オプション） | `confidential`, `public` |

### メタデータ

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `metadata.uid` | string | メタデータ識別子（オプション） | メタデータID |
| `metadata.correlation_uid` | string | 相関識別子（オプション） | 関連イベントのグループID |
| `metadata.labels` | array<string> | イベントのラベル（タグ） | `["event_name:download", "risk:high"]` |
| `metadata.original_time` | string | 元のタイムスタンプ（オプション） | `2024-01-01T00:00:00Z` |
| `metadata.processed` | bigint | 処理時刻（オプション） | Unix ミリ秒 |
| `metadata.product_name` | string | 製品名（オプション） | `Google Workspace` |
| `metadata.version` | string | バージョン（オプション） | `1.0` |

### 監視対象情報（Observables）

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `observables` | array | 監視対象の配列（オプション） | IOC（Indicator of Compromise）など |
| `observables[n].name` | string | 監視対象の名前 | `suspicious_ip` |
| `observables[n].type` | string | 監視対象のタイプ | `ip_address`, `domain`, `hash` |
| `observables[n].value` | string | 監視対象の値 | `192.0.2.100` |

### パーティション用フィールド

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `aws_region` | string | AWS リージョン | `ap-northeast-1` |
| `account_id` | string | AWS アカウント ID | `123456789012` |
| `event_hour` | string | イベント時刻（時間単位） | `2024-01-01-09` |
| `eventday` | string | イベント日付（パーティション） | `20240101` |
| `region` | string | リージョン（パーティション用） | `ap-northeast-1` |
| `accountid` | string | アカウントID（パーティション用） | `123456789012` |

## よく使うクエリパターン

### 時刻の扱い方

```sql
-- Unix ミリ秒を人間が読める形式に変換
SELECT 
    from_unixtime(time/1000) as event_time,
    actor.user.email_addr
FROM your_table

-- 日本時間での表示
SELECT 
    from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo' as event_time_jst
FROM your_table

-- 時刻範囲での検索
SELECT *
FROM your_table
WHERE time BETWEEN 1640995200000 AND 1641081600000
```

### アクティビティ別の集計

```sql
-- アクティビティタイプ別のイベント数
SELECT 
    activity_id,
    CASE activity_id
        WHEN 1 THEN 'Create'
        WHEN 2 THEN 'Read'
        WHEN 3 THEN 'Update'
        WHEN 4 THEN 'Delete'
        WHEN 7 THEN 'Export/Download'
        WHEN 8 THEN 'Share'
        ELSE 'Other'
    END as activity_name,
    COUNT(*) as event_count
FROM your_table
GROUP BY activity_id
ORDER BY event_count DESC
```

### ユーザー別アクティビティ分析

```sql
-- ユーザー別のアクティビティサマリー
SELECT 
    actor.user.email_addr,
    actor.user.type_id,
    COUNT(*) as total_events,
    COUNT(DISTINCT activity_id) as unique_activities,
    COUNT(DISTINCT DATE(from_unixtime(time/1000))) as active_days
FROM your_table
WHERE eventday BETWEEN '20240801' AND '20240831'
GROUP BY actor.user.email_addr, actor.user.type_id
ORDER BY total_events DESC
```

### 失敗イベントの検索

```sql
-- ログイン失敗の検索
SELECT 
    actor.user.email_addr,
    src_endpoint.ip,
    src_endpoint.location.country,
    from_unixtime(time/1000) as failure_time,
    api.response.message
FROM your_table
WHERE status_id = 2  -- Failure
    AND api.service.name = 'Google Identity'
    AND eventday = '20240801'
ORDER BY time DESC

-- 失敗率の高いユーザー
SELECT 
    actor.user.email_addr,
    COUNT(*) as total_attempts,
    SUM(CASE WHEN status_id = 2 THEN 1 ELSE 0 END) as failures,
    CAST(SUM(CASE WHEN status_id = 2 THEN 1 ELSE 0 END) AS DOUBLE) / COUNT(*) * 100 as failure_rate
FROM your_table
WHERE api.operation = 'login'
GROUP BY actor.user.email_addr
HAVING failure_rate > 20
ORDER BY failure_rate DESC
```

### リソースアクセスの分析

```sql
-- ダウンロードされたファイルの一覧
SELECT 
    actor.user.email_addr,
    web_resources[1].name as file_name,
    web_resources[1].type as file_type,
    from_unixtime(time/1000) as download_time
FROM your_table
WHERE activity_id = 7  -- Export/Download
    AND CARDINALITY(web_resources) > 0
    AND eventday = '20240801'

-- 配列データの展開（UNNEST使用）
SELECT 
    actor.user.email_addr,
    resource.name as resource_name,
    resource.uid as resource_id,
    resource.type as resource_type,
    from_unixtime(time/1000) as access_time
FROM your_table
CROSS JOIN UNNEST(web_resources) AS t(resource)
WHERE activity_id IN (2, 7)  -- Read or Download
    AND eventday = '20240801'
```

### 管理者アクティビティの監視

```sql
-- 管理者による重要操作
SELECT 
    actor.user.email_addr,
    api.operation,
    severity_id,
    from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo' as event_time_jst
FROM your_table
WHERE actor.user.type_id = 2  -- Admin
    AND severity_id >= 3  -- Medium以上
    AND eventday = '20240801'
ORDER BY time DESC

-- 深夜の管理者操作を検出
SELECT 
    actor.user.email_addr,
    api.operation,
    from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo' as event_time_jst,
    EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') as hour_jst
FROM your_table
WHERE actor.user.type_id = 2  -- Admin
    AND (
        EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') >= 22
        OR EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') < 6
    )
    AND eventday = '20240801'
```

### 地理的分析

```sql
-- 国別アクセス統計
SELECT 
    src_endpoint.location.country,
    COUNT(*) as access_count,
    COUNT(DISTINCT actor.user.email_addr) as unique_users
FROM your_table
WHERE src_endpoint.location.country IS NOT NULL
    AND eventday BETWEEN '20240801' AND '20240831'
GROUP BY src_endpoint.location.country
ORDER BY access_count DESC

-- 異常な場所からのアクセス検知
WITH user_locations AS (
    SELECT 
        actor.user.email_addr,
        src_endpoint.location.country,
        COUNT(*) as access_count
    FROM your_table
    WHERE eventday BETWEEN '20240801' AND '20240831'
    GROUP BY actor.user.email_addr, src_endpoint.location.country
),
user_primary_location AS (
    SELECT 
        email_addr,
        country,
        ROW_NUMBER() OVER (PARTITION BY email_addr ORDER BY access_count DESC) as rank
    FROM user_locations
)
SELECT 
    t.actor.user.email_addr,
    t.src_endpoint.location.country as access_country,
    upl.country as primary_country,
    from_unixtime(t.time/1000) as access_time
FROM your_table t
JOIN user_primary_location upl ON t.actor.user.email_addr = upl.email_addr
WHERE upl.rank = 1
    AND t.src_endpoint.location.country != upl.country
    AND t.eventday = '20240831'
```

### メタデータとラベルの活用

```sql
-- 特定のラベルを持つイベントの検索
SELECT 
    actor.user.email_addr,
    api.operation,
    metadata.labels,
    from_unixtime(time/1000) as event_time
FROM your_table
WHERE CONTAINS(metadata.labels, 'risk:high')
    OR CONTAINS(metadata.labels, 'sensitive')

-- 相関IDによる関連イベントの追跡
WITH target_correlation AS (
    SELECT DISTINCT metadata.correlation_uid
    FROM your_table
    WHERE actor.user.email_addr = 'suspicious@example.com'
        AND eventday = '20240801'
)
SELECT 
    t.actor.user.email_addr,
    t.api.operation,
    from_unixtime(t.time/1000) as event_time
FROM your_table t
JOIN target_correlation tc ON t.metadata.correlation_uid = tc.correlation_uid
WHERE t.eventday = '20240801'
ORDER BY t.time
```

## 重要な値の意味

### activity_id（アクティビティID）

- **1 (Create)**: ファイルやリソースの新規作成
- **2 (Read)**: ファイルの閲覧、ログインなどのアクセス
- **3 (Update)**: 既存リソースの変更、設定の更新
- **4 (Delete)**: リソースの削除
- **7 (Export)**: ファイルのダウンロードやエクスポート
- **8 (Share)**: ファイルやフォルダの共有設定

### severity_id（重要度）

- **1 (Informational)**: 通常の操作、情報レベル
- **2 (Low)**: 軽微な異常や注意事項
- **3 (Medium)**: 要注意のイベント（権限変更、大量ダウンロードなど）
- **4 (High)**: 重要なセキュリティイベント（管理者権限での異常操作など）

### status_id（ステータス）

- **1 (Success)**: 操作が正常に完了
- **2 (Failure)**: 操作が失敗（アクセス拒否、認証失敗など）

### actor.user.type_id（ユーザータイプ）

- **1 (Regular User)**: 一般ユーザー
- **2 (Admin)**: 管理者権限を持つユーザー

### api.response.code（HTTPレスポンスコード）

- **200番台**: 成功
  - `200`: OK
  - `201`: Created
  - `204`: No Content
- **400番台**: クライアントエラー
  - `400`: Bad Request
  - `401`: Unauthorized
  - `403`: Forbidden
  - `404`: Not Found
- **500番台**: サーバーエラー
  - `500`: Internal Server Error
  - `503`: Service Unavailable

## パフォーマンスのベストプラクティス

### 1. パーティションを活用した効率的なクエリ

```sql
-- 良い例：パーティションを使用
SELECT *
FROM your_table
WHERE eventday = '20240801'  -- パーティションキーを指定
    AND actor.user.email_addr = 'user@example.com'

-- 悪い例：パーティションを使用しない
SELECT *
FROM your_table
WHERE actor.user.email_addr = 'user@example.com'  -- 全データをスキャン
```

### 2. 必要なカラムのみを選択

```sql
-- 良い例：必要なカラムのみ
SELECT 
    actor.user.email_addr,
    api.operation,
    time
FROM your_table
WHERE eventday = '20240801'

-- 悪い例：全カラムを選択
SELECT *
FROM your_table
WHERE eventday = '20240801'
```

### 3. 集計前のフィルタリング

```sql
-- 良い例：WHERE句でフィルタリング後に集計
SELECT 
    actor.user.email_addr,
    COUNT(*) as event_count
FROM your_table
WHERE eventday BETWEEN '20240801' AND '20240807'
    AND status_id = 1
GROUP BY actor.user.email_addr

-- 悪い例：集計後にHAVING句でフィルタリング
SELECT 
    actor.user.email_addr,
    COUNT(*) as event_count,
    SUM(CASE WHEN status_id = 1 THEN 1 ELSE 0 END) as success_count
FROM your_table
WHERE eventday BETWEEN '20240801' AND '20240807'
GROUP BY actor.user.email_addr
HAVING success_count > 0
```

## 注意事項

1. **時刻データ**: `time` フィールドは Unix ミリ秒で保存されているため、`from_unixtime(time/1000)` で変換が必要

2. **配列アクセス**: Athena では配列のインデックスは 1 から始まる（`web_resources[1]` が最初の要素）

3. **パーティション**: `eventday` を WHERE 句に含めることでクエリ性能が大幅に向上

4. **NULL 値**: オプショナルなフィールドは NULL の可能性があるため、適切な NULL チェックが必要

5. **大文字小文字**: Athena のクエリは大文字小文字を区別しないが、データ値は区別される

6. **タイムゾーン**: 時刻は UTC で保存されているため、日本時間で分析する場合は `AT TIME ZONE 'Asia/Tokyo'` を使用

7. **データ型の変換**: 必要に応じて CAST 関数を使用してデータ型を変換

8. **配列の展開**: 配列フィールドを分析する場合は UNNEST を使用して行に展開