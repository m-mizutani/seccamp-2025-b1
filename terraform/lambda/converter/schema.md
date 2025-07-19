# Google Workspace監査ログのOCSF変換スキーマ

セキュリティキャンプ2025 B1講義用のGoogle Workspace監査ログをOCSF (Open Cybersecurity Schema Framework) 形式に変換するためのスキーマ定義書。

## 1. OCSF イベントクラス選定

### 選定結果: Web Resources Activity (Class ID: 6001)

**選定理由:**
- Application Activity カテゴリのAudit系イベントクラス
- 全てのGoogle Workspaceサービス（Login、Drive、Admin、Calendar）の活動を単一のイベントクラスで表現可能
- CRUD操作に加えてSearch、Import、Export、Shareなど豊富なactivity_id
- Webリソースへのアクセス監査に特化した設計
- 監査ログとして適切な属性セット（actor、web_resources、api）

### 代替候補の検討結果

- **Authentication (3002)**: ログイン専用で他の活動を表現できない
- **User Access Management (3005)**: 管理者操作専用で汎用性に欠ける
- **API Activity (6003)**: API呼び出しに特化しており、監査ログとしての側面が弱い

## 2. ログ種別とマッピングルール

### 2.1 ログ種別の定義

Google Workspaceの監査ログは、以下の要素で一意に識別されます：

1. **applicationName** (id.applicationName): サービスを識別
2. **eventType** (events[].type): イベントのカテゴリ
3. **eventName** (events[].name): 具体的なアクション

### 2.2 詳細マッピングルール

#### Login イベント (applicationName: "login")

| eventType | eventName | OCSF activity_id | api.operation | severity_id | 備考 |
|-----------|-----------|------------------|---------------|-------------|------|
| login | login_success | 2 (Read) | login_success | 1 | 正常ログイン |
| login | login_failure | 2 (Read) | login_failure | 2 | ログイン失敗 |
| login | logout | 99 (Other) | logout | 1 | ログアウト |
| login | suspicious_login | 2 (Read) | suspicious_login | 3 | 不審なログイン |
| login | login_challenge | 2 (Read) | login_challenge | 1 | 2FA認証チャレンジ |

#### Drive イベント (applicationName: "drive")

| eventType | eventName | OCSF activity_id | api.operation | severity_id | 備考 |
|-----------|-----------|------------------|---------------|-------------|------|
| access | view | 2 (Read) | view | 1 | ファイル閲覧 |
| access | edit | 3 (Update) | edit | 1 | ファイル編集 |
| access | download | 7 (Export) | download | 1 | ファイルダウンロード |
| access | upload | 6 (Import) | upload | 1 | ファイルアップロード |
| access | print | 7 (Export) | print | 1 | 印刷 |
| access | preview | 2 (Read) | preview | 1 | プレビュー |
| creation | create | 1 (Create) | create | 1 | ファイル作成 |
| deletion | trash | 4 (Delete) | trash | 1 | ゴミ箱へ移動 |
| deletion | delete | 4 (Delete) | delete | 2 | 完全削除 |
| sharing | share | 8 (Share) | share | 2 | 共有設定 |
| sharing | unshare | 8 (Share) | unshare | 1 | 共有解除 |
| access | access_denied | 2 (Read) | access_denied | 2 | アクセス拒否 |
| move | move | 3 (Update) | move | 1 | ファイル移動 |
| rename | rename | 3 (Update) | rename | 1 | ファイル名変更 |

#### Admin イベント (applicationName: "admin")

| eventType | eventName | OCSF activity_id | api.operation | severity_id | 備考 |
|-----------|-----------|------------------|---------------|-------------|------|
| USER_SETTINGS | CREATE_USER | 1 (Create) | create_user | 2 | ユーザー作成 |
| USER_SETTINGS | DELETE_USER | 4 (Delete) | delete_user | 3 | ユーザー削除 |
| USER_SETTINGS | SUSPEND_USER | 3 (Update) | suspend_user | 3 | ユーザー停止 |
| USER_SETTINGS | UNSUSPEND_USER | 3 (Update) | unsuspend_user | 2 | ユーザー再開 |
| USER_SETTINGS | CHANGE_USER_PASSWORD | 3 (Update) | change_password | 3 | パスワード変更 |
| GROUP_SETTINGS | CREATE_GROUP | 1 (Create) | create_group | 2 | グループ作成 |
| GROUP_SETTINGS | DELETE_GROUP | 4 (Delete) | delete_group | 2 | グループ削除 |
| DOMAIN_SETTINGS | CHANGE_DOMAIN_SETTING | 3 (Update) | change_setting | 3 | ドメイン設定変更 |
| SECURITY_SETTINGS | CHANGE_2SV_SETTING | 3 (Update) | change_2sv | 4 | 2段階認証設定 |
| APPLICATION_SETTINGS | CHANGE_APPLICATION_SETTING | 3 (Update) | change_app_setting | 2 | アプリ設定変更 |

#### Calendar イベント (applicationName: "calendar")

| eventType | eventName | OCSF activity_id | api.operation | severity_id | 備考 |
|-----------|-----------|------------------|---------------|-------------|------|
| event | create_event | 1 (Create) | create_event | 1 | イベント作成 |
| event | view_event | 2 (Read) | view_event | 1 | イベント閲覧 |
| event | edit_event | 3 (Update) | edit_event | 1 | イベント編集 |
| event | delete_event | 4 (Delete) | delete_event | 1 | イベント削除 |
| event | invite_respond | 3 (Update) | invite_respond | 1 | 招待への返答 |
| sharing | share_calendar | 8 (Share) | share_calendar | 2 | カレンダー共有 |

### 2.3 severity_id 判定ルール

severity_idは以下の基準で自動判定されます：

| severity_id | レベル | 判定基準 |
|-------------|--------|----------|
| 1 | Informational | 通常の読み取り・作成操作 |
| 2 | Low | 共有・エクスポート・ログイン失敗 |
| 3 | Medium | 管理者操作・ユーザー削除・不審なログイン |
| 4 | High | セキュリティ設定変更・2FA設定変更 |

### 2.4 特殊フィールドのマッピング

#### actor.user.type_id の判定
- 管理者権限での操作: `2`
- 通常ユーザーの操作: `1`
- 判定方法: eventTypeが管理系（USER_SETTINGS、GROUP_SETTINGS等）の場合は`2`

#### status_id の判定
- 成功イベント（login_success、view、create等）: `1`
- 失敗イベント（login_failure、access_denied等）: `2`
- eventNameに"failure"、"denied"、"error"が含まれる場合: `2`

#### disposition_id の判定
- 許可された操作: `1` (Allowed)
- ブロックされた操作: `2` (Blocked)
- 隔離された操作: `3` (Quarantined)
- 判定方法: status_idが`2`の場合は`2`、suspicious_loginの場合は`3`

## 3. 詳細フィールドマッピング

### 3.1 必須フィールド

#### 基本分類属性
```json
{
  "category_uid": 6,                    // Application Activity
  "class_uid": 6001,                    // Web Resources Activity
  "type_uid": 600102,                   // class_uid * 100 + activity_id
  "activity_id": 2,                     // Create=1, Read=2, Update=3, Delete=4, etc.
  "severity_id": 1,                     // Informational=1 (デフォルト)
  "time": "2024-08-12T10:15:30.123Z"   // id.time から変換
}
```

#### Actor（実行者）マッピング
```json
{
  "actor": {
    "user": {
      "name": "user@muhai-academy.com",      // actor.email
      "uid": "114511147312345678901",        // actor.profileId
      "email_addr": "user@muhai-academy.com", // actor.email
      "domain": "muhai-academy.com",         // ownerDomain
      "type_id": 1                          // 通常ユーザー=1, 管理者=2
    },
    "session": {
      "uid": "358068855354",                 // id.uniqueQualifier
      "created_time": "2024-08-12T10:00:00Z" // セッション開始時刻（推定）
    },
    "app_name": "Google Workspace",          // 固定値
    "app_uid": "drive"                       // id.applicationName
  }
}
```

#### API情報マッピング
```json
{
  "api": {
    "service": {
      "name": "Google Drive API",             // applicationName から決定
      "version": "v3"                        // 固定値またはメタデータから
    },
    "operation": "view",                     // events[0].name
    "request": {
      "uid": "358068855354"                  // id.uniqueQualifier
    },
    "response": {
      "code": 200,                          // 成功=200, 失敗=400系
      "message": "Success"                  // ステータスメッセージ
    }
  }
}
```

#### クラウド環境情報
```json
{
  "cloud": {
    "provider": "Google Cloud",              // 固定値
    "account": {
      "uid": "C03az79cb",                   // id.customerId
      "name": "muhai-academy"               // ownerDomain から推定
    },
    "org": {
      "name": "muhai-academy.com",          // ownerDomain
      "uid": "muhai-academy"                // ownerDomain から推定
    },
    "region": "asia-northeast1"             // 日本リージョン（推定）
  }
}
```

#### ソースエンドポイント
```json
{
  "src_endpoint": {
    "ip": "203.0.113.255",                  // ipAddress
    "hostname": "client.example.com",        // 逆引き結果（オプション）
    "location": {
      "country": "JP",                      // IPアドレスから地理情報推定
      "region": "Tokyo",
      "city": "Shinjuku"
    }
  }
}
```

### 3.2 Webリソース情報マッピング

#### Google Driveファイル
```json
{
  "web_resources": [{
    "name": "学習進捗データ.xlsx",           // doc_title parameter
    "uid": "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", // doc_id parameter
    "type": "spreadsheet",                  // doc_type parameter
    "url_string": "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
    "data": {
      "classification": "confidential"      // visibility parameter から推定
    }
  }]
}
```

### 3.3 ステータスマッピング

| Google Workspaceイベント | status_id | disposition_id | 説明 |
|--------------------------|-----------|----------------|------|
| login_success | 1 (Success) | 1 (Allowed) | 正常ログイン |
| login_failure | 2 (Failure) | 2 (Blocked) | ログイン失敗 |
| access_denied | 2 (Failure) | 2 (Blocked) | アクセス拒否 |
| view, edit, download | 1 (Success) | 1 (Allowed) | 正常操作 |
| suspicious_login | 1 (Success) | 3 (Quarantined) | 疑わしいログイン |

## 4. 実装における考慮事項

### 4.1 データ型変換

| Google Workspaceフィールド | 型 | OCSFフィールド | 型 | 変換処理 |
|---------------------------|---|---------------|---|---------|
| id.time | String (ISO8601) | time | Timestamp | RFC3339パース |
| actor.profileId | String | actor.user.uid | String | そのまま |
| ipAddress | String | src_endpoint.ip | String | IPv4/IPv6検証 |
| events[].parameters[].boolValue | Boolean | 各種フラグ | Boolean | そのまま |
| events[].parameters[].multiStrValue | Array | metadata.labels | Array | 配列マージ |

### 4.2 エラーハンドリング

| エラー条件 | 対応方法 |
|-----------|---------|
| 必須フィールド不足 | デフォルト値設定またはイベント破棄 |
| 不正なタイムスタンプ | 現在時刻で補完 |
| 不明なapplicationName | "Unknown Service" として処理 |
| IPアドレス形式エラー | "0.0.0.0" で補完 |
| パラメータ解析エラー | web_resources を空配列で初期化 |

### 4.3 パフォーマンス最適化

- **バッチ処理**: 複数ログの一括変換（100件単位）
- **フィールドキャッシュ**: 同一ユーザー/リソース情報の再利用
- **並列処理**: goroutineによる並行変換（CPU数×2のワーカー）
- **メモリ効率**: ストリーミング処理によるメモリ使用量削減

## 5. 検証とテスト

### 5.1 変換精度チェック項目

- [ ] 全必須フィールドの存在確認
- [ ] activity_idの正確性（イベント名との整合性）
- [ ] タイムスタンプ形式の統一（RFC3339）
- [ ] web_resources配列の完全性
- [ ] IPアドレスの妥当性（形式チェック）
- [ ] severity_idの適切性（異常パターンとの整合性）

### 5.2 異常検知テストケース

- [ ] 時間外アクセスの適切なseverity_id設定（3以上）
- [ ] 外部IPアクセスのmetadata.labels設定
- [ ] 管理者操作のactor.user.type_id設定（=2）
- [ ] アクセス拒否イベントのstatus_id設定（=2）
- [ ] 機密データアクセスのweb_resources.data.classification設定

---

**作成日**: 2025年7月19日  
**対象**: セキュリティキャンプ2025 B1講義  
**OCSF Version**: 1.3.0  
**作成者**: Claude Code Assistant