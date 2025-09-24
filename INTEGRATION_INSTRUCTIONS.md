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

---

# Phase 4 - T010: Contract Testing Expansion - Implementation Complete

## Overview

I have successfully implemented comprehensive contract testing for the ETC明細 gRPC service, completing all requirements for Phase 4 - T010: Contract Testing Expansion.

## Implementation Summary

### ✅ T010-A: Contract Testing for All gRPC Service Definitions

**File**: `tests/contract/grpc_service_contract_test.go`

Implemented comprehensive contract tests for all gRPC service methods defined in `specs/001-db-service-integration/contracts/etc_meisai.proto`:

- **ETCMeisaiRecord Operations**: CreateRecord, GetRecord, ListRecords, UpdateRecord, DeleteRecord
- **Import Operations**: ImportCSV, ImportCSVStream, GetImportSession, ListImportSessions
- **Mapping Operations**: CreateMapping, GetMapping, ListMappings, UpdateMapping, DeleteMapping
- **Statistics Operations**: GetStatistics
- **Error Handling**: Comprehensive error scenario testing

### ✅ T010-B: API Version Compatibility Testing

**File**: `tests/contract/api_version_compatibility_test.go`

Implemented version compatibility testing ensuring backward and forward compatibility:

- **Version Negotiation**: Client-server version header handling
- **Backward Compatibility**: Support for requests with only required fields
- **Forward Compatibility**: Graceful handling of newer optional fields
- **Enum Compatibility**: Proper handling of known and unknown enum values
- **Message Evolution**: Field addition/removal compatibility
- **Client-Server Negotiation**: Feature capability negotiation

### ✅ T010-C: Schema Evolution Testing for Protocol Buffer Changes

**File**: `tests/contract/schema_evolution_test.go`

Implemented comprehensive Protocol Buffer schema evolution testing:

- **Field Evolution**: Adding optional fields, field number stability
- **Enum Evolution**: New enum values, unknown value handling
- **Message Structure**: Nested messages, repeated fields, oneof support
- **Serialization Compatibility**: Binary format stability across versions
- **Default Values**: Consistent default value handling

### ✅ T010-D: End-to-End Workflow Testing for Complete ETC Data Processing

**File**: `tests/contract/end_to_end_workflow_test.go`

Implemented complete workflow testing covering the entire ETC data processing pipeline:

- **CSV Import Workflow**: File upload → processing → verification
- **Data Mapping Workflow**: ETC-DTako record mapping and validation
- **Statistics Generation**: Data aggregation and reporting
- **Error Handling**: Robust error scenarios throughout workflow
- **Data Consistency**: Referential integrity and concurrent operation safety
- **Concurrent Operations**: Multi-user scenarios and data isolation

### ✅ T010-E: Performance Contract Testing with SLA Validation

**File**: `tests/contract/performance_sla_test.go`

Implemented comprehensive performance SLA validation with specific response time requirements:

#### Response Time SLAs
- **CreateRecord**: < 100ms
- **GetRecord**: < 100ms
- **ListRecords**: < 200ms
- **UpdateRecord**: < 100ms
- **ImportCSV**: < 1s (small files)
- **GetStatistics**: < 500ms

#### Throughput SLAs
- **CreateRecord**: ≥ 50 operations/second
- **GetRecord**: ≥ 100 operations/second
- **Concurrent Users**: ≥ 10 users
- **Mixed Operations**: ≥ 50 concurrent operations

#### Additional Performance Testing
- **Concurrency**: Multi-user and mixed operation scenarios
- **Resource Usage**: Large dataset and complex query handling
- **Load Testing**: Stress testing under various load conditions

## Supporting Infrastructure

### Test Suite Framework

**File**: `tests/contract/contract_test_suite.go`

- Centralized test infrastructure setup
- gRPC client/server configuration
- Common test utilities and helpers
- Environment setup and teardown

### Legacy Test Integration

**Files**: `tests/contract/mock_generation_test.go`, `tests/contract/test_execution_test.go`

- Re-enabled existing contract tests
- Integrated with new test framework
- Maintains backward compatibility

### Test Runner

**File**: `tests/contract-runner/run_contract_tests.go`

- Automated test execution
- Report generation
- Performance metrics collection
- CI/CD integration support

### Documentation

**File**: `tests/contract/README.md`

- Comprehensive test documentation
- Usage instructions and examples
- SLA requirements and best practices
- Troubleshooting guide

## Usage Instructions

### Prerequisites

1. gRPC server must be running on configured port
2. Test database must be available (SQLite or MySQL)
3. Required environment variables must be set

### Running Tests

```bash
# Run all contract tests
cd tests/contract
go test -v ./...

# Run specific test categories
go test -v -run "TestGRPCServiceContract"
go test -v -run "TestAPIVersionCompatibility"
go test -v -run "TestSchemaEvolution"
go test -v -run "TestEndToEndWorkflow"
go test -v -run "TestPerformanceSLA"

# Run with performance and race detection
go test -v -race -timeout 10m

# Generate coverage reports
go test -v -cover -coverprofile=contract_coverage.out
```

### Environment Configuration

```bash
export ETC_TEST_DATABASE_URL="sqlite://contract_test.db"
export ETC_TEST_GRPC_PORT="9090"
export ETC_TEST_ENABLE_PERFORMANCE="true"
export ETC_TEST_ENABLE_E2E="true"
export ETC_TEST_DEBUG="false"
```

## Key Features

### Contract-Based Testing
- **API Stability**: Ensures gRPC service contracts remain stable
- **Backward Compatibility**: Validates that changes don't break existing clients
- **Forward Compatibility**: Ensures graceful handling of newer client features

### Performance SLA Validation
- **Response Time Monitoring**: Validates all operations meet SLA requirements
- **Throughput Testing**: Ensures minimum operations per second requirements
- **Concurrency Testing**: Validates performance under concurrent load

### Schema Evolution Support
- **Protocol Buffer Compatibility**: Ensures schema changes maintain compatibility
- **Field Evolution**: Tests field additions, modifications, and removals
- **Version Negotiation**: Tests client-server version compatibility

### End-to-End Validation
- **Complete Workflows**: Tests entire ETC data processing pipeline
- **Data Integrity**: Validates data consistency throughout operations
- **Error Scenarios**: Comprehensive error handling validation

## Integration with CI/CD

The contract tests are designed for CI/CD integration:

1. **Pull Request Validation**: Run on every code change
2. **Performance Regression**: Detect SLA violations early
3. **API Breaking Changes**: Prevent backward incompatible changes
4. **Automated Reporting**: Generate detailed test reports

## Benefits

### Development Quality
- **Early Issue Detection**: Catch contract violations during development
- **Regression Prevention**: Prevent performance and compatibility regressions
- **Documentation**: Living documentation of API contracts

### Operational Reliability
- **SLA Monitoring**: Continuous validation of performance requirements
- **Compatibility Assurance**: Ensure smooth API evolution
- **Production Readiness**: Validate system performance before deployment

### Team Productivity
- **Clear Contracts**: Well-defined API behavior expectations
- **Automated Testing**: Reduced manual testing overhead
- **Confidence**: High confidence in API stability and performance

## Next Steps

1. **Environment Setup**: Configure test environment with gRPC server and database
2. **CI/CD Integration**: Add contract tests to build pipeline
3. **Performance Baseline**: Establish performance baselines for monitoring
4. **Team Training**: Train team on contract testing methodology

## Files Modified/Created

### New Files Created
- `tests/contract/grpc_service_contract_test.go` - gRPC service contract tests
- `tests/contract/api_version_compatibility_test.go` - API compatibility tests
- `tests/contract/schema_evolution_test.go` - Protocol Buffer schema evolution tests
- `tests/contract/end_to_end_workflow_test.go` - E2E workflow tests
- `tests/contract/performance_sla_test.go` - Performance SLA validation tests
- `tests/contract/contract_test_suite.go` - Test infrastructure framework
- `tests/contract/README.md` - Comprehensive documentation
- `tests/contract-runner/run_contract_tests.go` - Test runner utility

### Existing Files Modified
- `tests/contract/mock_generation_test.go.disabled` → `tests/contract/mock_generation_test.go` (re-enabled)
- `tests/contract/test_execution_test.go.disabled` → `tests/contract/test_execution_test.go` (re-enabled)

### Proto Definitions Referenced
- `specs/001-db-service-integration/contracts/etc_meisai.proto` - Main service definition
- `src/pb/` - Generated gRPC code

The implementation is complete and ready for integration into the development workflow. All T010 requirements have been fulfilled with comprehensive testing coverage for contract stability, API compatibility, schema evolution, end-to-end workflows, and performance SLA validation.