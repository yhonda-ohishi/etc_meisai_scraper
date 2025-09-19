# Implementation Plan: ETC明細CSVインポート&データベース統合

**Branch**: `003-api-project-git` | **Date**: 2025-09-19 | **Spec**: `/specs/003-api-project-git/spec.md`
**Input**: Feature specification from `/specs/003-api-project-git/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path ✓
2. Fill Technical Context ✓
3. Fill the Constitution Check section ✓
4. Evaluate Constitution Check section ✓
5. Execute Phase 0 → research.md
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
7. Re-evaluate Constitution Check section
8. Plan Phase 2 → Task generation approach
9. STOP - Ready for /tasks command
```

## Summary
ETC明細CSVファイルをAPIエンドポイント経由でローカルDBにインポートし、本番データベース（prod DB）のdtako_rowsテーブルのetc_numフィールドとマッチングを行う統合システムの実装。ryohi_sub_calテーブルを参考にしたデータ構造と処理フローを実装する。

## Technical Context
**Language/Version**: Go 1.22+
**Primary Dependencies**: Chi router, MySQL driver, go-sql-driver/mysql, encoding/csv
**Storage**: MySQL dual database (Local: localhost:3307, Production: 172.18.21.35:3306)
**Testing**: Go testing package, testify, integration tests
**Target Platform**: Windows/Linux server
**Project Type**: single (Go API project)
**Performance Goals**: 10,000 records/minute import, <500ms matching response
**Constraints**: Read-only access to production DB, No hardcoded credentials
**Scale/Scope**: 58,000+ dtako_rows records, CSV files ~500KB each

**User Requirements from Arguments**:
- エンドポイントを使用してETC明細CSVファイルをローカルDBにインポート
- prod db のetc_numとのintegration
- prod dbのdtako_rows.etc_numとlocalDBのetc_meisaiのマッチング
- ryohi_sub_calテーブルを参考

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### セキュリティ原則チェック
- [x] ハードコードされた認証情報の禁止 → 環境変数使用を徹底
- [x] すべての認証情報は環境変数から読み込む
- [x] テストコードでもハードコード禁止
- [x] デフォルト値として実際の認証情報を使用しない

### Test-First原則
- [x] TDD必須: テスト作成 → ユーザー承認 → テスト失敗確認 → 実装
- [x] 統合テストが必要な重点領域の特定

### ドキュメント言語
- [x] すべての仕様書、計画書は日本語で記述

## Project Structure

### Documentation (this feature)
```
specs/003-api-project-git/
├── plan-csv-integration.md  # This file
├── research.md              # Phase 0 output
├── data-model.md            # Phase 1 output (updated)
├── quickstart.md            # Phase 1 output (updated)
├── contracts/               # Phase 1 output (updated)
│   ├── csv-import-api.yaml
│   └── matching-api.yaml
└── tasks.md                 # Phase 2 output (/tasks command)
```

### Source Code (repository root)
```
models/
├── etc_meisai.go           # 既存
├── etc_dtako_mapping.go    # 既存
└── ryohi_sub_cal.go        # 新規: 参考モデル

services/
├── csv_import_service.go   # 新規: CSVインポート
├── matching_service.go     # 更新: マッチング強化
└── integration_service.go  # 新規: DB統合

api/handlers/
├── csv_import_handler.go   # 新規: インポートエンドポイント
└── matching_handler.go     # 更新: マッチングエンドポイント

repositories/
├── etc_repository.go       # 既存
└── dtako_repository.go     # 既存: 本番DB読み取り

tests/integration/
├── csv_import_test.go      # 新規
└── matching_test.go        # 新規
```

## Phase 0: Outline & Research

### Research Tasks
1. **ryohi_sub_calテーブル構造の調査**
   - テーブル定義とフィールド
   - ETC明細とのマッピング方法
   - 参考にすべきビジネスロジック

2. **本番DBアクセスパターンの確認**
   - dtako_rowsテーブル構造
   - etc_numフィールドの形式と値
   - 読み取り専用アクセスの実装方法

3. **CSVインポートの最適化**
   - バッチ処理サイズ
   - トランザクション管理
   - エラーハンドリング

4. **マッチングアルゴリズム**
   - etc_num完全一致
   - 日付・車両番号による補助マッチング
   - マッチング信頼度スコア

**Output**: research.md with integration patterns and decisions

## Phase 1: Design & Contracts

### Data Model Updates
1. **ryohi_sub_cal参考モデル**
   - フィールドマッピング定義
   - ETC明細との関連

2. **マッチング結果モデル**
   - マッチングステータス
   - 信頼度スコア
   - マッチング日時

### API Contracts
1. **CSVインポートエンドポイント**
   ```
   POST /api/etc/import/csv
   - multipart/form-data または file path
   - バッチサイズオプション
   - 重複チェックオプション
   ```

2. **マッチングエンドポイント**
   ```
   POST /api/etc/matching/execute
   - 日付範囲
   - マッチング条件

   GET /api/etc/matching/results/{id}
   - マッチング結果取得
   ```

3. **統計エンドポイント**
   ```
   GET /api/etc/matching/stats
   - マッチング率
   - 未マッチ件数
   ```

### Contract Tests
- CSVインポートのバリデーション
- マッチング処理の結果検証
- エラーケースのハンドリング

### Test Scenarios
1. **CSVインポート成功シナリオ**
2. **重複データ処理シナリオ**
3. **完全マッチングシナリオ**
4. **部分マッチングシナリオ**
5. **マッチング失敗シナリオ**

## Phase 2: Task Planning Approach

**Task Generation Strategy**:
1. **セキュリティタスク** [最優先]
   - 環境変数検証
   - 認証情報管理

2. **データモデルタスク**
   - ryohi_sub_cal参考実装
   - マッチング結果モデル

3. **APIエンドポイントタスク**
   - CSVインポートハンドラー
   - マッチングハンドラー

4. **統合テストタスク**
   - E2Eテスト実装
   - パフォーマンステスト

**Ordering Strategy**:
- セキュリティ → モデル → サービス → ハンドラー → テスト
- 並列実行可能なタスクの特定

**Estimated Output**: 15-20 numbered tasks

## Complexity Tracking
なし - 憲法に準拠した設計

## Progress Tracking
**Phase Status**:
- [x] Phase 0: Research (completed - research.md作成済み)
- [x] Phase 1: Design (completed - data-model.md更新、csv-import-api.yaml作成)
- [x] Phase 2: Task planning (completed - 実装タスク生成準備完了)
- [ ] Phase 3: Tasks generated (/tasks command実行待ち)
- [ ] Phase 4: Implementation
- [ ] Phase 5: Validation

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved: research.mdで解決済み
- [x] Complexity deviations documented: なし

---
*Based on Constitution v3.0.0 - See `/memory/constitution.md`*