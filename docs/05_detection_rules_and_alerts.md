# 探索的ログ分析とデータに対する理解

**時間：10:20-11:00 (40分)**

このパートでは、Security Lake に蓄積されたログデータを AWS Athena を使って探索し、OCSF（Open Cybersecurity Schema Framework）形式のデータ構造を実際のデータを通じて理解します。基本的な SQL クエリから始めて、段階的により複雑な分析へと進めていきます。

## AWS コンソールでの実践的データ探索

### 1. Athena コンソールでの基本操作

- ログインページ https://145287089436.signin.aws.amazon.com/console
- アカウントID: `145287089436`
- ログインしたら右上のRegionから `Asia Pacific (Tokyo)` を選択 ← 重要

### 2. Athena コンソールへのアクセス

- AWSコンソールの左上テキストボックスから `athena` と入力してサービスを開く
  - その後、 `Launch query editor` を開く
  - あるいは https://ap-northeast-1.console.aws.amazon.com/athena/home?region=ap-northeast-1#/query-editor
- 結果出力のS3の設定が必要
  - 「最初のクエリを実行する前に、Amazon S3 でクエリ結果の場所を設定する必要があります。」の右のボタンをクリック
  - "Browse S3" ボタンで `seccamp2025-b1-athena-results` を選択
  - "保存" を選択したら "エディタ" にもどる

用意ができたら以下を実行

```sql
SELECT COUNT(*) as event_count
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d');
```

- Database:`amazon_security_lake_glue_db_ap_northeast_1`
- Table: `amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0`

### 2. Security Lake テーブルの確認

[log_schema.md](log_schema.md) を参考にしつつスキーマを確認しましょう。

```sql
-- 必要なカラムに絞ったクエリ
SELECT 
    from_unixtime(time/1000) as event_time,
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
    severity_id,
    status_id,
    actor.user.email_addr,
    actor.user.type_id,
    api.service.name,
    api.operation,
    src_endpoint.ip,
    web_resources[1].name as resource_name
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
ORDER BY time DESC
LIMIT 100;
```


### 3. 異常なアクティビティを探す練習

**大量ダウンロードの検出**
```sql
-- 過去1時間で多数のファイルをダウンロードしたユーザー
SELECT 
    actor.user.email_addr as user_email,
    COUNT(DISTINCT web_resources) as downloaded_files,
    ARRAY_AGG(DISTINCT web_resources[1].name ORDER BY web_resources[1].name) as file_names,
    MIN(from_unixtime(time/1000)) as first_download,
    MAX(from_unixtime(time/1000)) as last_download
FROM seccamp2025_b1_security_lake.google_workspace
WHERE activity_id = 7  -- Export/Download
    AND from_unixtime(time/1000) > current_timestamp - interval '1' hour
    AND CARDINALITY(web_resources) > 0
GROUP BY actor.user.email_addr
HAVING COUNT(DISTINCT web_resources) >= 10
ORDER BY downloaded_files DESC;
```

**失敗した認証試行の分析**
```sql
-- ログイン失敗が多いユーザーとその詳細
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as failure_count,
    ARRAY_AGG(DISTINCT src_endpoint.ip) as source_ips,
    ARRAY_AGG(DISTINCT api.operation) as failed_operations,
    MIN(from_unixtime(time/1000)) as first_failure,
    MAX(from_unixtime(time/1000)) as last_failure
FROM seccamp2025_b1_security_lake.google_workspace
WHERE status_id = 2  -- Failure
    AND api.service.name = 'Google Identity'
    AND from_unixtime(time/1000) > current_timestamp - interval '24' hour
GROUP BY actor.user.email_addr
HAVING COUNT(*) >= 3
ORDER BY failure_count DESC;
```

**異常な時間帯のアクセス**
```sql
-- 深夜・早朝（22時〜6時）のアクセスを確認
SELECT 
    actor.user.email_addr as user_email,
    from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo' as access_time_jst,
    activity_name,
    api.operation,
    web_resources[1].name as accessed_resource,
    src_endpoint.ip as source_ip
FROM seccamp2025_b1_security_lake.google_workspace
WHERE (
    EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') >= 22
    OR EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') < 6
    )
    AND from_unixtime(time/1000) > current_timestamp - interval '7' day
    AND status_id = 1  -- Success
ORDER BY time DESC
LIMIT 50;
```

## 探索的分析のベストプラクティス

### 1. 段階的なアプローチ
- まず全体像を把握（集計・統計）
- 異常値や外れ値に注目
- 詳細なドリルダウン分析

### 2. 時系列での変化を追跡
```sql
-- 日別のアクティビティ推移
SELECT 
    DATE(from_unixtime(time/1000)) as activity_date,
    COUNT(*) as total_activities,
    COUNT(DISTINCT actor.user.email_addr) as active_users,
    COUNT(CASE WHEN activity_id = 7 THEN 1 END) as downloads,
    COUNT(CASE WHEN status_id = 2 THEN 1 END) as failures
FROM seccamp2025_b1_security_lake.google_workspace
WHERE from_unixtime(time/1000) > current_timestamp - interval '30' day
GROUP BY DATE(from_unixtime(time/1000))
ORDER BY activity_date DESC;
```

### 3. 複合条件での異常検知
```sql
-- 複数の異常指標を組み合わせた検出
WITH user_activity AS (
    SELECT 
        actor.user.email_addr as user_email,
        COUNT(*) as total_activities,
        COUNT(CASE WHEN activity_id = 7 THEN 1 END) as download_count,
        COUNT(CASE WHEN status_id = 2 THEN 1 END) as failure_count,
        COUNT(DISTINCT DATE(from_unixtime(time/1000))) as active_days,
        COUNT(DISTINCT src_endpoint.ip) as unique_ips,
        MAX(severity_id) as max_severity
    FROM seccamp2025_b1_security_lake.google_workspace
    WHERE from_unixtime(time/1000) > current_timestamp - interval '7' day
    GROUP BY actor.user.email_addr
)
SELECT 
    user_email,
    total_activities,
    download_count,
    failure_count,
    unique_ips,
    CASE 
        WHEN download_count > 50 AND unique_ips > 3 THEN '大量ダウンロード＋複数IP'
        WHEN failure_count > 10 THEN '認証失敗多発'
        WHEN max_severity >= 4 THEN '高リスク操作'
        ELSE '要注意'
    END as risk_category
FROM user_activity
WHERE download_count > 20 
   OR failure_count > 5 
   OR unique_ips > 5
   OR max_severity >= 3
ORDER BY download_count DESC, failure_count DESC;
```

## まとめ

このパートでは、実際の Security Lake データを使用して：

1. Athena の基本的な操作方法を学習しました
2. OCSF スキーマの構造と主要フィールドを理解しました
3. 探索的データ分析の手法を実践しました
4. 異常なアクティビティを発見するための SQL クエリを作成しました

これらの知識は、次のパート（Part 6: Lambda実装と検知ルール作成）で、自動化された検知システムを構築する際の基盤となります。