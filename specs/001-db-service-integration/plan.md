# Implementation Plan: etc_meisai Server Repository Integration

**Branch**: `001-db-service-integration` | **Date**: 2025-09-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `C:/go/etc_meisai/specs/001-db-service-integration/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → ✓ Loaded successfully
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → ✓ No clarifications needed - instructions are comprehensive
   → Project Type: server (gRPC service integration)
   → Structure Decision: Option 1 (single project)
3. Fill the Constitution Check section based on the content of the constitution document.
   → ✓ Filled based on ETC明細システム Constitution v3.0.0
4. Evaluate Constitution Check section below
   → ✓ No violations identified
   → Update Progress Tracking: Initial Constitution Check PASS
5. Execute Phase 0 → research.md
   → Executing research phase
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
7. Re-evaluate Constitution Check section
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
etc_meisaiをserver_repoに統合し、Swagger UIにETC明細管理エンドポイントを自動表示させる。go-chiからgRPC+Protocol Buffersアーキテクチャへの移行により、db_serviceと統一された設計パターンを実現する。

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: gRPC, Protocol Buffers, grpc-gateway, GORM, Playwright-go
**Storage**: MySQL/SQLite (db_service経由)
**Testing**: Go test, 統合テスト, 契約テスト
**Target Platform**: Linux/Windows サーバー
**Project Type**: single (gRPCサービス)
**Performance Goals**: CSVファイル1万行を5秒以内で処理
**Constraints**: メモリ使用量500MB以下、同時ダウンロード5アカウントまで
**Scale/Scope**: 複数アカウント対応、バルク処理対応

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ドキュメント言語
- [x] すべての仕様書、計画書、設計書は日本語で記述

### セキュリティ原則
- [x] ハードコードされた認証情報の禁止を徹底
- [x] 環境変数からの認証情報読み込み設計

### Test-First
- [x] TDD必須: テスト作成 → ユーザー承認 → テスト失敗確認 → 実装
- [x] Red-Green-Refactorサイクルの適用計画

### 統合テスト
- [x] gRPCサービスの契約テスト計画
- [x] サービス間通信のテスト計画

### 可観測性
- [x] 構造化ログ記録の実装計画
- [x] 明確なエラーメッセージ設計

## Project Structure

### Documentation (this feature)
```
specs/001-db-service-integration/
├── plan.md              # This file (/plan command output)
├── spec.md              # Feature specification (created)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (選択)
src/
├── proto/               # Protocol Buffers定義
├── pb/                  # 生成されたgRPCコード
├── grpc/                # gRPCサーバー実装
├── models/              # データモデル
├── services/            # ビジネスロジック
└── adapters/            # 既存コードとの互換性レイヤー

tests/
├── contract/            # 契約テスト
├── integration/         # 統合テスト
└── unit/                # 単体テスト
```

**Structure Decision**: Option 1 (Single project) - gRPCサービス統合に最適

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - Protocol Buffers定義のベストプラクティス
   - grpc-gateway設定方法
   - go-chiからgRPCへの段階的移行戦略
   - bufコンパイラの設定最適化

2. **Generate and dispatch research agents**:
   ```
   Task: "Research Protocol Buffers service definition best practices for ETC toll system"
   Task: "Find grpc-gateway configuration for Swagger auto-generation"
   Task: "Research migration pattern from go-chi to gRPC"
   Task: "Find buf compiler optimization for Go code generation"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: Protocol Buffersファースト設計
   - Rationale: Swagger自動生成、型安全性、一貫性
   - Alternatives considered: RESTのまま手動Swagger定義、GraphQL

**Output**: research.md with all technical decisions documented

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - ETCMeisaiRecord (ETC明細レコード)
   - ETCMapping (マッピング情報)
   - ImportSession (インポートセッション)
   - 各エンティティのバリデーションルール

2. **Generate API contracts** from functional requirements:
   - CreateETCMeisai (POST /api/v1/etc-meisai/records)
   - GetETCMeisai (GET /api/v1/etc-meisai/records/{id})
   - ListETCMeisai (GET /api/v1/etc-meisai/records)
   - ImportCSV (POST /api/v1/etc-meisai/import)
   - CreateMapping (POST /api/v1/etc-meisai/mappings)
   - Output to `/contracts/etc_meisai.proto`

3. **Generate contract tests** from contracts:
   - test_create_etc_meisai.go
   - test_get_etc_meisai.go
   - test_list_etc_meisai.go
   - test_import_csv.go
   - test_create_mapping.go

4. **Extract test scenarios** from user stories:
   - Swagger UI表示確認テスト
   - CSVインポート統合テスト
   - マッピング作成統合テスト
   - エラーハンドリングテスト

5. **Update CLAUDE.md incrementally**:
   - gRPC統合情報を追加
   - Protocol Buffers使用を記載
   - 最新の変更履歴を更新

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, CLAUDE.md更新

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- protoファイル作成タスク [P]
- buf.gen.yaml設定タスク [P]
- 各gRPCメソッド実装タスク
- 既存サービスのラップタスク
- server_repo統合タスク
- テスト実装タスク（契約テスト、統合テスト）

**Ordering Strategy**:
- Proto定義 → コード生成 → gRPCサーバー実装 → 統合
- テストファーストアプローチの徹底
- 並列実行可能なタスクには[P]マーク

**Estimated Output**: 25-30個の順序付けられたタスク in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*No violations identified - standard gRPC integration pattern*

## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (なし)

---
*Based on ETC明細システム Constitution v3.0.0 - See `.specify/memory/constitution.md`*