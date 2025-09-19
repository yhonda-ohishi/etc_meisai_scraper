# Implementation Tasks: ETC明細スクレイピング＆DBインポート（シンプル版）

**Feature**: ETC明細スクレイピング＆DBインポート
**Date**: 2025-09-19
**Status**: Ready for execution

## Overview

シンプル化されたETC明細スクレイピングとDBインポート機能の実装タスク。APIやマッチング機能を削除し、CLIツールとして実装。

## Task List

### Phase 1: クリーンアップ (Clean up unnecessary code)

**T001** [P] 不要なAPIハンドラー削除
- **File**: `api/handlers/*.go` → 削除
- **Action**: APIハンドラーディレクトリ全体を削除
- **Test**: ビルドが成功すること

**T002** [P] HTTPサーバー関連コード削除
- **File**: `routes.go`, `registry.go`, `cmd/server/` → 削除
- **Action**: HTTPサーバー起動コードとルーティング削除
- **Test**: 不要な依存関係が削除されること

**T003** [P] 本番DB接続コード削除
- **File**: `config/prod_database.go` → 削除
- **Action**: 本番データベース接続設定を削除
- **Test**: ローカルDBのみ接続可能

**T004** マッチング機能削除
- **File**: `services/matching_service.go`, `services/auto_matching_service.go` → 削除
- **Action**: ETC-Dtakoマッチング機能を削除
- **Dependencies**: T001完了後

### Phase 2: データモデル実装 (Data model implementation)

**T005** [P] ETCMeisaiモデル簡素化
- **File**: `models/etc_meisai.go`
- **Action**: 不要フィールド削除、シンプルな構造に更新
```go
type ETCMeisai struct {
    ID             int64     `db:"id"`
    UsageDate      time.Time `db:"usage_date"`
    EntryIC        string    `db:"entry_ic"`
    ExitIC         string    `db:"exit_ic"`
    TollAmount     int       `db:"toll_amount"`
    VehicleNumber  string    `db:"vehicle_number"`
    ETCCardNumber  string    `db:"etc_card_number"`
    AccountType    string    `db:"account_type"`
    ImportedAt     time.Time `db:"imported_at"`
    CreatedAt      time.Time `db:"created_at"`
}
```

**T006** [P] データベーススキーマ作成
- **File**: `schema.sql`
- **Action**: シンプルなテーブル定義作成
- **Test**: MySQLでテーブル作成可能

### Phase 3: リポジトリ層実装 (Repository layer)

**T007** ETCリポジトリ簡素化
- **File**: `repositories/etc_repository.go`
- **Action**: CRUD操作を簡素化、不要メソッド削除
- **Methods**: Insert, BulkInsert, Exists, FindByDateRange
- **Dependencies**: T005完了後

### Phase 4: スクレイピング実装 (Scraping implementation)

**T008** スクレイパー整理
- **File**: `scraper/scraper.go`
- **Action**: 既存スクレイピングロジックの整理
- **Test**: Playwright-goでログイン可能

**T009** CSV解析機能確認
- **File**: `parser/csv_parser.go`
- **Action**: CSVパース機能の動作確認と調整
- **Dependencies**: T008完了後

### Phase 5: CLIツール実装 (CLI tool implementation)

**T010** CLIエントリーポイント作成
- **File**: `cmd/scraper/main.go`
- **Action**: コマンドライン引数処理とメイン処理フロー実装
```go
// 基本構造
func main() {
    // 引数パース (--account, --from, --to, --all)
    // 環境変数読み込み
    // スクレイピング実行
    // DB保存
}
```
- **Dependencies**: T007, T008完了後

**T011** 環境変数設定処理
- **File**: `config/env.go`
- **Action**: 環境変数からの設定読み込み実装
- **Variables**: DB_*, ETC_*_USER, ETC_*_PASS

### Phase 6: 統合テスト (Integration tests)

**T012** [P] スクレイピング統合テスト
- **File**: `tests/integration/scraper_test.go`
- **Action**: E2Eテスト作成（スクレイピング→DB保存）
- **Test Cases**:
  - 正常系: データ取得成功
  - 異常系: ログイン失敗
- **Dependencies**: T010完了後

**T013** [P] 重複検出テスト
- **File**: `tests/integration/duplicate_test.go`
- **Action**: 重複データ検出のテスト作成
- **Test Cases**:
  - 同一データの重複インポート防止
  - ユニークインデックス動作確認

### Phase 7: ドキュメント更新 (Documentation)

**T014** [P] README更新
- **File**: `README.md`
- **Action**: シンプル化された機能の説明に更新
- **Content**: インストール、使い方、設定方法

**T015** [P] 環境変数サンプル作成
- **File**: `.env.example`
- **Action**: 環境変数のテンプレート作成
- **Content**: 必要な環境変数のサンプル値

## Execution Order

```
Phase 1 (並列実行可能):
Task agent T001 T002 T003

Phase 2 (並列実行可能):
Task agent T005 T006

Phase 3-5 (順次実行):
Task agent T007
Task agent T008
Task agent T009
Task agent T010
Task agent T011

Phase 6 (並列実行可能):
Task agent T012 T013

Phase 7 (並列実行可能):
Task agent T014 T015
```

## Task Dependencies Graph

```
T001 ─┐
T002 ─┼─→ T004 ─┐
T003 ─┘         ├─→ T005 → T007 ─┐
                │                  ├─→ T010 → T011 → T012
T006 ───────────┘   T008 → T009 ─┘                  ↓
                                                    T013
                    T014 (独立)
                    T015 (独立)
```

## Success Criteria

- [ ] 不要なAPIコード削除完了
- [ ] CLIツールとしてビルド成功
- [ ] スクレイピング→DB保存が動作
- [ ] 重複検出が機能
- [ ] 統合テストがPASS
- [ ] ドキュメント更新完了

## Notes for Implementation

1. **セキュリティ**: パスワードは絶対にハードコードしない
2. **エラー処理**: 各フェーズで適切なエラーハンドリング実装
3. **ログ出力**: 重要な処理ステップでログ出力
4. **トランザクション**: DB保存時はトランザクション使用

---
*Task list generated for simplified ETC scraping and DB import implementation*