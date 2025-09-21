# etc_meisai統合指示書

## 概要
etc_meisaiをserver_repoに統合し、Swagger UIに自動的に表示されるようにする

## 現状の問題点
1. etc_meisaiは`go-chi`ルーターを使用しているが、server_repoは`Fiber`を使用している
2. etc_meisaiにはprotoファイルが存在しない（gRPCサービス定義がない）
3. etc_meisaiのエンドポイントがSwagger UIに表示されない

## 解決方法

### proto定義の追加によるgRPCサービス化
db_serviceと同じアーキテクチャにすることで一貫性を保ち、Swagger自動生成を実現

#### 必要な変更:

1. **protoファイルの作成** (`src/proto/etc_meisai.proto`)
```protobuf
syntax = "proto3";

package etc_meisai;

import "google/api/annotations.proto";

option go_package = "github.com/yhonda-ohishi/etc_meisai/src/proto";

// ETCMeisaiService - ETC明細管理サービス
service ETCMeisaiService {
  // ETC明細データ作成
  rpc Create(CreateETCMeisaiRequest) returns (ETCMeisaiResponse) {
    option (google.api.http) = {
      post: "/api/v1/etc-meisai/records"
      body: "etc_meisai"
    };
  }

  // ETC明細データ取得
  rpc Get(GetETCMeisaiRequest) returns (ETCMeisaiResponse) {
    option (google.api.http) = {
      get: "/api/v1/etc-meisai/records/{id}"
    };
  }

  // ETC明細データ一覧取得
  rpc List(ListETCMeisaiRequest) returns (ListETCMeisaiResponse) {
    option (google.api.http) = {
      get: "/api/v1/etc-meisai/records"
    };
  }

  // CSVインポート
  rpc ImportCSV(ImportCSVRequest) returns (ImportCSVResponse) {
    option (google.api.http) = {
      post: "/api/v1/etc-meisai/import"
      body: "*"
    };
  }

  // マッピング作成
  rpc CreateMapping(CreateMappingRequest) returns (MappingResponse) {
    option (google.api.http) = {
      post: "/api/v1/etc-meisai/mappings"
      body: "mapping"
    };
  }
}

// メッセージ定義
message ETCMeisaiRecord {
  int64 id = 1;
  string hash = 2;
  string date = 3;
  string time = 4;
  string entrance_ic = 5;
  string exit_ic = 6;
  int32 toll_amount = 7;
  string car_number = 8;
  string etc_card_number = 9;
}

// リクエスト/レスポンス定義...
```

2. **buf.gen.yamlの作成**
```yaml
version: v1
plugins:
  - plugin: go
    out: src/pb
    opt: paths=source_relative
  - plugin: go-grpc
    out: src/pb
    opt: paths=source_relative
  - plugin: grpc-gateway
    out: src/pb
    opt: paths=source_relative
  - plugin: openapiv2
    out: swagger
    opt:
      - logtostderr=true
      - allow_merge=true
      - merge_file_name=etc_meisai
```

3. **gRPCサーバーの実装**
```go
// src/grpc/server.go
package grpc

import (
    "context"
    pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
    "github.com/yhonda-ohishi/etc_meisai/src/services"
)

type ETCMeisaiServer struct {
    pb.UnimplementedETCMeisaiServiceServer
    etcService     *services.ETCService
    mappingService *services.MappingService
    importService  *services.ImportService
}

func NewETCMeisaiServer(services ...) *ETCMeisaiServer {
    // 実装
}

// gRPCメソッドの実装
func (s *ETCMeisaiServer) Create(ctx context.Context, req *pb.CreateETCMeisaiRequest) (*pb.ETCMeisaiResponse, error) {
    // 既存のETCServiceを呼び出し
}
```

4. **server_repoでの統合**
- ServiceRegistryにETCMeisaiServerを追加
- bufconn経由で登録

## 利点
1. **一貫性**: db_serviceと同じアーキテクチャで統一性が保てる
2. **自動生成**: SwaggerがProtobufから自動生成される
3. **型安全性**: gRPCによる型チェック
4. **将来性**: マイクロサービス化が容易

### 実装手順：
1. protoファイルを作成
2. `buf generate`でコード生成
3. 既存サービスをgRPCサーバーでラップ
4. server_repoのServiceRegistryに登録
5. テスト実施

## テスト項目
- [ ] gRPCサービスが正しく登録される
- [ ] REST APIエンドポイントが動作する
- [ ] Swagger UIにエンドポイントが表示される
- [ ] CSVインポート機能が動作する
- [ ] マッピング機能が動作する

## 注意事項
- 既存のビジネスロジックは変更しない
- go-chiの依存は段階的に削除可能
- データベース接続はdb_service経由で行う

## 優先度
高 - Swagger統合により、API仕様が明確になり、開発効率が向上する