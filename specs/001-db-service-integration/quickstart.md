# Quickstart Guide: etc_meisai Server Repository Integration

**Branch**: `001-db-service-integration`
**Date**: 2025-09-21

## 概要

このクイックスタートガイドでは、統合されたetc_meisaiサービスの基本的な使用方法とテスト手順を説明します。

## 前提条件

- Go 1.21+ がインストールされていること
- Protocol Buffers コンパイラ (protoc) がインストールされていること
- buf CLI がインストールされていること
- Docker/Docker Compose（オプション：ローカルテスト用）

## セットアップ

### 1. 依存関係のインストール

```bash
# Go依存関係
go mod download

# Protocol Buffers プラグイン
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# buf CLI
curl -sSL https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Windows-x86_64.exe -o buf.exe
```

### 2. Protocol Buffersのコンパイル

```bash
# protoファイルからコード生成
buf generate

# 生成されたファイルの確認
ls src/pb/
```

### 3. 環境変数の設定

```bash
# .envファイルを作成
cat > .env << EOF
DATABASE_URL=mysql://user:password@localhost:3306/etc_meisai
GRPC_SERVER_PORT=50051
HTTP_SERVER_PORT=8080
ETC_CORPORATE_ACCOUNTS=account1,account2
ETC_PERSONAL_ACCOUNTS=personal1,personal2
EOF
```

## 基本的な使用方法

### 1. サーバーの起動

```bash
# gRPCサーバーの起動
go run cmd/server/main.go

# 別のターミナルでHTTPゲートウェイの起動
go run cmd/gateway/main.go
```

### 2. Swagger UIへのアクセス

ブラウザで以下のURLにアクセス：
```
http://localhost:8080/swagger-ui/
```

### 3. 基本的なAPI操作

#### ETC明細レコードの作成

```bash
# HTTPリクエスト
curl -X POST http://localhost:8080/api/v1/etc-meisai/records \
  -H "Content-Type: application/json" \
  -d '{
    "record": {
      "hash": "abc123def456",
      "date": "2025-09-21",
      "time": "10:30:00",
      "entrance_ic": "東京IC",
      "exit_ic": "横浜IC",
      "toll_amount": 1200,
      "car_number": "品川 300 あ 1234",
      "etc_card_number": "1234567890123456"
    }
  }'

# gRPCリクエスト (grpcurlを使用)
grpcurl -plaintext \
  -d '{
    "record": {
      "hash": "abc123def456",
      "date": "2025-09-21",
      "time": "10:30:00",
      "entrance_ic": "東京IC",
      "exit_ic": "横浜IC",
      "toll_amount": 1200,
      "car_number": "品川 300 あ 1234",
      "etc_card_number": "1234567890123456"
    }
  }' \
  localhost:50051 \
  etc_meisai.v1.ETCMeisaiService/CreateRecord
```

#### ETC明細レコードの取得

```bash
# HTTPリクエスト
curl http://localhost:8080/api/v1/etc-meisai/records/1

# gRPCリクエスト
grpcurl -plaintext \
  -d '{"id": 1}' \
  localhost:50051 \
  etc_meisai.v1.ETCMeisaiService/GetRecord
```

#### ETC明細レコードの一覧取得

```bash
# HTTPリクエスト（ページネーション付き）
curl "http://localhost:8080/api/v1/etc-meisai/records?page=1&page_size=10&date_from=2025-09-01&date_to=2025-09-30"

# gRPCリクエスト
grpcurl -plaintext \
  -d '{
    "page": 1,
    "page_size": 10,
    "date_from": "2025-09-01",
    "date_to": "2025-09-30"
  }' \
  localhost:50051 \
  etc_meisai.v1.ETCMeisaiService/ListRecords
```

### 4. CSVインポート

#### 単一ファイルインポート

```bash
# CSVファイルの準備
cat > test_data.csv << EOF
利用日,利用時刻,入口IC,出口IC,通行料金,車両番号,ETCカード番号
2025-09-21,10:30:00,東京IC,横浜IC,1200,品川 300 あ 1234,1234567890123456
2025-09-21,14:15:00,横浜IC,静岡IC,2500,品川 300 あ 1234,1234567890123456
EOF

# HTTPリクエストでインポート
curl -X POST http://localhost:8080/api/v1/etc-meisai/import \
  -H "Content-Type: multipart/form-data" \
  -F "account_type=corporate" \
  -F "account_id=account1" \
  -F "file=@test_data.csv"
```

#### ストリーミングインポート（大容量ファイル用）

```go
// Go クライアントコード例
package main

import (
    "context"
    "io"
    "os"
    pb "github.com/yhonda-ohishi/etc_meisai_scraper/src/pb"
    "google.golang.org/grpc"
)

func streamImportCSV(client pb.ETCMeisaiServiceClient, filePath string) error {
    stream, err := client.ImportCSVStream(context.Background())
    if err != nil {
        return err
    }

    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    buffer := make([]byte, 4096)
    chunkNumber := 1

    for {
        n, err := file.Read(buffer)
        if err == io.EOF {
            // 最後のチャンク送信
            stream.Send(&pb.ImportCSVChunk{
                SessionId:    "session-123",
                Data:        buffer[:n],
                IsLast:      true,
                ChunkNumber: int32(chunkNumber),
            })
            break
        }

        // チャンク送信
        stream.Send(&pb.ImportCSVChunk{
            SessionId:    "session-123",
            Data:        buffer[:n],
            IsLast:      false,
            ChunkNumber: int32(chunkNumber),
        })
        chunkNumber++
    }

    // 進捗受信
    for {
        progress, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        fmt.Printf("Progress: %.2f%% (Processed: %d, Success: %d, Errors: %d)\n",
            progress.ProgressPercentage,
            progress.ProcessedRows,
            progress.SuccessRows,
            progress.ErrorRows)
    }

    return stream.CloseSend()
}
```

### 5. マッピング機能

#### マッピングの作成

```bash
# HTTPリクエスト
curl -X POST http://localhost:8080/api/v1/etc-meisai/mappings \
  -H "Content-Type: application/json" \
  -d '{
    "mapping": {
      "etc_record_id": 1,
      "mapping_type": "dtako",
      "mapped_entity_id": 100,
      "mapped_entity_type": "dtako_record",
      "confidence": 0.95,
      "status": "MAPPING_STATUS_ACTIVE"
    }
  }'
```

#### マッピングの一覧取得

```bash
# HTTPリクエスト
curl "http://localhost:8080/api/v1/etc-meisai/mappings?etc_record_id=1&status=MAPPING_STATUS_ACTIVE"
```

## テスト実行

### 1. 単体テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付きテスト
go test -cover ./...

# 特定のパッケージのテスト
go test ./src/services/...
```

### 2. 統合テスト

```bash
# 統合テスト環境の起動
docker-compose -f docker-compose.test.yml up -d

# 統合テスト実行
go test ./tests/integration/... -tags=integration

# テスト環境のクリーンアップ
docker-compose -f docker-compose.test.yml down
```

### 3. 契約テスト

```bash
# Protocol Buffers契約テスト
go test ./tests/contract/... -tags=contract

# APIレスポンス検証
go test ./tests/contract/api_test.go -v
```

### 4. パフォーマンステスト

```bash
# ベンチマークテスト
go test -bench=. ./tests/benchmark/...

# 負荷テスト（vegeta使用）
echo "GET http://localhost:8080/api/v1/etc-meisai/records" | \
  vegeta attack -rate=100 -duration=30s | \
  vegeta report
```

## トラブルシューティング

### 1. gRPCサーバーが起動しない

```bash
# ポートが使用されていないか確認
netstat -an | grep 50051

# プロセスを強制終了
taskkill /F /PID <process_id>  # Windows
kill -9 <process_id>            # Linux/Mac
```

### 2. Protocol Buffersコンパイルエラー

```bash
# bufの設定確認
buf lint

# 破壊的変更の確認
buf breaking --against '.git#branch=main'

# 依存関係の更新
buf mod update
```

### 3. データベース接続エラー

```bash
# 接続テスト
go run cmd/dbtest/main.go

# マイグレーション実行
go run cmd/migrate/main.go up

# データベースリセット
go run cmd/migrate/main.go reset
```

### 4. CSVインポート失敗

```bash
# ログ確認
tail -f logs/import.log

# インポートセッション状態確認
curl http://localhost:8080/api/v1/etc-meisai/import-sessions/<session_id>

# エラーログ詳細
curl http://localhost:8080/api/v1/etc-meisai/import-sessions/<session_id> | jq '.session.error_log'
```

## 検証シナリオ

### シナリオ1: 基本的なCRUD操作

1. ✅ ETC明細レコードを作成する
2. ✅ 作成したレコードを取得する
3. ✅ レコードを更新する
4. ✅ レコード一覧を取得する（フィルタ付き）
5. ✅ レコードを削除する

### シナリオ2: CSVインポート

1. ✅ 小規模CSVファイル（100行）をインポート
2. ✅ 大規模CSVファイル（10000行）をストリーミングインポート
3. ✅ 重複データの処理確認
4. ✅ エラー行の処理確認
5. ✅ インポートセッションの状態確認

### シナリオ3: マッピング機能

1. ✅ ETCレコードとデジタコデータのマッピング作成
2. ✅ マッピング信頼度の設定と更新
3. ✅ マッピング状態の遷移（pending → active）
4. ✅ 複数マッピングの管理
5. ✅ マッピングの削除とカスケード処理

### シナリオ4: Swagger UI統合

1. ✅ Swagger UIにアクセスできる
2. ✅ 全エンドポイントが表示される
3. ✅ Try it outで各APIを実行できる
4. ✅ レスポンススキーマが正しく表示される
5. ✅ 認証ヘッダーが正しく設定される

### シナリオ5: エラーハンドリング

1. ✅ 存在しないレコードへのアクセス（404）
2. ✅ 不正なリクエストボディ（400）
3. ✅ 認証エラー（401）
4. ✅ 権限エラー（403）
5. ✅ サーバーエラー（500）の適切な処理

## 次のステップ

1. **本番環境へのデプロイ**
   - Kubernetes manifestsの作成
   - CI/CDパイプラインの設定
   - 監視・アラートの設定

2. **セキュリティ強化**
   - TLS/SSL証明書の設定
   - API rate limiting
   - 監査ログの実装

3. **パフォーマンス最適化**
   - インデックスの最適化
   - キャッシュレイヤーの追加
   - CDNの設定

4. **機能拡張**
   - リアルタイムストリーミング
   - WebSocket通知
   - バッチ処理ジョブ

---
*Quickstart Guide v1.0 - 2025-09-21*