# 開発環境向けリソース無効化機能の実装計画

## 実装方針

デフォルト値を `true` にすることで、最初の `terraform apply` では既存環境に変更が発生しないことを確認してから、開発環境で `false` を設定して動作を検証します。

## 実装ステップ

### Phase 1: 変数定義と基本設定

- [x] 1. `terraform/variables.tf` に `enable_active_resources` 変数を追加
- [x] 2. `terraform/terraform.tfvars.example` を作成し、使用例を記載

### Phase 2: Lambda Function URL の条件付き作成

- [x] 3. 現在の `aws_lambda_function_url.auditlog` リソースを確認
- [x] 4. `count` メタ引数を追加して条件付き作成に変更
- [x] 5. 関連する output (`auditlog_lambda_url`) を更新

### Phase 3: Importer 定期実行の条件付き作成

- [x] 6. 現在の EventBridge 関連リソースの構成を確認
- [x] 7. `aws_cloudwatch_event_rule.importer_schedule` に `count` を追加
- [x] 8. `aws_cloudwatch_event_target.importer_target` に `count` を追加
- [x] 9. `aws_lambda_permission.allow_eventbridge` に `count` を追加

### Phase 4: Output とドキュメントの更新

- [x] 10. `terraform/outputs.tf` に `active_resources_enabled` output を追加
- [x] 11. README.md に開発環境での使用方法を追記

### Phase 5: 動作確認

- [x] 12. `terraform plan` を実行し、デフォルト設定（true）で差分がないことを確認
- [ ] 13. `terraform apply` を実行し、エラーがないことを確認
- [ ] 14. `terraform.tfvars` に `enable_active_resources = false` を設定
- [ ] 15. `terraform plan` で無効化される リソースを確認
- [ ] 16. `terraform apply` で リソースが正しく無効化されることを確認
- [ ] 17. 再度 `enable_active_resources = true` に戻して元の状態に復元できることを確認

## 実装時の注意事項

### 1. 後方互換性の確保
- 変数のデフォルト値を `true` にすることで、既存の動作を維持
- 最初の apply で差分が出ないことを必ず確認

### 2. count 使用時の参照変更
- `aws_lambda_function_url.auditlog` → `aws_lambda_function_url.auditlog[0]`
- 参照箇所での存在チェックが必要

### 3. 段階的な適用
- まず一つのリソース（Lambda URL）から実装し、動作確認後に他のリソースへ展開
- 各フェーズごとに `terraform plan` で影響範囲を確認

### 4. Glue Crawler について
- 現在のコードベースに Glue Crawler の定義がない場合は、将来の拡張として設計のみ残し、実装はスキップ

## 検証項目

1. **デフォルト設定での動作**
   - `terraform plan` で差分なし
   - 既存のリソースがすべて維持される

2. **無効化時の動作**
   - Lambda Function URL が削除される
   - EventBridge ルールが削除される
   - Lambda 関数自体は残る

3. **再有効化時の動作**
   - すべてのリソースが元通りに作成される
   - エンドポイント URL などが新しくなることの確認