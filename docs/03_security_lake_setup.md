# 実習環境とSecurity Lake概要

## 🛠️ 実習環境の構成

### 事前準備済みの環境

**アーキテクチャ概要**
```
データソース → Security Lake → 検知・分析
├─ CloudTrail      ├─ OCSF正規化    ├─ Athena クエリ
├─ VPC Flow Logs   ├─ S3 保存       ├─ Lambda 検知
└─ DNS Logs        └─ パーティション └─ SNS 通知
```

**主要コンポーネント**
- **Security Lake**: AWS マネージドサービス、OCSF標準フォーマット
- **データソース**: CloudTrail、VPC Flow Logs、DNS Logs
- **検知基盤**: Athena + Lambda (Go) + SNS

**権限設定**
- Security Lake読み取り権限
- Athena実行権限  
- SNS発行権限

## 📊 Security Lake とは

### AWS Security Lake の特徴

**OCSF (Open Cybersecurity Schema Framework) 標準化**
- **統一フォーマット**: 異なるデータソースを標準スキーマに正規化
- **SQL分析効率化**: 統一されたフィールド名でのクエリ作成
- **マルチベンダー対応**: AWS、Azure、GCP等のクラウド横断分析

**基本的なOCSF構造**
```json
{
  "time": 1691836800000,
  "class_uid": 4001,
  "activity_id": 1,
  "src_endpoint": {"ip": "10.0.1.100"},
  "dst_endpoint": {"ip": "203.0.113.10"},
  "connection_info": {"protocol_name": "TCP"}
}
```

### データ保存とパーティション

**S3上の効率的データ構造**
```
s3://security-lake-bucket/
├── aws-cloudtrail-logs/
│   └── region=us-east-1/year=2024/month=08/day=12/
├── vpc-flow-logs/
└── dns-logs/
```

**Parquet形式の利点**
- **圧縮効率**: JSON比75%削減
- **クエリ高速化**: 7.5倍高速  
- **コスト削減**: ストレージ・クエリ両方でコスト削減

## 🎯 本日の実習で使用するデータ

### 実習用サンプルデータ

**VPC Flow Logs (ネットワーク通信)**
- **src_endpoint**: 送信元IP・ポート
- **dst_endpoint**: 宛先IP・ポート  
- **connection_info**: プロトコル・バイト数・パケット数
- **disposition**: 通信許可/拒否

**DNS Logs (DNS解決)**
- **query**: ドメイン名・クエリタイプ
- **answer**: 解決されたIPアドレス
- **response_time**: 応答時間

**CloudTrail Logs (API呼び出し)**
- **actor**: 実行ユーザー・ロール
- **api**: 呼び出されたAPIとパラメータ  
- **target_resource**: 操作対象リソース

### 実習で作成する検知ルール

**1. 異常なネットワーク通信検知**
- 通常と異なる大量データ転送
- 業務時間外の通信パターン

**2. 特権操作の監視**
- 管理者権限での異常な操作
- 重要リソースへの不正アクセス

**3. DNS異常パターン検知**
- DGA（自動生成）ドメインへのクエリ
- 異常に大きなDNSクエリ

---

**🎯 次のステップ**  
実際にGo言語でLambda関数を実装し、これらの検知ルールを作成します！ 