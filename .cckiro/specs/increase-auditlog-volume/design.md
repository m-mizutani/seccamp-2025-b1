# 設計ファイル - auditlogのログ量10倍化

## アーキテクチャ概要

### 現在の構成
```
Lambda (auditlog)
  └─ 埋め込みseedデータ (day_2024-08-12.bin.gz)
      └─ メモリ展開 → ログ生成
```

### 新しい構成
```
S3 (専用バケット: seccamp2025-b1-auditlog-seeds)
  └─ seedデータ (large-seed.bin.gz)
      ↓ ダウンロード
Lambda (auditlog)
  └─ メモリキャッシュ
      └─ メモリ展開 → ログ生成
```

## 詳細設計

### 1. S3バケット構成
```
seccamp2025-b1-auditlog-seeds/
└── seeds/
    └── large-seed.bin.gz  # 10倍のデータ量を含むseedファイル
```

- バケット名: `seccamp2025-b1-auditlog-seeds`
- プレフィックス: `seeds/`
- ファイル名: `large-seed.bin.gz`

### 2. Lambda関数の改修

#### 2.1 グローバル変数での管理
```go
var (
    // キャッシュされたseedデータ
    cachedSeedData []byte
    cacheMutex     sync.RWMutex
    
    // S3クライアント（再利用）
    s3Client *s3.Client
)
```

#### 2.2 初期化処理
```go
func init() {
    // S3クライアントの初期化
    cfg, _ := config.LoadDefaultConfig(context.Background())
    s3Client = s3.NewFromConfig(cfg)
}
```

#### 2.3 Seedデータ取得ロジック
```go
func getSeedData(ctx context.Context) ([]byte, error) {
    // キャッシュチェック（warm start対応）
    cacheMutex.RLock()
    if cachedSeedData != nil {
        cacheMutex.RUnlock()
        return cachedSeedData, nil
    }
    cacheMutex.RUnlock()
    
    // S3からダウンロード
    data, err := downloadFromS3(ctx)
    if err != nil {
        // エラーをそのまま返す
        return nil, fmt.Errorf("failed to download seed data from S3: %w", err)
    }
    
    // キャッシュに保存
    cacheMutex.Lock()
    cachedSeedData = data
    cacheMutex.Unlock()
    
    return data, nil
}
```

#### 2.4 S3ダウンロード処理
```go
func downloadFromS3(ctx context.Context) ([]byte, error) {
    bucketName := os.Getenv("SEED_BUCKET_NAME")
    objectKey := "seeds/large-seed.bin.gz"
    
    result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(objectKey),
    })
    if err != nil {
        return nil, err
    }
    defer result.Body.Close()
    
    return io.ReadAll(result.Body)
}
```

### 3. Terraform設計

#### 3.1 S3バケット
```hcl
resource "aws_s3_bucket" "auditlog_seeds" {
  bucket = "${var.basename}-auditlog-seeds"
  
  tags = merge(local.common_tags, {
    Name = "${var.basename}-auditlog-seeds"
    Type = "seed-storage"
  })
}



# パブリックアクセスブロック
resource "aws_s3_bucket_public_access_block" "auditlog_seeds" {
  bucket = aws_s3_bucket.auditlog_seeds.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
```

#### 3.2 IAMポリシー追加
```hcl
resource "aws_iam_policy" "auditlog_lambda_s3_seeds" {
  name        = "${var.basename}-auditlog-lambda-s3-seeds-policy"
  description = "S3 permissions for auditlog Lambda to access seed data"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:GetObjectVersion"
        ]
        Resource = "${aws_s3_bucket.auditlog_seeds.arn}/seeds/*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket"
        ]
        Resource = aws_s3_bucket.auditlog_seeds.arn
      }
    ]
  })
}

# ポリシーアタッチ
resource "aws_iam_role_policy_attachment" "auditlog_lambda_s3_seeds" {
  policy_arn = aws_iam_policy.auditlog_lambda_s3_seeds.arn
  role       = aws_iam_role.auditlog_lambda.name
}
```

#### 3.3 Lambda環境変数
```hcl
# lambda.tfの既存のauditlog Lambda定義に追加
environment {
  variables = {
    SEED_BUCKET_NAME = aws_s3_bucket.auditlog_seeds.bucket
  }
}
```

### 4. エラーハンドリング

#### 4.1 S3アクセスエラー
- S3からのダウンロードに失敗した場合、エラーを返して処理を終了
- CloudWatch Logsにエラーを記録
- Lambda関数はエラーレスポンスを返す

#### 4.2 メモリ制限対策
- 大容量seedデータの圧縮を維持（gzip形式）
- 必要に応じてLambdaのメモリサイズを調整（現在の設定を確認後）

### 5. 運用考慮事項

#### 5.1 Seedデータの更新
1. 新しいseedファイルを`large-seed.bin.gz`としてローカルで作成
2. AWS CLIまたはコンソールから手動でS3にアップロード：
   ```bash
   aws s3 cp large-seed.bin.gz s3://seccamp2025-b1-auditlog-seeds/seeds/
   ```

#### 5.2 モニタリング
- S3ダウンロードエラーはCloudWatch Logsで監視
- Lambda関数のコールドスタート時間をメトリクスで確認

## セキュリティ考慮事項

1. **最小権限の原則**
   - Lambda関数は特定のS3バケット/プレフィックスのみアクセス可能
   - GetObjectとListBucketのみ許可

2. **データ保護**
   - パブリックアクセスを完全にブロック

3. **アクセス監査**
   - S3アクセスログを有効化（必要に応じて）
   - CloudTrailでAPIコールを記録

この設計でよろしいでしょうか？問題がある場合はフィードバックをお願いします。