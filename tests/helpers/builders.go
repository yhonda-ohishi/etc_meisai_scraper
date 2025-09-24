package helpers

import (
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

// ETCMeisaiRecordBuilder builds test ETC Meisai records
type ETCMeisaiRecordBuilder struct {
	record *models.ETCMeisaiRecord
}

// NewETCMeisaiRecordBuilder creates a new builder with default values
func NewETCMeisaiRecordBuilder() *ETCMeisaiRecordBuilder {
	etcNum := "1234567890"
	return &ETCMeisaiRecordBuilder{
		record: &models.ETCMeisaiRecord{
			ETCNum:        &etcNum,
			Date:          time.Now(),
			Time:          "10:00",
			EntranceIC:    "東京IC",
			ExitIC:        "大阪IC",
			TollAmount:    1000,
			CarNumber:     "品川300あ1234",
			ETCCardNumber: "4321098765432109",
			Hash:          "testhash123",
		},
	}
}

// WithETCNum sets the ETC number
func (b *ETCMeisaiRecordBuilder) WithETCNum(etcNum string) *ETCMeisaiRecordBuilder {
	b.record.ETCNum = &etcNum
	return b
}

// WithDate sets the date
func (b *ETCMeisaiRecordBuilder) WithDate(date time.Time) *ETCMeisaiRecordBuilder {
	b.record.Date = date
	return b
}

// WithTollAmount sets the toll amount
func (b *ETCMeisaiRecordBuilder) WithTollAmount(amount int) *ETCMeisaiRecordBuilder {
	b.record.TollAmount = amount
	return b
}

// WithEntranceIC sets the entrance IC
func (b *ETCMeisaiRecordBuilder) WithEntranceIC(ic string) *ETCMeisaiRecordBuilder {
	b.record.EntranceIC = ic
	return b
}

// WithExitIC sets the exit IC
func (b *ETCMeisaiRecordBuilder) WithExitIC(ic string) *ETCMeisaiRecordBuilder {
	b.record.ExitIC = ic
	return b
}

// Build returns the built record
func (b *ETCMeisaiRecordBuilder) Build() *models.ETCMeisaiRecord {
	return b.record
}

// ETCMappingBuilder builds test ETC mappings
type ETCMappingBuilder struct {
	mapping *models.ETCMapping
}

// NewETCMappingBuilder creates a new builder with default values
func NewETCMappingBuilder() *ETCMappingBuilder {
	return &ETCMappingBuilder{
		mapping: &models.ETCMapping{
			ETCRecordID:      1,
			MappingType:      "vehicle",
			MappedEntityID:   1,
			MappedEntityType: "vehicle",
			Confidence:       1.0,
			Status:           "active",
			CreatedBy:        "test",
		},
	}
}

// WithETCRecordID sets the ETC record ID
func (b *ETCMappingBuilder) WithETCRecordID(id int64) *ETCMappingBuilder {
	b.mapping.ETCRecordID = id
	return b
}

// WithMappingType sets the mapping type
func (b *ETCMappingBuilder) WithMappingType(mappingType string) *ETCMappingBuilder {
	b.mapping.MappingType = mappingType
	return b
}

// WithMappedEntityID sets the mapped entity ID
func (b *ETCMappingBuilder) WithMappedEntityID(id int64) *ETCMappingBuilder {
	b.mapping.MappedEntityID = id
	return b
}

// WithStatus sets the status
func (b *ETCMappingBuilder) WithStatus(status string) *ETCMappingBuilder {
	b.mapping.Status = status
	return b
}

// Build returns the built mapping
func (b *ETCMappingBuilder) Build() *models.ETCMapping {
	return b.mapping
}

// ImportSessionBuilder builds test import sessions
type ImportSessionBuilder struct {
	session *models.ImportSession
}

// NewImportSessionBuilder creates a new builder with default values
func NewImportSessionBuilder() *ImportSessionBuilder {
	return &ImportSessionBuilder{
		session: &models.ImportSession{
			ID:            "test-session-001",
			AccountType:   "corporate",
			AccountID:     "ACC001",
			FileName:      "test.csv",
			FileSize:      1024,
			Status:        "pending",
			TotalRows:     100,
			ProcessedRows: 0,
		},
	}
}

// WithID sets the session ID
func (b *ImportSessionBuilder) WithID(id string) *ImportSessionBuilder {
	b.session.ID = id
	return b
}

// WithStatus sets the status
func (b *ImportSessionBuilder) WithStatus(status string) *ImportSessionBuilder {
	b.session.Status = status
	return b
}

// WithTotalRows sets the total rows count
func (b *ImportSessionBuilder) WithTotalRows(total int) *ImportSessionBuilder {
	b.session.TotalRows = total
	return b
}

// Build returns the built session
func (b *ImportSessionBuilder) Build() *models.ImportSession {
	return b.session
}