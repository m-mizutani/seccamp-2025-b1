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
LIMIT 10;
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

**2-2. 認証失敗回数の多いアカウント一覧を見てみる**

**この演習で新たに使う機能**
- `HAVING`: GROUP BY後の絞り込み条件（WHEREはグループ化前）
- `COUNT(DISTINCT src_endpoint.ip)`: ユニークなIPアドレス数のカウント

<details>
<summary>💡 ヒント1: 認証失敗の特定</summary>

- `status_id = 2` が失敗を表します
- Google Identityサービスで絞り込みます

</details>

<details>
<summary>💡 ヒント2: HAVING句の使い方</summary>

- HAVINGはGROUP BY後の絞り込み条件
- `HAVING COUNT(*) >= 3` でグループ化後の件数でフィルタリング

</details>

<details>
<summary>💡 ヒント3: 集計したい情報</summary>

- 失敗回数
- 異なるIPアドレスの数（複数場所からのアタックの可能性）
- 最初と最後の失敗時刻

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as failure_count,
    COUNT(DISTINCT src_endpoint.ip) as unique_ips,
    MIN(from_unixtime(time/1000)) as first_failure,
    MAX(from_unixtime(time/1000)) as last_failure
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND api.service.name = 'Google Identity'
    AND status_id = 2
GROUP BY actor.user.email_addr
HAVING COUNT(*) >= 3
ORDER BY failure_count DESC;
```

</details>

**2-3. 認証失敗回数の多いIPアドレス一覧を見てみる**

**この演習で新たに使う機能**
- `ARRAY_AGG(DISTINCT カラム)`: 値を配列として集約
- 複数カラムでのGROUP BY

<details>
<summary>💡 ヒント1: IPアドレスでのグループ化</summary>

- IPアドレスと国の両方でグループ化
- これによりIPアドレスごとの統計が取れます

</details>

<details>
<summary>💡 ヒント2: ARRAY_AGG関数</summary>

- `ARRAY_AGG(DISTINCT カラム)` で値を配列にまとめます
- どのユーザーが標的になったか一目でわかります

</details>

<details>
<summary>💡 ヒント3: 攻撃元の分析</summary>

- 同一IPから複数ユーザーへの認証失敗はブルートフォース攻撃の可能性
- 国情報も含めて確認

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    src_endpoint.ip as source_ip,
    src_endpoint.location.country as country,
    COUNT(*) as failure_count,
    COUNT(DISTINCT actor.user.email_addr) as target_users,
    ARRAY_AGG(DISTINCT actor.user.email_addr) as user_list
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND api.service.name = 'Google Identity'
    AND status_id = 2
GROUP BY src_endpoint.ip, src_endpoint.location.country
HAVING COUNT(*) >= 5
ORDER BY failure_count DESC
LIMIT 20;
```

</details>

#### Step 3: 時間帯による傾向を見てみる

**3-1. 特定ユーザーの時間帯ごとのAPI呼び出しを見てみる**

まず、アクティブなユーザーを特定しましょう。

**この演習で新たに使う機能**
- `EXTRACT(HOUR FROM timestamp)`: タイムスタンプから時間を抽出
- `AT TIME ZONE 'Asia/Tokyo'`: タイムゾーン変換

<details>
<summary>💡 ヒント1: アクティブユーザーの特定</summary>

- ユーザーごとにイベント数を集計
- 上位5人を表示

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as total_events
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
GROUP BY actor.user.email_addr
ORDER BY total_events DESC
LIMIT 5;
```

</details>

上記で見つかったユーザーのメールアドレスを使って、時間帯別の分析をしてみましょう。

<details>
<summary>💡 ヒント1: 時間の抽出</summary>

- `EXTRACT(HOUR FROM ...)` で時間を取り出します
- `AT TIME ZONE 'Asia/Tokyo'` で日本時間に変換

</details>

<details>
<summary>💡 ヒント2: 特定ユーザーで絞り込み</summary>

- WHERE句に `actor.user.email_addr = 'ユーザーメール'` を追加
- 実際のユーザーメールアドレスに置き換えてください

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') as hour_jst,
    COUNT(*) as event_count,
    COUNT(DISTINCT api.operation) as unique_operations,
    ARRAY_AGG(DISTINCT api.operation) as operations_list
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND actor.user.email_addr = 'user@example.com'  -- 実際のユーザーに変更
GROUP BY EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo')
ORDER BY hour_jst;
```

</details>

**3-2. 全体的な時間帯別の傾向を見る**

**この演習で新たに使う機能**
- `COUNT(CASE WHEN 条件 THEN 1 END)`: 条件付きカウント

<details>
<summary>💡 ヒント1: CASE文を使った集計</summary>

```sql
COUNT(CASE WHEN 条件 THEN 1 END)
```
この方法で、特定の条件に一致するレコードだけをカウントできます。

</details>

<details>
<summary>💡 ヒント2: 集計したい情報</summary>

- 総イベント数
- アクティブユーザー数
- ダウンロード数
- 失敗数

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo') as hour_jst,
    COUNT(*) as total_events,
    COUNT(DISTINCT actor.user.email_addr) as active_users,
    COUNT(CASE WHEN activity_id = 7 THEN 1 END) as downloads,
    COUNT(CASE WHEN status_id = 2 THEN 1 END) as failures
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
GROUP BY EXTRACT(HOUR FROM from_unixtime(time/1000) AT TIME ZONE 'Asia/Tokyo')
ORDER BY hour_jst;
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

**2. 過去1時間に10個以上のファイルをダウンロードしたユーザーはいますか？**

<details>
<summary>💡 ヒント</summary>

- 過去1時間: `time >= (unix_timestamp() - 3600) * 1000`
- activity_id = 7 がダウンロード
- HAVING句で`COUNT(*) >= 10`を指定

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as download_count,
    MIN(from_unixtime(time/1000)) as first_download,
    MAX(from_unixtime(time/1000)) as last_download
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND activity_id = 7
    AND time >= (unix_timestamp() - 3600) * 1000
GROUP BY actor.user.email_addr
HAVING COUNT(*) >= 10
ORDER BY download_count DESC;
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

