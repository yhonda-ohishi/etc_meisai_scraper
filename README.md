# ETC明細ダウンロードサービス

ETCの利用明細をWebスクレイピングで自動取得するGoモジュールです。

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

### サービスとして実行

このモジュールは主に他のサービスから呼び出されるライブラリとして設計されています。

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

## 🔒 セキュリティ

- パスワードは環境変数で管理
- Headlessモードでの実行推奨
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

## 📝 ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 📧 お問い合わせ

問題や質問がある場合は、[Issues](https://github.com/yhonda-ohishi/etc_meisai_scraper/issues)でお知らせください。

---

Built with ❤️ and 100% test coverage