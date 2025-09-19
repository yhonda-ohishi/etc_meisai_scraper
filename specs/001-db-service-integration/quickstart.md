# クイックスタート: データベースサービス統合

**対象**: 開発者・システム管理者
**前提条件**: Go 1.21+, Git, 基本的なgRPCの知識

## 統合後の開発環境セットアップ

### 1. 依存関係の確認

```bash
# 現在のプロジェクトディレクトリ
cd C:/go/etc_meisai

# Go モジュールの状態確認
go mod tidy

# 必要な依存関係が追加されていることを確認
cat go.mod | grep -E "(gorm|grpc|protobuf)"
```

### 2. db_service統合設定

```bash
# db_serviceの場所を確認
ls C:/go/db_service/

# go.modでローカル依存関係を設定
go mod edit -replace github.com/db_service=C:/go/db_service

# 依存関係の更新
go mod tidy
```

### 3. 環境変数設定

```bash
# .env ファイルの作成（例）
cat > .env << EOF
# データベース設定（統合後）
DATABASE_URL=mysql://user:password@localhost:3306/etc_meisai
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5

# gRPCサービス設定
GRPC_SERVER_PORT=50051
GRPC_CLIENT_TIMEOUT=30s

# ETC明細設定（既存）
ETC_CORPORATE_ACCOUNTS=account1,account2
ETC_PERSONAL_ACCOUNTS=personal1,personal2
EOF
```

## 統合後のアーキテクチャ理解

### レイヤー構造
```
HTTP API (handlers/)
    ↓
Service Layer (services/)
    ↓
gRPC Client → db_service (gRPC Server)
    ↓
Repository Layer (GORM)
    ↓
Database (MySQL/SQLite)
```

### 主要コンポーネント

1. **統合Repository**: `src/repositories/etc_integrated_repo.go`
2. **gRPCクライアント**: `src/clients/db_service_client.go`
3. **互換性レイヤー**: `src/adapters/etc_compat_adapter.go`
4. **モデル変換**: `src/models/etc_converter.go`

## 基本的な使用方法

### 1. サーバー起動

```bash
# db_serviceを先に起動
cd C:/go/db_service
go run cmd/server/main.go

# 別ターミナルでETC明細サーバー起動
cd C:/go/etc_meisai
go run cmd/simple_server/main.go
```

### 2. API テスト

```bash
# ヘルスチェック
curl http://localhost:8080/health

# アカウント一覧（既存API）
curl http://localhost:8080/api/accounts

# ETC明細一覧（統合後）
curl "http://localhost:8080/api/etc-meisai?limit=10&start_date=2025-01-01"
```

### 3. CSVインポートテスト

```bash
# サンプルCSVファイルを準備
cat > sample.csv << EOF
利用日,利用時間,入口IC,出口IC,料金,車両番号,ETC番号
2025-01-15,10:30,東京IC,横浜IC,850,123-4567,1234567890
2025-01-15,15:45,横浜IC,東京IC,850,123-4567,1234567890
EOF

# CSVアップロード
curl -X POST \
  -F "file=@sample.csv" \
  http://localhost:8080/api/parse
```

## 開発フロー

### 1. 新機能開発

```bash
# 1. テスト作成（TDD）
touch src/tests/unit/test_new_feature.go

# 2. インターフェース定義
# contracts/repository_interface.go に追加

# 3. gRPC契約更新
# contracts/grpc_service.proto に追加

# 4. 実装
# src/repositories/, src/services/ に実装

# 5. テスト実行
go test ./src/tests/unit/
go test ./src/tests/integration/
```

### 2. データベースマイグレーション

```bash
# 1. マイグレーションファイル作成
touch migrations/$(date +%Y%m%d%H%M%S)_add_new_field.sql

# 2. マイグレーション実行
go run cmd/migrate/main.go up

# 3. テストデータ投入
go run cmd/seed/main.go
```

### 3. パフォーマンステスト

```bash
# ベンチマークテスト実行
go test -bench=. ./src/tests/benchmark/

# メモリ使用量チェック
go test -benchmem -memprofile=mem.prof ./src/tests/benchmark/

# プロファイル確認
go tool pprof mem.prof
```

## トラブルシューティング

### よくある問題

#### 1. gRPC接続エラー
```bash
# エラー: "connection refused"
# 解決: db_serviceが起動しているか確認
ps aux | grep db_service

# ポート確認
netstat -an | grep 50051
```

#### 2. データベース接続エラー
```bash
# エラー: "database connection failed"
# 解決: 環境変数とDB状態を確認
echo $DATABASE_URL
mysql -u user -p -h localhost -e "SHOW DATABASES;"
```

#### 3. モジュール依存関係エラー
```bash
# エラー: "module not found"
# 解決: go.modの依存関係を確認・修正
go mod verify
go mod download
go clean -modcache
```

### ログ確認方法

```bash
# アプリケーションログ
tail -f logs/app.log

# db_serviceログ
tail -f C:/go/db_service/logs/service.log

# 構造化ログのフィルタリング
cat logs/app.log | jq '.level == "ERROR"'
```

## テスト実行ガイド

### 1. 単体テスト
```bash
# 全テスト実行
go test ./...

# カバレッジ付き
go test -cover ./src/...

# 詳細出力
go test -v ./src/tests/unit/
```

### 2. 統合テスト
```bash
# データベースを使用した統合テスト
go test -tags=integration ./src/tests/integration/

# E2Eテスト
go test -tags=e2e ./src/tests/e2e/
```

### 3. 契約テスト
```bash
# Repository契約テスト
go test ./src/tests/contract/repository/

# gRPC契約テスト
go test ./src/tests/contract/grpc/
```

## 監視・メトリクス

### 1. ヘルスチェック
```bash
# アプリケーション状態
curl http://localhost:8080/health

# データベース接続
curl http://localhost:8080/health/db

# gRPCサービス接続
curl http://localhost:8080/health/grpc
```

### 2. メトリクス確認
```bash
# パフォーマンスメトリクス
curl http://localhost:8080/metrics

# データベース統計
curl http://localhost:8080/api/stats/database

# 処理統計
curl http://localhost:8080/api/stats/processing
```

## 本番デプロイ準備

### 1. 設定ファイル確認
```bash
# 本番用設定
cp .env.example .env.production
# 本番用値を設定

# 設定値バリデーション
go run cmd/validate-config/main.go .env.production
```

### 2. セキュリティチェック
```bash
# ハードコードされた認証情報チェック
grep -r "password\|secret\|key" src/ --exclude-dir=tests

# 依存関係の脆弱性チェック
go list -json -m all | nancy sleuth
```

### 3. パフォーマンス検証
```bash
# 本番相当データでのテスト
go test -tags=production ./src/tests/performance/

# メモリリーク検証
go test -memprofile=mem.prof ./src/tests/stress/
```

## 次のステップ

1. **機能拡張**: 新しいRepositoryメソッドの追加
2. **最適化**: クエリパフォーマンスの改善
3. **監視強化**: より詳細なメトリクス収集
4. **自動化**: CI/CDパイプラインの構築

このガイドに従って、統合されたシステムの開発・運用を効率的に進めることができます。