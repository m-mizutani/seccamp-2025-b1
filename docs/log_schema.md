# Security Lake ログスキーマ・リファレンス（分析者向け）

このドキュメントでは、AWS Athena で Security Lake のデータを分析する際に使用するフィールドについて説明します。

## 概要

Security Lake には Google Workspace のログが OCSF (Open Cybersecurity Schema Framework) 形式で保存されています。OCSF は異なるセキュリティ製品のログを標準化するフレームワークで、統一的な分析を可能にします。

## 主要フィールド一覧

### 基本分類フィールド

| カラム名 | 型 | 説明 | 主な値 |
|---------|---|------|--------|
| `activity_id` | int | アクティビティの種類 | `1`=Create（作成）<br>`2`=Read（読み取り）<br>`3`=Update（更新）<br>`4`=Delete（削除）<br>`7`=Export（ダウンロード）<br>`8`=Share（共有） |
| `severity_id` | int | イベントの重要度 | `1`=Informational（情報）<br>`2`=Low（低）<br>`3`=Medium（中）<br>`4`=High（高） |
| `status_id` | int | 操作の成功/失敗 | `1`=Success（成功）<br>`2`=Failure（失敗） |
| `time` | bigint | イベント発生時刻（Unix ミリ秒） | 例: `1640995200000` |

### ユーザー情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `actor.user.email_addr` | string | 操作を実行したユーザーのメールアドレス | `user@example.com` |
| `actor.user.type_id` | int | ユーザーの種別 | `1`=一般ユーザー<br>`2`=管理者 |
| `actor.user.domain` | string | ユーザーのドメイン | `example.com` |

### API 情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `api.service.name` | string | 使用されたサービス | `Google Identity`<br>`Google Drive API`<br>`Google Admin API` |
| `api.operation` | string | 実行された操作 | `login`, `download`, `share`, etc. |
| `api.response.code` | int | HTTP レスポンスコード | `200`=成功<br>`403`=アクセス拒否 |

### アクセス元情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `src_endpoint.ip` | string | アクセス元 IP アドレス | `192.0.2.1` |

### リソース情報

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `web_resources` | array | アクセスされたリソースの配列 | ドキュメント情報など |
| `web_resources[1].name` | string | リソース名（配列の最初の要素） | `重要資料.xlsx` |
| `web_resources[1].uid` | string | リソースID | ドキュメントID |

### メタデータ

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `metadata.labels` | array | イベントのラベル（タグ） | `event_name:download` |
| `metadata.original_time` | string | 元のタイムスタンプ | `2024-01-01T00:00:00Z` |

### パーティション用フィールド

| カラム名 | 型 | 説明 | 例 |
|---------|---|------|-----|
| `eventday` | string | イベント日付（パーティション） | `20240101` |
| `region` | string | AWS リージョン | `ap-northeast-1` |
| `accountid` | string | AWS アカウント ID | `123456789012` |

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
```

### 失敗イベントの検索

```sql
-- ログイン失敗の検索
SELECT 
    actor.user.email_addr,
    src_endpoint.ip,
    from_unixtime(time/1000) as failure_time
FROM your_table
WHERE status_id = 2  -- Failure
    AND api.service.name = 'Google Identity'
```

### リソースアクセスの分析

```sql
-- ダウンロードされたファイルの一覧
SELECT 
    actor.user.email_addr,
    web_resources[1].name as file_name,
    from_unixtime(time/1000) as download_time
FROM your_table
WHERE activity_id = 7  -- Export/Download
    AND CARDINALITY(web_resources) > 0
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

- **1 (Informational)**: 通常の操作
- **2 (Low)**: 軽微な異常や注意事項
- **3 (Medium)**: 要注意のイベント（権限変更など）
- **4 (High)**: 重要なセキュリティイベント（管理者操作など）

### status_id（ステータス）

- **1 (Success)**: 操作が正常に完了
- **2 (Failure)**: 操作が失敗（アクセス拒否、認証失敗など）

### actor.user.type_id（ユーザータイプ）

- **1 (Regular User)**: 一般ユーザー
- **2 (Admin)**: 管理者権限を持つユーザー

## 分析のヒント

### パーティションを活用した効率的なクエリ

```sql
-- 特定の日付のデータのみを検索（高速）
SELECT *
FROM your_table
WHERE eventday = '20240801'  -- YYYYMMDD形式

-- 日付範囲での検索
SELECT *
FROM your_table
WHERE eventday BETWEEN '20240801' AND '20240807'
```

### 複合条件での異常検知

```sql
-- 深夜の管理者操作を検出
SELECT 
    actor.user.email_addr,
    api.operation,
    from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo' as event_time_jst
FROM your_table
WHERE actor.user.type_id = 2  -- Admin
    AND (
        EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') >= 22
        OR EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') < 6
    )
```

### 配列データの扱い方

```sql
-- web_resources 配列から特定の情報を抽出
SELECT 
    actor.user.email_addr,
    web_resource.name as resource_name,
    web_resource.uid as resource_id
FROM your_table
CROSS JOIN UNNEST(web_resources) AS t(web_resource)
WHERE activity_id = 7  -- Download
```

## 注意事項

1. **時刻データ**: `time` フィールドは Unix ミリ秒で保存されているため、`from_unixtime(time/1000)` で変換が必要

2. **配列アクセス**: Athena では配列のインデックスは 1 から始まる（`web_resources[1]` が最初の要素）

3. **パーティション**: `eventday` を WHERE 句に含めることでクエリ性能が大幅に向上

4. **NULL 値**: オプショナルなフィールドは NULL の可能性があるため、適切な NULL チェックが必要