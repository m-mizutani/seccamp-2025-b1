# Lambda実装と検知ルール作成

**時間：11:00-11:50 (50分)**

## 概要

このパートでは、実際にコードを書いて Security Lake を活用したセキュリティ監視システムの一部を実装します。前のパートで学んだ OCSF スキーマの知識を活かして検知ルールを作成し、自動化された脅威検知システムを構築します。

## ログ収集 Lambda 実装

ここからは個人に分かれてコードを書きます。各リソースは「チーム」という単位として分けています。（今回は1人ちーむとなります）アサインは https://github.com/seccamp2025-b/b1-secmon を参照してください。

### 環境準備・理解

#### 1. GitHub リポジトリのクローンとブランチ作成

```bash
# リポジトリのクローン（まだの場合）
git clone git@github.com:seccamp2025-b/b1-secmon.git
cd b1-secmon

git checkout -b {team_id}/init
git checkout -b blue/init # 例
```

原則として変更は以下の手順にしてください。

1. branch作成
2. GitHubへpushし、Pull Requestを作成
3. `@m-mizutani` にレビューリクエストを設定
4. Approve されたらマージ

慣れてきたら `main` へ直pushありにします。

#### 2. プロジェクト構造の理解

```
b1-secmon/
├── lambda/
│   ├── blue/          # チームごとのLambda実装
│   │   ├── main.go    # Lambda ハンドラー
│   │   ├── queries/   # SQLクエリファイル
│   │   └── go.mod
│   ├── green/
│   └── red/
│   └── ...
├── internal/          # 共有パッケージ
│   ├── athena/       # Athena操作ユーティリティ
│   ├── query/        # クエリローダー
│   └── sns/          # SNS通知ユーティリティ
├── .github/
│   └── workflows/    # GitHub Actions (自動デプロイ)
└── scripts/
    └── create-team.sh # 新チーム作成スクリプト
```

### 初期デプロイ

チームのLambda環境をセットアップします。初心者の方でも迷わないよう、一つ一つ丁寧に説明していきます。

#### 1. チーム用ディレクトリの作成

まず、自分のチーム用のディレクトリを作成します。`blue` チームのコードをベースにコピーします。

```bash
# scripts ディレクトリにあるスクリプトを実行
# {team_id} を自分のチーム名に置き換えてください
# 例: ./scripts/create-team.sh red
./scripts/create-team.sh {team_id}
```

このスクリプトは `blue` チームのコードをコピーして、新しいチーム用のディレクトリを作成します。

その後、以下の "Hello, I'm blue" を自分のものとわかる内容に書き換える。

```go
// アラートメッセージを作成
alertMessage := AlertMessage{
    // TODO(4): ここに必要な情報を埋め込む
    Title: "Hello, I'm blue",
}
```

#### 2. 変更をGitに記録する

作成したファイルをGitで管理します。

```bash
# 現在の状態を確認（新しく作成されたファイルが表示されます）
git status

# 作成したチームディレクトリをGitに追加
# 例: git add lambda/red
git add lambda/{team_id}

# 変更をコミット（記録）
# -m のあとのメッセージは変更内容を説明するもの
git commit -m "Add {team_id} team initial setup"
```

#### 3. GitHubにプッシュ（アップロード）する

ローカルの変更をGitHubに送信します。

```bash
# GitHubにブランチをプッシュ
# 例: git push origin red/init
git push origin {team_id}/init
```

#### 4. Pull Request（PR）を作成する

GitHubのWebサイトでPull Requestを作成します。

1. ブラウザで https://github.com/seccamp2025-b/b1-secmon を開く
2. 黄色いバナーが表示されている場合：
   - 「Compare & pull request」ボタンをクリック
3. バナーが表示されていない場合：
   - 「Pull requests」タブをクリック
   - 「New pull request」ボタンをクリック
   - 「compare:」のドロップダウンから `{team_id}/init` を選択

4. Pull Request作成画面で：
   - **Title**: `[{team_id}] Initial setup` など分かりやすいタイトルを入力
5. 右側の「Reviewers」で `m-mizutani` を選択
6. 「Create pull request」ボタンをクリック

#### 5. マージを待つ

レビュアーがコードを確認し、Approveされたら「Merge pull request」ボタンが押せるようになります。
マージが完了したら、GitHub Actionsが自動的にLambda関数をデプロイします。

#### 6. デプロイの確認とテスト実行

マージ後、GitHub Actionsが自動でデプロイを開始します。まずはデプロイの状況を確認しましょう。

##### 6-1. GitHub Actionsでデプロイ状況を確認

1. GitHubのリポジトリページ ( https://github.com/seccamp2025-b/b1-secmon ) を開く
2. 「Actions」タブをクリック
3. 最新のワークフロー実行を確認：
   - **緑のチェックマーク ✓**: デプロイ成功
   - **黄色の丸 ●**: 実行中（1-2分程度かかります）
   - **赤の×**: デプロイ失敗

4. ワークフローをクリックして詳細を確認
   - 「Deploy Lambda Functions」のステップで自分のチーム名が表示されていることを確認
   - 例: `Deploying team: red`

デプロイが失敗した場合は、エラーログを確認して講師に相談してください。

##### 6-2. AWS Lambda コンソールへアクセス

1. AWS Management Console にログイン
2. サービス検索で「Lambda」を入力して選択
3. 関数一覧から `{team_id}-detector` を探してクリック
   - 例: `blue-detector`

##### 6-3. テスト実行

1. Lambda関数の詳細画面で「テスト」タブをクリック
2. 「テスト」ボタンをクリックして実行

実行結果が表示され、「実行結果: 成功」と表示されれば正常に動作しています。

##### 6-3. アラート確認

Lambda関数が正常に実行されると、アラートが送信されます。

1. ブラウザで https://warren-171198963743.asia-northeast1.run.app/alerts を開く
2. ページ上部の検索ボックスに自分のチーム名（例: `red`）を入力
3. 自分のチームから送信されたアラートが表示されることを確認

アラートには以下のような情報が含まれているはずです：
- **Title**: 検知ルールのタイトル
- **Team**: 自分のチーム名
- **Time**: アラート送信時刻

もしアラートが表示されない場合は：
- Lambda関数の実行ログを確認（CloudWatch Logs）
- 数分待ってからページをリロード
- 講師に相談

これで初期デプロイは完了です。次は実際の検知ルールを実装していきましょう。

### Lambda 実装の概要

以降は `{team_id}/{任意の名前}` でブランチを作って作業し、それをPull Requestにするようにしてください。

#### 1. 基本的な処理フローの理解

b1-secmon の Lambda 関数は以下の流れで動作します：

1. **Athena クライアントの初期化** - Security Lake のデータを検索
2. **SNS パブリッシャーの初期化** - アラート通知用
3. **SQL クエリファイルの読み込み** - `queries/` ディレクトリから
4. **各クエリの実行とアラート判定**
5. **必要に応じてアラート送信**

#### 2. コード実装のポイント

`lambda/blue/main.go` を例に、実装すべき TODO を確認します：

```go
// TODO(1): queries ディレクトリに検知用SQLを配置
//go:embed queries/*.sql
var queryFS embed.FS

// HandleRequest - Lambda のメイン処理
func detect(ctx context.Context) error {
    // ... 初期化処理 ...
    
    // 各クエリファイルを処理
    for _, query := range queries {
        // Athenaでクエリ実行
        results, err := athenaClient.Query(ctx, query.Content)
        
        // TODO(2): 結果を検証してアラート送信を判断
        // 例: 特定の条件を満たさない場合は continue
        
        // TODO(3): アラートメッセージ構造を定義
        type AlertMessage struct {
            Title       string `json:"title"`
            // 追加フィールド...
        }
        
        // TODO(4): アラート内容を設定
        alertMessage := AlertMessage{
            Title: "検知タイトル",
        }
        
        // SNS へアラート送信
        snsPublisher.Publish(ctx, alertMessage)
    }
}
```

#### 3. SQLクエリの作成

`queries/` ディレクトリに検知ルールとなる SQL を配置します：

```sql
-- queries/suspicious_login.sql
SELECT 
    unmapped['email'] as user_email,
    src_endpoint.ip as source_ip,
    COUNT(*) as failed_count
FROM 
    amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE 
    eventday = date_format(current_date, '%Y%m%d')
    AND activity_name = 'login'
    AND status_code != '1'  -- 失敗
GROUP BY 
    unmapped['email'], src_endpoint.ip
HAVING 
    COUNT(*) >= 5  -- 5回以上の失敗
```

### ローカルでのテストとデバッグ

#### 1. 構文チェック

```bash
cd lambda/blue
go vet ./...
```

#### 2. ビルド確認

```bash
# Lambda 用にビルド（エラーがないか確認）
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap .
```

#### 3. SQL クエリの検証

AWS コンソールの Athena でクエリを直接実行して動作確認：

1. AWS コンソール → Athena を開く
2. データベース: `amazon_security_lake_glue_db_ap_northeast_1` を選択
3. クエリエディタでSQLを実行

### GitHub Actions によるデプロイ

#### 1. コミットとプッシュ

```bash
git add lambda/blue/
git commit -m "feat: 不審なログイン試行の検知ルールを追加"
git push origin feature/my-detector
```

#### 2. プルリクエスト作成

GitHub でプルリクエストを作成し、main ブランチにマージすると自動デプロイが実行されます。

#### 3. デプロイの確認

GitHub Actions のログで以下を確認：
- ビルド成功
- Lambda 関数の更新完了
- 環境変数の設定

### 実装のベストプラクティス

1. **エラーハンドリング**
   - 必ず `goerr.Wrap` でエラーをラップ
   - エラー時は `logger.Error` でログ出力

2. **アラートメッセージ**
   - トリアージに必要な情報を含める
   - 誤検知を減らすため、条件や閾値は慎重に設定

3. **パフォーマンス**
   - クエリは必要最小限のデータを取得
   - `eventday` でパーティション絞り込み必須

4. **セキュリティ**
   - 秘密情報はコードに含めない
   - 環境変数は GitHub Actions が自動設定

## 検知ルールの作成

検知ルールを作成し、定期実行される（想定の）Lambda関数に組み込み、検知システムを完成させます。SQLによって発見された事象をアラートとして発報する際、アラートを調査する担当者の行動をイメージして必要な情報を報告することが重要です。

### 調査のためにアラートに含めるべき情報

基本的には5W(1H)を抑える形式が望ましいです。そこからさらに調査の足がかりになるような情報を付与できるとよいでしょう。

- Who: そのアラートを発生させた主体（principle, subject）
  - ユーザID: 認証が済んでいるシステムならユーザ名があるとよいでしょう
  - IPアドレス: リモートからのアクセスにおいて利用価値のある識別情報です
- What: どのリソースに対してのアクセス、あるいは攻撃だったか
  - 今回だと対象となるドキュメントであったり、ログイン試行なら対象ユーザIDなどが相当します
- When: 検知をした時刻だけでなく、検知対象となったイベントが発生した時刻がより重要である点に注意です
- Where: 今回はGoogle Workspaceなのでそのサービス内であることは自明ですが、例えばkubernetesのようなシステムではどのPod、Nodeで発生した事象なのかという情報は重要になります

このほか、IoC（Indicator of Compromise）になりそうな情報や、アラートの概況（特に複数アラートを束ねる場合）を伝える情報を組み込むのが良いでしょう。

### 検知ルールの作成課題

ここからは、実際にセキュリティイベントを検知するSQLクエリを作成していきます。以下の課題から選んで実装してみましょう。

**注意事項**:
- 以下のSQL例では時刻フィルタリングに`to_unixtime(current_timestamp)`を使用しています（AWS Athenaで利用可能）
- WITH句を使用して段階的にデータを処理しています

**WITH句について（初心者向け説明）**:
WITH句は複雑なクエリを分かりやすく書くための機能です。一時的な結果セットに名前を付けて、後で参照できます。
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

#### 🎯 課題1: 継続的な認証攻撃の検知

**シナリオ解説（初心者向け）**:
認証攻撃とは、悪意のある攻撃者が正規ユーザーになりすまそうとする行為です。最も一般的な手法として「ブルートフォース攻撃」や「パスワードスプレー攻撃」があります。

- **ブルートフォース攻撃**: 特定のユーザーに対して様々なパスワードを試す
- **パスワードスプレー攻撃**: 複数のユーザーに対して一般的なパスワード（例: password123）を試す

これらの攻撃は、短時間に大量の認証失敗を生み出すという特徴があります。

**検知したい状況の詳細**:
1. **攻撃の兆候**
   - 同一のIPアドレスから5分以内に10回以上のログイン失敗
   - 複数の異なるユーザーアカウントへの試行（パスワードスプレーの可能性）
   - 通常の業務時間外でのアクセス試行

2. **なぜこのパターンが危険か**
   - 攻撃者がアカウントの乗っ取りを試みている可能性が高い
   - 成功すれば組織の機密情報にアクセスされる恐れ
   - 他の攻撃の前段階である可能性（偵察行為）

3. **正常な行動との区別**
   - 通常のユーザーは2-3回の失敗後にパスワードリセットを行う
   - 同一IPから複数ユーザーへの認証試行は通常発生しない

<details>
<summary>💡 ヒント1: 使えそうなフィールド</summary>

- `api.service.name = 'Google Identity'`: 認証サービス
- `status_id = 2`: 失敗を表す
- `src_endpoint.ip`: 攻撃元IP
- `actor.user.email_addr`: 標的となったユーザー
- `time`: 時刻でのフィルタリング

</details>

<details>
<summary>💡 ヒント2: 時間窓の設定</summary>

```sql
-- 過去5分間のログを対象にする
WHERE time >= (to_unixtime(current_timestamp) - 300) * 1000
```

</details>

<details>
<summary>💡 ヒント3: 攻撃パターンの特徴</summary>

1. 同一IPから複数ユーザーへの試行
2. 短時間に高頻度の失敗
3. IPアドレスごとにグループ化して集計

</details>

<details>
<summary>✅ 回答例</summary>

```sql
WITH login_failures AS (
    -- まず、ログイン失敗のレコードを抽出
    SELECT 
        src_endpoint.ip,
        actor.user.email_addr,
        from_unixtime(time/1000) as attempt_time
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
    WHERE eventday = date_format(current_date, '%Y%m%d')
        AND api.service.name = 'Google Identity'
        AND (status_id = 2 OR api.operation = 'login_failure')  -- 失敗の判定
        AND time >= (to_unixtime(current_timestamp) - 300) * 1000  -- 過去5分間
),
attack_summary AS (
    -- 次に、IPアドレスごとに集計
    SELECT 
        ip as attacker_ip,
        COUNT(*) as failure_count,
        COUNT(DISTINCT email_addr) as target_users,
        MIN(attempt_time) as first_attempt,
        MAX(attempt_time) as last_attempt,
        ARRAY_AGG(DISTINCT email_addr) as targeted_users_list
    FROM login_failures
    GROUP BY ip
)
-- 最後に、10回以上の失敗があったIPを抽出
SELECT * 
FROM attack_summary
WHERE failure_count >= 10
ORDER BY failure_count DESC;
```

</details>

#### 🎯 課題2: 大量データ窃取の検知

**シナリオ解説（初心者向け）**:
データ窃取（Data Exfiltration）とは、組織の機密情報を不正に外部へ持ち出す行為です。内部犯行と外部からの侵入の両方で発生する可能性があります。

典型的なデータ窃取の手口：
- **一括ダウンロード**: 短時間で大量のファイルをダウンロード
- **重要ファイルの選別**: 機密性の高いファイルを狙い撃ち
- **時間外アクセス**: 監視が手薄な深夜や休日に実行

**検知したい状況の詳細**:
1. **異常なダウンロードパターン**
   - 10分間で50件以上のファイルダウンロード（通常業務では考えられない量）
   - 異なるフォルダやプロジェクトから無差別にファイルを収集
   - 同一IPアドレスからの連続的なアクセス

2. **なぜこのパターンが危険か**
   - 機密情報の大量流出につながる可能性
   - 競合他社への情報漏洩のリスク
   - 個人情報保護法違反などコンプライアンス違反の恐れ
   - 知的財産の損失による経済的損害

3. **正常な行動との区別**
   - 通常業務では必要なファイルを選択的にダウンロード
   - プロジェクトメンバーは関連ファイルのみアクセス
   - 大量ダウンロードが必要な場合は事前申請がある

<details>
<summary>💡 ヒント1: 使えそうなフィールド</summary>

- `activity_id = 7`: ダウンロード操作
- `actor.user.email_addr`: ダウンロードユーザー
- `src_endpoint.ip`: アクセス元IP
- `web_resources[1].name`: ダウンロードされたファイル
- `time`: 時間フィルタリング

</details>

<details>
<summary>💡 ヒント2: 異常なダウンロードパターン</summary>

1. 短時間に大量のファイル
2. 異なるディレクトリからのファイル収集
3. 通常の業務パターンと異なる時間帯

</details>

<details>
<summary>💡 ヒント3: 時間窓と集計</summary>

```sql
-- 過去10分間
WHERE time >= (to_unixtime(current_timestamp) - 600) * 1000
-- ユーザーとIPでグループ化
GROUP BY actor.user.email_addr, src_endpoint.ip
```

</details>

<details>
<summary>✅ 回答例</summary>

```sql
WITH download_activities AS (
    -- まず、ダウンロード活動を抽出
    SELECT 
        actor.user.email_addr,
        src_endpoint.ip,
        web_resources[1].name as file_name,
        from_unixtime(time/1000) as download_time
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
    WHERE eventday = date_format(current_date, '%Y%m%d')
        AND activity_id = 7  -- ダウンロード操作
        AND time >= (to_unixtime(current_timestamp) - 600) * 1000  -- 過去10分間
),
suspicious_downloads AS (
    -- ユーザーとIPアドレスごとに集計
    SELECT 
        email_addr as user_email,
        ip as source_ip,
        COUNT(*) as download_count,
        COUNT(DISTINCT file_name) as unique_files,
        MIN(download_time) as first_download,
        MAX(download_time) as last_download,
        ARRAY_AGG(DISTINCT substr(file_name, 1, 50)) as sample_files
    FROM download_activities
    GROUP BY email_addr, ip
)
-- 50件以上のダウンロードを検知
SELECT *
FROM suspicious_downloads
WHERE download_count >= 50
ORDER BY download_count DESC;
```

</details>

#### 🎯 課題3: 異常なサービスアクセスパターンの検知

**シナリオ解説（初心者向け）**:
攻撃者が組織のシステムに侵入した後、最初に行うのが「偵察活動」です。どのようなサービスやデータにアクセスできるかを探索し、価値の高い情報を見つけようとします。

典型的な偵察活動のパターン：
- **横展開（Lateral Movement）**: 侵入後、他のシステムやサービスへアクセスを試みる
- **権限昇格の試み**: 管理者権限が必要なサービスへのアクセスを試行
- **情報収集**: 様々なサービスを巡回して組織の構造を把握

**検知したい状況の詳細**:
1. **不審なアクセスパターン**
   - 5分以内に3つ以上の異なるサービスへアクセス試行
   - アクセス試行の70%以上が権限エラーで失敗
   - 通常業務では使用しないサービスへのアクセス（特にAdmin系）

2. **なぜこのパターンが危険か**
   - 攻撃者が次の攻撃対象を探している可能性
   - 権限昇格や横展開の前兆
   - システム全体の脆弱性を探る偵察行為
   - 成功すれば重要なシステムへの侵入につながる

3. **正常な行動との区別**
   - 通常のユーザーは自分の業務に必要なサービスのみ使用
   - 権限エラーは稀にしか発生しない（設定ミス程度）
   - 新しいサービスへのアクセスは段階的に増える

<details>
<summary>💡 ヒント1: 使えそうなフィールド</summary>

- `api.service.name`: アクセスされたサービス名
- `status_id = 2`: アクセス失敗
- `actor.user.email_addr`: アクセスユーザー
- `api.operation`: 実行しようとした操作

</details>

<details>
<summary>💡 ヒント2: 不審なアクセスパターン</summary>

1. 異なるサービスへの短時間アクセス
2. 高い失敗率（70%以上）
3. 管理系サービスへのアクセス試行

</details>

<details>
<summary>💡 ヒント3: サービス横断的な分析</summary>

```sql
-- サービスごとの成功/失敗を集計
COUNT(DISTINCT api.service.name) as services_accessed
SUM(CASE WHEN status_id = 2 THEN 1 ELSE 0 END) as failed_attempts
```

</details>

<details>
<summary>✅ 回答例</summary>

```sql
WITH service_access AS (
    -- 過去5分間のすべてのサービスアクセスを抽出
    SELECT 
        actor.user.email_addr,
        api.service.name as service_name,
        api.operation,
        status_id,
        from_unixtime(time/1000) as access_time
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
    WHERE eventday = date_format(current_date, '%Y%m%d')
        AND time >= (to_unixtime(current_timestamp) - 600) * 1000  -- 過去5分間
),
user_service_summary AS (
    -- ユーザーごとにサービスアクセスを集計
    SELECT 
        email_addr as user_email,
        COUNT(DISTINCT service_name) as services_accessed,
        COUNT(*) as total_attempts,
        SUM(CASE 
            WHEN status_id = 2 OR operation IN ('access_denied', 'permission_denied') 
            THEN 1 ELSE 0 
        END) as failed_attempts,
        ARRAY_AGG(DISTINCT service_name) as service_list,
        MIN(access_time) as first_attempt,
        MAX(access_time) as last_attempt
    FROM service_access
    GROUP BY email_addr
),
suspicious_users AS (
    -- 失敗率を計算
    SELECT 
        *,
        CAST(failed_attempts AS DOUBLE) / total_attempts as failure_rate
    FROM user_service_summary
)
-- 3つ以上のサービスにアクセスし、70%以上が失敗しているユーザーを検知
SELECT *
FROM suspicious_users
WHERE services_accessed >= 3 
    AND failure_rate >= 0.7
ORDER BY services_accessed DESC, failure_rate DESC;
```

</details>

#### 🎯 課題4: 地理的に不可能なアクセスの検知

**シナリオ解説（初心者向け）**:
「不可能な移動（Impossible Travel）」は、アカウント乗っ取りを検知する重要な指標です。物理的に移動不可能な速度で異なる場所からアクセスがあった場合、それは同一ユーザーによるものではなく、攻撃者による不正アクセスの可能性が高いです。

このような状況が発生する理由：
- **盗まれた認証情報**: パスワードやトークンが漏洩し、攻撃者が別の場所から使用
- **セッションハイジャック**: 正規ユーザーのセッションを攻撃者が乗っ取り
- **多要素認証の回避**: 何らかの方法で認証を突破された

**検知したい状況の詳細**:
1. **物理的に不可能な移動パターン**
   - 30分以内に異なる国（例：日本→アメリカ）からのアクセス
   - 飛行機でも移動不可能な速度での地点間移動
   - 同時刻に複数の地理的に離れた場所からのアクティビティ

2. **なぜこのパターンが危険か**
   - アカウントが確実に侵害されている証拠
   - 攻撃者と正規ユーザーが同時にアクセスしている可能性
   - 早急な対応をしないと被害が拡大する恐れ
   - 多要素認証が突破されている可能性も示唆

3. **正常な行動との区別**
   - VPN使用時は事前に申請や設定がある
   - 出張時は移動時間を考慮した妥当なアクセスパターン
   - プロキシサービスの利用は組織ポリシーで制限

4. **誤検知を避けるための考慮事項**
   - VPNやプロキシサービスの正規利用
   - クラウドサービスからの自動アクセス
   - モバイルデバイスのローミング

<details>
<summary>💡 ヒント1: 使えそうなフィールド</summary>

- `actor.user.email_addr`: ユーザー識別
- `src_endpoint.location.country`: アクセス元の国
- `src_endpoint.ip`: IPアドレス
- `time`: アクセス時刻

</details>

<details>
<summary>💡 ヒント2: 不可能な移動の判定</summary>

1. 30分以内に異なる国からアクセス
2. 同じ時間帯に複数のIPアドレス
3. 国コードの変化を検出

</details>

<details>
<summary>💡 ヒント3: 複数地点の検出</summary>

```sql
-- ユーザーごとに国とIPを集計
GROUP BY actor.user.email_addr
HAVING COUNT(DISTINCT src_endpoint.location.country) >= 2
```

</details>

<details>
<summary>✅ 回答例</summary>

```sql
WITH user_access_locations AS (
    -- 過去30分間のユーザーアクセスと位置情報を抽出
    SELECT 
        actor.user.email_addr,
        src_endpoint.ip,
        src_endpoint.location.country,
        from_unixtime(time/1000) as access_time
    FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
    WHERE eventday = date_format(current_date, '%Y%m%d')
        AND time >= (to_unixtime(current_timestamp) - 1800) * 1000  -- 過去30分間
        AND src_endpoint.location.country IS NOT NULL
),
user_location_summary AS (
    -- ユーザーごとに国とアクセス時間を集計
    SELECT 
        email_addr as user_email,
        COUNT(DISTINCT country) as country_count,
        COUNT(DISTINCT ip) as unique_ips,
        COUNT(*) as total_access,
        ARRAY_AGG(DISTINCT country) as countries,
        ARRAY_AGG(DISTINCT ip) as ip_addresses,
        MIN(access_time) as first_access,
        MAX(access_time) as last_access,
        date_diff('minute', MIN(access_time), MAX(access_time)) as time_span_minutes
    FROM user_access_locations
    GROUP BY email_addr
)
-- 30分以内に2カ国以上からアクセスがあったユーザーを検知
SELECT *
FROM user_location_summary
WHERE country_count >= 2 
    AND time_span_minutes <= 30
ORDER BY country_count DESC, total_access DESC;
```

</details>

### 実装のポイント

#### 1. SQLファイルの作成

```bash
# lambda/{team_id}/queries/ディレクトリに.sqlファイルを作成
# 例: lambda/red/queries/suspicious_login.sql
```

#### 2. クエリのテスト

書いたSQLはAthenaで直接テストできます：
1. AWSコンソールでAthenaを開く
2. データベース: `amazon_security_lake_glue_db_ap_northeast_1`
3. クエリを実行して結果を確認

#### 3. アラートメッセージの実装

`main.go`の`AlertMessage`構造体に、トリアージに必要な情報を追加：

```go
type AlertMessage struct {
    Title       string `json:"title"`
    // TODO: 以下のようなフィールドを追加
    // UserEmail   string `json:"user_email"`
    // Count       int    `json:"count"`
    // SourceIP    string `json:"source_ip"`
    // Country     string `json:"country"`
}
```

#### 4. 閾値判定の実装

クエリ結果に基づいてアラートを送信するか判定：

```go
// 例: 結果が0件の場合はアラートを送信しない
if len(results) == 0 {
    continue
}
```

### 提出チェックリスト

- [ ] `queries/`ディレクトリに最低1つのSQLファイルを作成
- [ ] Athenaでクエリが正常に動作することを確認
- [ ] AlertMessage構造体に必要なフィールドを追加
- [ ] アラート送信の閾値判定を実装
- [ ] テスト実行でアラートがWarrenに表示されることを確認

### テストとデバッグのヒント

#### 1. Athena でのクエリテスト

AWSコンソールで直接テストする方法：
1. AWS Management Console → Athena
2. データベース `amazon_security_lake_glue_db_ap_northeast_1` を選択
3. 作成したSQLクエリを貼り付けて実行

#### 2. ローカルでの単体テスト

ローカル環境でLambda関数をテストしたい場合は、AWS APIキーを発行できます。

**APIキーの発行を希望する場合**：
- 講師に「ローカルテスト用のAWS APIキーが必要です」と伝えてください
- 以下の権限が付与されたキーを発行します：
  - Athenaクエリ実行権限
  - S3読み取り権限（Security Lake）
  - SNS発行権限（アラート送信）

**ローカルテストの環境設定**：
```bash
# 発行されたキーを環境変数に設定
export AWS_ACCESS_KEY_ID="発行されたキー"
export AWS_SECRET_ACCESS_KEY="発行されたシークレット"
export AWS_REGION="ap-northeast-1"

# ローカルでテスト実行
cd lambda/{team_id}
go test -v
```

**単体テストの例**：
```go
// main_test.go
func TestAlertMessageCreation(t *testing.T) {
    // クエリ結果のモックデータ
    mockResults := []map[string]string{
        {
            "user_email": "test@example.com",
            "count": "10",
            "source_ip": "192.0.2.1",
        },
    }
    
    // アラートメッセージの生成をテスト
    alert := createAlertMessage("suspicious_login", mockResults)
    
    if alert.Title == "" {
        t.Error("Alert title should not be empty")
    }
    
    // 他のフィールドもテスト
}
```

### ボーナス課題

余裕がある方は以下のチャレンジにも挑戦してみましょう：

#### 🎆 チャレンジ1: 複数の検知ルールの組み合わせ

複数のクエリを作成し、それぞれの結果を統合してより高度な検知を実現してみましょう。

#### 🎆 チャレンジ2: アラートの重要度判定

検知結果に基づいてアラートの重要度（Critical/High/Medium/Low）を自動判定するロジックを実装してみましょう。

#### 🎆 チャレンジ3: WITH句を使った高度な分析

SQLのWITH句（CTE: Common Table Expression）を使って、過去のデータと比較した異常検知を実装してみましょう。

### デバッグのポイント

1. **CloudWatch Logsでログ確認**
   - Lambda関数の実行ログを確認
   - エラーメッセージやクエリ結果をチェック

2. **Athenaでクエリを直接実行**
   - SQLが正しいか確認
   - 期待する結果が返ってくるか確認

3. **アラートが表示されない場合**
   - クエリ結果が0件でないか確認
   - アラート送信ロジックを確認
   - Warrenへの通信が成功しているか確認
