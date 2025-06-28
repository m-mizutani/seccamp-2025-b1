# 実習環境とSecurity Lake概要

## 🛠️ 実習環境の説明

### 事前準備済みのTerraform構成

#### インフラ構成図
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Data Sources  │    │  Security Lake  │    │   Analytics     │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ VPC Flow Logs   │────│ Raw Data (S3)   │────│ Athena          │
│ DNS Logs        │    │ OCSF Format     │    │ Lambda (Go)     │
│ CloudTrail      │    │ Partitioned     │    │ SNS Notifications│
│ Application     │    │ Compressed      │    │ Slack Integration│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

#### 主要コンポーネント
```hcl
# Terraform設定例（概要）
resource "aws_securitylake_data_lake" "main" {
  region = "us-east-1"
  
  configuration {
    encryption_configuration {
      kms_key_id = aws_kms_key.security_lake.arn
    }
  }
}

resource "aws_securitylake_subscriber" "lambda" {
  data_lake_arn = aws_securitylake_data_lake.main.arn
  
  source {
    aws_log_source_resource {
      source_name    = "CLOUD_TRAIL_MGMT"
      source_version = "2.0"
    }
  }
}
```

### IAMロール・ポリシー設定

#### Lambda実行ロール
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "athena:GetQueryExecution",
        "athena:GetQueryResults", 
        "athena:StartQueryExecution"
      ],
      "Resource": "arn:aws:athena:*:*:workgroup/security-lake"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::aws-security-data-lake-*",
        "arn:aws:s3:::aws-security-data-lake-*/*"
      ]
    }
  ]
}
```

#### Security Lake設定済み環境
- **データソース**: VPC Flow Logs、DNS Logs、CloudTrail
- **パーティション**: 年/月/日/時間でのパーティション設定
- **保存期間**: Hot Data (30日)、Warm Data (1年)
- **暗号化**: AWS KMS による暗号化

## 📊 Security Lake とは

### Open Cybersecurity Schema Framework (OCSF)

#### OCSFの特徴
- **標準化**: セキュリティデータの統一フォーマット
- **拡張性**: 新しいデータタイプの追加が容易
- **相互運用性**: 異なるツール間でのデータ交換
- **クエリ効率**: 標準化によるSQL分析の高速化

#### OCSFデータ構造例
```json
{
  "metadata": {
    "version": "1.0.0",
    "product": {
      "name": "AWS VPC Flow Logs",
      "vendor_name": "Amazon Web Services"
    }
  },
  "time": 1691836800000,
  "severity_id": 1,
  "class_name": "Network Activity",
  "activity_id": 5,
  "src_endpoint": {
    "ip": "10.0.1.100",
    "port": 443
  },
  "dst_endpoint": {
    "ip": "203.0.113.10", 
    "port": 80
  },
  "connection_info": {
    "protocol_name": "TCP",
    "bytes": 1024,
    "packets": 10
  }
}
```

### パーケットベースのデータレイク

#### S3上のデータ構造
```
s3://aws-security-data-lake-us-east-1-123456789012/
├── aws-cloudtrail-logs/
│   ├── region=us-east-1/
│   │   ├── year=2024/
│   │   │   ├── month=08/
│   │   │   │   ├── day=12/
│   │   │   │   │   ├── hour=10/
│   │   │   │   │   │   └── cloudtrail_logs.parquet
│   │   │   │   │   └── hour=11/
│   │   │   │   └── day=13/
│   │   │   └── month=09/
│   │   └── year=2025/
├── vpc-flow-logs/
│   └── (同様のパーティション構造)
└── dns-logs/
    └── (同様のパーティション構造)
```

#### パーティション戦略の利点

| パーティション | クエリ性能 | コスト削減 | 使用例 |
|---------------|------------|------------|--------|
| 年 | 大幅改善 | 90%削減 | 年次レポート |
| 月 | 大幅改善 | 70%削減 | 月次分析 |
| 日 | 改善 | 50%削減 | 日次監視 |
| 時 | 改善 | 30%削減 | リアルタイム分析 |

### S3上のデータ構造詳細

#### Parquet形式の利点
```
🔍 Parquet vs JSON比較

データサイズ:
・JSON: 100GB
・Parquet: 25GB (75%削減)

クエリ速度:
・JSON: 60秒
・Parquet: 8秒 (7.5倍高速)

コスト:
・Storage: 75%削減
・Query: 87%削減
```

#### 圧縮とエンコーディング
- **圧縮アルゴリズム**: Snappy (高速) / GZIP (高圧縮)
- **カラムナーエンコーディング**: Dictionary、RLE、Delta
- **プッシュダウン述語**: WHERE句での効率的フィルタリング

## 🎯 本日の実習で使用するデータ

### VPC Flow Logs
```json
{
  "metadata": {
    "product": {"name": "Amazon VPC Flow Logs"}
  },
  "time": 1691836800000,
  "src_endpoint": {"ip": "10.0.1.100", "port": 443},
  "dst_endpoint": {"ip": "203.0.113.10", "port": 80},
  "connection_info": {
    "protocol_name": "TCP",
    "bytes": 2048,
    "packets": 15
  },
  "disposition": "Allowed"
}
```

### DNS Logs
```json
{
  "metadata": {
    "product": {"name": "Amazon Route 53 Resolver"}
  },
  "time": 1691836800000,
  "query": {
    "hostname": "suspicious-domain.example.com",
    "type": "A"
  },
  "answers": [
    {"rdata": "192.0.2.100", "type": "A"}
  ],
  "src_endpoint": {"ip": "10.0.1.50"}
}
```

### CloudTrail Logs
```json
{
  "metadata": {
    "product": {"name": "AWS CloudTrail"}
  },
  "time": 1691836800000,
  "api": {
    "operation": "AssumeRole",
    "service": {"name": "sts"}
  },
  "actor": {
    "user": {
      "type": "IAMUser",
      "name": "admin-user"
    }
  },
  "resources": [
    {"uid": "arn:aws:iam::123456789012:role/PowerUser"}
  ]
}
```

## 🔧 実習環境へのアクセス方法

### 1. AWSアカウントログイン
```bash
# 実習用認証情報（講師より配布）
AWS_ACCESS_KEY_ID=AKIA...
AWS_SECRET_ACCESS_KEY=...
AWS_DEFAULT_REGION=us-east-1
```

### 2. Security Lake確認
```bash
# AWS CLIでの確認
aws securitylake list-data-lakes
aws securitylake list-subscribers
```

### 3. Athenaワークブック確認
```sql
-- Security Lakeテーブル一覧確認
SHOW TABLES IN amazon_security_lake_glue_db_us_east_1;

-- サンプルクエリ実行
SELECT * FROM amazon_security_lake_table_us_east_1_cloud_trail_mgmt_2_0 
LIMIT 10;
```

## 📋 実習で使用するツール

### 開発環境
- **IDE**: Visual Studio Code、GoLand
- **言語**: Go 1.21+
- **AWS SDK**: aws-sdk-go-v2

### 監視・通知
- **ログ監視**: CloudWatch Logs
- **メトリクス**: CloudWatch Metrics
- **通知**: SNS → Slack Webhook

### CI/CD
- **リポジトリ**: GitHub
- **自動化**: GitHub Actions
- **デプロイ**: AWS SAM、Terraform

## 🚀 次のステップ

### これから実装するもの
1. **Go Lambda関数**: SQL実行とアラート生成
2. **検知ルール**: 3つのシナリオから選択
3. **CI/CDパイプライン**: 自動ビルド・デプロイ
4. **通知システム**: Slackへのアラート通知

### 学習ポイント
- **OCSF形式**: セキュリティデータの標準フォーマット理解
- **パフォーマンス**: パーティション活用による高速クエリ
- **運用考慮**: コスト効率とセキュリティのバランス

---

**準備完了！Go言語でのLambda実装に進みましょう！** 🚀 