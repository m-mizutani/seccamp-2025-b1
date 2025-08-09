# loggen - Log Seed Generator

Google Workspace監査ログのシードデータを生成するツール。

## 機能

- ログシードの生成（時間分布、異常パターンを含む）
- ローカルファイルまたはS3への出力
- 複数の出力フォーマット対応（JSON、バイナリ、圧縮バイナリ）
- 常時発生型異常パターンの生成（24時間継続的に検知可能）

## インストール

```bash
go build -o loggen
```

## 使用方法

### 基本的な使用（ローカル出力）

```bash
# デフォルト設定でシードを生成
./loggen generate

# 特定の日付でシードを生成
./loggen generate --date 2024-08-12

# 異常ログの比率を変更
./loggen generate --anomaly-ratio 0.20
```

### S3への出力

```bash
# S3バケットに直接出力
./loggen generate --output s3://seccamp2025-b1-auditlog-seeds/

# 圧縮バイナリ形式でS3に出力（推奨）
./loggen generate --output s3://seccamp2025-b1-auditlog-seeds/ --format binary-compressed
```

### その他のコマンド

```bash
# シードデータの検証
./loggen validate --input ./output/seeds/day_2024-08-12.bin.gz

# 統計情報の表示
./loggen stats --input ./output/seeds/day_2024-08-12.bin.gz

# ログのプレビュー
./loggen preview --input ./output/seeds/day_2024-08-12.bin.gz --time-range "10:00-11:00"
```

## オプション

### generate コマンド

- `--date`: 生成する日付（YYYY-MM-DD形式、デフォルト: 本日）
- `--output`: 出力先（ローカルディレクトリまたはs3://bucket/prefix/、デフォルト: ./output）
- `--anomaly-ratio`: 異常ログの比率（0.0-1.0、デフォルト: 0.15）
- `--format`: 出力フォーマット（json, binary, binary-compressed、デフォルト: binary-compressed）
- `--dry-run`: 実際にファイルを書き込まずに実行

## S3出力について

### 必要な権限

S3に出力する場合、以下のIAM権限が必要です：

- `s3:PutObject` - 指定バケットへの書き込み権限
- `s3:ListBucket` - バケットの存在確認（オプション）

### 認証設定

AWS認証は以下の優先順位で解決されます：

1. 環境変数（`AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY`）
2. AWS認証情報ファイル（`~/.aws/credentials`）
3. IAMロール（EC2、Lambda等で実行時）

### リージョン設定

AWSリージョンは以下の優先順位で解決されます：

1. 環境変数 `AWS_REGION`
2. 環境変数 `AWS_DEFAULT_REGION`
3. デフォルト値 `ap-northeast-1`

例：
```bash
export AWS_REGION=ap-northeast-1
./loggen generate --output s3://seccamp2025-b1-auditlog-seeds/
```

### 出力パス

S3に出力する場合、以下のパスに保存されます：

```
s3://bucket-name/prefix/seeds/large-seed.bin.gz  # binary-compressed形式
s3://bucket-name/prefix/seeds/day_YYYY-MM-DD.json  # json形式
s3://bucket-name/prefix/seeds/day_YYYY-MM-DD.bin   # binary形式
```

## auditlog Lambdaとの連携

このツールで生成したシードデータは、auditlog Lambda関数で使用されます：

1. `binary-compressed`形式で生成（推奨）
2. S3バケット `seccamp2025-b1-auditlog-seeds` にアップロード
3. ファイル名を `large-seed.bin.gz` とする（Lambda側の期待値）

例：
```bash
./loggen generate \
  --output s3://seccamp2025-b1-auditlog-seeds/ \
  --format binary-compressed
```

## トラブルシューティング

### S3アップロードエラー

- AWS認証情報が正しく設定されているか確認
- S3バケットへの書き込み権限があるか確認
- バケット名とリージョンが正しいか確認

## 異常パターンについて

loggenは以下の異常パターンを生成します：

### 時間帯限定パターン（既存）
1. **夜間の管理者による大量ダウンロード** - 18:00-9:00に発生
2. **外部リンクアクセスバースト** - 10:00-16:00の15分間隔で発生
3. **VPN経由の水平移動攻撃** - 9:00-18:00の業務時間内に発生

### 常時発生型パターン（新規）
4. **高頻度認証攻撃** - 24時間継続、1分に3-5回の認証失敗
5. **超高速データ窃取** - 24時間継続、1分に10-15件のダウンロード
6. **マルチサービス不正アクセス** - 24時間継続、複数サービスへの探索
7. **地理的同時アクセス** - 24時間継続、2カ国から同時操作

常時発生型パターンにより、5分間隔でのクエリでも確実に異常を検知できるようになりました。