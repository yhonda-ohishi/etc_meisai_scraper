# 実装タスクリスト: ETCDtakoMappingテーブル実装

**機能**: ETCDtakoMappingテーブル実装
**ブランチ**: `003-api-project-git`
**作成日**: 2025-09-18
**技術スタック**: Go 1.22+, MySQL (ローカルDB), github.com/jmoiron/sqlx

## タスクリスト

### セットアップタスク
- [ ] **T001**: プロジェクト依存関係の確認
  - ファイル: `go.mod`
  - 内容: MySQLドライバー、sqlxの依存関係確認

- [ ] **T002**: データベース接続設定の確認
  - ファイル: `config/database.go`
  - 内容: ローカルMySQLの接続設定確認

### テストタスク (TDD - 実装前に作成) [P]
- [ ] **T003** [P]: マッピング作成APIの契約テスト
  - ファイル: `tests/contract/mapping_create_test.go`
  - 内容: POST /api/etc/mapping のリクエスト/レスポンススキーマテスト

- [ ] **T004** [P]: マッピング取得APIの契約テスト
  - ファイル: `tests/contract/mapping_get_test.go`
  - 内容: GET /api/etc/mapping/{id} のテスト

- [ ] **T005** [P]: マッピング検索APIの契約テスト
  - ファイル: `tests/contract/mapping_search_test.go`
  - 内容: by-meisai, by-dtako エンドポイントのテスト

- [ ] **T006** [P]: マッピング更新・削除APIの契約テスト
  - ファイル: `tests/contract/mapping_update_delete_test.go`
  - 内容: PUT, DELETE エンドポイントのテスト

- [ ] **T007** [P]: マッピング統合テスト
  - ファイル: `tests/integration/mapping_integration_test.go`
  - 内容: 全エンドポイントの統合テスト、重複チェック、外部キー制約テスト

### コアタスク

#### データベース
- [ ] **T008**: マッピングテーブルのSQL定義追加
  - ファイル: `schema.sql`
  - 内容: etc_dtako_mappingテーブルのCREATE TABLE文追加
  ```sql
  CREATE TABLE IF NOT EXISTS etc_dtako_mapping (
      id INT AUTO_INCREMENT PRIMARY KEY,
      etc_meisai_id INT NOT NULL,
      dtako_row_id VARCHAR(100) NOT NULL,
      vehicle_id VARCHAR(50),
      mapping_type VARCHAR(20) NOT NULL DEFAULT 'manual',
      notes TEXT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
      created_by VARCHAR(100),
      CONSTRAINT fk_etc_meisai FOREIGN KEY (etc_meisai_id) REFERENCES etc_meisai(id) ON DELETE CASCADE,
      INDEX idx_etc_meisai_id (etc_meisai_id),
      INDEX idx_dtako_row_id (dtako_row_id),
      INDEX idx_vehicle_id (vehicle_id),
      UNIQUE KEY unique_meisai_dtako (etc_meisai_id, dtako_row_id)
  )
  ```

#### モデル [P]
- [ ] **T009** [P]: ETCDtakoMappingモデルの修正
  - ファイル: `models/etc_dtako_mapping.go`
  - 内容: MatchScoreフィールドとValidateMatchScore()メソッドの削除

- [ ] **T010** [P]: マッピングリクエスト/レスポンス型の更新
  - ファイル: `models/etc_dtako_mapping.go`
  - 内容: ETCDtakoMappingRequest, ETCDtakoMappingResponseからMatchScore関連を削除

#### リポジトリ
- [ ] **T011**: マッピングリポジトリのCRUD実装
  - ファイル: `repositories/etc_dtako_mapping_repository.go`
  - 内容: Create, Get, Update, Delete メソッドの実装（MatchScore削除対応）

- [ ] **T012**: マッピング検索メソッドの実装
  - ファイル: `repositories/etc_dtako_mapping_repository.go`
  - 内容: GetByMeisaiID, GetByDtakoRowID, ListWithPagination メソッドの実装

- [ ] **T013**: 重複チェックメソッドの実装
  - ファイル: `repositories/etc_dtako_mapping_repository.go`
  - 内容: CheckDuplicate メソッドの実装

#### APIハンドラー
- [ ] **T014**: マッピング作成ハンドラーの実装
  - ファイル: `api_handlers_mapping.go`
  - 内容: CreateMappingHandler の実装（重複チェック含む）

- [ ] **T015**: マッピング取得ハンドラーの実装
  - ファイル: `api_handlers_mapping.go`
  - 内容: GetMappingHandler, GetMappingByMeisaiHandler, GetMappingByDtakoHandler の実装

- [ ] **T016**: マッピング更新・削除ハンドラーの実装
  - ファイル: `api_handlers_mapping.go`
  - 内容: UpdateMappingHandler, DeleteMappingHandler の実装

- [ ] **T017**: マッピング一覧ハンドラーの実装
  - ファイル: `api_handlers_mapping.go`
  - 内容: ListMappingsHandler の実装（ページネーション対応）

### 統合タスク
- [ ] **T018**: ハンドラーのルーティング登録
  - ファイル: `registry.go`
  - 内容: 全マッピングハンドラーの登録追加

- [ ] **T019**: エラーハンドリングの統一
  - ファイル: `api_handlers_mapping.go`
  - 内容: 統一的なエラーレスポンス形式の実装

- [ ] **T020**: ロギングとメトリクスの追加
  - ファイル: `api_handlers_mapping.go`, `repositories/etc_dtako_mapping_repository.go`
  - 内容: 各操作のログ出力とパフォーマンスメトリクス

### ポリッシュタスク [P]
- [ ] **T021** [P]: 単体テストの作成（モデル）
  - ファイル: `models/etc_dtako_mapping_test.go`
  - 内容: GetMappingType, IsAutomatic, IsManual のテスト

- [ ] **T022** [P]: 単体テストの作成（リポジトリ）
  - ファイル: `repositories/etc_dtako_mapping_repository_test.go`
  - 内容: 各リポジトリメソッドのテスト

- [ ] **T023** [P]: パフォーマンステストの作成
  - ファイル: `tests/performance/mapping_perf_test.go`
  - 内容: 大量データでのレスポンス時間測定

- [ ] **T024**: Swaggerドキュメントの更新
  - ファイル: `docs/swagger.yaml`
  - 内容: マッピングAPIエンドポイントの追加

- [ ] **T025**: READMEの更新
  - ファイル: `README.md`
  - 内容: マッピング機能の使用方法追加

## 実行順序

### フェーズ1: セットアップ（T001-T002）
```bash
# 依存関係の確認
go mod tidy
```

### フェーズ2: テスト作成（T003-T007）- 並列実行可能
```bash
# 契約テストを並列で作成
Task agent: "T003 マッピング作成APIの契約テスト作成"
Task agent: "T004 マッピング取得APIの契約テスト作成"
Task agent: "T005 マッピング検索APIの契約テスト作成"
Task agent: "T006 マッピング更新・削除APIの契約テスト作成"
Task agent: "T007 マッピング統合テスト作成"
```

### フェーズ3: データベース（T008）
```bash
# SQLファイルの更新と実行
mysql -u root -p etc_meisai_db < schema.sql
```

### フェーズ4: モデル（T009-T010）- 並列実行可能
```bash
Task agent: "T009 ETCDtakoMappingモデル修正"
Task agent: "T010 リクエスト/レスポンス型更新"
```

### フェーズ5: リポジトリ（T011-T013）- 順次実行
```bash
# リポジトリメソッドは同一ファイルなので順次実行
go test ./repositories -run TestETCDtakoMapping
```

### フェーズ6: APIハンドラー（T014-T017）- 順次実行
```bash
# ハンドラーは同一ファイルなので順次実行
go test ./tests/contract -v
```

### フェーズ7: 統合（T018-T020）
```bash
# 統合テスト実行
go test ./tests/integration -v
```

### フェーズ8: ポリッシュ（T021-T025）- 並列実行可能
```bash
Task agent: "T021 モデル単体テスト作成"
Task agent: "T022 リポジトリ単体テスト作成"
Task agent: "T023 パフォーマンステスト作成"
# T024, T025は順次実行
```

## 並列実行の例

### テストタスクの並列実行
```bash
# 5つの契約テストを同時に作成
Task agent --parallel \
  "T003: tests/contract/mapping_create_test.go作成" \
  "T004: tests/contract/mapping_get_test.go作成" \
  "T005: tests/contract/mapping_search_test.go作成" \
  "T006: tests/contract/mapping_update_delete_test.go作成" \
  "T007: tests/integration/mapping_integration_test.go作成"
```

### モデルタスクの並列実行
```bash
# 異なるファイルなので並列可能
Task agent --parallel \
  "T009: models/etc_dtako_mapping.go修正" \
  "T021: models/etc_dtako_mapping_test.go作成"
```

## 完了基準

1. 全テストがパスする
2. コードカバレッジ80%以上
3. レスポンス時間200ms以下
4. Swaggerドキュメント更新済み
5. READMEに使用方法記載

## 注意事項

- **T008** (SQL) は他のタスクの前提条件
- **T011-T013** は同一ファイルなので順次実行
- **T014-T017** も同一ファイルなので順次実行
- [P]マークのタスクは並列実行可能（異なるファイル）

## 次のステップ

タスク完了後：
1. コードレビュー
2. パフォーマンス測定
3. 本番環境へのデプロイ準備
4. ユーザードキュメント作成

---
*このタスクリストは`specs/003-api-project-git/`の設計文書に基づいて生成されました*