# Research Document: etc_meisai Server Repository Integration

**Branch**: `001-db-service-integration`
**Date**: 2025-09-21
**Status**: Complete

## 技術調査概要

etc_meisaiをserver_repoに統合するための技術調査結果をまとめる。主要な技術選択、アーキテクチャパターン、移行戦略について検証済みの決定事項を記録する。

## 1. Protocol Buffers Service定義

### Decision: Protocol Buffersファースト設計
### Rationale:
- Swagger/OpenAPIドキュメントの自動生成が可能
- 強力な型安全性とコンパイル時チェック
- db_serviceと統一されたアーキテクチャ
- バージョン管理と後方互換性の保証

### Alternatives considered:
- **RESTのまま手動Swagger定義**: メンテナンス負荷が高い、型安全性が弱い
- **GraphQL**: 過度な複雑性、既存システムとの不整合
- **JSON-RPC**: エコシステムが限定的、ツール支援が少ない

### ベストプラクティス:
```protobuf
// パッケージ命名規則
package etc_meisai.v1;

// import管理
import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "validate/validate.proto";

// フィールド番号予約
reserved 10 to 15; // 将来の拡張用

// 明確なメッセージ命名
message ETCMeisaiRecord {
  int64 id = 1 [(validate.rules).int64.gt = 0];
  string hash = 2 [(validate.rules).string.min_len = 1];
  // ...
}
```

## 2. grpc-gateway Configuration

### Decision: grpc-gateway v2採用
### Rationale:
- OpenAPI v3サポート
- カスタムHTTPルーティング
- JSON/Protocol Buffers自動変換
- Swaggerドキュメント自動生成

### 設定最適化:
```yaml
# buf.gen.yaml
version: v1
plugins:
  - plugin: go
    out: src/pb
    opt: paths=source_relative
  - plugin: go-grpc
    out: src/pb
    opt:
      - paths=source_relative
      - require_unimplemented_servers=false
  - plugin: grpc-gateway
    out: src/pb
    opt:
      - paths=source_relative
      - generate_unbound_methods=true
  - plugin: openapiv2
    out: swagger
    opt:
      - allow_merge=true
      - merge_file_name=etc_meisai
      - json_names_for_fields=false
```

### HTTP Mapping戦略:
```protobuf
service ETCMeisaiService {
  rpc ListRecords(ListRecordsRequest) returns (ListRecordsResponse) {
    option (google.api.http) = {
      get: "/api/v1/etc-meisai/records"
      additional_bindings {
        get: "/api/v1/etc-meisai/records/search"
      }
    };
  }
}
```

## 3. go-chi → gRPC移行戦略

### Decision: 段階的移行アプローチ
### Rationale:
- サービス中断を最小化
- 既存クライアントの後方互換性維持
- 段階的なテストとバリデーション
- ロールバック可能性の確保

### 移行フェーズ:
1. **Phase 1: Dual Stack**
   - gRPCサーバーを並行実装
   - 既存go-chiルーターを維持
   - リクエストを内部でgRPCに転送

2. **Phase 2: Primary gRPC**
   - gRPCをプライマリとして設定
   - go-chiをレガシーラッパーに
   - メトリクスで使用状況監視

3. **Phase 3: Deprecation**
   - go-chi依存の段階的削除
   - クライアント移行完了確認
   - 完全gRPC移行

### 互換性アダプター設計:
```go
// adapters/chi_to_grpc.go
type ChiToGRPCAdapter struct {
    grpcClient pb.ETCMeisaiServiceClient
}

func (a *ChiToGRPCAdapter) HandleGetRecord(w http.ResponseWriter, r *http.Request) {
    // chi request → gRPC request変換
    req := &pb.GetRecordRequest{
        Id: chi.URLParam(r, "id"),
    }

    // gRPC呼び出し
    resp, err := a.grpcClient.GetRecord(r.Context(), req)

    // gRPC response → HTTP response変換
    json.NewEncoder(w).Encode(resp)
}
```

## 4. buf Compiler最適化

### Decision: buf CLI v1採用
### Rationale:
- 高速なコンパイル
- リント・フォーマット機能内蔵
- 破壊的変更の検出
- CI/CD統合容易

### 設定最適化:
```yaml
# buf.yaml
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
  except:
    - ENUM_VALUE_PREFIX
    - PACKAGE_VERSION_SUFFIX
  enum_zero_value_suffix: _UNSPECIFIED
  rpc_allow_same_request_response: false
  service_suffix: Service
```

### ビルド最適化:
```makefile
# Makefile
.PHONY: proto
proto:
	buf generate
	buf lint
	buf breaking --against '.git#branch=main'
```

## 5. データモデル統合

### Decision: GORM統一モデル
### Rationale:
- db_serviceとの完全互換性
- マイグレーション自動化
- 関連マッピング強化
- トランザクション管理統一

### モデル変換戦略:
```go
// models/converter.go
func ETCMeisaiRecordToProto(m *models.ETCMeisaiRecord) *pb.ETCMeisaiRecord {
    return &pb.ETCMeisaiRecord{
        Id:           m.ID,
        Hash:         m.Hash,
        Date:         m.Date.Format("2006-01-02"),
        Time:         m.Time,
        EntranceIc:   m.EntranceIC,
        ExitIc:       m.ExitIC,
        TollAmount:   int32(m.TollAmount),
        CarNumber:    m.CarNumber,
        EtcCardNumber: m.ETCCardNumber,
    }
}
```

## 6. エラーハンドリング統一

### Decision: gRPC Status + 構造化エラー
### Rationale:
- 標準化されたエラーコード
- クライアント側での適切な処理
- デバッグ情報の豊富さ
- 国際化対応

### エラー設計:
```go
// errors/grpc_errors.go
func NotFoundError(resource string, id string) error {
    return status.Errorf(
        codes.NotFound,
        "リソース %s (ID: %s) が見つかりません",
        resource, id,
    )
}

func ValidationError(field string, reason string) error {
    return status.Errorf(
        codes.InvalidArgument,
        "バリデーションエラー: %s - %s",
        field, reason,
    )
}
```

## 7. 認証・認可統合

### Decision: JWT + gRPC Interceptor
### Rationale:
- server_repoの既存認証システム活用
- ステートレスな認証
- マイクロサービス間の認証統一
- スケーラビリティ

### Interceptor実装:
```go
// interceptors/auth.go
func UnaryAuthInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    // メタデータからトークン取得
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "認証情報がありません")
    }

    // JWT検証
    token := md.Get("authorization")
    claims, err := validateJWT(token[0])
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "無効なトークン")
    }

    // コンテキストに認証情報を追加
    ctx = context.WithValue(ctx, "user", claims)
    return handler(ctx, req)
}
```

## 8. パフォーマンス最適化

### Decision: Connection Pooling + Streaming
### Rationale:
- 大量CSVインポートの効率化
- メモリ使用量の最適化
- レスポンス時間の改善
- 同時接続数の制御

### ストリーミング実装:
```protobuf
service ETCMeisaiService {
  rpc ImportCSVStream(stream ImportCSVChunk) returns (stream ImportProgress) {}
}
```

### Connection Pool設定:
```go
// clients/grpc_pool.go
var connPool = &grpc.ClientConnPool{
    MaxIdle:     10,
    MaxActive:   50,
    IdleTimeout: 5 * time.Minute,
}
```

## 9. 監視・ロギング

### Decision: OpenTelemetry + 構造化ログ
### Rationale:
- 分散トレーシング対応
- メトリクス自動収集
- 既存の監視システムとの統合
- 問題の迅速な特定

### 実装:
```go
// telemetry/setup.go
func SetupTelemetry() {
    // Tracer
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource),
    )

    // Metrics
    mp := metric.NewMeterProvider(
        metric.WithReader(reader),
    )

    // Logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
```

## 10. テスト戦略

### Decision: 契約テスト + 統合テスト重視
### Rationale:
- API契約の保証
- サービス間統合の検証
- リグレッション防止
- CI/CDパイプライン統合

### テスト構成:
```
tests/
├── contract/           # Protocol Buffers契約テスト
│   ├── grpc/          # gRPCサービステスト
│   └── rest/          # REST APIテスト
├── integration/       # 統合テスト
│   ├── db/           # データベース統合
│   ├── csv/          # CSVインポート
│   └── mapping/      # マッピング機能
└── benchmark/         # パフォーマンステスト
```

## まとめ

### 主要な技術決定:
1. **Protocol Buffers** - API定義の単一情報源
2. **grpc-gateway v2** - REST/gRPCブリッジ
3. **段階的移行** - リスク最小化
4. **buf CLI** - 効率的なprotobuf管理
5. **GORM統一** - データ層の一貫性
6. **gRPC Status** - エラーハンドリング標準化
7. **JWT Interceptor** - 認証統合
8. **Streaming API** - 大量データ処理
9. **OpenTelemetry** - 可観測性強化
10. **契約テスト** - API互換性保証

### 次のステップ:
- Phase 1: data-model.md作成
- Phase 1: contracts/作成
- Phase 1: quickstart.md作成
- Phase 1: CLAUDE.md更新

---
*Research completed: 2025-09-21*