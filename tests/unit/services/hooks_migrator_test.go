package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
	"github.com/yhonda-ohishi/etc_meisai/src/services"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHooksMigratorService_ETCMeisaiRecordBeforeCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil) // Use default logger
	service := services.NewHooksMigratorService(validationService, auditService)

	tests := []struct {
		name        string
		record      *pb.ETCMeisaiRecord
		expectError bool
		expectHash  bool
	}{
		{
			name: "valid record without hash",
			record: &pb.ETCMeisaiRecord{
				Date:           "2023-12-01",
				Time:           "10:30:00",
				EntranceIc:     "Tokyo IC",
				ExitIc:         "Osaka IC",
				TollAmount:     1500,
				CarNumber:      "123-45",
				EtcCardNumber:  "1234567890123456",
			},
			expectError: false,
			expectHash:  true,
		},
		{
			name: "valid record with existing hash",
			record: &pb.ETCMeisaiRecord{
				Hash:           "existing_hash",
				Date:           "2023-12-01",
				Time:           "10:30:00",
				EntranceIc:     "Tokyo IC",
				ExitIc:         "Osaka IC",
				TollAmount:     1500,
				CarNumber:      "123-45",
				EtcCardNumber:  "1234567890123456",
			},
			expectError: false,
			expectHash:  false, // Should not overwrite existing hash
		},
		{
			name: "invalid record - missing date",
			record: &pb.ETCMeisaiRecord{
				Time:          "10:30:00",
				EntranceIc:    "Tokyo IC",
				ExitIc:        "Osaka IC",
				TollAmount:    1500,
				CarNumber:     "123-45",
				EtcCardNumber: "1234567890123456",
			},
			expectError: true,
		},
		{
			name: "invalid record - invalid time format",
			record: &pb.ETCMeisaiRecord{
				Date:          "2023-12-01",
				Time:          "25:30:00", // Invalid hour
				EntranceIc:    "Tokyo IC",
				ExitIc:        "Osaka IC",
				TollAmount:    1500,
				CarNumber:     "123-45",
				EtcCardNumber: "1234567890123456",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalHash := tt.record.Hash
			err := service.ETCMeisaiRecordBeforeCreate(tt.record)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectHash {
					assert.NotEmpty(t, tt.record.Hash)
					assert.NotEqual(t, originalHash, tt.record.Hash)
				} else {
					assert.Equal(t, originalHash, tt.record.Hash)
				}
			}
		})
	}
}

func TestHooksMigratorService_ETCMeisaiRecordBeforeSave(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	tests := []struct {
		name        string
		record      *pb.ETCMeisaiRecord
		expectError bool
	}{
		{
			name: "valid record",
			record: &pb.ETCMeisaiRecord{
				Hash:           "existing_hash",
				Date:           "2023-12-01",
				Time:           "10:30:00",
				EntranceIc:     "Tokyo IC",
				ExitIc:         "Osaka IC",
				TollAmount:     1500,
				CarNumber:      "123-45",
				EtcCardNumber:  "1234567890123456",
			},
			expectError: false,
		},
		{
			name: "invalid record - negative toll amount",
			record: &pb.ETCMeisaiRecord{
				Hash:          "existing_hash",
				Date:          "2023-12-01",
				Time:          "10:30:00",
				EntranceIc:    "Tokyo IC",
				ExitIc:        "Osaka IC",
				TollAmount:    -100, // Negative amount
				CarNumber:     "123-45",
				EtcCardNumber: "1234567890123456",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ETCMeisaiRecordBeforeSave(tt.record)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHooksMigratorService_ImportSessionBeforeCreate(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	tests := []struct {
		name        string
		session     *pb.ImportSession
		expectError bool
		checkUUID   bool
		checkTime   bool
		checkStatus bool
	}{
		{
			name: "valid session without defaults",
			session: &pb.ImportSession{
				AccountType: "corporate",
				AccountId:   "test_account",
				FileName:    "test.csv",
				FileSize:    1024,
				CreatedBy:   "user123",
			},
			expectError: false,
			checkUUID:   true,
			checkTime:   true,
			checkStatus: true,
		},
		{
			name: "session with existing ID",
			session: &pb.ImportSession{
				Id:          "existing-uuid-1234-5678-9abc-def123456789",
				AccountType: "personal",
				AccountId:   "test_account",
				FileName:    "data.csv",
				FileSize:    2048,
				Status:      pb.ImportStatus_IMPORT_STATUS_PROCESSING,
				StartedAt:   timestamppb.New(time.Now()),
				CreatedBy:   "user456",
			},
			expectError: false,
			checkUUID:   false, // Should not overwrite existing ID
			checkTime:   false, // Should not overwrite existing time
			checkStatus: false, // Should not overwrite existing status
		},
		{
			name: "invalid session - invalid account type",
			session: &pb.ImportSession{
				AccountType: "invalid_type",
				AccountId:   "test_account",
				FileName:    "test.csv",
				FileSize:    1024,
				CreatedBy:   "user123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalID := tt.session.Id
			originalTime := tt.session.StartedAt
			originalStatus := tt.session.Status

			err := service.ImportSessionBeforeCreate(tt.session)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.checkUUID {
					assert.NotEmpty(t, tt.session.Id)
					assert.NotEqual(t, originalID, tt.session.Id)
				} else {
					assert.Equal(t, originalID, tt.session.Id)
				}

				if tt.checkTime {
					assert.NotNil(t, tt.session.StartedAt)
					assert.NotEqual(t, originalTime, tt.session.StartedAt)
				} else if originalTime != nil {
					assert.Equal(t, originalTime, tt.session.StartedAt)
				}

				if tt.checkStatus {
					assert.Equal(t, pb.ImportStatus_IMPORT_STATUS_PENDING, tt.session.Status)
				} else if originalStatus != pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED {
					assert.Equal(t, originalStatus, tt.session.Status)
				}
			}
		})
	}
}

func TestHooksMigratorService_ImportSessionBeforeSave(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	validSession := &pb.ImportSession{
		Id:            "12345678-1234-4abc-8901-123456789abc",
		AccountType:   "corporate",
		AccountId:     "test_account",
		FileName:      "test.csv",
		FileSize:      1024,
		TotalRows:     100,
		ProcessedRows: 50,
		SuccessRows:   40,
		ErrorRows:     5,
		DuplicateRows: 5,
		Status:        pb.ImportStatus_IMPORT_STATUS_PROCESSING,
		StartedAt:     timestamppb.New(time.Now()),
		CreatedBy:     "user123",
	}

	err := service.ImportSessionBeforeSave(validSession)
	assert.NoError(t, err)

	// Test invalid session
	invalidSession := &pb.ImportSession{
		Id:            "invalid-uuid",
		AccountType:   "corporate",
		AccountId:     "test_account",
		FileName:      "test.csv",
		FileSize:      1024,
		CreatedBy:     "user123",
	}

	err = service.ImportSessionBeforeSave(invalidSession)
	assert.Error(t, err)
}

func TestHooksMigratorService_ETCMappingBeforeCreate(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	tests := []struct {
		name            string
		mapping         *pb.ETCMapping
		expectError     bool
		checkStatus     bool
		checkTimestamps bool
	}{
		{
			name: "valid mapping without defaults",
			mapping: &pb.ETCMapping{
				EtcRecordId:      123,
				MappingType:      "manual",
				MappedEntityId:   456,
				MappedEntityType: "transaction",
				Confidence:       0.95,
				CreatedBy:        "user123",
			},
			expectError:     false,
			checkStatus:     true,
			checkTimestamps: true,
		},
		{
			name: "mapping with existing status",
			mapping: &pb.ETCMapping{
				EtcRecordId:      123,
				MappingType:      "automatic",
				MappedEntityId:   456,
				MappedEntityType: "transaction",
				Confidence:       0.85,
				Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
				CreatedBy:        "user456",
			},
			expectError:     false,
			checkStatus:     false, // Should not overwrite existing status
			checkTimestamps: true,
		},
		{
			name: "invalid mapping - confidence out of range",
			mapping: &pb.ETCMapping{
				EtcRecordId:      123,
				MappingType:      "manual",
				MappedEntityId:   456,
				MappedEntityType: "transaction",
				Confidence:       1.5, // Invalid confidence > 1.0
				CreatedBy:        "user123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalStatus := tt.mapping.Status
			originalCreatedAt := tt.mapping.CreatedAt

			err := service.ETCMappingBeforeCreate(tt.mapping)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.checkStatus {
					assert.Equal(t, pb.MappingStatus_MAPPING_STATUS_PENDING, tt.mapping.Status)
				} else if originalStatus != pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED {
					assert.Equal(t, originalStatus, tt.mapping.Status)
				}

				if tt.checkTimestamps {
					assert.NotNil(t, tt.mapping.CreatedAt)
					assert.NotEqual(t, originalCreatedAt, tt.mapping.CreatedAt)
				}
			}
		})
	}
}

func TestHooksMigratorService_ExecuteBeforeCreateHook(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	tests := []struct {
		name       string
		entityType string
		entity     interface{}
		expectPass bool
	}{
		{
			name:       "valid ETC record",
			entityType: "ETCMeisaiRecord",
			entity: &pb.ETCMeisaiRecord{
				Date:           "2023-12-01",
				Time:           "10:30:00",
				EntranceIc:     "Tokyo IC",
				ExitIc:         "Osaka IC",
				TollAmount:     1500,
				CarNumber:      "123-45",
				EtcCardNumber:  "1234567890123456",
			},
			expectPass: true,
		},
		{
			name:       "valid import session",
			entityType: "ImportSession",
			entity: &pb.ImportSession{
				AccountType: "corporate",
				AccountId:   "test_account",
				FileName:    "test.csv",
				FileSize:    1024,
				CreatedBy:   "user123",
			},
			expectPass: true,
		},
		{
			name:       "valid ETC mapping",
			entityType: "ETCMapping",
			entity: &pb.ETCMapping{
				EtcRecordId:      123,
				MappingType:      "manual",
				MappedEntityId:   456,
				MappedEntityType: "transaction",
				Confidence:       0.95,
				CreatedBy:        "user123",
			},
			expectPass: true,
		},
		{
			name:       "unsupported entity type",
			entityType: "UnsupportedEntity",
			entity:     "some data",
			expectPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ExecuteBeforeCreateHook(tt.entityType, tt.entity)

			require.NotNil(t, result)
			assert.Equal(t, "BeforeCreate", result.HookType)
			assert.Equal(t, tt.entityType, result.EntityType)
			assert.False(t, result.ExecutedAt.IsZero())

			if tt.expectPass {
				assert.True(t, result.Success)
				assert.Empty(t, result.ErrorMessage)
			} else {
				assert.False(t, result.Success)
				assert.NotEmpty(t, result.ErrorMessage)
			}
		})
	}
}

func TestHooksMigratorService_ExecuteBeforeSaveHook(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	validRecord := &pb.ETCMeisaiRecord{
		Hash:           "test_hash",
		Date:           "2023-12-01",
		Time:           "10:30:00",
		EntranceIc:     "Tokyo IC",
		ExitIc:         "Osaka IC",
		TollAmount:     1500,
		CarNumber:      "123-45",
		EtcCardNumber:  "1234567890123456",
	}

	result := service.ExecuteBeforeSaveHook("ETCMeisaiRecord", validRecord)

	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "BeforeSave", result.HookType)
	assert.Equal(t, "ETCMeisaiRecord", result.EntityType)
	assert.Empty(t, result.ErrorMessage)
}

func TestHooksMigratorService_GetSupportedEntityTypes(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	supportedTypes := service.GetSupportedEntityTypes()

	expectedTypes := []string{
		"ETCMeisaiRecord",
		"ImportSession",
		"ETCMapping",
	}

	assert.ElementsMatch(t, expectedTypes, supportedTypes)
}

func TestHooksMigratorService_HashGeneration(t *testing.T) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	record1 := &pb.ETCMeisaiRecord{
		Date:           "2023-12-01",
		Time:           "10:30:00",
		EntranceIc:     "Tokyo IC",
		ExitIc:         "Osaka IC",
		TollAmount:     1500,
		CarNumber:      "123-45",
		EtcCardNumber:  "1234567890123456",
	}

	record2 := &pb.ETCMeisaiRecord{
		Date:           "2023-12-01",
		Time:           "10:30:00",
		EntranceIc:     "Tokyo IC",
		ExitIc:         "Osaka IC",
		TollAmount:     1500,
		CarNumber:      "123-45",
		EtcCardNumber:  "1234567890123456",
	}

	record3 := &pb.ETCMeisaiRecord{
		Date:           "2023-12-01",
		Time:           "10:30:00",
		EntranceIc:     "Tokyo IC",
		ExitIc:         "Osaka IC",
		TollAmount:     2000, // Different amount
		CarNumber:      "123-45",
		EtcCardNumber:  "1234567890123456",
	}

	// Generate hashes
	err1 := service.ETCMeisaiRecordBeforeCreate(record1)
	err2 := service.ETCMeisaiRecordBeforeCreate(record2)
	err3 := service.ETCMeisaiRecordBeforeCreate(record3)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	// Same records should have same hash
	assert.Equal(t, record1.Hash, record2.Hash)

	// Different records should have different hash
	assert.NotEqual(t, record1.Hash, record3.Hash)

	// Hashes should be non-empty and reasonable length
	assert.NotEmpty(t, record1.Hash)
	assert.True(t, len(record1.Hash) >= 32) // SHA256 should be at least 32 chars in hex
}

// Benchmark tests
func BenchmarkHooksMigratorService_ETCMeisaiRecordBeforeCreate(b *testing.B) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	record := &pb.ETCMeisaiRecord{
		Date:           "2023-12-01",
		Time:           "10:30:00",
		EntranceIc:     "Tokyo IC",
		ExitIc:         "Osaka IC",
		TollAmount:     1500,
		CarNumber:      "123-45",
		EtcCardNumber:  "1234567890123456",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset hash for each iteration
		record.Hash = ""
		service.ETCMeisaiRecordBeforeCreate(record)
	}
}

func BenchmarkHooksMigratorService_ValidationOnly(b *testing.B) {
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(nil)
	service := services.NewHooksMigratorService(validationService, auditService)

	record := &pb.ETCMeisaiRecord{
		Hash:           "existing_hash",
		Date:           "2023-12-01",
		Time:           "10:30:00",
		EntranceIc:     "Tokyo IC",
		ExitIc:         "Osaka IC",
		TollAmount:     1500,
		CarNumber:      "123-45",
		EtcCardNumber:  "1234567890123456",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ETCMeisaiRecordBeforeSave(record)
	}
}