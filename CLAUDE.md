# CLAUDE.md - ETC明細Goモジュール プロジェクトコンテキスト

## プロジェクト概要
ETC明細データをWebスクレイピングで取得し、データベースに保存するGoモジュール。
db-handler-serverパターンに従ったハンドラー実装への移行中。

## 技術スタック
- **言語**: Go 1.21+
- **フレームワーク**: chi (HTTPルーティング)
- **データベース**: GORM + MySQL/SQLite (db_service統合)
- **通信**: gRPC (database_repo統合)
- **スクレイピング**: Playwright-go
- **依存管理**: Go Modules
- **アーキテクチャ**: db-handler-serverパターン (統合中)

## プロジェクト構造
```
etc_meisai/
├── api.go               # メインクライアントインターフェース
├── module.go            # モジュール初期化
├── handlers/            # HTTPハンドラー（db-handler-serverパターン）
├── services/            # ビジネスロジック層
├── repositories/        # データアクセス層
├── models/              # データモデル
├── parser/              # CSV解析
├── config/              # 設定管理
└── downloads/           # CSVファイル保存先
```

## 主要機能
1. **ETC明細ダウンロード**: 複数アカウント対応、非同期処理
2. **データ処理**: CSV解析、データ変換、バルク保存
3. **マッピング管理**: ETC明細とデジタコデータの関連付け（etc_num活用）
4. **進捗追跡**: リアルタイム進捗通知（SSE対応）
5. **自動マッチング**: dtako_row_idとの精密マッチング

## 最近の変更 (v0.0.18 - 統合進行中)
- **database_repo統合**: C:\go\db_service のdb-handler-serverパターン統合
- **アーキテクチャ移行**: GORM + gRPCベースのデータアクセス層
- **統合テスト**: Repository契約テスト・統合テストスイート実装

## 開発中の機能 (統合フェーズ)
- **モデル統合**: db_serviceのGORMモデル + 互換性レイヤー実装
- **Repository統合**: 統合Repository interface + gRPCクライアント実装
- **サービス統合**: 既存services/のgRPCクライアント化

## スコープ外の機能
- Excel/PDF エクスポート機能
- 統計情報生成機能
- キャッシュ機能（ユーザー要求により除外）

## パフォーマンス目標
- CSVファイル1万行を5秒以内で処理
- メモリ使用量500MB以下
- 同時ダウンロード5アカウントまで

## SQLite最適化設定
```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = normal;
PRAGMA cache_size = -32000;
```

## 統合アーキテクチャ (database_repo)
```
HTTP API (handlers/) → Service Layer (services/) → gRPC Client → db_service
                                                      ↓
                                               Repository (GORM) → Database
```

### 統合コンポーネント
- **統合Repository**: `src/repositories/etc_integrated_repo.go`
- **gRPCクライアント**: `src/clients/db_service_client.go`
- **互換性レイヤー**: `src/adapters/etc_compat_adapter.go`
- **統合テスト**: `tests/contract/`, `tests/integration/`

### 統合仕様
- **データモデル**: [data-model.md](specs/001-db-service-integration/data-model.md)
- **API契約**: [contracts/](specs/001-db-service-integration/contracts/)
- **開発ガイド**: [quickstart.md](specs/001-db-service-integration/quickstart.md)

## 環境変数
- `ETC_CORPORATE_ACCOUNTS`: 法人アカウント（カンマ区切り）
- `ETC_PERSONAL_ACCOUNTS`: 個人アカウント（カンマ区切り）
- `DATABASE_URL`: データベース接続URL (統合後)
- `GRPC_SERVER_PORT`: gRPCサーバーポート (統合後)

## テストコマンド
```bash
go test ./...                    # 単体テスト
go test ./tests/integration -v   # 統合テスト
```

## ビルド＆実行
```bash
go build -o etc_meisai
./etc_meisai
```

---
*最終更新: 2025-09-19 | v0.0.17*