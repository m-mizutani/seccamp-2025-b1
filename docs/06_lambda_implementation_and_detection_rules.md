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

#### 🎯 課題1: 夜間の大量ダウンロード検知

**目標**: 深夜から早朝にかけて管理者権限を持つユーザーが大量のファイルをダウンロードする異常パターンを検出する

**検知したい状況**:
- 18:00から翌9:00の間のアクティビティ
- 管理者ロールを持つユーザーによる操作
- 短時間に複数のファイルダウンロード

<details>
<summary>💡 ヒント1: 使えそうなフィールド</summary>

- `activity_id = 7`: ダウンロード操作
- `actor.user.email_addr`: ユーザーメール（adminを含むメールアドレス）
- `EXTRACT(HOUR FROM from_unixtime(time/1000))`: 時間抽出（UTC）
- `web_resources[1].name`: ダウンロードされたファイル名

</details>

<details>
<summary>💡 ヒント2: SQLの構成</summary>

1. WHERE句で時間帯とダウンロード操作を絞り込み
2. adminユーザーの特定（メールアドレスにadminを含む）
3. GROUP BYでユーザーごとに集計
4. 時間帯の条件は `HOUR >= 18 OR HOUR <= 9` で指定

</details>

<details>
<summary>💡 ヒント3: 時間帯の判定</summary>

```sql
-- UTC時間での18:00-9:00の判定
WHERE EXTRACT(HOUR FROM from_unixtime(time/1000)) >= 18 
   OR EXTRACT(HOUR FROM from_unixtime(time/1000)) <= 9
```

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    COUNT(*) as download_count,
    COUNT(DISTINCT web_resources[1].name) as unique_files,
    MIN(from_unixtime(time/1000)) as first_download,
    MAX(from_unixtime(time/1000)) as last_download,
    ARRAY_AGG(DISTINCT web_resources[1].name) as file_list
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND activity_id = 7
    AND actor.user.email_addr LIKE '%admin%'
    AND (EXTRACT(HOUR FROM from_unixtime(time/1000)) >= 18 
         OR EXTRACT(HOUR FROM from_unixtime(time/1000)) <= 9)
GROUP BY actor.user.email_addr
HAVING COUNT(*) >= 5
ORDER BY download_count DESC;
```

</details>

#### 🎯 課題2: 外部リンクへのバーストアクセス検知

**目標**: 外部公開設定（anyone with link）されたファイルへの短時間での集中的なアクセスを検出

**検知したい状況**:
- 10:00-16:00の業務時間帯
- 15分間隔でのバーストパターン（自動化されたアクセスの可能性）
- 外部ユーザーまたは異常なIPからのアクセス

<details>
<summary>💡 ヒント1: 使えそうなフィールド</summary>

- `activity_id = 2`: 読み取り操作
- `actor.user.type_id`: ユーザータイプ（99=外部ユーザー）
- `time`: アクセス時刻（分単位での集計に使用）
- `src_endpoint.ip`: アクセス元IP

</details>

<details>
<summary>💡 ヒント2: 時間窓での集計</summary>

```sql
-- 15分単位での集計例
date_trunc('minute', from_unixtime(time/1000)) - 
    interval '1' minute * (minute(from_unixtime(time/1000)) % 15)
```

</details>

<details>
<summary>💡 ヒント3: バーストパターンの検出</summary>

1. 15分単位で時間を区切る
2. 各時間窓でのアクセス数を集計
3. 短時間での異常なアクセス集中を検出

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    date_format(from_unixtime(time/1000), '%Y-%m-%d %H:%i') as time_window,
    COUNT(*) as access_count,
    COUNT(DISTINCT src_endpoint.ip) as unique_ips,
    COUNT(DISTINCT actor.user.email_addr) as unique_users,
    ARRAY_AGG(DISTINCT web_resources[1].name) as accessed_files
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND activity_id = 2
    AND EXTRACT(HOUR FROM from_unixtime(time/1000)) BETWEEN 10 AND 16
    AND (actor.user.type_id = 99 OR actor.user.email_addr LIKE '%external%')
GROUP BY date_format(from_unixtime(time/1000), '%Y-%m-%d %H:%i')
HAVING COUNT(*) >= 20
ORDER BY time_window DESC;
```

</details>

#### 🎯 課題3: VPN経由の水平移動攻撃検知

**目標**: VPN接続元からの異常なリソースアクセスパターン（水平移動の兆候）を検出

**検知したい状況**:
- 9:00-18:00の業務時間内
- VPN経由のアクセス（IPアドレスパターンで判定）
- 短時間に複数の異なるリソースへアクセス

<details>
<summary>💡 ヒント1: 使えそうなフィールド</summary>

- `src_endpoint.ip`: VPN IPアドレスの特定
- `web_resources[1].name`: アクセスされたリソース
- `api.operation`: 実行された操作
- VPNユーザーの特定方法を考えてみましょう

</details>

<details>
<summary>💡 ヒント2: 水平移動の特徴</summary>

1. 短時間に多くの異なるリソースにアクセス
2. 通常とは異なるアクセスパターン
3. 複数のサービスやファイルを横断的に探索

</details>

<details>
<summary>💡 ヒント3: VPNアクセスの判定</summary>

- 特定のIPレンジ（10.x.x.x, 172.16-31.x.x, 192.168.x.x）
- または特定のlocationパターン
- emailアドレスにvpnを含むユーザー

</details>

<details>
<summary>✅ 回答例</summary>

```sql
SELECT 
    actor.user.email_addr as user_email,
    src_endpoint.ip as source_ip,
    COUNT(DISTINCT web_resources[1].name) as unique_resources,
    COUNT(DISTINCT api.operation) as unique_operations,
    COUNT(*) as total_access,
    MIN(from_unixtime(time/1000)) as first_access,
    MAX(from_unixtime(time/1000)) as last_access,
    ARRAY_AGG(DISTINCT web_resources[1].name) as resource_list
FROM amazon_security_lake_glue_db_ap_northeast_1.amazon_security_lake_table_ap_northeast_1_ext_google_workspace_1_0
WHERE eventday = date_format(current_date, '%Y%m%d')
    AND EXTRACT(HOUR FROM from_unixtime(time/1000)) BETWEEN 9 AND 18
    AND (src_endpoint.ip LIKE '10.%' 
         OR src_endpoint.ip LIKE '172.%' 
         OR src_endpoint.ip LIKE '192.168.%'
         OR actor.user.email_addr LIKE '%vpn%')
GROUP BY actor.user.email_addr, src_endpoint.ip
HAVING COUNT(DISTINCT web_resources[1].name) >= 10
    AND time_diff('minute', MIN(from_unixtime(time/1000)), MAX(from_unixtime(time/1000))) <= 30
ORDER BY unique_resources DESC;
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
