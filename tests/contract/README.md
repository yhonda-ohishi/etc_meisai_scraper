# Contract Testing Suite - Phase 4 T010

This directory contains comprehensive contract tests for the ETC明細 gRPC service, implementing all requirements for Phase 4 - T010: Contract Testing Expansion.

## Overview

The contract testing suite validates API stability, backward compatibility, and performance SLAs for the gRPC-based ETC明細 service.

## Test Structure

### T010-A: gRPC Service Contract Testing
**File**: `grpc_service_contract_test.go`

Tests all gRPC service definitions in `src/proto/etc_meisai.proto`:

- **ETCMeisaiRecord Operations**:
  - `CreateRecord` - Record creation with validation
  - `GetRecord` - Record retrieval by ID
  - `ListRecords` - Paginated listing with filters
  - `UpdateRecord` - Record modification
  - `DeleteRecord` - Record deletion

- **Import Operations**:
  - `ImportCSV` - CSV file import and processing
  - `GetImportSession` - Import session status
  - `ListImportSessions` - Import history

- **Mapping Operations**:
  - `CreateMapping` - ETC-DTako mapping creation
  - `GetMapping` - Mapping retrieval
  - `ListMappings` - Mapping queries
  - `UpdateMapping` - Mapping updates
  - `DeleteMapping` - Mapping deletion

- **Statistics Operations**:
  - `GetStatistics` - Aggregated data analysis

### T010-B: API Version Compatibility Testing
**File**: `api_version_compatibility_test.go`

Validates compatibility across API versions:

- **Version Negotiation**:
  - Client version declaration
  - Server version response
  - Feature capability negotiation

- **Backward Compatibility**:
  - Required field preservation
  - Optional field handling
  - Default value consistency

- **Forward Compatibility**:
  - New optional field support
  - Unknown enum value handling
  - Message field evolution

- **Error Handling**:
  - Version mismatch scenarios
  - Deprecation warnings
  - Graceful degradation

### T010-C: Schema Evolution Testing
**File**: `schema_evolution_test.go`

Tests Protocol Buffer schema changes:

- **Field Evolution**:
  - Adding optional fields
  - Field number stability
  - Field type compatibility

- **Enum Evolution**:
  - New enum value addition
  - Unknown enum value handling
  - Default enum behavior

- **Message Structure**:
  - Nested message changes
  - Repeated field evolution
  - Oneof field support

- **Serialization Compatibility**:
  - Binary format stability
  - Cross-version deserialization
  - Field preservation

### T010-D: End-to-End Workflow Testing
**File**: `end_to_end_workflow_test.go`

Complete ETC data processing pipeline:

- **CSV Import Workflow**:
  - File upload and validation
  - Processing progress monitoring
  - Error handling and reporting
  - Data verification

- **Data Mapping Workflow**:
  - ETC-DTako record mapping
  - Automatic matching algorithms
  - Manual mapping confirmation
  - Referential integrity

- **Statistics Generation**:
  - Data aggregation
  - Report generation
  - Performance optimization

- **Concurrent Operations**:
  - Multi-user scenarios
  - Data consistency under load
  - Error isolation

### T010-E: Performance SLA Testing
**File**: `performance_sla_test.go`

Service Level Agreement validation:

- **Response Time SLAs**:
  - CreateRecord: < 100ms
  - GetRecord: < 100ms
  - ListRecords: < 200ms
  - UpdateRecord: < 100ms
  - ImportCSV: < 1s (small files)
  - GetStatistics: < 500ms

- **Throughput SLAs**:
  - CreateRecord: ≥ 50 ops/sec
  - GetRecord: ≥ 100 ops/sec
  - Concurrent users: ≥ 10 users
  - Mixed operations: ≥ 50 concurrent ops

- **Resource Usage**:
  - Large dataset handling
  - Complex query performance
  - Memory efficiency
  - Connection pooling

## Test Infrastructure

### Test Suite Setup
**File**: `contract_test_suite.go`

Provides common infrastructure:

- gRPC server setup and teardown
- Test client configuration
- Database initialization
- Performance monitoring
- Error collection and reporting

### Legacy Tests
**Files**: `mock_generation_test.go.disabled`, `test_execution_test.go.disabled`

Previous contract tests (currently disabled):
- Mock generation testing
- Service execution contracts
- Repository interface testing

## Running Contract Tests

### Prerequisites

1. **gRPC Server**: Ensure the ETC明細 gRPC server is running
2. **Database**: Set up test database (SQLite or MySQL)
3. **Environment**: Configure test environment variables

### Environment Variables

```bash
export ETC_TEST_DATABASE_URL="sqlite://contract_test.db"
export ETC_TEST_GRPC_PORT="9090"
export ETC_TEST_ENABLE_PERFORMANCE="true"
export ETC_TEST_ENABLE_E2E="true"
```

### Running Tests

```bash
# Run all contract tests
go test ./tests/contract -v

# Run specific test categories
go test ./tests/contract -v -run "TestGRPCServiceContract"
go test ./tests/contract -v -run "TestAPIVersionCompatibility"
go test ./tests/contract -v -run "TestSchemaEvolution"
go test ./tests/contract -v -run "TestEndToEndWorkflow"
go test ./tests/contract -v -run "TestPerformanceSLA"

# Run with race detection
go test ./tests/contract -v -race

# Run with coverage
go test ./tests/contract -v -cover -coverprofile=contract_coverage.out
```

### Performance Testing

```bash
# Run only performance tests
go test ./tests/contract -v -run "TestPerformanceSLA"

# Run with custom timeout
go test ./tests/contract -v -timeout 5m -run "TestPerformanceSLA"

# Run with memory profiling
go test ./tests/contract -v -memprofile=mem.prof -run "TestPerformanceSLA"
```

## SLA Requirements

### Response Time SLAs
- **Individual Operations**: Must complete within specified time limits
- **Concurrent Operations**: Performance degradation must be minimal
- **Large Datasets**: Must handle efficiently within reasonable time

### Throughput SLAs
- **Create Operations**: Minimum 50 operations per second
- **Read Operations**: Minimum 100 operations per second
- **Concurrent Users**: Support at least 10 concurrent users
- **Mixed Workload**: Handle 50+ concurrent operations

### Availability SLAs
- **Uptime**: 99.9% availability during business hours
- **Error Rate**: Less than 1% error rate under normal load
- **Recovery Time**: Service recovery within 30 seconds

## Best Practices

### Contract Design
1. **Stable APIs**: Ensure field numbers and message structure remain stable
2. **Backward Compatibility**: New fields must be optional
3. **Forward Compatibility**: Handle unknown fields gracefully
4. **Error Handling**: Provide meaningful error messages and codes

### Testing Strategy
1. **Comprehensive Coverage**: Test all gRPC methods and scenarios
2. **Real-world Data**: Use realistic test data and scenarios
3. **Performance Focus**: Validate SLAs under various load conditions
4. **Automation**: Integrate into CI/CD pipeline

### Maintenance
1. **Regular Updates**: Keep tests current with API changes
2. **Performance Monitoring**: Track SLA compliance over time
3. **Documentation**: Maintain clear test documentation
4. **Review Process**: Regular review of test effectiveness

## Troubleshooting

### Common Issues

1. **Test Environment**: Ensure gRPC server and database are properly configured
2. **Timeouts**: Adjust test timeouts for slower environments
3. **Concurrency**: Check for race conditions in concurrent tests
4. **Memory**: Monitor memory usage during performance tests

### Debug Mode

```bash
# Enable debug logging
export ETC_TEST_DEBUG="true"
go test ./tests/contract -v

# Run single test with verbose output
go test ./tests/contract -v -run "TestSpecificTest"
```

### Performance Analysis

```bash
# Generate CPU profile
go test ./tests/contract -cpuprofile=cpu.prof -run "TestPerformanceSLA"

# Generate memory profile
go test ./tests/contract -memprofile=mem.prof -run "TestPerformanceSLA"

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Integration with CI/CD

The contract tests should be integrated into the CI/CD pipeline:

1. **Pull Request Validation**: Run contract tests on every PR
2. **Performance Regression**: Detect performance degradation
3. **API Breaking Changes**: Prevent backward incompatible changes
4. **SLA Monitoring**: Track SLA compliance over time

## Future Enhancements

1. **Streaming Support**: Add tests for bidirectional streaming
2. **Multi-region Testing**: Test across different deployment regions
3. **Load Testing**: Extended load testing scenarios
4. **Chaos Engineering**: Fault injection and resilience testing