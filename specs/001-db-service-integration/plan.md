# Implementation Plan: データベースサービス統合

**Branch**: `001-db-service-integration` | **Date**: 2025-09-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-db-service-integration/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → ✅ COMPLETE: Feature spec loaded and analyzed
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → ✅ COMPLETE: All technical context resolved through research
3. Fill the Constitution Check section
   → ✅ COMPLETE: Constitution requirements verified
4. Evaluate Constitution Check section
   → ✅ COMPLETE: No violations, proceeding with implementation
5. Execute Phase 0 → research.md
   → ✅ COMPLETE: Research findings documented
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, CLAUDE.md
   → ✅ COMPLETE: All Phase 1 artifacts created
7. Re-evaluate Constitution Check section
   → ✅ COMPLETE: Design maintains constitutional compliance
8. Plan Phase 2 → Describe task generation approach
   → ✅ COMPLETE: Task planning approach documented
9. STOP - Ready for /tasks command
   → ✅ COMPLETE: Plan execution finished
```

## Summary
C:\go\db_service に作成されたdb-handler-serverパターンのdatabase_repoを現在のETC明細システムに統合。既存のデータベース操作をリポジトリパターンに移行し、コードの保守性を向上させる。段階的な移行により、既存APIの互換性を保ちながらアーキテクチャを改善する。

## Technical Context
**Language/Version**: Go 1.21+
**Primary Dependencies**: chi (HTTPルーティング), Playwright-go, GORM, gRPC, database_repo
**Storage**: database_repo経由でのデータアクセス (GORM + MySQL/SQLite)
**Testing**: go test, 統合テスト, 契約テスト
**Target Platform**: Windows (MINGW64), サーバー環境
**Project Type**: single (モノリシックGoアプリケーション)
**Performance Goals**: CSVファイル1万行を5秒以内で処理、現在のレスポンス時間維持
**Constraints**: <500MB メモリ使用量、既存API互換性、段階的移行必須
**Scale/Scope**: 現在のコードベース、5アカウント同時ダウンロード対応

**統合対象**: C:\go\db_service のdatabase_repo (db-handler-serverパターン実装)

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### 初期チェック ✅ PASS
✅ **セキュリティ原則**: 認証情報のハードコード禁止 - 既存の環境変数使用を維持
✅ **Test-First**: TDD必須 - リポジトリ統合前にテスト作成
✅ **統合テスト**: 新規ライブラリ(database_repo)の契約テスト必須
✅ **可観測性**: 構造化ログ記録、明確なエラーメッセージ

### Phase 1後の再チェック ✅ PASS
✅ **データモデル設計**: GORMベースの正規化されたモデル、適切な検証機能
✅ **API契約**: gRPC契約定義、型安全な通信、エラーハンドリング
✅ **テスト戦略**: 契約テスト、統合テスト、E2Eテストの包括的カバレッジ
✅ **段階的移行**: 既存システムへの影響を最小化する設計

**結果**: 憲法要件に完全準拠。統合テストと段階的移行によりリスクを軽減。

## Project Structure

### Documentation (this feature) ✅ COMPLETE
```
specs/001-db-service-integration/
├── plan.md              # This file (/plan command output) ✅
├── research.md          # Phase 0 output (/plan command) ✅
├── data-model.md        # Phase 1 output (/plan command) ✅
├── quickstart.md        # Phase 1 output (/plan command) ✅
├── contracts/           # Phase 1 output (/plan command) ✅
│   ├── repository_interface.go  # 統合Repository interface
│   └── grpc_service.proto       # gRPCサービス契約
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Integrated structure (post-implementation)
src/
├── models/              # → 統合モデル (GORM + 互換性レイヤー)
├── services/            # → gRPCクライアント化
├── repositories/        # → 統合Repository実装
├── handlers/            # → 維持 (service層経由でアクセス)
├── adapters/            # → 新規 (互換性レイヤー)
├── clients/             # → 新規 (gRPCクライアント)
├── parser/              # → 維持
└── config/              # → db_service設定管理統合

tests/
├── contract/            # → 新規 (Repository契約テスト)
├── integration/         # → 拡張 (E2Eテスト)
└── unit/                # → 維持
```

**Structure Decision**: 既存構造を維持しつつ、段階的にdb-handler-serverパターンに移行

## Phase 0: Outline & Research ✅ COMPLETE

### 実施内容
1. **C:\go\db_service の実装調査**: 完全なdb-handler-serverパターン実装を確認
   - GORMベースのRepository層
   - gRPCサービス実装
   - 包括的なテスト体制
   - Protocol Buffers契約定義

2. **ETC明細システム分析**: 現在の3層アーキテクチャと統合ポイントを特定
   - handlers/の7ファイル構成
   - services/の6サービス実装状況
   - 既存38フィールドモデルの互換性課題
   - database/sql直接使用からの移行要件

3. **統合戦略決定**: 4段階の段階的移行アプローチを決定
   - Phase 1: モデル層統合
   - Phase 2: Repository層移行
   - Phase 3: サービス統合
   - Phase 4: 最適化・クリーンアップ

### 成果物
- **research.md**: 統合アプローチと技術的発見を文書化
- **統合ポイント特定**: 明確な統合戦略と実装方針
- **リスク評価**: 段階的移行によるリスク軽減戦略

## Phase 1: Design & Contracts ✅ COMPLETE

### 1. データモデル統合設計 → `data-model.md` ✅
- **統合ETCMeisaiモデル**: db_serviceのGORMモデルをベース
- **互換性レイヤー**: 既存38フィールド対応のETCMeisaiCompat構造体
- **マッピングモデル**: ETCMeisaiMapping for DTako関連付け
- **バッチ処理モデル**: ETCImportBatch for CSV処理管理
- **バリデーション規則**: 包括的なデータ検証ルール
- **パフォーマンス最適化**: インデックス戦略とバッチ処理設計

### 2. API契約定義 → `/contracts/` ✅
- **repository_interface.go**: 統合Repository interface定義
  - 基本CRUD操作 (db_service互換)
  - ETC固有機能 (バルク操作、集計)
  - ハッシュベース重複検出
  - トランザクション支援
- **grpc_service.proto**: gRPCサービス契約
  - ETCService: メインデータ操作
  - ETCMappingService: マッピング管理
  - ETCImportService: バッチ処理
  - 包括的なメッセージ定義

### 3. 契約テスト設計
- **Repository契約テスト**: 全CRUD操作の動作保証
- **gRPCサービステスト**: 型安全な通信の検証
- **API互換性テスト**: 既存エンドポイントの動作保証
- **パフォーマンステスト**: 要件達成の検証

### 4. 統合ガイド → `quickstart.md` ✅
- **環境セットアップ**: 統合開発環境の構築手順
- **API使用方法**: 統合後のAPIテスト方法
- **開発フロー**: TDD従った開発プロセス
- **トラブルシューティング**: よくある問題と解決方法
- **監視・メトリクス**: 統合後の監視方法

### 5. エージェントコンテキスト更新 → `CLAUDE.md` ✅
- **統合アーキテクチャ図**: gRPCベースのデータフロー
- **統合コンポーネント**: 新規作成予定ファイル一覧
- **統合仕様リンク**: Phase 1成果物への参照
- **技術スタック更新**: GORM + gRPC追加
- **環境変数更新**: 統合後の設定項目

## Phase 2: Task Planning Approach ✅ COMPLETE

**Task Generation Strategy**:
統合の複雑性と段階的移行の要件に基づき、以下の4フェーズでタスクを生成:

### Phase 1: モデル層統合タスク [P] (並列実行可能)
- ETCMeisai統合モデル実装
- 互換性レイヤー (ETCMeisaiCompat) 実装
- モデル変換ユーティリティ作成
- バリデーション機能統合
- 単体テスト作成 (TDD)

### Phase 2: Repository層移行タスク (順次実行)
- 統合Repository interface実装
- gRPCクライアント実装
- GORM統合とクエリ変換
- 既存database/sqlからの移行
- 契約テスト実装

### Phase 3: サービス統合タスク (順次実行)
- services/層のgRPCクライアント化
- handlers/層の統合Repository使用
- エラーハンドリング統一
- ログ・メトリクス統合
- 統合テスト実装

### Phase 4: 最適化・クリーンアップタスク [P] (並列実行可能)
- パフォーマンス最適化
- 不要コードの削除
- ドキュメント更新
- E2Eテスト実装
- 本番環境準備

**Ordering Strategy**:
- **TDD順序**: テスト作成 → 実装 → 検証
- **依存順序**: モデル → Repository → サービス → API
- **リスク軽減**: 段階的移行、ロールバック戦略
- **互換性確保**: 既存機能テスト → 統合 → 新機能テスト

**Estimated Output**: 22-25個の順序付きタスク
- Phase 1: 6-7タスク (モデル統合)
- Phase 2: 6-7タスク (Repository移行)
- Phase 3: 5-6タスク (サービス統合)
- Phase 4: 5-6タスク (最適化・検証)

**特別な考慮事項**:
- 既存システムの無停止移行
- データ整合性の保証
- パフォーマンス要件の維持
- セキュリティ要件の遵守

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| なし | 憲法完全準拠 | 統合テストとTDDにより複雑性を管理 |

**設計複雑性の管理**:
- 互換性レイヤーによる既存システム保護
- 段階的移行によるリスク分散
- 包括的テスト戦略による品質保証
- 明確な責任分離によるコード保守性確保

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
- [x] Complexity deviations documented (none required)

**Artifact Generation Status**:
- [x] research.md: 統合アプローチと技術調査
- [x] data-model.md: 統合データモデル設計
- [x] contracts/repository_interface.go: Repository契約定義
- [x] contracts/grpc_service.proto: gRPCサービス契約
- [x] quickstart.md: 統合開発ガイド
- [x] CLAUDE.md: エージェントコンテキスト更新

**次のステップ**: /tasks コマンドで詳細実装タスクを生成

---
*Based on Constitution v3.0.0 - See `/.specify/memory/constitution.md`*