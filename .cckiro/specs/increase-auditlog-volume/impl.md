# 実装計画ファイル - auditlogのログ量10倍化

## 実装ステップ

### Phase 1: Terraform インフラストラクチャ

#### [x] 1. S3バケットの作成
- `terraform/s3.tf`に専用のS3バケットリソースを追加
- バケット名: `seccamp2025-b1-auditlog-seeds`
- パブリックアクセスブロック設定

#### [x] 2. IAMポリシーの作成と付与
- `terraform/iam.tf`にauditlog Lambda用のS3読み取りポリシーを追加
- 必要な権限: `s3:GetObject`, `s3:GetObjectVersion`, `s3:ListBucket`
- 既存のauditlog Lambdaロールにポリシーをアタッチ

#### [x] 3. Lambda環境変数の追加
- `terraform/lambda.tf`の既存auditlog Lambda定義に環境変数を追加
- `SEED_BUCKET_NAME`環境変数にS3バケット名を設定

#### [ ] 4. Terraform適用
- `terraform plan`で変更内容を確認
- `terraform apply`でリソースを作成

### Phase 2: Lambda関数の改修

#### [x] 5. 依存パッケージの追加
- `terraform/lambda/auditlog/go.mod`にAWS SDK for Go v2の依存を追加
- 必要なパッケージ: `github.com/aws/aws-sdk-go-v2/config`, `github.com/aws/aws-sdk-go-v2/service/s3`

#### [x] 6. S3クライアントの初期化
- `main.go`にグローバル変数としてS3クライアントを定義
- `init()`関数でS3クライアントを初期化

#### [x] 7. キャッシュ機構の実装
- グローバル変数でseedデータのキャッシュを管理
- `sync.RWMutex`で並行アクセスを制御

#### [x] 8. S3ダウンロード関数の実装
- `downloadFromS3()`関数を実装
- 環境変数からバケット名を取得
- S3からオブジェクトをダウンロード

#### [x] 9. getSeedData関数の実装
- キャッシュチェックロジック
- S3ダウンロード呼び出し
- エラーハンドリング（エラー時は処理終了）

#### [x] 10. generateLogs関数の修正
- 埋め込みseedデータの代わりに`getSeedData()`を呼び出す
- エラーハンドリングを追加

#### [x] 11. エラーレスポンスの改善
- S3アクセスエラー時の適切なエラーメッセージ
- HTTPステータスコード500でレスポンス

### Phase 3: ビルドとデプロイ

#### [x] 12. Lambda関数のビルド
- `cd terraform/lambda/auditlog`
- `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap main.go`
- ビルドエラーがないことを確認

#### [ ] 13. Terraform再適用
- Lambda関数の更新をTerraformで適用
- `terraform apply`でLambda関数を更新

### Phase 4: seedデータの準備とテスト

#### [x] 14. 大容量seedデータの生成
- tools/putseedツールを作成
- 既存のseedデータを基に10倍のデータ量を持つファイルを生成可能
- ファイル名: `large-seed.bin.gz`

#### [x] 15. S3へのseedデータアップロード
- tools/putseedツールで自動アップロード
- `cd tools/putseed && go run main.go`
- オプションで手動アップロードも可能

#### [ ] 16. Lambda関数のテスト
- Lambda関数のテスト実行
- S3からseedデータが正常にダウンロードされることを確認
- ログ生成が正常に動作することを確認

### Phase 5: クリーンアップとドキュメント

#### [ ] 17. 不要な埋め込みデータの削除（オプション）
- 必要に応じて`//go:embed`ディレクティブを削除
- ただし、完全に削除せず、コメントアウトに留めることを推奨

#### [x] 18. README更新
- seedデータの管理方法について記載
- S3バケットへのアップロード手順を記載
- tools/putseedの使用方法を記載

## 実装時の注意事項

1. **後方互換性の維持**
   - APIインターフェースは変更しない
   - レスポンス形式も維持

2. **エラーハンドリング**
   - S3アクセスエラーは即座に処理を終了
   - 詳細なエラーログをCloudWatch Logsに出力

3. **パフォーマンス考慮**
   - warm start時はキャッシュを活用
   - S3ダウンロードは初回のみ

4. **セキュリティ**
   - IAMポリシーは最小権限
   - S3バケットはパブリックアクセス不可

この実装計画でよろしいでしょうか？問題がある場合はフィードバックをお願いします。