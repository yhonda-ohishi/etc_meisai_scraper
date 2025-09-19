# CLAUDE.md - ETC明細Goモジュール プロジェクトコンテキスト

## プロジェクト概要
ETC明細データをWebスクレイピングで取得し、データベースに保存するGoモジュール。
db-handler-serverパターンに従ったハンドラー実装への移行中。

## 技術スタック
- **言語**: Go 1.21+
- **フレームワーク**: chi (HTTPルーティング)
- **データベース**: SQLite (WALモード)
- **スクレイピング**: Playwright-go
- **依存管理**: Go Modules
- **アーキテクチャ**: db-handler-serverパターン

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

## 最近の変更 (v0.0.17)
- db-handler-serverパターンへの移行開始
- 不要ファイルの削除とコードベース整理
- APIエンドポイント仕様の完成（OpenAPI）

## 開発中の機能
- ParseCSVHandler の実装（優先度: 高）
- 自動マッチングアルゴリズム（etc_numベース）
- 手動マッピング管理UI

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

## 環境変数
- `ETC_CORPORATE_ACCOUNTS`: 法人アカウント（カンマ区切り）
- `ETC_PERSONAL_ACCOUNTS`: 個人アカウント（カンマ区切り）
- `DATABASE_PATH`: SQLiteデータベースパス

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