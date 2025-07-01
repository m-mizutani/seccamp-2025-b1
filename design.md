# 設計

## 概要

- このプロジェクトはセキュリティ監視を実践するためのシステムを構築します
- 各所からログを収集し、それを分析しアラートを検知します
- 基本的なデータフローは以下のとおりです
  - Rawログデータを保管するためのS3バケットにログデータがPutされる
  - Putされると SNS → SQS と通知が送信されLambdaが起動し、対象オブジェクトをSecurity Lake に格納可能なparquetファイルに変換する
  - これらのログとは別に、CloudTrail のログが Security Lake に格納される
  - Athena に対してクエリを定期実行するLambdaがあり、その結果を通知するSNSがある

このようなシステムを以下の要求を満たすterraformの構築をしてください。

## 要求事項

- Rawログデータを保管するためのS3バケットを作成する
  - バケット名は `${basename}-raw-logs` とする
  - Public access は禁止する
  - versioning は有効にする
  - このバケットに書き込みおよびListBucketができるRoleを作成する
- RawログデータがPutされるとSNSに通知が送信される
  - 通知先のSNSの名前は `${basename}-raw-logs-sns` とする
  - Create Object のイベントのみ通知する
  - 通知先は SQS とする
  - 通知先のSQSの名前は `${basename}-raw-logs-queue` とする
  - このSNSに対してLambdaが起動する
- RawログデータをParquetへ変換するLambdaを作成する。このLambdaは以下のようなロジックで動作する
  - 対象オブジェクトを取得する
  - 対象オブジェクトのデータをParseする。スキーマは "Rawログのスキーマ" 節を参照すること
  - パースされたデータをparquetオブジェクトに変換する。変換したファイルをS3バケットにアップロードする
    - アップロードするパスはSecurity Lakeのルールに準拠する
    - カスタムログソースとして変換したデータを登録する。ログソースの名前は `service-logs` とする
- Security Lake を構築する
  - 以下のサービスのログを収集する
    - CloudTrail
  - カスタムログソースとしてRawログデータを変換したものを登録する
    - ログソースの名前は `service-logs` とする
  - Glue crawler は1日1度のみ実行されるようにする
- アラート検知用のLambdaを作成する。このLambdaは以下のようなロジックで動作する
  - Security Lake のログにAthena経由でクエリを実行する
  - クエリはSQLで記述され、そのLambdaにgo embedによって同梱されている
  - クエリが終了したら、その結果を取得する
  - 実行結果が空でない場合はSNSに通知する
  - 実行結果が空の場合は何もしない
  - SNSの名前は `${basename}-alerts-sns` とする
  - SNSへの通知のスキーマは以下のとおり
    - `title`: アラートを一言で説明する
    - `description`: アラートの詳細を説明する。何を見つけようとしたのか、どう見つかったのかなどを記載する
    - `attrs`: アラートに関連する属性情報をmap形式で記述する
- IAM Identity Provider を作成する
  - GitHub Actions からの認証を行うためのIdentity Providerを作成する
  - 許可するリポジトリは `github.com/m-mizutani/seccamp-2025-b1` とする
  - GitHub Actions からはterraformのapplyを行うことができるようにする。そのためAdministratorAccess 、あるいはそれに相当する適切な権限を付与する


## Rawログのスキーマ

以下のようなスキーマのログがJSONL形式で保存されます。

- `id` : ログごとにユニークなID
- `timestamp`: RFC3339形式の時刻
- `user`: 操作をした認証済みユーザ名。 `login` の場合は認証試行したユーザ名
- `action`: `read`, `write` はドキュメントの操作、 `login` はログイン試行
- `target`: 対象となるドキュメント。 `login` の場合は空
- `success`: `read`, `write` は認可の成否、 `login` は認証の成否
- `remote` : 操作元のIPアドレス（v4のみ）

## 制約事項

- 変数
  - basenameは seccamp2025-b1-poc とする。これは variables.tf に定義され、変更可能にする
  - regionは ap-northeast-1 とする。これは variables.tf に定義され、変更可能にする
- terraformのコードは /terraform ディレクトリ以下に配置する
- lambda はそれぞれ /terraform/lambda/ ディレクトリ以下に配置する
  - parquet 変換用の lambda は /terraform/lambda/converter/ 以下に作成する
  - アラート検知用の lambda は /terraform/lambda/detector/ 以下に作成する
- Lambda はGoで作成する
  - Goのバージョンは 1.23.4 とする
  - terraform の apply によってデプロイされるようにする
  - それぞれ必要最低限の権限が付与されたロールが作成され、割り当てられるようにする
  - 簡略化のためにディレクトリはフラットな構造にする
- terraformのファイルはリソースごとに分割する