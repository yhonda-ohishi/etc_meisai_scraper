// Package mocks provides mock generation for all gRPC service interfaces
// This file contains go:generate directives to create mocks from Protocol Buffer generated services
//
// Usage:
//   cd tests && go generate ./mocks
//
// Generated mocks will be placed in tests/mocks/ and follow the pattern mock_*_client.go
package mocks

// Generate mocks for all ETCMeisaiService client interfaces from Protocol Buffers
//go:generate mockgen -source=../../src/pb/etc_meisai_grpc.pb.go -destination=mock_etc_meisai_service.go -package=mocks

// Generate mocks for gRPC client connection interface
//go:generate mockgen -package=mocks -destination=mock_grpc_client_conn.go google.golang.org/grpc ClientConnInterface

// Generate mocks for context package for testing timeout and cancellation scenarios
//go:generate mockgen -package=mocks -destination=mock_context.go context Context

// Generate mocks for common testing utilities
//go:generate mockgen -package=mocks -destination=mock_testing.go testing TB

// Repository Interface Mocks (will be generated after interface definitions are created)

// ETCMappingRepository mock - generates after src/repositories/etc_mapping_repository.go is implemented
//go:generate sh -c "if [ -f ../../src/repositories/etc_mapping_repository.go ]; then mockgen -source=../../src/repositories/etc_mapping_repository.go -destination=mock_etc_mapping_repository.go -package=mocks; fi"

// ETCMeisaiRecordRepository mock - generates after src/repositories/etc_meisai_record_repository.go is implemented
//go:generate sh -c "if [ -f ../../src/repositories/etc_meisai_record_repository.go ]; then mockgen -source=../../src/repositories/etc_meisai_record_repository.go -destination=mock_etc_meisai_record_repository.go -package=mocks; fi"

// ImportRepository mock - generates after src/repositories/import_repository.go is implemented
//go:generate sh -c "if [ -f ../../src/repositories/import_repository.go ]; then mockgen -source=../../src/repositories/import_repository.go -destination=mock_import_repository.go -package=mocks; fi"

// StatisticsRepository mock - generates after src/repositories/statistics_repository.go is implemented
//go:generate sh -c "if [ -f ../../src/repositories/statistics_repository.go ]; then mockgen -source=../../src/repositories/statistics_repository.go -destination=mock_statistics_repository.go -package=mocks; fi"

// Service Interface Mocks (will be generated after service interfaces are created)

// ETCMappingService mock - generates after src/services/etc_mapping_service.go interface is defined
//go:generate sh -c "if [ -f ../../src/services/etc_mapping_service.go ] && grep -q 'type.*Service.*interface' ../../src/services/etc_mapping_service.go; then mockgen -source=../../src/services/etc_mapping_service.go -destination=mock_etc_mapping_service.go -package=mocks; fi"

// ETCMeisaiService mock - generates after src/services/etc_meisai_service.go interface is defined
//go:generate sh -c "if [ -f ../../src/services/etc_meisai_service.go ] && grep -q 'type.*Service.*interface' ../../src/services/etc_meisai_service.go; then mockgen -source=../../src/services/etc_meisai_service.go -destination=mock_etc_meisai_service.go -package=mocks; fi"

// Database Adapter Mocks (will be generated after adapter interfaces are created)

// ProtoDBAdapter mock - generates after src/adapters/proto_db_adapter.go is implemented
//go:generate sh -c "if [ -f ../../src/adapters/proto_db_adapter.go ]; then mockgen -source=../../src/adapters/proto_db_adapter.go -destination=mock_proto_db_adapter.go -package=mocks; fi"

// Common utilities for mock testing

// MockGenerationMetadata contains information about mock generation process
type MockGenerationMetadata struct {
	GeneratedAt   string   `json:"generated_at"`
	SourceFiles   []string `json:"source_files"`
	GeneratedMocks []string `json:"generated_mocks"`
	Errors        []string `json:"errors"`
	Warnings      []string `json:"warnings"`
}

// CommonMockSetup provides common setup functionality for all mock-based tests
type CommonMockSetup struct {
	// Add common fields for mock setup if needed in the future
}

// NewCommonMockSetup creates a new instance of CommonMockSetup
func NewCommonMockSetup() *CommonMockSetup {
	return &CommonMockSetup{}
}

// Protocol Buffer Service Mock Generation Instructions:
//
// This file uses go:generate directives to create mocks for all gRPC services generated
// from Protocol Buffer definitions. The generated mocks follow these patterns:
//
// 1. Client Interfaces: Mock*ServiceClient - for testing gRPC client code
// 2. Server Interfaces: Mock*ServiceServer - for testing gRPC server implementations
// 3. Connection Interfaces: MockClientConnInterface - for testing connection management
//
// Mock Naming Conventions:
// - Protocol Buffer services: Mock{ServiceName}Client, Mock{ServiceName}Server
// - Repository interfaces: Mock{RepositoryName}
// - Service interfaces: Mock{ServiceName}Service
// - Adapter interfaces: Mock{AdapterName}Adapter
//
// Usage in Tests:
//   import "github.com/yhonda-ohishi/etc_meisai/tests/mocks"
//
//   // Create mock
//   mockClient := mocks.NewMockETCMeisaiServiceClient(ctrl)
//
//   // Set expectations
//   mockClient.EXPECT().CreateRecord(gomock.Any(), gomock.Any()).Return(response, nil)
//
//   // Use in tests
//   service := NewServiceWithClient(mockClient)
//   result := service.DoSomething()
//
// Regeneration:
// Mocks are automatically regenerated when source interfaces change.
// To manually regenerate all mocks:
//   cd tests && go generate ./mocks
//
// Constitutional Compliance:
// - All mock files are generated in tests/mocks/ (not in src/)
// - Mock generation follows the "no manual proto creation" principle
// - Generated mocks are ignored in .gitignore to prevent manual modification