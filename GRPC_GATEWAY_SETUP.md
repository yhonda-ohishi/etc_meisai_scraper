# gRPC-Gateway セットアップ指示書

## 概要
このドキュメントは、`etc_meisai_scraper` の gRPC サービスを REST API として公開し、Swagger UI で自動的にドキュメント化するための設定手順です。

## 必要な変更

### 1. proto ファイルの更新

#### download.proto の修正
`src/proto/download.proto` を以下のように更新してください：

```proto
syntax = "proto3";

package etc_meisai.download.v1;

option go_package = "github.com/yhonda-ohishi/etc_meisai_scraper/src/pb";

import "google/protobuf/timestamp.proto";
import "google/api/annotations.proto";
import "google/api/field_behavior.proto";
import "protoc-gen-openapiv2/options/annotations.proto";

// Swagger/OpenAPI の設定
option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: {
    title: "ETC Meisai Download API";
    version: "1.0";
    description: "ETC明細のダウンロードと管理を行うAPI";
  };
  schemes: [HTTP, HTTPS];
  consumes: "application/json";
  produces: "application/json";
  responses: {
    key: "default";
    value: {
      description: "エラーレスポンス";
      schema: {
        json_schema: {
          ref: "#/definitions/rpcStatus";
        };
      };
    };
  };
};

// ダウンロードサービス
service DownloadService {
  // 同期ダウンロード
  rpc DownloadSync(DownloadRequest) returns (DownloadResponse) {
    option (google.api.http) = {
      post: "/api/v1/download/sync"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "同期ダウンロード";
      description: "ETC明細データを同期的にダウンロードします";
      tags: "Download";
    };
  }

  // 非同期ダウンロード開始
  rpc DownloadAsync(DownloadRequest) returns (DownloadJobResponse) {
    option (google.api.http) = {
      post: "/api/v1/download/async"
      body: "*"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "非同期ダウンロード開始";
      description: "ETC明細データの非同期ダウンロードジョブを開始します";
      tags: "Download";
    };
  }

  // ジョブステータス取得
  rpc GetJobStatus(GetJobStatusRequest) returns (JobStatus) {
    option (google.api.http) = {
      get: "/api/v1/download/jobs/{job_id}"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "ジョブステータス取得";
      description: "非同期ダウンロードジョブのステータスを取得します";
      tags: "Download";
    };
  }

  // 全アカウントID取得
  rpc GetAllAccountIDs(GetAllAccountIDsRequest) returns (GetAllAccountIDsResponse) {
    option (google.api.http) = {
      get: "/api/v1/accounts"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      summary: "全アカウントID取得";
      description: "登録されているすべてのETCアカウントIDを取得します";
      tags: "Account";
    };
  }
}

// メッセージ定義は既存のものをそのまま使用
// ... (既存のメッセージ定義)
```

### 2. buf.gen.yaml の更新

`buf.gen.yaml` に以下のプラグインを追加してください：

```yaml
version: v1
plugins:
  # 既存のプラグイン
  - plugin: go
    out: src/pb
    opt:
      - paths=source_relative

  - plugin: go-grpc
    out: src/pb
    opt:
      - paths=source_relative

  # 新規追加: grpc-gateway プラグイン
  - plugin: grpc-gateway
    out: src/pb
    opt:
      - paths=source_relative
      - generate_unbound_methods=true
      - allow_repeated_fields_in_body=true

  # 新規追加: OpenAPI v2 生成
  - plugin: openapiv2
    out: swagger
    opt:
      - allow_repeated_fields_in_body=true
      - generate_unbound_methods=true
```

### 3. 必要な依存関係の追加

`go.mod` に以下を追加：

```go
require (
    // 既存の依存関係...
    github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2
    google.golang.org/genproto/googleapis/api v0.0.0-20250908214217-97024824d090
)
```

### 4. proto 依存ファイルのダウンロード

以下のコマンドを実行して、必要な proto ファイルをダウンロード：

```bash
# buf.yaml がある場合
buf mod update

# または手動でダウンロード
mkdir -p third_party/googleapis
git clone https://github.com/googleapis/googleapis.git third_party/googleapis
```

### 5. コード生成

```bash
# buf を使用する場合
buf generate

# または protoc を直接使用する場合
protoc -I src/proto \
  -I third_party/googleapis \
  --go_out=src/pb --go_opt=paths=source_relative \
  --go-grpc_out=src/pb --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=src/pb --grpc-gateway_opt=paths=source_relative \
  --openapiv2_out=swagger \
  src/proto/download.proto
```

### 6. 生成されるファイル

以下のファイルが生成されます：
- `src/pb/download.pb.go` - Protocol Buffers のGoコード（既存）
- `src/pb/download_grpc.pb.go` - gRPC サーバー/クライアントコード（既存）
- `src/pb/download.pb.gw.go` - **新規: grpc-gateway コード**
- `swagger/download.swagger.json` - **新規: OpenAPI/Swagger 定義**

## 統合確認

1. コード生成が成功したら、`server_repo` を再ビルド
2. サーバーを起動
3. Swagger UI (http://localhost:8080/docs) でAPIドキュメントを確認
4. REST エンドポイントが利用可能になる：
   - POST `/api/v1/download/sync`
   - POST `/api/v1/download/async`
   - GET `/api/v1/download/jobs/{job_id}`
   - GET `/api/v1/accounts`

## トラブルシューティング

### proto ファイルのインポートエラー
- `google/api/annotations.proto` が見つからない場合は、googleapis をダウンロードしてインクルードパスに追加

### コード生成エラー
- buf または protoc のバージョンを最新に更新
- プラグインがインストールされているか確認：
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
  ```

## 参考資料
- [gRPC-Gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Protocol Buffers Style Guide](https://developers.google.com/protocol-buffers/docs/style)
- [OpenAPI Specification](https://swagger.io/specification/)