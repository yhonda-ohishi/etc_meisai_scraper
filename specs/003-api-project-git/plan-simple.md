# Implementation Plan: ETC明細スクレイピング＆DBインポート（シンプル版）

**Branch**: `003-api-project-git-simple` | **Date**: 2025-09-19 | **Spec**: `/specs/003-api-project-git/spec-simple.md`
**Input**: スクレイピングとDBインポートのみに機能を限定

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path ✓
2. Fill Technical Context ✓
3. Fill the Constitution Check section ✓
4. Evaluate Constitution Check section ✓
5. Execute Phase 0 → research-simple.md
6. Execute Phase 1 → data-model-simple.md, quickstart-simple.md
7. Re-evaluate Constitution Check section
8. Plan Phase 2 → Task generation approach
9. STOP - Ready for /tasks command
```

## Summary
ETCポータルサイトから明細データをスクレイピングし、ローカルMySQLデータベースに保存するシンプルな実装。APIやマッチング機能は除外し、コア機能に集中。

## Technical Context
**Language/Version**: Go 1.22+
**Primary Dependencies**: Playwright-go (スクレイピング), MySQL driver, encoding/csv
**Storage**: MySQL (localhost:3307のみ、本番DB接続なし)
**Testing**: Go testing package, testify
**Target Platform**: Windows/Linux (CLI実行)
**Project Type**: single (CLI tool)
**Performance Goals**: 1アカウント5分以内
**Constraints**: 環境変数による設定、ハードコードなし
**Scale/Scope**: 法人・個人アカウント対応、バッチ処理

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### セキュリティ原則チェック
- [x] ハードコードされた認証情報の禁止 → 環境変数使用
- [x] パスワードは環境変数から読み込む
- [x] テストコードでもハードコード禁止

### シンプル化原則
- [x] 不要な機能を削除（API、マッチング、統合）
- [x] コア機能に集中（スクレイピング、DBインポート）
- [x] 複雑な依存関係を避ける

## Project Structure

### Documentation (this feature)
```
specs/003-api-project-git/
├── plan-simple.md           # This file
├── research-simple.md       # Phase 0 output
├── data-model-simple.md     # Phase 1 output
├── quickstart-simple.md     # Phase 1 output
└── tasks-simple.md          # Phase 2 output (/tasks command)
```

### Source Code (repository root)
```
cmd/
└── scraper/
    └── main.go              # CLIエントリーポイント

scraper/
├── scraper.go              # スクレイピングロジック
└── login.go                # ログイン処理

parser/
└── csv_parser.go           # CSV解析

models/
└── etc_meisai.go           # データモデル

repositories/
└── etc_repository.go       # DB操作

config/
└── database.go             # DB接続設定

tests/
└── integration/
    └── scraper_test.go     # 統合テスト
```

## Phase 0: Outline & Research

### Research Tasks
1. **現在のスクレイピング実装の確認**
   - Playwright-goの使用状況
   - ログイン処理の実装
   - CSV取得ロジック

2. **データベース構造の簡素化**
   - 必要最小限のテーブル
   - インデックス設計
   - トランザクション処理

3. **不要機能の削除対象**
   - APIハンドラー
   - マッチング処理
   - 本番DB接続

**Output**: research-simple.md with simplification decisions

## Phase 1: Design & Contracts

### Data Model Updates
1. **ETCMeisaiテーブル（簡素化）**
   - 基本フィールドのみ
   - 重複検出用インデックス

2. **ETCAccountテーブル**
   - アカウント管理用
   - 環境変数からの読み込み

### CLI Interface Design
```bash
# 実行例
./etc_scraper --account corporate --from 2025-01-01 --to 2025-01-31
./etc_scraper --all-accounts --last-month
```

### Test Scenarios
1. **スクレイピング成功**
2. **DB保存成功**
3. **重複スキップ**
4. **エラーハンドリング**

## Phase 2: Task Planning Approach

**Task Generation Strategy**:
1. **クリーンアップタスク**
   - 不要なAPIコード削除
   - 依存関係整理

2. **コア機能タスク**
   - スクレイピング実装
   - DB保存処理
   - 重複検出

3. **テストタスク**
   - 統合テスト作成
   - 実行確認

**Ordering Strategy**:
- クリーンアップ → コア実装 → テスト

**Estimated Output**: 10-15 numbered tasks

## Complexity Tracking
なし - シンプル化により複雑性を排除

## Progress Tracking
**Phase Status**:
- [x] Phase 0: Research (completed - research-simple.md)
- [x] Phase 1: Design (completed - data-model-simple.md, quickstart-simple.md)
- [x] Phase 2: Task planning (completed - ready for /tasks)
- [ ] Phase 3: Tasks generated (/tasks command待ち)
- [ ] Phase 4: Implementation
- [ ] Phase 5: Validation

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All clarifications resolved
- [x] Simplification complete

---
*Based on Constitution v3.0.0 - Simplified for core functionality*