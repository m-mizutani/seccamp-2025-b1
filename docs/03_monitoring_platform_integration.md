# セキュリティ監視基盤の統合と段階的構築

## 🔗 基盤統一の必要性と効果

前のセクションで学んだように、現代のセキュリティ監視では多様なデータソースからの情報収集が必要です。しかし、これらのデータが分散していると、真の脅威を見逃す可能性が高くなります。

### ログ同士の突合（相関分析）の重要性

**単一ソースでは見えない攻撃パターン**

攻撃者は複数のシステムを跨いで巧妙に行動するため、単一のログソースだけでは全体像を把握できません。

🏫 **無敗塾での実例**: Google Workspace ログイン失敗 + AWS 権限昇格試行 = 総合的脅威判定

```
時系列での相関例:
10:15 - Google Workspace: user@muhaijuku.com ログイン失敗 (パスワード間違い)
10:16 - Google Workspace: 同じユーザーでログイン失敗 (5回目)  
10:20 - AWS CloudTrail: 同じユーザーでIAM権限変更試行
10:22 - VPC Flow Logs: 内部サーバーへの異常ポートスキャン

個別の判定:
→ Google Workspace: 一般的なログイン失敗
→ AWS CloudTrail: 権限変更は正常な範囲内
→ VPC Flow Logs: 内部通信のため低リスク

統合分析の判定:
→ 複数システムでの連続した異常行動パターン = 高リスク脅威
```

**時系列相関の威力**

異なるシステムでの時間的に関連する異常イベントを組み合わせることで、高精度な脅威検知が可能になります。

```mermaid
gantt
    title 複数システムの時系列相関分析による攻撃検知
    dateFormat HH:mm
    axisFormat %H:%M
    
    section Google Workspace
    正常ログイン                :done, gws1, 10:10, 10:11
    ログイン失敗(1回目)         :crit, gws2, 10:15, 10:16
    ログイン失敗(2回目)         :crit, gws3, 10:16, 10:17
    ログイン失敗(3-5回目)       :crit, gws4, 10:17, 10:20
    
    section AWS CloudTrail
    正常API呼び出し             :done, aws1, 10:18, 10:19
    IAM権限変更試行            :active, aws2, 10:20, 10:22
    S3バケット一覧取得         :active, aws3, 10:22, 10:23
    
    section VPC Flow Logs
    正常通信                   :done, vpc1, 10:15, 10:20
    内部ポートスキャン開始      :crit, vpc2, 10:22, 10:25
    異常ポート通信             :crit, vpc3, 10:25, 10:28
    
    section GitHub
    正常コード参照             :done, gh1, 10:20, 10:21
    リポジトリ一覧取得         :active, gh2, 10:24, 10:25
    機密リポジトリアクセス試行  :crit, gh3, 10:26, 10:27
    
    section 統合分析
    個別システム判定(正常)      :done, analysis1, 10:15, 10:25
    時系列相関分析実行         :active, analysis2, 10:25, 10:28
    高リスク判定・アラート      :milestone, alert, 10:28, 0m
```
*複数システムの時系列イベントを並べて、攻撃の全体像を可視化した図*

### 管理の集約による運用効率化

**一元的なアラート管理**

分散したアラートシステムでは、重要な脅威が埋もれてしまう可能性があります。

**無敗塾フェーズ2での課題**:
- AWS CloudWatch Alarms: 15件/日
- Google Workspace セキュリティアラート: 8件/日  
- GitHub セキュリティアドバイザリ: 3件/日
- ファイアウォール異常検知: 12件/日

**問題**: 合計38件のアラートが異なるシステムから通知され、優先度判定が困難

**統合後の改善**:
- 統一されたダッシュボードで全アラートを可視化
- 重要度に基づく自動優先度付け
- 関連アラートの自動グループ化

**統一されたダッシュボード**

複数システムの状況を単一画面で把握することで、運用チームの負荷を大幅に削減できます。

```mermaid
graph TB
    subgraph "統合セキュリティダッシュボード"
        subgraph "リアルタイム監視パネル"
            STATUS[全体ステータス<br/>🟢 正常運用中]
            ALERT_COUNT[アクティブアラート: 3件<br/>🔴 高: 1件 🟡 中: 2件]
            TREND[24時間トレンド<br/>📈 イベント増加傾向]
        end
        
        subgraph "AWS監視"
            AWS_STATUS[AWS環境<br/>🟢 正常]
            AWS_METRICS[• CloudTrail: 1,234 API calls/h<br/>• GuardDuty: 脅威なし<br/>• VPC: 異常通信 2件]
        end
        
        subgraph "Google Workspace監視"
            GWS_STATUS[Google Workspace<br/>🟡 注意]
            GWS_METRICS[• ログイン失敗: 15件/h<br/>• ファイル共有: 234件<br/>• access_denied: 8件 ⚠️]
        end
        
        subgraph "GitHub監視"
            GH_STATUS[GitHub<br/>🟢 正常]
            GH_METRICS[• コミット: 45件<br/>• PR作成: 12件<br/>• セキュリティアラート: 0件]
        end
        
        subgraph "ネットワーク監視"
            NET_STATUS[ネットワーク<br/>🔴 警告]
            NET_METRICS[• ファイアウォール: 正常<br/>• VPN接続: 異常地域 1件 🚨<br/>• Wi-Fi: 未許可デバイス 1件]
        end
        
        subgraph "統合分析結果"
            RISK_SCORE[総合リスクスコア: 65/100<br/>🟡 中リスク]
            TOP_ALERTS[最優先アラート<br/>1. 海外VPN + GWS access_denied<br/>2. 未許可デバイス接続<br/>3. 異常時間帯API使用]
        end
        
        subgraph "対応状況"
            INCIDENT[インシデント対応<br/>📋 進行中: 1件<br/>📝 調査中: 2件]
            SOC_STATUS[SOCチーム状況<br/>👥 対応可能: 3名<br/>⏰ 平均対応時間: 12分]
        end
    end
    
    subgraph "アクションパネル"
        INVESTIGATE[🔍 詳細調査]
        ESCALATE[📞 エスカレーション]
        SUPPRESS[🔇 アラート抑制]
        REPORT[📊 レポート生成]
    end
    
    AWS_METRICS -.-> RISK_SCORE
    GWS_METRICS -.-> RISK_SCORE
    GH_METRICS -.-> RISK_SCORE
    NET_METRICS -.-> RISK_SCORE
    
    RISK_SCORE --> TOP_ALERTS
    TOP_ALERTS --> INVESTIGATE
    TOP_ALERTS --> ESCALATE
    
    style STATUS fill:#c8e6c9
    style ALERT_COUNT fill:#ffcdd2
    style GWS_STATUS fill:#fff3e0
    style NET_STATUS fill:#ffcdd2
    style RISK_SCORE fill:#fff3e0
    style TOP_ALERTS fill:#ffcdd2
```
*AWS、Google Workspace、GitHub、ネットワーク機器の状況を一画面に統合したダッシュボード*

**スキル・知識の集約**

個別のツールそれぞれの習得コストを削減し、チーム全体のセキュリティ知識を向上させます。

- **従来**: 各ツール専門者が必要（AWS専門者、Google Workspace専門者、ネットワーク専門者）
- **統合後**: 統一インターフェースにより、少数の専門者で幅広いカバレッジが可能

## ⚖️ スキーマ共通化のメリット・デメリット

### 👍 メリット（Pros）

**クエリの汎用性**

共通スキーマ（OCSF等）を採用することで、同一のSQLクエリで複数のデータソースを検索できます。

```sql
-- OCSF共通フォーマットでの例
SELECT user_name, activity_name, count(*) 
FROM security_events 
WHERE time > NOW() - INTERVAL '1 HOUR'
  AND risk_level = 'High'
GROUP BY user_name, activity_name
HAVING count(*) > 5

-- このクエリで以下のデータソースを横断検索:
-- - AWS CloudTrail (activity_name = 'AssumeRole')  
-- - Google Workspace (activity_name = 'Login')
-- - GitHub (activity_name = 'PushCode')
```

**検知ルールの再利用**

共通フォーマットにより、一度作成した検知ルールを複数の環境で再利用できます。

**分析ツールの統一**

Business Intelligence（BI）ツールや機械学習基盤を一つのスキーマに最適化することで、開発・運用コストを削減できます。

### 👎 デメリット（Cons）

**変換処理のオーバーヘッド**

各データソースから共通フォーマットへの変換（ETL処理）により、遅延とコストが発生します。

**無敗塾での実測例**:
- **生ログ投入**: Google Workspace → S3 (平均5秒)
- **OCSF変換処理**: S3 → Security Lake (平均45秒)
- **合計遅延**: 50秒（リアルタイム検知に影響）

**元データの情報損失**

標準化により、ソース固有の重要な情報が失われる可能性があります。

**例**: Google Workspace の豊富なファイル共有権限情報
- **生ログ**: 詳細な権限設定、継承関係、有効期限
- **OCSF変換後**: 基本的なアクセス可否のみ

**ソース固有機能の制約**

各システム特有の豊富な分析機能を活用できなくなる場合があります。

- **AWS CloudTrail Insights**: AWS特有の異常検知機能
- **Google Workspace Security Investigation**: Google特有の調査機能

```mermaid
graph LR
    subgraph "生ログ（詳細情報）"
        GWS_RAW[Google Workspace<br/>生ログ]
        GWS_RAW --> GWS_DETAIL[詳細情報<br/>• ファイル権限継承関係<br/>• 共有リンク有効期限<br/>• 編集履歴詳細<br/>• IP地理情報詳細<br/>• ユーザーエージェント<br/>• 参照元情報]
        
        AWS_RAW[AWS CloudTrail<br/>生ログ]
        AWS_RAW --> AWS_DETAIL[詳細情報<br/>• IAM条件詳細<br/>• リソースタグ情報<br/>• API呼び出し元情報<br/>• エラー詳細コード<br/>• レスポンス詳細<br/>• セッション情報]
    end
    
    subgraph "ETL変換処理"
        TRANSFORM[OCSF変換<br/>データ正規化]
        TRANSFORM --> MAPPING[フィールドマッピング<br/>• 標準化<br/>• 型変換<br/>• 必須項目抽出]
        MAPPING --> LOSS[情報損失<br/>• 非標準フィールド削除<br/>• 詳細情報省略<br/>• データ型制約]
    end
    
    subgraph "OCSF標準スキーマ（情報損失後）"
        OCSF_OUT[OCSF形式ログ]
        OCSF_OUT --> OCSF_BASIC[基本情報のみ<br/>• タイムスタンプ<br/>• ユーザーID<br/>• アクション<br/>• リソース名<br/>• 成功/失敗<br/>• 基本的な地理情報]
    end
    
    subgraph "影響例"
        subgraph "失われる情報"
            LOST1[Google Workspace<br/>• ファイル共有の詳細権限<br/>• 編集協力者情報<br/>• リンク共有設定詳細]
            LOST2[AWS<br/>• IAM条件の詳細評価<br/>• リソース依存関係<br/>• セッション継続情報]
        end
        
        subgraph "分析への影響"
            IMPACT1[高度な分析が困難<br/>• 権限継承の追跡不可<br/>• 詳細な因果関係分析不可]
            IMPACT2[ベンダー固有機能<br/>• CloudTrail Insights使用不可<br/>• Workspace調査機能使用不可]
        end
    end
    
    GWS_DETAIL --> TRANSFORM
    AWS_DETAIL --> TRANSFORM
    LOSS --> OCSF_OUT
    
    OCSF_BASIC -.-> LOST1
    OCSF_BASIC -.-> LOST2
    LOST1 --> IMPACT1
    LOST2 --> IMPACT2
    
    style GWS_DETAIL fill:#c8e6c9
    style AWS_DETAIL fill:#c8e6c9
    style LOSS fill:#ffcdd2
    style OCSF_BASIC fill:#fff3e0
    style IMPACT1 fill:#ffcdd2
    style IMPACT2 fill:#ffcdd2
```
*元データの詳細情報が共通スキーマ変換で失われる様子を視覚化*

## 🏫 無敗塾における段階的統合戦略

### フェーズ1（創業期）: システムごとのログ保全

**基本的なログ収集と保存**
- 各システムの標準ログ機能を有効化
- 長期保存のためのS3への自動アーカイブ
- 手動での定期確認（週次）

**実装内容**:
- AWS CloudTrail の全API監査ログ
- Google Workspace の基本監査ログ
- アプリケーションログのCloudWatch Logs保存

### フェーズ2（事業拡大期）: 可能な範囲でのログの収集と保全、マニュアルによるログの検査

**重要システムの監視強化**
- セキュリティインシデント対応での教訓を活用
- 手動での相関分析プロセス確立
- アラート通知の基本的な統合

**実装の特徴**:
- インシデント発生時の手動調査フロー確立
- 重要アラートの Slack 通知統合
- 月次でのログ分析レポート作成

### フェーズ3（市場拡大期）: 包括的統合によるログ収集・保全と検知システムの導入

**Security Lake基盤の構築**
- OCSF フォーマットでの統合データレイク
- 自動化された検知ルールの実装
- リアルタイム相関分析の開始

**対応システム**:
- AWS（CloudTrail、VPC Flow Logs、GuardDuty）
- Google Workspace（監査ログ、セキュリティログ）
- GitHub（監査ログ、セキュリティアラート）
- Okta（認証ログ、プロビジョニングログ）

### フェーズ4-5（専門・グローバル展開）: 専門チームの編成や、検知システムの改善サイクルの実施

**高度なセキュリティ運用**
- 専門的なセキュリティチームによる24時間監視
- 機械学習を活用した高度な異常検知
- 継続的な検知精度向上プロセス

**グローバル統合**:
- マルチリージョンでのログ統合
- タイムゾーンを考慮した相関分析
- 各国の法規制要件を満たすデータ管理

```mermaid
timeline
    title 無敗塾の段階的統合ロードマップ
    
    section フェーズ1: 創業期
        システムごとのログ保全
        : AWS CloudTrail基本ログ
        : Google Workspace基本監査
        : 手動での週次確認
        : S3への基本アーカイブ
        
    section フェーズ2: 事業拡大期  
        重要システム統合
        : VPC Flow Logs追加
        : X-Ray分散トレース
        : Slack通知統合
        : 手動相関分析プロセス
        : インシデント対応フロー確立
        
    section フェーズ3: 市場拡大期
        包括的統合基盤
        : Security Lake構築
        : OCSF形式での統合
        : Okta SSO監視追加
        : GitHub監査ログ統合
        : 自動検知ルール実装
        : リアルタイム相関分析
        
    section フェーズ4: 専門教育期
        AI/ML活用高度化
        : 機械学習検知導入
        : サプライチェーン監視
        : 行動ベースライン学習
        : 専門チーム細分化
        : 高度な脅威対応
        
    section フェーズ5: グローバル展開
        国際統合監視
        : マルチリージョン統合
        : 24時間SOC運用
        : 脅威インテリジェンス
        : 文化的コンテキスト対応
        : 国際水準セキュリティ認証
```
*フェーズ1-5での監視システムの進化を時系列で図示、各段階での対象システムとカバレッジを表示*

## 🎯 統合戦略の判断基準

### コスト vs 効果の評価

**Phase 1評価ポイント**: 
- 既存ツールの活用度
- 手動作業の許容範囲
- インシデント対応時間の要件

**Phase 2-3移行の判断基準**:
- 1日あたりのアラート件数（20件超で統合検討）
- インシデント調査時間（4時間超で統合メリット大）
- 法規制要件の追加（GDPR、FERPA等）

### 段階的実装のベストプラクティス

**重要システムから順次統合**
1. **Tier 1**: 本番環境、決済システム、顧客データ
2. **Tier 2**: 開発環境、社内システム
3. **Tier 3**: 検証環境、ログ保存のみシステム

**誤検知を最小化する導入**
- パイロット運用での精度検証
- 既存アラートとの並行運用期間設定
- 段階的な閾値調整

## 🚀 次のセクションへ

**これまでのまとめ**
- **基盤統一の価値**: 相関分析による高精度脅威検知、運用効率化
- **スキーマ共通化**: メリット（汎用性、再利用性）vs デメリット（情報損失、変換コスト）
- **段階的統合**: 無敗塾の成長に合わせた現実的な実装戦略

**次は実習**
- AWS Security Lake を使った実際の統合基盤構築
- Google Workspace 監査ログを使ったカスタム検知実装
- 無敗塾シナリオでの実践的な脅威検知ルール作成