# Research: Full gRPC Architecture Migration

**Feature**: 006-refactor-src-to
**Date**: 2025-09-26
**Status**: Complete

## Overview
This document consolidates research findings for migrating the etc_meisai codebase from a hybrid architecture to a fully Protocol Buffer-based system.

## Research Tasks Completed

### 1. Buf Tooling Configuration for Go gRPC Projects

**Decision**: Use buf for Protocol Buffer management and code generation

**Rationale**:
- Industry standard for modern Protocol Buffer workflows
- Built-in linting and breaking change detection
- Seamless integration with Go module ecosystem
- Supports multiple code generation plugins

**Configuration Structure**:
```yaml
# buf.yaml - for linting and breaking change detection
version: v1
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT

# buf.gen.yaml - for code generation
version: v1
managed:
  enabled: true
  go_package_prefix:
    default: github.com/yhonda-ohishi/etc_meisai/src/pb
plugins:
  - name: go
  - name: go-grpc
  - name: grpc-gateway
  - name: openapiv2
```

**Alternatives Considered**:
- protoc directly: More manual configuration, lacks modern tooling
- prototool: Deprecated in favor of buf
- Bazel: Too complex for our project size

### 2. Mock Generation from gRPC Interfaces

**Decision**: Use mockgen with gRPC client interfaces

**Rationale**:
- Official Google mock generation tool
- Direct support for gRPC service interfaces
- Generates type-safe mocks
- Integrates with existing gomock framework

**Implementation Pattern**:
```bash
# Generate mocks from compiled proto
mockgen -source=src/pb/repository_grpc.pb.go \
        -destination=tests/mocks/mock_repository.go \
        -package=mocks

# Or generate from reflection
mockgen -destination=tests/mocks/mock_repository.go \
        -package=mocks \
        github.com/yhonda-ohishi/etc_meisai/src/pb ETCMappingRepositoryClient
```

**Alternatives Considered**:
- testify/mock: Manual mock creation, more maintenance
- counterfeiter: Less Go ecosystem adoption
- Custom mocks: Too much boilerplate

### 3. Database Adapter Patterns for Protocol Buffers

**Decision**: Create adapter layer for proto<->database mapping

**Rationale**:
- Clean separation of concerns
- Maintains database column naming conventions
- Allows gradual migration
- Handles type conversions (timestamp, enums)

**Pattern Example**:
```go
// src/adapters/db_adapter.go
type ProtoDBAdapter struct {
    db *gorm.DB
}

func (a *ProtoDBAdapter) ETCMappingToDB(pb *pb.ETCMappingEntity) map[string]interface{} {
    return map[string]interface{}{
        "id": pb.Id,
        "etc_record_id": pb.EtcRecordId,
        "status": pb.Status.String(),
        "created_at": pb.CreatedAt.AsTime(),
    }
}

func (a *ProtoDBAdapter) DBToETCMapping(row map[string]interface{}) *pb.ETCMappingEntity {
    // Conversion logic
}
```

**Alternatives Considered**:
- Direct GORM with proto structs: Field naming conflicts
- ORM replacement: Too invasive for migration
- Manual SQL: Loss of type safety

### 4. Migration Strategy from GORM to Proto-Based Models

**Decision**: Phased migration with adapter pattern

**Rationale**:
- Minimal risk to existing functionality
- Allows incremental testing
- Maintains backward compatibility
- Rollback capability at each phase

**Migration Phases**:
1. **Phase 1**: Define all protos alongside existing code
2. **Phase 2**: Create adapters for database operations
3. **Phase 3**: Replace repository interfaces with gRPC
4. **Phase 4**: Replace service interfaces with gRPC
5. **Phase 5**: Remove GORM models and manual interfaces

**Key Considerations**:
- Database migrations: Not required (column mapping in adapter)
- Performance impact: Minimal (binary serialization faster than JSON)
- Testing strategy: Parallel tests during migration
- Rollback plan: Git revert at any phase

**Alternatives Considered**:
- Big bang migration: Too risky
- Dual model maintenance: Increases complexity
- Database schema change: Unnecessary disruption

## Technical Recommendations

### 1. Directory Structure
```
src/
├── proto/          # All .proto definitions
├── pb/             # Generated code (git-ignored)
├── adapters/       # Proto<->DB adapters
├── repositories/   # Repository implementations
├── services/       # Service implementations
└── grpc/           # gRPC server setup
```

### 2. Code Generation Workflow
```bash
# One-time setup
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/golang/mock/mockgen@latest

# Generate code
cd src/proto
buf generate

# Generate mocks
go generate ./...
```

### 3. Naming Conventions
- Proto files: snake_case (e.g., `repository_service.proto`)
- Proto messages: PascalCase with suffix (e.g., `ETCMappingEntity`)
- Proto fields: snake_case (e.g., `etc_record_id`)
- Generated Go: Automatic conversion to Go conventions

### 4. Testing Strategy
- Unit tests: Mock gRPC clients
- Integration tests: In-memory gRPC server
- Contract tests: Schema validation
- Performance tests: Benchmark proto vs current

## Risk Mitigation

### Identified Risks
1. **Learning curve**: Team unfamiliar with Protocol Buffers
   - Mitigation: Quickstart guide, pair programming
2. **Build time increase**: Code generation adds overhead
   - Mitigation: Cache generated code, parallel generation
3. **Debugging complexity**: Generated code harder to debug
   - Mitigation: Good logging, clear adapter layer
4. **Version conflicts**: Proto changes affect all layers
   - Mitigation: Buf breaking change detection

### Success Metrics
- Build time: < 60 seconds (including generation)
- Test coverage: Maintain 100%
- Performance: Response time ±10%
- Code consistency: 0 manual interface definitions

## Conclusion

The research confirms that migrating to a fully Protocol Buffer-based architecture is feasible and beneficial. The phased approach with adapter pattern minimizes risk while the buf toolchain provides modern developer experience. The migration will eliminate naming inconsistencies, reduce maintenance overhead, and improve type safety across all layers.

## Next Steps

1. Create comprehensive data model documentation
2. Define all service contracts in Protocol Buffers
3. Generate initial contract tests
4. Update quickstart guide for developers
5. Update CLAUDE.md with new architecture

---
*Research completed: 2025-09-26*