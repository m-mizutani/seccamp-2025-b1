# Google Workspace 監査ログフォーマット

セキュリティキャンプ2025 B1講義で使用するGoogle Workspace監査ログの構造とフィールドについて説明します。

## 1. ログの基本構造

Google Workspace監査ログは、Admin SDK Reports APIの形式を基本とし、以下の構造を持ちます：

```json
{
  "kind": "audit#activity",
  "id": {
    "time": "2024-08-12T10:15:30.123456Z",
    "uniqueQualifier": "358068855354",
    "applicationName": "drive",
    "customerId": "C03az79cb"
  },
  "actor": {
    "callerType": "USER",
    "email": "user@muhai-academy.com",
    "profileId": "114511147312345678901"
  },
  "ownerDomain": "muhai-academy.com",
  "ipAddress": "203.0.113.255",
  "events": [
    {
      "type": "access",
      "name": "view",
      "parameters": [
        {
          "name": "doc_id",
          "value": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
        },
        {
          "name": "doc_title",
          "value": "学習進捗データ"
        }
      ]
    }
  ]
}
```

## 2. 主要フィールドの説明

### 2.1 基本情報（id）
- **time**: イベント発生時刻（ISO 8601形式）
- **uniqueQualifier**: イベントの一意識別子
- **applicationName**: サービス名（login, drive, calendar, admin等）
- **customerId**: Google Workspaceの顧客ID

### 2.2 実行者情報（actor）
- **callerType**: 呼び出し元タイプ（通常は"USER"）
- **email**: 実行者のメールアドレス
- **profileId**: ユーザーのプロフィールID

### 2.3 環境情報
- **ownerDomain**: 組織のドメイン名
- **ipAddress**: アクセス元IPアドレス

### 2.4 イベント詳細（events）
- **type**: イベントタイプ（access, admin, login等）
- **name**: 具体的なアクション名
- **parameters**: イベント固有のパラメータ

## 3. アプリケーション別のイベントタイプ

### 3.1 Login イベント
```json
{
  "applicationName": "login",
  "events": [
    {
      "type": "login",
      "name": "login_success",
      "parameters": [
        {
          "name": "login_type",
          "value": "google_password"
        },
        {
          "name": "login_challenge_method",
          "multiStrValue": ["password"]
        }
      ]
    }
  ]
}
```

**主要なイベント名**:
- `login_success`: ログイン成功
- `login_failure`: ログイン失敗
- `logout`: ログアウト
- `suspicious_login`: 不審なログイン

### 3.2 Drive イベント
```json
{
  "applicationName": "drive",
  "events": [
    {
      "type": "access",
      "name": "view",
      "parameters": [
        {
          "name": "doc_id",
          "value": "document_identifier"
        },
        {
          "name": "doc_title",
          "value": "ファイル名"
        },
        {
          "name": "doc_type",
          "value": "spreadsheet"
        },
        {
          "name": "owner",
          "value": "owner@muhai-academy.com"
        },
        {
          "name": "visibility",
          "value": "private"
        }
      ]
    }
  ]
}
```

**主要なイベント名**:
- `view`: ファイル閲覧
- `edit`: ファイル編集
- `download`: ファイルダウンロード
- `upload`: ファイルアップロード
- `share`: 共有設定変更
- `create`: ファイル作成
- `delete`: ファイル削除
- `move`: ファイル移動

### 3.3 Admin イベント
```json
{
  "applicationName": "admin",
  "events": [
    {
      "type": "USER_SETTINGS",
      "name": "CREATE_USER",
      "parameters": [
        {
          "name": "USER_EMAIL",
          "value": "newuser@muhai-academy.com"
        },
        {
          "name": "DOMAIN_NAME",
          "value": "muhai-academy.com"
        }
      ]
    }
  ]
}
```

**主要なイベント名**:
- `CREATE_USER`: ユーザー作成
- `DELETE_USER`: ユーザー削除
- `SUSPEND_USER`: ユーザー停止
- `CHANGE_USER_PASSWORD`: パスワード変更

## 4. 講義で扱う異常パターン

### 4.1 実例1: 夜間の管理者による大量データダウンロード
```json
{
  "id": {
    "time": "2024-08-12T22:30:15.123456Z",
    "applicationName": "drive"
  },
  "actor": {
    "email": "admin@muhai-academy.com"
  },
  "ipAddress": "192.168.1.10",
  "events": [
    {
      "type": "access",
      "name": "download",
      "parameters": [
        {
          "name": "doc_title",
          "value": "学習進捗データ_202408.xlsx"
        },
        {
          "name": "doc_type",
          "value": "spreadsheet"
        },
        {
          "name": "billable",
          "boolValue": true
        }
      ]
    }
  ]
}
```

### 4.2 実例2: 外部からの機密情報アクセス
```json
{
  "id": {
    "time": "2024-08-12T14:15:30.789012Z",
    "applicationName": "drive"
  },
  "actor": {
    "email": "unknown@external-domain.com"
  },
  "ipAddress": "203.0.113.45",
  "events": [
    {
      "type": "access",
      "name": "view",
      "parameters": [
        {
          "name": "doc_title",
          "value": "学籍管理データベース.xlsx"
        },
        {
          "name": "visibility",
          "value": "anyone_with_link"
        },
        {
          "name": "primary_event",
          "boolValue": true
        }
      ]
    }
  ]
}
```

### 4.3 実例3: VPN経由の不審なアクセス試行
```json
{
  "id": {
    "time": "2024-08-12T11:45:20.456789Z",
    "applicationName": "drive"
  },
  "actor": {
    "email": "compromised.user@muhai-academy.com"
  },
  "ipAddress": "10.0.100.50",
  "events": [
    {
      "type": "access",
      "name": "access_denied",
      "parameters": [
        {
          "name": "doc_title",
          "value": "財務データ/予算計画.xlsx"
        },
        {
          "name": "doc_type",
          "value": "spreadsheet"
        },
        {
          "name": "denied_reason",
          "value": "insufficient_permissions"
        }
      ]
    }
  ]
}
```

## 5. 検知すべき異常パターンの特徴

### 5.1 時間的異常
- **業務時間外アクセス**: 18:00-9:00の時間帯での活動
- **休日アクセス**: 土日での不審な活動
- **深夜活動**: 深夜1:00-5:00での管理者操作

### 5.2 アクセス元異常
- **地理的異常**: 通常と異なる国・地域からのアクセス
- **外部IPアクセス**: 組織外IPアドレスからの機密データアクセス
- **VPN経由の異常**: 内部IPだが不審なデバイスからのアクセス

### 5.3 行動パターン異常
- **大量アクセス**: 短時間での大量ファイルアクセス
- **権限昇格試行**: 権限外リソースへの連続アクセス試行
- **共有設定変更**: 機密ファイルの外部共有設定

### 5.4 データ種別異常
- **機密データアクセス**: 学籍データ、財務データ、人事データ
- **システムファイル**: バックアップ、設定ファイルへのアクセス
- **外部共有**: "anyone_with_link"設定での機密ファイル共有

## 6. SQL検知クエリの観点

実習では以下の観点からSQLクエリを作成します：

1. **時間窓集計**: 特定時間帯でのイベント数集計
2. **ユーザー行動分析**: 個別ユーザーの異常行動検知
3. **リソースアクセス分析**: 機密データへのアクセスパターン
4. **IPアドレス分析**: 地理的・ネットワーク的異常の検知
5. **相関分析**: 複数イベント間の時系列相関

---

**作成日**: 2024年7月14日  
**対象**: セキュリティキャンプ2025 B1講義用資料  
**更新者**: Claude Code Assistant