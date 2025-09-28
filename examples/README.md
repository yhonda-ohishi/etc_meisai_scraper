# ETC明細スクレイパー サンプルコード

このディレクトリには、ETC明細スクレイパーの使用例が含まれています。

## ディレクトリ構造

```
examples/
├── grpc_client/     # gRPCクライアントサンプル
│   └── main.go
└── scraper_usage/   # 直接スクレイパー使用サンプル
    └── main.go
```

## ビルドと実行

### gRPCクライアント (grpc_client)

gRPC経由でDownloadBufferServiceを使用する例です。

```bash
# ディレクトリに移動
cd examples/grpc_client

# ビルド
go build

# 実行（gRPCサーバーが起動している必要があります）
./grpc_client
```

**機能:**
- CSVデータを直接バイナリで取得
- ストリーミングでチャンク受信
- 構造化データ（Protocol Buffers形式）として取得

### スクレイパー直接使用 (scraper_usage)

ETCScraperを直接使用してCSVデータを処理する例です。

```bash
# ディレクトリに移動
cd examples/scraper_usage

# ビルド
go build

# 実行
./scraper_usage
```

**機能:**
- バッファとして直接取得
- io.Readerとしてストリーミング風に処理
- CSVデータをパースして構造化データとして処理
- メモリ内でのデータ変換例

## 必要な設定

### grpc_client
- gRPCサーバーが`localhost:50051`で起動している必要があります
- サーバー側で`DownloadBufferService`が実装されている必要があります

### scraper_usage
- 有効なETC明細サイトのログイン情報が必要です
- コード内の以下の部分を実際の認証情報に変更してください：
  ```go
  config := &scraper.ScraperConfig{
      UserID:       "your-user-id",
      Password:     "your-password",
      DownloadPath: "./temp",
      Headless:     true,
  }
  ```

## トラブルシューティング

### ビルドエラーが発生する場合
```bash
# 依存関係を更新
go mod tidy

# モジュールキャッシュをクリア
go clean -modcache
```

### gRPCクライアントが接続できない場合
1. gRPCサーバーが起動しているか確認
2. ポート番号（50051）が正しいか確認
3. ファイアウォールでポートがブロックされていないか確認