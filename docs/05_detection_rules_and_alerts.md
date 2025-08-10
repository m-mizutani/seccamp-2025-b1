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

### 3. データ検索の練習

#### Step 1: まずログとして見てみる

**1-1. ファイルダウンロードのログ一覧を見てみる**

まずは、今日のファイルダウンロードのログを10件だけ見てみましょう。

**この演習で使う主要フィールド**
- `time`: イベント発生時刻（UNIXタイムスタンプ、ミリ秒単位、UTC）
  - 読める形式に変換: `from_unixtime(time/1000)`
- `eventday`: パーティションキー（YYYYMMDD形式）
- `actor.user.email_addr`: アクションを実行したユーザー
- `src_endpoint.ip`: アクセス元IP
- `web_resources[1].name`: ファイル名
- `activity_id = 7`: ダウンロード操作を表す

<details>
<summary>💡 ヒント1: SELECTの基本構造</summary>

```sql
SELECT 
    カラム1,
    カラム2
FROM テーブル名
WHERE 条件
ORDER BY 並び替えカラム
LIMIT 件数;
```

</details>

<details>
<summary>💡 ヒント2: 必要な情報</summary>

- activity_id = 7 がダウンロード/エクスポートを表します
- 時刻: `from_unixtime(time/1000)` でUnix時刻を読みやすい形式に
- ユーザー: `actor.user.email_addr`
- IPアドレス: `src_endpoint.ip`
- ファイル名: `web_resources[1].name`

</details>

<details>
<summary>💡 ヒント3: 時刻の区切り方と表示</summary>

時刻データの扱い方の例：
```sql
-- UTC時刻をそのまま表示
from_unixtime(time/1000) as event_time_utc,

-- 日本時間に変換して表示
from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo' as event_time_jst,

-- 日付部分のみ取得
date(from_unixtime(time/1000)) as event_date,

-- 特定の時間範囲でフィルタ（過去1時間）
WHERE time >= (unix_timestamp() - 3600) * 1000
```

注意: Security LakeのタイムスタンプはすべてUTC（協定世界時）で記録されています。
日本時間との差は+9時間です。

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    from_unixtime(time/1000) as event_time,
    actor.user.email_addr as user_email,
    src_endpoint.ip as source_ip,
    web_resources[1].name as file_name
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND activity_id = 7
ORDER BY time DESC
LIMIT 100;
```

</details>

**1-2. 認証のログ一覧を見てみる**

次に、認証（ログイン）関連のログを見てみましょう。

**この演習で使う主要フィールド**
- `api.service.name = 'Google Identity'`: 認証サービスを特定
- `status_id`: 認証結果
  - 1: 成功
  - 2: 失敗
- `api.operation`: 実行された操作種別

<details>
<summary>💡 ヒント1: 認証サービスの特定</summary>

- Google Workspaceの認証は `api.service.name = 'Google Identity'` で絞り込めます
- WHERE句に条件を追加しましょう

</details>

<details>
<summary>💡 ヒント2: 表示したい情報</summary>

認証ログで重要な情報：
- 時刻
- ユーザーメール
- アクセス元IP
- 成功/失敗のステータス
- 操作タイプ

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    from_unixtime(time/1000) as event_time,
    actor.user.email_addr as user_email,
    src_endpoint.ip as source_ip,
    CASE status_id 
        WHEN 1 THEN '成功'
        WHEN 2 THEN '失敗'
    END as status,
    api.operation as operation_type
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND api.service.name = 'Google Identity'
ORDER BY time DESC
LIMIT 20;
```

</details>

#### Step 2: 統計情報を見てみる

**2-1. ファイルダウンロードの回数が多い人、上位20人を見てみる**

GROUP BYを使って集計してみましょう。

**この演習で使う集計関連の機能**
- `GROUP BY`: 指定したカラムでグループ化
- `COUNT(*)`: グループ内のレコード数
- `COUNT(DISTINCT カラム)`: グループ内のユニークな値の数
- `MIN()/MAX()`: 最小値/最大値

<details>
<summary>💡 ヒント1: GROUP BYの基本</summary>

```sql
SELECT 
    グループ化するカラム,
    COUNT(*) as カウント数
FROM テーブル名
WHERE 条件
GROUP BY グループ化するカラム
ORDER BY カウント数 DESC
```

</details>

<details>
<summary>💡 ヒント2: 集計関数</summary>

- `COUNT(*)`: 行数をカウント
- `COUNT(DISTINCT カラム)`: ユニークな値の数をカウント
- `MIN()`: 最小値
- `MAX()`: 最大値

</details>

<details>
<summary>💡 ヒント3: 必要な情報</summary>

1. ユーザーメールアドレスでグループ化
2. ダウンロード回数をカウント
3. ユニークなファイル数をカウント
4. 最初と最後のダウンロード時刻を取得

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as download_count,
    COUNT(DISTINCT web_resources[1].name) as unique_files,
    MIN(from_unixtime(time/1000)) as first_download,
    MAX(from_unixtime(time/1000)) as last_download
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND activity_id = 7
GROUP BY actor.user.email_addr
ORDER BY download_count DESC
LIMIT 20;
```

</details>

**2-2. サービス別の利用状況を見てみる**

各Google Workspaceサービスがどれくらい使われているか統計を取ってみましょう。

**この演習で新たに使う機能**
- `COUNT(DISTINCT actor.user.email_addr)`: ユニークなユーザー数のカウント
- `HAVING`: GROUP BY後の絞り込み条件（WHEREはグループ化前）

<details>
<summary>💡 ヒント1: サービス名でグループ化</summary>

- `api.service.name` でグループ化
- NULL値を除外することを忘れずに

</details>

<details>
<summary>💡 ヒント2: 集計したい情報</summary>

- 総アクセス数
- アクティブユーザー数
- 各サービスで実行された操作の種類数

</details>

<details>
<summary>💡 ヒント3: 意味のあるサービスのみ表示</summary>

- HAVING句で最低限のアクセス数でフィルタリング
- アクセス数の多い順に並べ替え

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    api.service.name as service_name,
    COUNT(*) as total_access,
    COUNT(DISTINCT actor.user.email_addr) as unique_users,
    COUNT(DISTINCT api.operation) as operation_types,
    MIN(from_unixtime(time/1000)) as first_access,
    MAX(from_unixtime(time/1000)) as last_access
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND api.service.name IS NOT NULL
GROUP BY api.service.name
HAVING COUNT(*) >= 10
ORDER BY total_access DESC;
```

</details>

**2-3. ファイル共有の統計情報を見てみる**

WITH句を使って、ファイル共有に関する統計情報を段階的に分析してみましょう。

**この演習で新たに使う機能**
- `WITH`: 一時的な結果セットに名前を付ける（CTE: Common Table Expression）
- `ARRAY_AGG(DISTINCT カラム)`: 値を配列として集約

**WITH句について**
WITH句を使うと、複雑なクエリを段階的に書くことができます：
```sql
WITH 名前1 AS (
    -- 最初の処理
),
名前2 AS (
    -- 名前1の結果を使った処理  
)
-- 最終的な結果を取得
SELECT * FROM 名前2;
```

<details>
<summary>💡 ヒント1: WITH句の構造</summary>

1. まず共有アクティビティを抽出（activity_id = 8）
2. その結果を使ってユーザーごとに集計
3. 最終的に共有数の多い順に表示

</details>

<details>
<summary>💡 ヒント2: 段階的な処理</summary>

- 第1段階: 共有イベントのみを抽出
- 第2段階: ユーザーごとに集計して統計を作成

</details>

<details>
<summary>💡 ヒント3: 集計したい情報</summary>

- 共有したファイル数
- ユニークなファイル数
- よく共有されるファイルのリスト

</details>

<details>
<summary>✅ 回答例</summary>

```sql
WITH share_activities AS (
    -- 第1段階: ファイル共有アクティビティを抽出
    SELECT 
        actor.user.email_addr as user_email,
        web_resources[1].name as file_name,
        from_unixtime(time/1000) as share_time
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
    WHERE eventday = date_format(current_date, '%Y%m%d')
        AND activity_id = 8  -- 共有操作
),
user_share_stats AS (
    -- 第2段階: ユーザーごとに集計
    SELECT 
        user_email,
        COUNT(*) as total_shares,
        COUNT(DISTINCT file_name) as unique_files_shared,
        MIN(share_time) as first_share,
        MAX(share_time) as last_share,
        ARRAY_AGG(DISTINCT substr(file_name, 1, 50)) as sample_files
    FROM share_activities
    GROUP BY user_email
)
-- 最終結果: 共有数の多いユーザーTOP10
SELECT * 
FROM user_share_stats
ORDER BY total_shares DESC
LIMIT 10;
```

</details>

#### Step 3: 時間帯による傾向の詳細分析

**3-1. 時間帯別のアクティビティ分析**

複数の観点から時間帯別の傾向を分析してみましょう。

**この演習で新たに使う機能**
- `EXTRACT(HOUR FROM timestamp)`: タイムスタンプから時間を抽出
- `AT TIME ZONE 'Asia/Tokyo'`: タイムゾーン変換
- 複数のWITH句を連鎖させる

**なぜ時間帯分析が重要か**
- 通常の業務時間外のアクティビティは不審な可能性
- システムの負荷パターンを理解
- 異常検知の基準となるベースラインの把握

<details>
<summary>💡 ヒント1: 段階的な分析の構造</summary>

1. 全イベントを日本時間の時間帯付きで抽出
2. 時間帯ごとに集計
3. ビジネスアワーかどうかの判定を追加

</details>

<details>
<summary>💡 ヒント2: 時間の抽出と変換</summary>

- `EXTRACT(HOUR FROM ...)` で時間を取り出します
- `AT TIME ZONE 'Asia/Tokyo'` で日本時間に変換
- ビジネスアワー: 9時〜18時

</details>

<details>
<summary>💡 ヒント3: 複数の観点での集計</summary>

- 総イベント数
- アクティブユーザー数
- 主要なアクティビティ種別
- サービス別の利用状況

</details>

<details>
<summary>✅ 回答例</summary>

```sql
WITH hourly_events AS (
    -- 第1段階: イベントを時間帯付きで抽出
    SELECT 
        EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') as hour_jst,
        actor.user.email_addr,
        activity_id,
        api.service.name as service_name,
        status_id
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
    WHERE eventday = date_format(current_date, '%Y%m%d')
),
hourly_summary AS (
    -- 第2段階: 時間帯ごとに集計
    SELECT 
        hour_jst,
        COUNT(*) as total_events,
        COUNT(DISTINCT email_addr) as active_users,
        COUNT(CASE WHEN activity_id = 7 THEN 1 END) as downloads,
        COUNT(CASE WHEN activity_id = 8 THEN 1 END) as shares,
        COUNT(CASE WHEN status_id = 2 THEN 1 END) as failures,
        COUNT(DISTINCT service_name) as services_used
    FROM hourly_events
    GROUP BY hour_jst
)
-- 最終結果: ビジネスアワーの判定を追加
SELECT 
    hour_jst,
    CASE 
        WHEN hour_jst BETWEEN 9 AND 18 THEN 'Business Hours'
        ELSE 'After Hours'
    END as time_category,
    total_events,
    active_users,
    downloads,
    shares,
    failures,
    services_used
FROM hourly_summary
ORDER BY hour_jst;
```

</details>

**3-2. アクティブユーザーの時間帯パターン分析**

特定のアクティブユーザーの行動パターンを詳しく分析してみましょう。複数の分析結果を組み合わせることで、より深い洞察を得ることができます。

<details>
<summary>💡 ヒント1: まずアクティブユーザーを特定</summary>

- 最初のWITH句でアクティブユーザーTOP5を抽出
- 次のWITH句でそのユーザーの詳細な活動を分析

</details>

<details>
<summary>💡 ヒント2: ユーザーごとの時間帯パターン</summary>

- どの時間帯に最も活発か
- 通常と異なる時間帯の活動はあるか
- 主にどんな操作をしているか

</details>

<details>
<summary>✅ 回答例</summary>

```sql
WITH active_users AS (
    -- 最もアクティブなユーザーTOP5を特定
    SELECT 
        actor.user.email_addr as user_email,
        COUNT(*) as total_events
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
    WHERE eventday = date_format(current_date, '%Y%m%d')
    GROUP BY actor.user.email_addr
    ORDER BY total_events DESC
    LIMIT 5
),
user_hourly_pattern AS (
    -- アクティブユーザーの時間帯別活動を分析
    SELECT 
        t.actor.user.email_addr as user_email,
        EXTRACT(HOUR FROM from_unixtime(t.time/1000) AT TIME ZONE 'Asia/Tokyo') as hour_jst,
        COUNT(*) as event_count,
        COUNT(DISTINCT t.api.operation) as operation_types,
        COUNT(CASE WHEN t.activity_id = 7 THEN 1 END) as downloads,
        COUNT(CASE WHEN t.activity_id = 8 THEN 1 END) as shares
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0 t
    INNER JOIN active_users au ON t.actor.user.email_addr = au.user_email
    WHERE t.eventday = date_format(current_date, '%Y%m%d')
    GROUP BY t.actor.user.email_addr, EXTRACT(HOUR FROM from_unixtime(t.time/1000) AT TIME ZONE 'Asia/Tokyo')
)
-- 最終結果: ユーザーごとの時間帯パターンを表示
SELECT 
    user_email,
    hour_jst,
    event_count,
    operation_types,
    downloads,
    shares,
    CASE 
        WHEN hour_jst < 6 OR hour_jst > 22 THEN '深夜・早朝'
        WHEN hour_jst BETWEEN 9 AND 18 THEN '業務時間内'
        ELSE '業務時間外'
    END as time_category
FROM user_hourly_pattern
ORDER BY user_email, hour_jst;
```

</details>

### 練習問題

上記のクエリを参考に、以下の質問に答えるクエリを書いてみましょう。

**1. 今日最も多くのファイルを共有（activity_id = 8）したユーザーは誰ですか？**

<details>
<summary>💡 ヒント</summary>

- activity_id = 8 が共有を表します
- GROUP BYでユーザーごとに集計
- COUNT(*)で共有回数をカウント

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as share_count,
    COUNT(DISTINCT web_resources[1].name) as unique_files_shared
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND activity_id = 8
GROUP BY actor.user.email_addr
ORDER BY share_count DESC
LIMIT 10;
```

</details>

**2. 特定のサービスを最も多く利用しているユーザーTOP5を見つけてください（例: Google Drive）**

<details>
<summary>💡 ヒント</summary>

- `api.service.name = 'Google Drive'` で特定サービスを絞り込み
- ユーザーごとにアクセス回数を集計
- 上位5人を表示

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as access_count,
    COUNT(DISTINCT api.operation) as operation_types,
    MIN(from_unixtime(time/1000)) as first_access,
    MAX(from_unixtime(time/1000)) as last_access
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND api.service.name = 'Google Drive'
GROUP BY actor.user.email_addr
ORDER BY access_count DESC
LIMIT 5;
```

</details>

**3. 日本以外の国からアクセスしているユーザーを見つけてください**

<details>
<summary>💡 ヒント</summary>

- `src_endpoint.location.country`で国情報を取得
- `!= 'JP'`または`NOT IN ('JP', 'Japan')`で日本以外をフィルタリング

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT DISTINCT
    actor.user.email_addr as user_email,
    src_endpoint.location.country as country,
    src_endpoint.ip as source_ip,
    COUNT(*) as access_count
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND src_endpoint.location.country IS NOT NULL
    AND src_endpoint.location.country NOT IN ('JP', 'Japan')
GROUP BY actor.user.email_addr, src_endpoint.location.country, src_endpoint.ip
ORDER BY access_count DESC;
```

</details>

### SQLクエリを書く際のヒント

1. **まずは小さく始める**: LIMIT 10 などで結果を制限して、データの形を確認
2. **段階的に条件を追加**: WHERE句の条件を1つずつ追加して結果を確認
3. **エラーが出たら**: カラム名のスペルミス、括弧の対応、クォートの閉じ忘れをチェック
4. **パフォーマンスを意識**: eventday でパーティションを指定することを忘れずに
5. **段階的な分析**: 複雑なクエリは段階的に分解して、読みやすく保守しやすいコードに

