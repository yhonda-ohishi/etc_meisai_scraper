# Full gRPC Migration Refactoring Plan

## Overview
This document outlines the comprehensive refactoring plan to migrate the entire etc_meisai codebase from the current hybrid architecture (gRPC API + manual Go interfaces/GORM models) to a fully gRPC-based architecture using Protocol Buffers throughout.

## Current Architecture Issues
1. **Mixed Paradigms**: gRPC at API layer but manual interfaces at repository layer
2. **Naming Inconsistencies**: Manual models don't match gRPC messages (ETCMapping vs ETCMapping with different fields)
3. **Maintenance Overhead**: Duplicate definitions between proto files and Go interfaces
4. **Testing Complexity**: Mock generation issues due to mixed approaches

## Target Architecture
- **All layers use gRPC**: Repository, Service, and API layers all use Protocol Buffer generated code
- **Single Source of Truth**: All interfaces and data models defined in `.proto` files
- **Consistent Naming**: Generated code ensures consistent naming across all layers
- **Simplified Testing**: Mocks generated directly from Protocol Buffer definitions

## Migration Phases

### Phase 1: Define All Interfaces in Protocol Buffers

#### 1.1 Create Repository Service Definitions
Create `src/proto/repository.proto`:
```proto
syntax = "proto3";
package etc_meisai.repository.v1;

// Define repository services for each entity
service ETCMappingRepository {
  rpc Create(CreateMappingRequest) returns (CreateMappingResponse);
  rpc GetByID(GetByIDRequest) returns (ETCMapping);
  rpc Update(UpdateMappingRequest) returns (UpdateMappingResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
  rpc List(ListMappingsRequest) returns (ListMappingsResponse);
  rpc GetByETCRecordID(GetByETCRecordIDRequest) returns (GetMappingsResponse);
  rpc GetByMappedEntity(GetByMappedEntityRequest) returns (GetMappingsResponse);
  rpc BulkCreate(BulkCreateRequest) returns (BulkCreateResponse);
  rpc UpdateStatus(UpdateStatusRequest) returns (UpdateStatusResponse);
}

service ETCMeisaiRecordRepository {
  rpc Create(CreateRecordRequest) returns (CreateRecordResponse);
  rpc GetByID(GetByIDRequest) returns (ETCMeisaiRecord);
  rpc Update(UpdateRecordRequest) returns (UpdateRecordResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
  rpc List(ListRecordsRequest) returns (ListRecordsResponse);
  rpc GetByHash(GetByHashRequest) returns (ETCMeisaiRecord);
  rpc BulkCreate(BulkCreateRecordsRequest) returns (BulkCreateResponse);
}
```

#### 1.2 Create Service Layer Definitions
Create `src/proto/services.proto`:
```proto
syntax = "proto3";
package etc_meisai.services.v1;

// Service layer interfaces
service ETCMappingService {
  rpc CreateMapping(CreateMappingRequest) returns (CreateMappingResponse);
  rpc GetMapping(GetMappingRequest) returns (GetMappingResponse);
  rpc UpdateMapping(UpdateMappingRequest) returns (UpdateMappingResponse);
  rpc DeleteMapping(DeleteMappingRequest) returns (DeleteMappingResponse);
  rpc ListMappings(ListMappingsRequest) returns (ListMappingsResponse);
  rpc ProcessBulkMappings(ProcessBulkRequest) returns (ProcessBulkResponse);
}

service ETCMeisaiService {
  rpc ImportCSV(ImportCSVRequest) returns (ImportCSVResponse);
  rpc ProcessRecords(ProcessRecordsRequest) returns (ProcessRecordsResponse);
  rpc GetStatistics(GetStatisticsRequest) returns (GetStatisticsResponse);
}
```

### Phase 2: Generate Go Code from Protocol Buffers

#### 2.1 Setup Code Generation
Create `src/proto/buf.yaml`:
```yaml
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```

Create `src/proto/buf.gen.yaml`:
```yaml
version: v1
managed:
  enabled: true
  go_package_prefix:
    default: github.com/yhonda-ohishi/etc_meisai/src/pb
plugins:
  - name: go
    out: ../pb
    opt:
      - paths=source_relative
  - name: go-grpc
    out: ../pb
    opt:
      - paths=source_relative
  - name: grpc-gateway
    out: ../pb
    opt:
      - paths=source_relative
      - generate_unbound_methods=true
  - name: openapiv2
    out: ../api
    opt:
      - allow_merge=true
      - merge_file_name=api
```

#### 2.2 Generate Code
```bash
cd src/proto
buf generate
```

### Phase 3: Replace Manual Interfaces with gRPC Services

#### 3.1 Replace Repository Interfaces
Transform `src/repositories/etc_mapping_repository.go`:
```go
// OLD: Manual interface
type ETCMappingRepository interface {
    Create(ctx context.Context, mapping *models.ETCMapping) error
    // ...
}

// NEW: Use generated gRPC client
type ETCMappingRepository struct {
    client pb.ETCMappingRepositoryClient
}

func NewETCMappingRepository(conn *grpc.ClientConn) *ETCMappingRepository {
    return &ETCMappingRepository{
        client: pb.NewETCMappingRepositoryClient(conn),
    }
}

func (r *ETCMappingRepository) Create(ctx context.Context, mapping *pb.ETCMapping) error {
    _, err := r.client.Create(ctx, &pb.CreateMappingRequest{
        Mapping: mapping,
    })
    return err
}
```

#### 3.2 Replace Service Layer
Transform `src/services/etc_mapping_service.go`:
```go
// OLD: Using manual models
func (s *ETCMappingService) CreateMapping(mapping *models.ETCMapping) error {
    // ...
}

// NEW: Using Protocol Buffer messages
func (s *ETCMappingService) CreateMapping(ctx context.Context, req *pb.CreateMappingRequest) (*pb.CreateMappingResponse, error) {
    // Implementation using gRPC
}
```

### Phase 4: Replace GORM Models with Protocol Buffer Messages

#### 4.1 Remove GORM Models
- Delete `src/models/etc_mapping.go`
- Delete `src/models/etc_meisai_record.go`
- Delete other manual model files

#### 4.2 Update Database Layer
Create adapter layer for database operations:
```go
// src/adapters/db_adapter.go
package adapters

import (
    "github.com/yhonda-ohishi/etc_meisai/src/pb"
    "gorm.io/gorm"
)

// Convert Protocol Buffer message to database-compatible format
func PBToDBMapping(mapping *pb.ETCMapping) map[string]interface{} {
    return map[string]interface{}{
        "id":                 mapping.Id,
        "etc_record_id":      mapping.EtcRecordId,
        "mapping_type":       mapping.MappingType,
        "mapped_entity_id":   mapping.MappedEntityId,
        "mapped_entity_type": mapping.MappedEntityType,
        "confidence":         mapping.Confidence,
        "status":            mapping.Status.String(),
        "created_at":        mapping.CreatedAt.AsTime(),
        "updated_at":        mapping.UpdatedAt.AsTime(),
    }
}

// Convert database result to Protocol Buffer message
func DBToPBMapping(result map[string]interface{}) *pb.ETCMapping {
    // Conversion logic
}
```

### Phase 5: Update Mock Generation

#### 5.1 Generate Mocks from gRPC Interfaces
```bash
# Generate mocks for repository clients
mockgen -source=src/pb/repository_grpc.pb.go \
        -destination=tests/mocks/mock_repository_client.go \
        -package=mocks

# Generate mocks for service clients
mockgen -source=src/pb/services_grpc.pb.go \
        -destination=tests/mocks/mock_service_client.go \
        -package=mocks
```

#### 5.2 Update Test Files
Update all test files to use new gRPC-based mocks:
```go
// tests/unit/repositories/etc_mapping_repository_test.go
func TestETCMappingRepository_Create(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockClient := mocks.NewMockETCMappingRepositoryClient(ctrl)
    repo := repositories.NewETCMappingRepository(mockClient)

    mapping := &pb.ETCMapping{
        Id:          1,
        EtcRecordId: 100,
        MappingType: "dtako",
        // ...
    }

    mockClient.EXPECT().
        Create(gomock.Any(), &pb.CreateMappingRequest{Mapping: mapping}).
        Return(&pb.CreateMappingResponse{Mapping: mapping}, nil)

    err := repo.Create(context.Background(), mapping)
    assert.NoError(t, err)
}
```

### Phase 6: Implementation Steps

#### 6.1 Repository Layer Migration
1. Create repository.proto with all repository operations
2. Generate Go code using buf
3. Create gRPC server implementations for repositories
4. Create gRPC client wrappers for repository access
5. Update all repository usages to use gRPC clients

#### 6.2 Service Layer Migration
1. Create services.proto with all service operations
2. Generate Go code using buf
3. Implement service gRPC servers
4. Update handlers to use service gRPC clients
5. Remove manual service interfaces

#### 6.3 Model Migration
1. Define all data models in Protocol Buffers
2. Create database adapter layer
3. Update all model usages to use pb messages
4. Remove GORM model files

#### 6.4 Testing Migration
1. Generate mocks from gRPC interfaces
2. Update all test files to use new mocks
3. Ensure constitution compliance (tests in tests/ directory)
4. Verify 100% coverage target

## Benefits of Migration

1. **Consistency**: All layers use the same generated code
2. **Type Safety**: Protocol Buffers provide strong typing
3. **Maintainability**: Single source of truth in .proto files
4. **Performance**: gRPC provides efficient binary serialization
5. **Testing**: Simplified mock generation from gRPC interfaces
6. **Documentation**: Proto files serve as API documentation
7. **Versioning**: Protocol Buffers support backward compatibility

## Migration Timeline

- **Week 1**: Define all .proto files (repository, services, models)
- **Week 2**: Generate code and create adapter layers
- **Week 3**: Migrate repository layer to gRPC
- **Week 4**: Migrate service layer to gRPC
- **Week 5**: Update all tests and achieve 100% coverage
- **Week 6**: Performance testing and optimization

## Risk Mitigation

1. **Backward Compatibility**: Maintain adapter layer during transition
2. **Testing**: Comprehensive test coverage before switching
3. **Rollback Plan**: Version control allows reverting if issues arise
4. **Phased Approach**: Migrate one component at a time
5. **Monitoring**: Track performance metrics during migration

## Success Criteria

- [ ] All interfaces defined in Protocol Buffers
- [ ] No manual Go interfaces in src/repositories/
- [ ] No GORM models in src/models/
- [ ] All tests passing with ≥90% coverage
- [ ] Consistent naming across all layers
- [ ] Mock generation working from gRPC interfaces
- [ ] Performance metrics maintained or improved
- [ ] Constitution compliance (no tests in src/)

## Next Steps

1. Review and approve this plan
2. Create detailed .proto file specifications
3. Set up buf tooling and CI/CD integration
4. Begin Phase 1 implementation
5. Track progress through todo list and monitoring