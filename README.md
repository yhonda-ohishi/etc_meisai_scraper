# ETC明細管理システム

高速道路ETC明細データの取得・管理・分析を行うGoモジュール

## 🚀 特徴

- **高性能バッチ処理**: 10,000件のCSVレコードを5秒以内で処理
- **重複検出**: SHA256ハッシュベースの確実な重複防止
- **レガシー互換**: 既存38フィールドAPIとの完全互換性
- **gRPC統合**: db_serviceとのシームレスな連携
- **Docker対応**: 即座にデプロイ可能なコンテナ設定
- **包括的テスト**: 単体・統合テストによる品質保証

## 📋 必要要件

- Go 1.21以上
- Docker & Docker Compose (オプション)
- PostgreSQL 15+ または SQLite (開発用)
- gRPCサーバー (db_service)

## 🛠️ インストール

### 1. リポジトリのクローン

```bash
git clone https://github.com/yhonda-ohishi/etc_meisai.git
cd etc_meisai
```

### 2. 依存関係のインストール

```bash
go mod download
```

### 3. 環境設定

```bash
cp .env.example .env
# .envファイルを編集して設定値を調整
```

### 4. データベースマイグレーション

```bash
go run cmd/migrate/main.go migrate
```

## 🚦 クイックスタート

### ローカル実行

```bash
# サーバー起動
go run cmd/server/main.go

# ヘルスチェック
curl http://localhost:8080/health
```

### Docker実行

```bash
# ビルドと起動
docker-compose up -d

# ログ確認
docker-compose logs -f etc_meisai

# 停止
docker-compose down
```

---

**Version**: 1.0.0
**Last Updated**: 2025-01-20
**Maintainer**: yhonda-ohishi
