package models_test

import (
	"testing"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestMappingStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   models.MappingStatus
		expected string
	}{
		{
			name:     "active status",
			status:   models.MappingStatusActive,
			expected: "active",
		},
		{
			name:     "inactive status",
			status:   models.MappingStatusInactive,
			expected: "inactive",
		},
		{
			name:     "pending status",
			status:   models.MappingStatusPending,
			expected: "pending",
		},
		{
			name:     "rejected status",
			status:   models.MappingStatusRejected,
			expected: "rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.AssertEqual(t, tt.expected, string(tt.status))
		})
	}
}

func TestMappingType_Constants(t *testing.T) {
	tests := []struct {
		name        string
		mappingType models.MappingType
		expected    string
	}{
		{
			name:        "dtako type",
			mappingType: models.MappingTypeDtako,
			expected:    "dtako",
		},
		{
			name:        "expense type",
			mappingType: models.MappingTypeExpense,
			expected:    "expense",
		},
		{
			name:        "invoice type",
			mappingType: models.MappingTypeInvoice,
			expected:    "invoice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.AssertEqual(t, tt.expected, string(tt.mappingType))
		})
	}
}

func TestMappedEntityType_Constants(t *testing.T) {
	tests := []struct {
		name       string
		entityType models.MappedEntityType
		expected   string
	}{
		{
			name:       "dtako record type",
			entityType: models.EntityTypeDtakoRecord,
			expected:   "dtako_record",
		},
		{
			name:       "expense record type",
			entityType: models.EntityTypeExpenseRecord,
			expected:   "expense_record",
		},
		{
			name:       "invoice record type",
			entityType: models.EntityTypeInvoiceRecord,
			expected:   "invoice_record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.AssertEqual(t, tt.expected, string(tt.entityType))
		})
	}
}

func TestETCMapping_BeforeCreate(t *testing.T) {
	mapping := &models.ETCMapping{
		ETCRecordID:      1,
		MappingType:      string(models.MappingTypeDtako),
		MappedEntityID:   123,
		MappedEntityType: string(models.EntityTypeDtakoRecord),
		Confidence:       0.95,
	}

	err := mapping.BeforeCreate(nil)
	helpers.AssertNoError(t, err)

	// Should set timestamps
	helpers.AssertFalse(t, mapping.CreatedAt.IsZero())
	helpers.AssertFalse(t, mapping.UpdatedAt.IsZero())

	// Should set default status if empty
	helpers.AssertEqual(t, string(models.MappingStatusActive), mapping.Status)
}

func TestETCMapping_BeforeUpdate(t *testing.T) {
	mapping := &models.ETCMapping{
		ETCRecordID:      1,
		MappingType:      string(models.MappingTypeDtako),
		MappedEntityID:   123,
		MappedEntityType: string(models.EntityTypeDtakoRecord),
		Confidence:       0.95,
		Status:           string(models.MappingStatusActive),
		CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	oldUpdatedAt := mapping.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure time difference

	err := mapping.BeforeUpdate()
	helpers.AssertNoError(t, err)

	// Should update timestamp
	helpers.AssertTrue(t, mapping.UpdatedAt.After(oldUpdatedAt))

	// Should not change created timestamp
	helpers.AssertEqual(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), mapping.CreatedAt)
}

func TestETCMapping_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mapping *models.ETCMapping
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mapping",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: false,
		},
		{
			name: "zero ETC record ID",
			mapping: &models.ETCMapping{
				ETCRecordID:      0,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "ETC record ID must be positive",
		},
		{
			name: "empty mapping type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "",
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "mapping type cannot be empty",
		},
		{
			name: "zero mapped entity ID",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   0,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "mapped entity ID must be positive",
		},
		{
			name: "empty mapped entity type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: "",
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "mapped entity type cannot be empty",
		},
		{
			name: "invalid confidence (negative)",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       -0.1,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "invalid confidence (greater than 1)",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       1.5,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "invalid mapping type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      "invalid_type",
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "invalid mapping type:",
		},
		{
			name: "invalid entity type",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: "invalid_entity",
				Confidence:       0.95,
				Status:           string(models.MappingStatusActive),
			},
			wantErr: true,
			errMsg:  "invalid mapped entity type:",
		},
		{
			name: "invalid status",
			mapping: &models.ETCMapping{
				ETCRecordID:      1,
				MappingType:      string(models.MappingTypeDtako),
				MappedEntityID:   123,
				MappedEntityType: string(models.EntityTypeDtakoRecord),
				Confidence:       0.95,
				Status:           "invalid_status",
			},
			wantErr: true,
			errMsg:  "invalid status:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mapping.Validate()

			if tt.wantErr {
				helpers.AssertError(t, err)
				if tt.errMsg != "" {
					helpers.AssertContains(t, err.Error(), tt.errMsg)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestETCMapping_IsValidMappingType(t *testing.T) {
	tests := []struct {
		name        string
		mappingType string
		expected    bool
	}{
		{
			name:        "valid dtako type",
			mappingType: string(models.MappingTypeDtako),
			expected:    true,
		},
		{
			name:        "valid expense type",
			mappingType: string(models.MappingTypeExpense),
			expected:    true,
		},
		{
			name:        "valid invoice type",
			mappingType: string(models.MappingTypeInvoice),
			expected:    true,
		},
		{
			name:        "invalid type",
			mappingType: "invalid",
			expected:    false,
		},
		{
			name:        "empty type",
			mappingType: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidMappingType(tt.mappingType)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestETCMapping_IsValidEntityType(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		expected   bool
	}{
		{
			name:       "valid dtako record type",
			entityType: string(models.EntityTypeDtakoRecord),
			expected:   true,
		},
		{
			name:       "valid expense record type",
			entityType: string(models.EntityTypeExpenseRecord),
			expected:   true,
		},
		{
			name:       "valid invoice record type",
			entityType: string(models.EntityTypeInvoiceRecord),
			expected:   true,
		},
		{
			name:       "invalid type",
			entityType: "invalid",
			expected:   false,
		},
		{
			name:       "empty type",
			entityType: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidEntityType(tt.entityType)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestETCMapping_IsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "valid active status",
			status:   string(models.MappingStatusActive),
			expected: true,
		},
		{
			name:     "valid inactive status",
			status:   string(models.MappingStatusInactive),
			expected: true,
		},
		{
			name:     "valid pending status",
			status:   string(models.MappingStatusPending),
			expected: true,
		},
		{
			name:     "valid rejected status",
			status:   string(models.MappingStatusRejected),
			expected: true,
		},
		{
			name:     "invalid status",
			status:   "invalid",
			expected: false,
		},
		{
			name:     "empty status",
			status:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidStatus(tt.status)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestETCMapping_GetTableName(t *testing.T) {
	mapping := &models.ETCMapping{}
	tableName := mapping.GetTableName()
	helpers.AssertEqual(t, "etc_mappings", tableName)
}

func TestETCMapping_String(t *testing.T) {
	mapping := &models.ETCMapping{
		ID:               1,
		ETCRecordID:      123,
		MappingType:      string(models.MappingTypeDtako),
		MappedEntityID:   456,
		MappedEntityType: string(models.EntityTypeDtakoRecord),
		Confidence:       0.95,
		Status:           string(models.MappingStatusActive),
	}

	str := mapping.String()

	// Should contain key information
	helpers.AssertContains(t, str, "123")   // ETC Record ID
	helpers.AssertContains(t, str, "dtako") // Mapping Type
	helpers.AssertContains(t, str, "456")   // Mapped Entity ID
	helpers.AssertContains(t, str, "0.95")  // Confidence
}