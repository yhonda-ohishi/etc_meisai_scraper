# Quickstart: Full gRPC Architecture Migration

**Feature**: 006-refactor-src-to
**Date**: 2025-09-26

## Overview
This guide helps developers understand and work with the new fully Protocol Buffer-based architecture after migration from the hybrid system.

## Prerequisites
- Go 1.21+ installed
- buf CLI tool installed
- mockgen tool installed
- Access to the etc_meisai repository

## Quick Setup

### 1. Install Required Tools
```bash
# Install buf for Protocol Buffer management
go install github.com/bufbuild/buf/cmd/buf@latest

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Install mockgen for test mocks
go install github.com/golang/mock/mockgen@latest
```

### 2. Generate Code from Protocol Buffers
```bash
# Navigate to proto directory
cd src/proto

# Generate all code
buf generate

# Verify generated code
ls ../pb/
```

### 3. Generate Mocks for Testing
```bash
# Generate mocks from gRPC interfaces
cd ../..
go generate ./...

# Verify mocks
ls tests/mocks/
```

## Test Scenarios

### Scenario 1: Update a Data Structure
**Given**: A developer needs to add a new field to ETCMappingEntity
**When**: They update the protocol buffer definition
**Then**: All related code is automatically regenerated

```bash
# 1. Edit the proto file
vi src/proto/repository.proto
# Add new field: string notes = 12;

# 2. Regenerate code
cd src/proto && buf generate

# 3. Run tests to verify
go test ./tests/contract/...
```

**Expected Result**: New field available in all layers without manual interface updates

### Scenario 2: Add a New Service Method
**Given**: A developer needs to add a new repository method
**When**: They define it in the protocol buffer file
**Then**: The method is available across all layers

```bash
# 1. Add method to proto
vi src/proto/repository.proto
# Add: rpc GetActiveCount(GetActiveCountRequest) returns (CountResponse);

# 2. Regenerate
cd src/proto && buf generate

# 3. Implement the method
vi src/repositories/etc_mapping_repository.go

# 4. Test
go test ./tests/unit/repositories/...
```

**Expected Result**: Method signature automatically available without manual interface definition

### Scenario 3: Generate Mocks for Testing
**Given**: A developer needs to test with mocked repositories
**When**: They run the mock generation tool
**Then**: Mocks are generated from Protocol Buffer definitions

```bash
# 1. Generate mocks
mockgen -source=src/pb/repository_grpc.pb.go \
        -destination=tests/mocks/mock_repository.go \
        -package=mocks

# 2. Use in tests
cat > test_example.go << 'EOF'
func TestWithMock(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockETCMappingRepositoryClient(ctrl)
    // Use mock...
}
EOF

# 3. Run tests
go test ./tests/unit/...
```

**Expected Result**: Type-safe mocks generated without manual maintenance

### Scenario 4: Verify Complete Migration
**Given**: The system is fully migrated
**When**: A developer searches for interface definitions
**Then**: They find only protocol buffer files

```bash
# Check for manual interfaces (should be empty)
find src -name "*.go" -exec grep -l "^type.*Repository interface" {} \;

# Check for GORM models (should be empty)
find src/models -name "*.go" -exec grep -l "gorm.Model" {} \;

# Verify proto files exist
ls src/proto/*.proto
```

**Expected Result**: No manual interfaces or GORM models found

### Scenario 5: Performance Validation
**Given**: The migration is complete
**When**: Performance tests are run
**Then**: Response times are within ±10% of original

```bash
# Run benchmark tests
go test -bench=. ./tests/performance/...

# Compare with baseline
# Original: BenchmarkCreateMapping-8  10000  120534 ns/op
# New:     BenchmarkCreateMapping-8  10000  115892 ns/op

# Calculate difference
# (115892 - 120534) / 120534 = -3.85% (within ±10%)
```

**Expected Result**: Performance maintained or improved

## Common Tasks

### Adding a New Entity
1. Define the entity in `src/proto/models.proto`
2. Define repository service in `src/proto/repository.proto`
3. Run `buf generate` in `src/proto/`
4. Implement the repository in `src/repositories/`
5. Create adapter in `src/adapters/`
6. Write tests in `tests/unit/repositories/`

### Modifying an Existing Entity
1. Update the proto message definition
2. Run `buf generate`
3. Update adapter mappings if needed
4. Run tests to verify compatibility

### Creating a New Service
1. Define service in `src/proto/services.proto`
2. Run `buf generate`
3. Implement service in `src/services/`
4. Write contract tests in `tests/contract/`

## Troubleshooting

### Issue: Generated code not found
**Solution**: Ensure buf.gen.yaml paths are correct and run `buf generate` from `src/proto/`

### Issue: Mock generation fails
**Solution**: Ensure generated pb files exist first, then run mockgen

### Issue: Database column mismatch
**Solution**: Check adapter mappings in `src/adapters/db_adapter.go`

### Issue: Enum values not recognized
**Solution**: Use `.String()` method on enum values for database storage

### Issue: Timestamp conversion errors
**Solution**: Use `timestamppb.New()` and `.AsTime()` for conversions

## Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ gRPC
┌──────▼──────┐
│  API Layer  │ (grpc-gateway)
└──────┬──────┘
       │
┌──────▼──────┐
│   Service   │ (business logic)
│    Layer    │
└──────┬──────┘
       │ gRPC
┌──────▼──────┐
│ Repository  │ (data access)
│    Layer    │
└──────┬──────┘
       │
┌──────▼──────┐
│   Adapter   │ (proto<->DB mapping)
│    Layer    │
└──────┬──────┘
       │
┌──────▼──────┐
│   Database  │
└─────────────┘
```

## Best Practices

1. **Never modify generated code** - Always change the proto files
2. **Use buf lint** before committing proto changes
3. **Run buf breaking** to check for breaking changes
4. **Keep adapters thin** - Only type conversions, no business logic
5. **Test with generated mocks** - Don't create manual mocks
6. **Version your proto files** - Use package versioning (v1, v2)
7. **Document proto fields** - Add comments in proto files

## Useful Commands

```bash
# Lint proto files
buf lint

# Check for breaking changes
buf breaking --against '.git#branch=main'

# Format proto files
buf format -w

# Generate specific service only
buf generate --path src/proto/repository.proto

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Next Steps

1. Complete any remaining manual interface removals
2. Update CI/CD pipeline to include buf generate
3. Train team on Protocol Buffer best practices
4. Monitor performance metrics post-migration
5. Document any custom adapter logic

## Support

For questions or issues:
- Check the [research.md](./research.md) for design decisions
- Review [data-model.md](./data-model.md) for entity details
- See [contracts/](./contracts/) for API specifications
- Contact the architecture team for complex issues

---
*Quickstart guide version 1.0 - 2025-09-26*