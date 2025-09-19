# Tasks: ETC明細APIシステム完成

**Input**: 設計ドキュメント from `C:/go/etc_meisai/specs/003-api-project-git/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## 実行フロー (main)
```
1. plan.mdから技術スタック、ライブラリ、構造を抽出
2. 設計ドキュメントから分析:
   → data-model.md: エンティティ → モデルタスク
   → contracts/: 各ファイル → 契約テストタスク
   → quickstart.md: テストシナリオ → 統合テストタスク
3. カテゴリ別タスク生成:
   → セットアップ: プロジェクト初期化、依存関係
   → テスト修正: 失敗テストの根本修正（最優先）
   → コア実装: モデル、サービス、エンドポイント
   → 統合: DB接続、ミドルウェア
   → 仕上げ: 単体テスト、ドキュメント
4. タスクルール適用:
   → 異なるファイル = [P] 並列実行
   → 同ファイル = 順次実行
   → テスト先行開発（TDD）
```

## フォーマット: `[ID] [P?] 説明`
- **[P]**: 並列実行可能（異なるファイル、依存関係なし）
- 説明には正確なファイルパスを含む

## 技術情報
- **言語**: Go 1.22+
- **フレームワーク**: Chi router, MySQL driver, testify
- **プロジェクト構造**: Single project (現在のGo API構造)
- **データベース**: MySQL dual setup (ローカル + 本番)

## 🚨 Phase 3.1: 失敗テスト修正（最重要・最優先）
**CRITICAL: テストを通すことがプロジェクトの成功条件**

- [ ] **T001** [P] CSVパーサーの複数日付フォーマット対応実装 `parser/etc_csv_parser.go`
  - `"06/01/02"`, `"2006/01/02"`, `"06/1/2"`, `"2006/1/2"` 形式サポート
  - 順次試行するフォールバック機能

- [ ] **T002** [P] エンコーディング自動検出機能実装 `parser/encoding_detector.go`
  - Shift-JIS/UTF-8自動判定
  - BOM検出機能
  - テストデータとの互換性確保

- [ ] **T003** [P] テストデータ修正と整備 `tests/integration/testdata/`
  - 実際のCSV形式に準拠
  - 適切な日付フォーマット使用
  - 15列構造への対応

- [ ] **T004** TestHashImportHandler修正 `tests/integration/hash_import_test.go`
  - ProcessedCount > 0 になるよう修正
  - パーサーエラーハンドリング改善
  - テストデータとパーサーの整合性確保

- [ ] **T005** TestDuplicateDetection修正 `tests/integration/hash_import_test.go`
  - 重複検出機能の動作確認
  - ハッシュ計算の正常動作検証
  - 重複データの適切な生成

- [ ] **T006** TestChangeDetection修正 `tests/integration/hash_import_test.go`
  - 変更検出機能の動作確認
  - コンテンツハッシュ比較機能
  - 変更シナリオの適切なテスト

## Phase 3.2: API契約テスト実装 ⚠️ IMPLEMENTATION前に完了必須

- [ ] **T007** [P] ハッシュインポートAPI契約テスト `tests/contract/hash_import_api_test.go`
  - `POST /api/etc/import/hash` エンドポイント
  - `GET /api/etc/hash/stats` エンドポイント
  - `GET /api/etc/duplicates` エンドポイント

- [ ] **T008** [P] マッピング管理API契約テスト `tests/contract/mapping_api_test.go`
  - `POST /api/etc/mapping` エンドポイント
  - `GET /api/etc/mapping/list` エンドポイント
  - `PUT /api/etc/mapping/{id}` エンドポイント

- [ ] **T009** [P] ダウンロードAPI契約テスト `tests/contract/download_api_test.go`
  - `POST /api/etc/download-async` エンドポイント
  - `GET /api/etc/download-status/{job_id}` エンドポイント
  - `POST /api/etc/download-sse` エンドポイント

- [ ] **T010** [P] 統合テストシナリオ実装 `tests/integration/api_integration_test.go`
  - ETCデータインポートフロー
  - 自動マッピング処理フロー
  - 非同期ダウンロードフロー

## Phase 3.3: コア実装（テスト失敗後のみ実行）

- [ ] **T011** [P] ETCDtakoMappingモデル完成 `models/etc_dtako_mapping.go`
  - バリデーションルール実装
  - 状態遷移ロジック
  - データベースマッピング

- [ ] **T012** [P] MappingBatchJobモデル完成 `models/mapping_batch_job.go`
  - ジョブステータス管理
  - 進捗計算ロジック
  - エラーハンドリング

- [ ] **T013** [P] ETCDtakoMappingRepository実装 `repositories/etc_dtako_mapping_repository.go`
  - CRUD操作完成
  - 複雑なクエリ実装
  - トランザクション管理

- [ ] **T014** [P] MappingBatchJobRepository実装 `repositories/mapping_batch_job_repository.go`
  - ジョブ管理機能
  - ステータス更新機能
  - クリーンアップ機能

- [ ] **T015** AutoMatchingService完成 `services/auto_matching_service.go`
  - バッチマッチング機能
  - 単一レコードマッチング
  - マッチング品質判定

- [ ] **T016** ManualMatchingService完成 `services/manual_matching_service.go`
  - 手動マッピング作成
  - マッピング検証
  - バッチ更新機能

- [ ] **T017** API自動マッピングハンドラー実装 `api/handlers/auto_matching_handler.go`
  - バッチマッチングエンドポイント
  - ステータス確認エンドポイント
  - エラーハンドリング

- [ ] **T018** API手動マッピングハンドラー実装 `api/handlers/manual_matching_handler.go`
  - マッピングCRUDエンドポイント
  - 検証エンドポイント
  - 候補取得エンドポイント

- [ ] **T019** APIマッピング一覧ハンドラー実装 `api/handlers/mapping_list_handler.go`
  - 一覧表示機能
  - フィルタリング機能
  - CSVエクスポート機能

## Phase 3.4: 統合とミドルウェア

- [ ] **T020** データベース接続統合テスト `tests/integration/database_test.go`
  - ローカルDB接続確認
  - 本番DB読み取り専用確認
  - トランザクション分離確認

- [ ] **T021** APIルーティング統合 `api_handlers_*.go`
  - 全エンドポイントのルーター登録
  - ミドルウェア設定
  - CORS設定

- [ ] **T022** エラーハンドリング統一 `api/handlers/response_helpers.go`
  - 統一エラーレスポンス
  - ログ出力標準化
  - HTTPステータスコード適正化

- [ ] **T023** ログ出力とモニタリング設定 `config/logging.go`
  - 構造化ログ実装
  - パフォーマンスメトリクス
  - エラー追跡機能

## Phase 3.5: 仕上げとドキュメント

- [ ] **T024** [P] 単体テスト拡張 `tests/unit/`
  - サービス層単体テスト
  - バリデーション単体テスト
  - ユーティリティ関数テスト

- [ ] **T025** [P] パフォーマンステスト実装 `tests/performance/`
  - 大量データ処理テスト
  - 同時リクエストテスト
  - メモリ使用量テスト

- [ ] **T026** [P] OpenAPI仕様更新 `specs/003-api-project-git/contracts/`
  - 実装に合わせた仕様調整
  - レスポンス例の追加
  - エラーレスポンス詳細化

- [ ] **T027** [P] クイックスタートガイド検証 `specs/003-api-project-git/quickstart.md`
  - セットアップ手順の検証
  - テスト実行手順の確認
  - トラブルシューティング更新

- [ ] **T028** コード品質改善とリファクタリング
  - 重複コード除去
  - パフォーマンス最適化
  - コメント・ドキュメント充実

## 依存関係グラフ

### 最優先（Phase 3.1）
- T001-T006: 並列実行可能（テスト修正が最重要）

### 契約テスト（Phase 3.2）
- T007-T010: 並列実行可能、T001-T006の後

### コア実装（Phase 3.3）
- T011, T012: 並列実行可能（異なるモデル）
- T013, T014: T011, T012の後、並列実行可能
- T015, T016: T013, T014の後、並列実行可能
- T017, T018, T019: T015, T016の後、順次実行（ハンドラー層）

### 統合（Phase 3.4）
- T020: T013, T014の後
- T021: T017, T018, T019の後
- T022, T023: T021の後、並列実行可能

### 仕上げ（Phase 3.5）
- T024, T025, T026, T027: 並列実行可能、T023の後
- T028: 全ての実装完了後

## 並列実行例

### Phase 3.1 (失敗テスト修正)
```bash
# T001-T003を並列実行:
Task: "CSVパーサーの複数日付フォーマット対応実装 parser/etc_csv_parser.go"
Task: "エンコーディング自動検出機能実装 parser/encoding_detector.go"
Task: "テストデータ修正と整備 tests/integration/testdata/"
```

### Phase 3.2 (契約テスト)
```bash
# T007-T009を並列実行:
Task: "ハッシュインポートAPI契約テスト tests/contract/hash_import_api_test.go"
Task: "マッピング管理API契約テスト tests/contract/mapping_api_test.go"
Task: "ダウンロードAPI契約テスト tests/contract/download_api_test.go"
```

### Phase 3.3 (モデル実装)
```bash
# T011-T012を並列実行:
Task: "ETCDtakoMappingモデル完成 models/etc_dtako_mapping.go"
Task: "MappingBatchJobモデル完成 models/mapping_batch_job.go"
```

## 注意事項

- **[P]** タスク = 異なるファイル、依存関係なし
- **テスト修正が最優先**: T001-T006は他のすべてに優先
- **TDD原則**: テストが失敗してから実装開始
- **各タスク後にコミット**: 段階的な進捗管理
- **成功条件**: 全統合テストがPASS

## タスク生成ルール検証
*main()実行時に確認済み*

- [x] すべての契約に対応するテストあり
- [x] すべてのエンティティにモデルタスクあり
- [x] すべてのテストが実装前に配置
- [x] 並列タスクが真に独立
- [x] 各タスクが正確なファイルパス指定
- [x] 同ファイル変更する[P]タスクなし

## 🎯 成功の定義

**最重要**: このプロジェクトの成功は「テストが通ること」
- [ ] `TestHashImportHandler`: ProcessedCount > 0
- [ ] `TestDuplicateDetection`: 重複検出数 > 0
- [ ] `TestChangeDetection`: 変更検出数 > 0
- [ ] 全API契約テストPASS
- [ ] 全統合テストPASS

---

**総タスク数**: 28タスク
**推定期間**: 失敗テスト修正 (T001-T006) を最優先で実行
**並列実行**: Phase毎に最大3-4タスクを並列実行可能