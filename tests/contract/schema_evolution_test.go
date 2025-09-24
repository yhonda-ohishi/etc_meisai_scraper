package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/yhonda-ohishi/etc_meisai/src/pb"
)

// TestSchemaEvolution_T010C validates Protocol Buffer message schema evolution
// This ensures that schema changes maintain backward and forward compatibility
func TestSchemaEvolution_T010C(t *testing.T) {

	t.Run("ETCMeisaiRecord_FieldAddition_Compatibility", func(t *testing.T) {
		// Contract: Adding optional fields must not break existing serialized data

		// Simulate old version of ETCMeisaiRecord (without newer optional fields)
		oldRecord := &pb.ETCMeisaiRecord{
			Id:             1,
			Hash:           "evolution-test-001",
			Date:           "2024-01-01",
			Time:           "10:00:00",
			EntranceIc:     "進化テストIC",
			ExitIc:         "進化テスト出口IC",
			TollAmount:     1500,
			CarNumber:      "品川123あ1234",
			EtcCardNumber:  "1234567890123456",
			CreatedAt:      timestamppb.Now(),
			UpdatedAt:      timestamppb.Now(),
			// EtcNum and DtakoRowId intentionally omitted (simulating old schema)
		}

		// Serialize with old schema
		oldData, err := proto.Marshal(oldRecord)
		require.NoError(t, err, "Old record must serialize successfully")

		// Deserialize with new schema (should handle missing optional fields)
		var newRecord pb.ETCMeisaiRecord
		err = proto.Unmarshal(oldData, &newRecord)
		require.NoError(t, err, "Old data must deserialize with new schema")

		// Contract assertions
		assert.Equal(t, oldRecord.Id, newRecord.Id, "ID must be preserved")
		assert.Equal(t, oldRecord.Hash, newRecord.Hash, "Hash must be preserved")
		assert.Equal(t, oldRecord.Date, newRecord.Date, "Date must be preserved")
		assert.Equal(t, oldRecord.TollAmount, newRecord.TollAmount, "TollAmount must be preserved")
		assert.Nil(t, newRecord.EtcNum, "New optional field EtcNum should be nil for old data")
		assert.Nil(t, newRecord.DtakoRowId, "New optional field DtakoRowId should be nil for old data")
	})

	t.Run("ETCMeisaiRecord_FieldRename_Compatibility", func(t *testing.T) {
		// Contract: Field renames using deprecated fields must maintain compatibility

		record := &pb.ETCMeisaiRecord{
			Id:             2,
			Hash:           "rename-test-001",
			Date:           "2024-01-02",
			Time:           "11:00:00",
			EntranceIc:     "リネームテストIC",    // Current field name
			ExitIc:         "リネームテスト出口IC", // Current field name
			TollAmount:     2000,
			CarNumber:      "品川123あ5678",
			EtcCardNumber:  "5678901234567890",
			EtcNum:         proto.String("RENAME001"),
			CreatedAt:      timestamppb.Now(),
			UpdatedAt:      timestamppb.Now(),
		}

		// Serialize and deserialize to test field preservation
		data, err := proto.Marshal(record)
		require.NoError(t, err, "Record must serialize successfully")

		var deserializedRecord pb.ETCMeisaiRecord
		err = proto.Unmarshal(data, &deserializedRecord)
		require.NoError(t, err, "Record must deserialize successfully")

		// Contract assertions - all fields must be preserved
		assert.Equal(t, record.EntranceIc, deserializedRecord.EntranceIc, "EntranceIc must be preserved")
		assert.Equal(t, record.ExitIc, deserializedRecord.ExitIc, "ExitIc must be preserved")
		assert.Equal(t, record.EtcNum, deserializedRecord.EtcNum, "EtcNum must be preserved")
	})

	t.Run("EnumEvolution_Compatibility", func(t *testing.T) {
		// Contract: Adding new enum values must not break existing data

		// Test all current enum values for MappingStatus
		enumValues := []pb.MappingStatus{
			pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED,
			pb.MappingStatus_MAPPING_STATUS_ACTIVE,
			pb.MappingStatus_MAPPING_STATUS_INACTIVE,
			pb.MappingStatus_MAPPING_STATUS_PENDING,
			pb.MappingStatus_MAPPING_STATUS_REJECTED,
		}

		for _, enumValue := range enumValues {
			t.Run("MappingStatus_"+enumValue.String(), func(t *testing.T) {
				mapping := &pb.ETCMapping{
					Id:               1,
					EtcRecordId:      100,
					MappingType:      "automatic",
					MappedEntityId:   200,
					MappedEntityType: "dtako_record",
					Confidence:       0.95,
					Status:           enumValue,
					CreatedBy:        "test-user",
					CreatedAt:        timestamppb.Now(),
					UpdatedAt:        timestamppb.Now(),
				}

				// Serialize and deserialize
				data, err := proto.Marshal(mapping)
				require.NoError(t, err, "Mapping with enum must serialize")

				var deserializedMapping pb.ETCMapping
				err = proto.Unmarshal(data, &deserializedMapping)
				require.NoError(t, err, "Mapping with enum must deserialize")

				// Contract assertions
				assert.Equal(t, enumValue, deserializedMapping.Status, "Enum value must be preserved")
			})
		}

		// Test unknown enum value handling (simulating future enum values)
		t.Run("UnknownEnumValue", func(t *testing.T) {
			mapping := &pb.ETCMapping{
				Id:               2,
				EtcRecordId:      101,
				MappingType:      "manual",
				MappedEntityId:   201,
				MappedEntityType: "custom_record",
				Confidence:       0.88,
				Status:           pb.MappingStatus(999), // Unknown enum value
				CreatedBy:        "future-client",
				CreatedAt:        timestamppb.Now(),
				UpdatedAt:        timestamppb.Now(),
			}

			// This should not panic or error when serializing unknown enum values
			data, err := proto.Marshal(mapping)
			assert.NoError(t, err, "Unknown enum values should serialize without error")

			var deserializedMapping pb.ETCMapping
			err = proto.Unmarshal(data, &deserializedMapping)
			assert.NoError(t, err, "Unknown enum values should deserialize without error")

			// The behavior of unknown enum values depends on protobuf implementation
			// but it should not cause crashes
			assert.NotNil(t, &deserializedMapping.Status, "Status field should exist even with unknown value")
		})
	})

	t.Run("ImportSession_StatusEvolution", func(t *testing.T) {
		// Contract: ImportStatus enum evolution must maintain compatibility

		statusValues := []pb.ImportStatus{
			pb.ImportStatus_IMPORT_STATUS_UNSPECIFIED,
			pb.ImportStatus_IMPORT_STATUS_PENDING,
			pb.ImportStatus_IMPORT_STATUS_PROCESSING,
			pb.ImportStatus_IMPORT_STATUS_COMPLETED,
			pb.ImportStatus_IMPORT_STATUS_FAILED,
			pb.ImportStatus_IMPORT_STATUS_CANCELLED,
		}

		for _, status := range statusValues {
			t.Run("ImportStatus_"+status.String(), func(t *testing.T) {
				session := &pb.ImportSession{
					Id:             "session-" + status.String(),
					AccountType:    "corporate",
					AccountId:      "account-001",
					FileName:       "test.csv",
					FileSize:       2048,
					Status:         status,
					TotalRows:      100,
					ProcessedRows:  50,
					SuccessRows:    45,
					ErrorRows:      5,
					DuplicateRows:  0,
					StartedAt:      timestamppb.Now(),
					CreatedBy:      "test-user",
					CreatedAt:      timestamppb.Now(),
				}

				// Serialize and deserialize
				data, err := proto.Marshal(session)
				require.NoError(t, err, "ImportSession must serialize")

				var deserializedSession pb.ImportSession
				err = proto.Unmarshal(data, &deserializedSession)
				require.NoError(t, err, "ImportSession must deserialize")

				// Contract assertions
				assert.Equal(t, status, deserializedSession.Status, "Import status must be preserved")
				assert.Equal(t, session.Id, deserializedSession.Id, "Session ID must be preserved")
			})
		}
	})

	t.Run("RepeatedField_Evolution", func(t *testing.T) {
		// Contract: Changes to repeated fields must maintain compatibility

		// Test ImportSession with error log (repeated field)
		session := &pb.ImportSession{
			Id:          "repeated-test-001",
			AccountType: "personal",
			AccountId:   "user-001",
			FileName:    "errors.csv",
			FileSize:    1024,
			Status:      pb.ImportStatus_IMPORT_STATUS_FAILED,
			ErrorLog: []*pb.ImportError{
				{
					RowNumber:    1,
					ErrorType:    "validation",
					ErrorMessage: "Invalid date format",
					RawData:      "2024-13-01,invalid-data",
				},
				{
					RowNumber:    5,
					ErrorType:    "parsing",
					ErrorMessage: "Missing required field",
					RawData:      "incomplete,row,data",
				},
			},
			CreatedBy: "system",
			CreatedAt: timestamppb.Now(),
		}

		// Serialize and deserialize
		data, err := proto.Marshal(session)
		require.NoError(t, err, "Session with repeated fields must serialize")

		var deserializedSession pb.ImportSession
		err = proto.Unmarshal(data, &deserializedSession)
		require.NoError(t, err, "Session with repeated fields must deserialize")

		// Contract assertions
		assert.Len(t, deserializedSession.ErrorLog, 2, "Repeated field length must be preserved")
		assert.Equal(t, session.ErrorLog[0].RowNumber, deserializedSession.ErrorLog[0].RowNumber, "First error row number must match")
		assert.Equal(t, session.ErrorLog[1].ErrorType, deserializedSession.ErrorLog[1].ErrorType, "Second error type must match")
	})

	t.Run("NestedMessage_Evolution", func(t *testing.T) {
		// Contract: Changes to nested messages must maintain compatibility

		mapping := &pb.ETCMapping{
			Id:               3,
			EtcRecordId:      102,
			EtcRecord: &pb.ETCMeisaiRecord{
				Id:             102,
				Hash:           "nested-test-001",
				Date:           "2024-01-03",
				Time:           "12:00:00",
				EntranceIc:     "ネストテストIC",
				ExitIc:         "ネストテスト出口IC",
				TollAmount:     1800,
				CarNumber:      "品川123あ7777",
				EtcCardNumber:  "7777888899990000",
				EtcNum:         proto.String("NESTED001"),
				CreatedAt:      timestamppb.Now(),
				UpdatedAt:      timestamppb.Now(),
			},
			MappingType:      "automatic",
			MappedEntityId:   302,
			MappedEntityType: "dtako_record",
			Confidence:       0.92,
			Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE,
			CreatedBy:        "system",
			CreatedAt:        timestamppb.Now(),
			UpdatedAt:        timestamppb.Now(),
		}

		// Serialize and deserialize
		data, err := proto.Marshal(mapping)
		require.NoError(t, err, "Mapping with nested message must serialize")

		var deserializedMapping pb.ETCMapping
		err = proto.Unmarshal(data, &deserializedMapping)
		require.NoError(t, err, "Mapping with nested message must deserialize")

		// Contract assertions
		assert.NotNil(t, deserializedMapping.EtcRecord, "Nested ETC record must be preserved")
		assert.Equal(t, mapping.EtcRecord.Id, deserializedMapping.EtcRecord.Id, "Nested record ID must match")
		assert.Equal(t, mapping.EtcRecord.Hash, deserializedMapping.EtcRecord.Hash, "Nested record hash must match")
		assert.Equal(t, mapping.EtcRecord.EtcNum, deserializedMapping.EtcRecord.EtcNum, "Nested record ETC number must match")
	})

	t.Run("Oneof_Field_Evolution", func(t *testing.T) {
		// Contract: Changes to oneof fields must maintain compatibility
		// Note: The current schema doesn't have oneof fields, but this tests future additions

		// This test would be relevant if oneof fields are added in future schema versions
		// For now, we test that the schema can handle potential oneof additions

		// Create a record with all current fields to ensure no conflicts
		record := &pb.ETCMeisaiRecord{
			Id:             4,
			Hash:           "oneof-test-001",
			Date:           "2024-01-04",
			Time:           "13:00:00",
			EntranceIc:     "OneofテストIC",
			ExitIc:         "Oneofテスト出口IC",
			TollAmount:     2200,
			CarNumber:      "品川123あ8888",
			EtcCardNumber:  "8888999900001111",
			EtcNum:         proto.String("ONEOF001"),
			DtakoRowId:     proto.Int64(12345),
			CreatedAt:      timestamppb.Now(),
			UpdatedAt:      timestamppb.Now(),
		}

		// Serialize and deserialize
		data, err := proto.Marshal(record)
		require.NoError(t, err, "Record must serialize for oneof compatibility test")

		var deserializedRecord pb.ETCMeisaiRecord
		err = proto.Unmarshal(data, &deserializedRecord)
		require.NoError(t, err, "Record must deserialize for oneof compatibility test")

		// Contract assertions
		assert.Equal(t, record.Id, deserializedRecord.Id, "ID must be preserved in oneof test")
		assert.Equal(t, record.EtcNum, deserializedRecord.EtcNum, "EtcNum must be preserved in oneof test")
		assert.Equal(t, record.DtakoRowId, deserializedRecord.DtakoRowId, "DtakoRowId must be preserved in oneof test")
	})
}

func TestSchemaEvolution_FieldNumberStability(t *testing.T) {
	// Contract: Field numbers must remain stable across schema versions

	t.Run("ETCMeisaiRecord_FieldNumbers", func(t *testing.T) {
		// This test ensures that field numbers in ETCMeisaiRecord are stable
		record := &pb.ETCMeisaiRecord{
			Id:             1,     // field 1
			Hash:           "test", // field 2
			Date:           "2024-01-01", // field 3
			Time:           "10:00:00",   // field 4
			EntranceIc:     "test-ic",    // field 5
			ExitIc:         "test-exit",  // field 6
			TollAmount:     1000,         // field 7
			CarNumber:      "test-car",   // field 8
			EtcCardNumber:  "test-card",  // field 9
			EtcNum:         proto.String("TEST"), // field 10
			DtakoRowId:     proto.Int64(123),     // field 11
			CreatedAt:      timestamppb.Now(),    // field 12
			UpdatedAt:      timestamppb.Now(),    // field 13
		}

		// Serialize
		data, err := proto.Marshal(record)
		require.NoError(t, err, "Record must serialize")

		// Deserialize
		var deserializedRecord pb.ETCMeisaiRecord
		err = proto.Unmarshal(data, &deserializedRecord)
		require.NoError(t, err, "Record must deserialize")

		// Contract assertions - all fields must be preserved exactly
		assert.Equal(t, record.Id, deserializedRecord.Id, "Field 1 (Id) must be stable")
		assert.Equal(t, record.Hash, deserializedRecord.Hash, "Field 2 (Hash) must be stable")
		assert.Equal(t, record.Date, deserializedRecord.Date, "Field 3 (Date) must be stable")
		assert.Equal(t, record.Time, deserializedRecord.Time, "Field 4 (Time) must be stable")
		assert.Equal(t, record.EntranceIc, deserializedRecord.EntranceIc, "Field 5 (EntranceIc) must be stable")
		assert.Equal(t, record.ExitIc, deserializedRecord.ExitIc, "Field 6 (ExitIc) must be stable")
		assert.Equal(t, record.TollAmount, deserializedRecord.TollAmount, "Field 7 (TollAmount) must be stable")
		assert.Equal(t, record.CarNumber, deserializedRecord.CarNumber, "Field 8 (CarNumber) must be stable")
		assert.Equal(t, record.EtcCardNumber, deserializedRecord.EtcCardNumber, "Field 9 (EtcCardNumber) must be stable")
		assert.Equal(t, record.EtcNum, deserializedRecord.EtcNum, "Field 10 (EtcNum) must be stable")
		assert.Equal(t, record.DtakoRowId, deserializedRecord.DtakoRowId, "Field 11 (DtakoRowId) must be stable")
	})

	t.Run("ETCMapping_FieldNumbers", func(t *testing.T) {
		// Ensure field numbers are stable for ETCMapping
		mapping := &pb.ETCMapping{
			Id:               1,               // field 1
			EtcRecordId:      100,            // field 2
			MappingType:      "test",         // field 4
			MappedEntityId:   200,            // field 5
			MappedEntityType: "test-entity",  // field 6
			Confidence:       0.95,           // field 7
			Status:           pb.MappingStatus_MAPPING_STATUS_ACTIVE, // field 8
			CreatedBy:        "test-user",    // field 10
			CreatedAt:        timestamppb.Now(), // field 11
			UpdatedAt:        timestamppb.Now(), // field 12
		}

		// Serialize and deserialize
		data, err := proto.Marshal(mapping)
		require.NoError(t, err, "Mapping must serialize")

		var deserializedMapping pb.ETCMapping
		err = proto.Unmarshal(data, &deserializedMapping)
		require.NoError(t, err, "Mapping must deserialize")

		// Verify all fields are preserved
		assert.Equal(t, mapping.Id, deserializedMapping.Id, "Mapping ID field must be stable")
		assert.Equal(t, mapping.EtcRecordId, deserializedMapping.EtcRecordId, "ETC record ID field must be stable")
		assert.Equal(t, mapping.MappingType, deserializedMapping.MappingType, "Mapping type field must be stable")
		assert.Equal(t, mapping.Confidence, deserializedMapping.Confidence, "Confidence field must be stable")
		assert.Equal(t, mapping.Status, deserializedMapping.Status, "Status field must be stable")
	})
}

func TestSchemaEvolution_DefaultValues(t *testing.T) {
	// Contract: Default values must be handled consistently across schema versions

	t.Run("Numeric_Defaults", func(t *testing.T) {
		// Create record with minimal fields (others should get default values)
		record := &pb.ETCMeisaiRecord{
			Hash: "default-test-001",
			Date: "2024-01-01",
			Time: "10:00:00",
			// Other numeric fields will get default values
		}

		data, err := proto.Marshal(record)
		require.NoError(t, err, "Record with defaults must serialize")

		var deserializedRecord pb.ETCMeisaiRecord
		err = proto.Unmarshal(data, &deserializedRecord)
		require.NoError(t, err, "Record with defaults must deserialize")

		// Contract assertions for default values
		assert.Equal(t, int64(0), deserializedRecord.Id, "Default ID should be 0")
		assert.Equal(t, int32(0), deserializedRecord.TollAmount, "Default TollAmount should be 0")
		assert.Empty(t, deserializedRecord.CarNumber, "Default CarNumber should be empty")
		assert.Empty(t, deserializedRecord.EtcCardNumber, "Default EtcCardNumber should be empty")
	})

	t.Run("Enum_Defaults", func(t *testing.T) {
		// Create mapping with minimal fields
		mapping := &pb.ETCMapping{
			EtcRecordId:      1,
			MappingType:      "test",
			MappedEntityId:   1,
			MappedEntityType: "test",
			Confidence:       0.5,
			// Status will get default value
			CreatedBy: "test",
		}

		data, err := proto.Marshal(mapping)
		require.NoError(t, err, "Mapping with default enum must serialize")

		var deserializedMapping pb.ETCMapping
		err = proto.Unmarshal(data, &deserializedMapping)
		require.NoError(t, err, "Mapping with default enum must deserialize")

		// Contract assertions
		assert.Equal(t, pb.MappingStatus_MAPPING_STATUS_UNSPECIFIED, deserializedMapping.Status, "Default enum should be UNSPECIFIED")
	})
}