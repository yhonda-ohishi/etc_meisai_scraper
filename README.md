# ETC明細ダウンロードサービス

ETCの利用明細をWebスクレイピングで自動取得するGoモジュールです。

**最新バージョン: v0.0.23** | [変更履歴](#-変更履歴)

## 📢 v0.0.23 リリース情報

- **セッションフォルダ共有**: 複数アカウントのCSVファイルを同一フォルダに保存
- **ファイル名プレフィックス**: アカウント名が自動的にファイル名に付与
- **整理されたダウンロード**: downloads/YYYYMMDD_HHMMSS/配下に各実行分を整理
- **使いやすさ向上**: 1回の実行で複数アカウントのCSVを一箇所に集約

## 🚀 特徴

- **自動ダウンロード**: ETC明細サービスからCSVファイルを自動取得
- **複数アカウント対応**: 法人・個人の複数アカウントを同時処理
- **非同期処理**: 効率的な並行ダウンロード
- **モック対応設計**: テスト容易なインターフェース設計
- **100%テストカバレッジ**: 高品質なコードベース（手書きコード）

## 📋 必要要件

- Go 1.21以上
- Playwright (自動インストール)

## 🔧 インストール

```bash
go get github.com/yhonda-ohishi/etc_meisai
```

## 🏃 クイックスタート

### 基本的な使い方

```go
package main

import (
    "github.com/yhonda-ohishi/etc_meisai_scraper/src/scraper"
    "log"
)

func main() {
    config := &scraper.ScraperConfig{
        UserID:   "your-user-id",
        Password: "your-password",
        Headless: true,
    }

    scraper, err := scraper.NewETCScraper(config, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer scraper.Close()

    // 初期化
    if err := scraper.Initialize(); err != nil {
        log.Fatal(err)
    }

    // ログイン
    if err := scraper.Login(); err != nil {
        log.Fatal(err)
    }

    // 明細ダウンロード（CSVファイル保存）
    csvPath, err := scraper.DownloadMeisai("2024-01-01", "2024-01-31")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("CSVファイル保存完了: %s", csvPath)
}
```

### スタンドアロンサーバーとして実行

このモジュールは別プロセスとして実行し、他のサービス（例: desktop-server）からgRPCで接続できます。

#### gRPCサーバーとして起動（推奨）

```bash
# デフォルト設定で起動（ポート: 50052）
./etc_meisai_scraper.exe

# カスタムポートで起動
./etc_meisai_scraper.exe --grpc-port 50052
```

#### レガシーHTTPサーバーとして起動

```bash
./etc_meisai_scraper.exe --grpc=false --http-port 8080
```

#### ヘルプの表示

```bash
./etc_meisai_scraper.exe --help
```

### desktop-server との統合

desktop-serverとの統合では、etc_meisai_scraperを**別プロセス**として実行することを推奨します：

```
┌─────────────────────────────────────┐
│ desktop-server.exe                  │
│ - db_service統合（同一プロセス）    │
│ - gRPC-Webプロキシ                  │
│ - フロントエンド提供                │
└─────────────────────────────────────┘
         │
         │ gRPC Client
         ↓
┌─────────────────────────────────────┐
│ etc_meisai_scraper.exe (別プロセス) │
│ - DownloadService提供               │
│ - Playwrightでスクレイピング        │
└─────────────────────────────────────┘
```

**メリット:**
- ✅ desktop-serverのバイナリサイズが小さいまま
- ✅ 環境依存性の分離（Playwright依存）
- ✅ スクレイピング処理がデスクトップアプリに影響しない
- ✅ 必要な時だけ起動可能（オンデマンド起動）

詳細な統合方法は [desktop-serverリポジトリ](https://github.com/yhonda-ohishi/desktop-server) を参照してください。

## 🌐 API エンドポイント

### gRPC-Gateway REST API

gRPC-Gatewayを使用してREST APIとして公開する場合、以下のエンドポイントが利用可能です：

- `POST /etc_meisai_scraper/v1/download/sync` - 同期ダウンロード
- `POST /etc_meisai_scraper/v1/download/async` - 非同期ダウンロード
- `GET /etc_meisai_scraper/v1/download/jobs/{job_id}` - ジョブステータス取得
- `GET /etc_meisai_scraper/v1/accounts` - 全アカウントID取得

### gRPC サービス

gRPCサービスとして利用する場合：
- `DownloadService.DownloadSync` - 同期ダウンロード
- `DownloadService.DownloadAsync` - 非同期ダウンロード
- `DownloadService.GetJobStatus` - ジョブステータス確認
- `DownloadService.GetAllAccountIDs` - 全アカウントID取得

## 📝 Swagger/OpenAPI ドキュメント生成

### 初期セットアップ

初回のみ、proto依存関係をダウンロードする必要があります：

```bash
# googleapis と grpc-gateway の proto ファイルを取得
mkdir -p third_party
git clone --depth=1 https://github.com/googleapis/googleapis.git third_party/googleapis
git clone --depth=1 https://github.com/grpc-ecosystem/grpc-gateway.git third_party/grpc-gateway

# protoc プラグインのインストール
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### コード生成とSwagger更新

```bash
# buf を使用してコード生成（gRPC-Gateway と Swagger を含む）
buf generate src/proto

# または protoc を直接使用
protoc -I src/proto \
  -I third_party/googleapis \
  -I third_party/grpc-gateway \
  --go_out=src/pb --go_opt=paths=source_relative \
  --go-grpc_out=src/pb --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=src/pb --grpc-gateway_opt=paths=source_relative,grpc_api_configuration=src/proto/download_api.yaml \
  --openapiv2_out=swagger --openapiv2_opt=grpc_api_configuration=src/proto/download_api.yaml \
  src/proto/download.proto
```

### 生成されるファイル

- `src/pb/download.pb.go` - Protocol Buffers のGoコード
- `src/pb/download_grpc.pb.go` - gRPC サーバー/クライアントコード
- `src/pb/download.pb.gw.go` - gRPC-Gateway コード
- `swagger/etc_meisai.swagger.json` - OpenAPI/Swagger 定義

### HTTPマッピングの変更

REST APIのパスを変更する場合は、`src/proto/download_api.yaml` を編集してから再生成：

```yaml
http:
  rules:
    - selector: etc_meisai.download.v1.DownloadService.DownloadSync
      post: /etc_meisai_scraper/v1/download/sync
      body: "*"
```

## 📊 テストカバレッジ

このプロジェクトは**100%のテストカバレッジ**を達成しています（自動生成コードを除く）。

### カバレッジレポートの確認

```bash
# カバレッジレポートの生成と表示
./show_coverage.sh
```

出力例：
```
📊 テストカバレッジレポート (Generated Codeを除く)
================================================
✅ etc_scraper.go:NewETCScraper                    100.0%
✅ etc_scraper.go:Initialize                        100.0%
✅ etc_scraper.go:Login                             100.0%
✅ etc_scraper.go:DownloadMeisai                    100.0%
...
============================================
📊 総合カバレッジ (PB除外): 100.0%
============================================
```

### テストの実行

```bash
# 全テスト実行
go test ./...

# カバレッジ付きテスト
go test -cover ./...

# 特定パッケージのテスト
go test ./tests/unit/scraper/...
```

## 📁 プロジェクト構造

```
etc_meisai/
├── src/
│   ├── scraper/         # Webスクレイピング機能
│   ├── services/        # ビジネスロジック
│   ├── handlers/        # HTTPハンドラー
│   ├── grpc/           # gRPCサーバー
│   └── models/         # データモデル
├── tests/
│   ├── unit/           # 単体テスト
│   ├── integration/    # 統合テスト
│   └── mocks/          # モック定義
└── show_coverage.sh    # カバレッジレポート生成
```

## ⚙️ 環境変数

| 変数名 | 説明 | デフォルト値 |
|--------|------|--------------|
| `ETC_CORPORATE_ACCOUNTS` | 法人アカウント（カンマ区切り） | - |
| `ETC_PERSONAL_ACCOUNTS` | 個人アカウント（カンマ区切り） | - |
| `ETC_HEADLESS` | Headlessモード | `true` |

### ETC_HEADLESS の使用例

```bash
# Headlessモード（デフォルト、ブラウザ非表示）
ETC_HEADLESS=true ./etc_meisai_scraper.exe

# ブラウザ表示モード（デバッグ用）
ETC_HEADLESS=false ./etc_meisai_scraper.exe
```

**推奨**: 本番環境では`ETC_HEADLESS=true`（デフォルト）を使用してください。

## 🔒 セキュリティ

- パスワードは環境変数で管理
- Headlessモードでの実行推奨（`ETC_HEADLESS=true`）
- ログに機密情報は出力されません

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'feat: Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

### コミットメッセージ規約

- `feat:` 新機能
- `fix:` バグ修正
- `test:` テスト追加・修正
- `docs:` ドキュメント更新
- `refactor:` リファクタリング

## 📜 変更履歴

### v0.0.23 (2025-10-18)
- **feat**: セッションフォルダ共有機能
  - 複数アカウントのCSVファイルを同一セッションフォルダに保存
  - ファイル名にアカウント名プレフィックスを自動付与
  - ScraperConfigにSessionFolderフィールドを追加
  - ProcessAsync()で共通のタイムスタンプ付きフォルダを生成
  - ファイル管理の改善: downloads/YYYYMMDD_HHMMSS/配下に整理

### v0.0.22 (2025-10-18)
- **fix**: CSV明細ダウンロード機能の修正
  - ダイアログ自動承認ハンドラーを追加（CSV確認ダイアログ対応）
  - ダウンロードハンドラーの安定化（SaveAs()のハング対策）
  - goroutineとタイムアウトを使用してダウンロードを監視
  - 不要なチェックボックス選択処理を削除
  - PageInterfaceにEvaluateメソッドを追加
  - BrowserContextInterfaceにOnメソッドを追加
  - 詳細なログ出力を追加（デバッグ用）

### v0.0.21 (2025-10-18)
- **fix**: ETCログイン処理の修正
  - ログインリンクをクリックして専用ログインページに遷移するように修正
  - 正しいフォームフィールド名 (`risLoginId`/`risPassword`) を使用
  - 正しいログインボタンセレクタを使用
  - 冗長なデバッグコードと複数セレクタ試行ロジックを削除
  - 全ユニットテストを更新し、100%テストカバレッジを維持

### v0.0.20 (2025-09-21)
- **feat**: desktop-server統合対応
  - Registry パッケージ追加
  - スタンドアロンgRPCサーバー機能
  - 別プロセス実行をサポート

## 📝 ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 📧 お問い合わせ

問題や質問がある場合は、[Issues](https://github.com/yhonda-ohishi/etc_meisai_scraper/issues)でお知らせください。

---

Built with ❤️ and 100% test coverage