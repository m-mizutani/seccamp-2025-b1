# 設計

## 概要

- このプロジェクトはセキュリティ監視を実践するためのシステムを構築します
- 各所からログを収集し、それを分析しアラートを検知します
- 基本的なデータフローは以下のとおりです
  - Rawログデータを保管するためのS3バケットにログデータがPutされる
  - Putされると SNS → SQS と通知が送信されLambdaが起動し、対象オブジェクトをSecurity Lake に格納可能なparquetファイルに変換する
  - Athena に対してクエリを定期実行するLambdaがあり、その結果を通知するSNSがある

このようなシステムを以下の要求を満たすterraformの構築をしてください。

## 前提

- このシステムは複数のチームが利用することを想定する
- チーム名は以下の通りとする
  - `blue`
  - `green`
  - `red`
  - `yellow`
  - `purple`
  - `orange`
  - `pink`
  - `brown`
  - `gray`
  - `black`
  - `white`
- 原則、チームごとに必要なリソースはモジュール化する


## 要求事項

- Rawログデータを保管するためのS3バケットを作成する
  - バケット名は `${basename}-${team_name}-raw-logs` とする
  - Public access は禁止する
  - versioning は有効にする
- このバケットに書き込みおよびListBucketができるLambda用のロールを作成する
  - Role名は `lambda-${TeamName}-importer-role` とする
    - `blue` の場合は `lambda-blue-importer-role` となる
  - このロールには以下の権限を付与する
    - `${basename}-${team_name}-raw-logs` に対してS3のデータ書き込み
    - `${basename}-${team_name}-raw-logs` に対してS3のデータList
    - Lambdaの基本権限
- RawログデータがPutされるとSNSに通知が送信される
  - 通知先のSNSの名前は `${basename}-${team_name}-raw-logs-sns` とする
  - Create Object のイベントのみ通知する
  - 通知先は SQS とする
  - 通知先のSQSの名前は `${basename}-${team_name}-raw-logs-queue` とする
  - このSNSに対してLambdaが起動する
- RawログデータをParquetへ変換するLambdaを作成する。このLambdaは以下のようなロジックで動作する
  - 対象オブジェクトを取得する
  - 対象オブジェクトのデータをParseする。スキーマは "Rawログのスキーマ" 節を参照すること
  - パースされたデータをparquetオブジェクトに変換する。変換したファイルをS3バケットにアップロードする
    - アップロードするパスはSecurity Lakeのルールに準拠する
    - カスタムログソースとして変換したデータを登録する。ログソースの名前は `${team_name}` とする
- Security Lake を構築する
  - カスタムログソースとしてRawログデータを変換したものを登録する
    - ログソースの名前は `${team_name}` とする
  - Glue crawler は1日1度のみ実行されるようにする
- Security Lake のログをAthenaでクエリを実行するLambdaのロールを作成する
  - Role名は `lambda-${TeamName}-detector-role` とする
    - `blue` の場合は `lambda-blue-detector-role` となる
  - このRoleは他にも以下の権限を付与する
    - Athena のクエリ実行
    - `${basename}-alerts-sns` に対してSNSの通知
    - Lambdaの基本権限
- アラート通知用SNSを作成する。名前は `${basename}-alerts-sns` とする
- IAM Identity Provider を作成する
  - GitHub Actions からの認証を行うためのIdentity Providerを作成する
  - 許可するリポジトリは以下の通り
    - `github.com/m-mizutani/seccamp-2025-b1`
    - `github.com/seccamp2025-b/b1-secmon`
  - GitHub Actions からはterraformのapplyを行うことができるようにする。そのためAdministratorAccess 、あるいはそれに相当する適切な権限を付与する
    - このIdentity Provider は `github.tf` に配置する
    - このためのRoleは `GitHubActionsRole` とする

## 制約事項

- 変数
  - basenameは seccamp2025-b1-poc とする。これは variables.tf に定義され、変更可能にする
  - regionは ap-northeast-1 とする。これは variables.tf に定義され、変更可能にする
- terraformのコードは /terraform ディレクトリ以下に配置する
- lambda はそれぞれ /terraform/lambda/ ディレクトリ以下に配置する
  - parquet 変換用の lambda は /terraform/lambda/converter/ 以下に作成する
- Lambda はGoで作成する
  - Goのバージョンは 1.23.4 とする
  - armアーキテクチャを使う
  - AWSのSDKはv2を利用する
  - terraform の apply によってデプロイされるようにする
  - それぞれ必要最低限の権限が付与されたロールが作成され、割り当てられるようにする
  - 簡略化のためにディレクトリはフラットな構造にする
  - 単体テストにおいて各AWSサービスへのアクセスはモックを利用する
    - 例えばRawログの取得や変換されたparquetオブジェクトのアップロードはモックを利用し、期待する通りにデータが処理されたかを検証する
- terraformのファイルはリソースごとに分割する