# テストディレクトリ構造

## ディレクトリ構成

```
tests/
├── unit/           # ユニットテスト
│   ├── services/   # サービス層のテスト
│   └── handlers/   # ハンドラー層のテスト
├── integration/    # 統合テスト
├── e2e/           # エンドツーエンドテスト
└── examples/       # サンプルコード・動作確認用

```

## テスト実行方法

### すべてのテストを実行
```bash
go test ./tests/...
```

### ユニットテストのみ実行
```bash
go test ./tests/unit/...
```

### 統合テストのみ実行
```bash
go test ./tests/integration/...
```

### カバレッジ付きでテスト実行
```bash
go test -cover ./tests/...
```

### 詳細出力付きでテスト実行
```bash
go test -v ./tests/...
```

## サンプルコード実行

### ダウンロード機能のテスト
```bash
# 環境変数設定
export ETC_USER_ID="your_user_id"
export ETC_PASSWORD="your_password"

# 実行
go run tests/examples/test_download_example.go
```

## テストファイル一覧

### Unit Tests
- `tests/unit/services/download_service_test.go` - ダウンロードサービスのユニットテスト
- `tests/unit/handlers/download_handler_test.go` - HTTPハンドラーのユニットテスト

### Integration Tests
- `tests/integration/grpc_server_test.go` - gRPCサーバーの統合テスト

### Examples
- `tests/examples/test_download_example.go` - スクレイピング機能の動作確認用サンプル