// T013-D: Test factory pattern demonstration and validation
package fixtures

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
)

func TestTestFactory_DeterministicGeneration(t *testing.T) {
	// Test deterministic generation with seed
	factory1 := NewTestFactory(12345)
	factory2 := NewTestFactory(12345)

	record1 := factory1.CreateETCMeisaiRecord()
	record2 := factory2.CreateETCMeisaiRecord()

	// Should generate identical random values with same seed
	assert.Equal(t, record1.CarNumber, record2.CarNumber)
	assert.Equal(t, record1.ETCCardNumber, record2.ETCCardNumber)
	assert.Equal(t, record1.TollAmount, record2.TollAmount)
}

func TestTestFactory_SequentialIDs(t *testing.T) {
	factory := NewTestFactory()

	// Test sequential ID generation
	id1 := factory.NextID("test_entity")
	id2 := factory.NextID("test_entity")
	id3 := factory.NextID("test_entity")

	assert.Equal(t, 1, id1)
	assert.Equal(t, 2, id2)
	assert.Equal(t, 3, id3)

	// Different entity types have separate counters
	otherId1 := factory.NextID("other_entity")
	assert.Equal(t, 1, otherId1)
}

func TestTestFactory_RandomGenerators(t *testing.T) {
	factory := NewTestFactory(42)

	// Test random string generation
	str1 := factory.RandomString(10)
	str2 := factory.RandomString(10)
	assert.Len(t, str1, 10)
	assert.Len(t, str2, 10)
	assert.NotEqual(t, str1, str2)

	// Test random int generation
	for i := 0; i < 100; i++ {
		num := factory.RandomInt(10, 20)
		assert.GreaterOrEqual(t, num, 10)
		assert.LessOrEqual(t, num, 20)
	}

	// Test random amount generation
	amount := factory.RandomAmount()
	assert.GreaterOrEqual(t, amount, 300)
	assert.LessOrEqual(t, amount, 5000)

	// Test random IC generation
	ic := factory.RandomIC()
	assert.NotEmpty(t, ic)
	assert.Contains(t, ic, "IC")

	// Test random car number generation
	carNumber := factory.RandomCarNumber()
	assert.NotEmpty(t, carNumber)
	assert.Greater(t, len(carNumber), 5)

	// Test random ETC number generation
	etcNumber := factory.RandomETCNumber()
	assert.Regexp(t, `^\d{4}-\d{4}-\d{4}-\d{4}$`, etcNumber)
}

func TestTestFactory_CreateWithOptions(t *testing.T) {
	factory := NewTestFactory()

	// Test creating with options
	customDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	record := factory.CreateETCMeisaiRecord(func(r *models.ETCMeisaiRecord) {
		r.Date = customDate
		r.CarNumber = "Custom123"
		r.TollAmount = 9999
	})

	assert.Equal(t, customDate, record.Date)
	assert.Equal(t, "Custom123", record.CarNumber)
	assert.Equal(t, 9999, record.TollAmount)
}

func TestTestFactory_BatchCreation(t *testing.T) {
	factory := NewTestFactory(100)

	// Test batch creation
	records := factory.CreateETCMeisaiRecordBatch(50)
	assert.Len(t, records, 50)

	// Verify all records are unique
	idSet := make(map[int64]bool)
	for _, record := range records {
		assert.False(t, idSet[record.ID], "Duplicate ID found")
		idSet[record.ID] = true
	}

	// Test batch creation with options
	batchWithOptions := factory.CreateETCMeisaiBatch(10, func(m *models.ETCMeisai) {
		m.Amount = 1000
	})
	assert.Len(t, batchWithOptions, 10)
	for _, record := range batchWithOptions {
		assert.Equal(t, int32(1000), record.Amount)
	}
}

func TestETCMeisaiBuilder(t *testing.T) {
	factory := NewTestFactory()

	// Test builder pattern
	date := time.Date(2024, 7, 1, 10, 30, 0, 0, time.UTC)
	record := factory.NewETCMeisaiBuilder().
		WithDate(date).
		WithRoute("Tokyo IC", "Yokohama IC").
		WithAmount(1500).
		WithCarNumber("品川500あ1234").
		WithETCNumber("1111-2222-3333-4444").
		Build()

	assert.Equal(t, date, record.UseDate)
	assert.Equal(t, "Tokyo IC", record.EntryIC)
	assert.Equal(t, "Yokohama IC", record.ExitIC)
	assert.Equal(t, int32(1500), record.Amount)
	assert.Equal(t, "品川500あ1234", record.CarNumber)
	assert.Equal(t, "1111-2222-3333-4444", record.ETCNumber)
}

func TestImportSessionBuilder(t *testing.T) {
	factory := NewTestFactory()

	// Test import session builder
	session := factory.NewImportSessionBuilder().
		WithFileName("test_import.csv").
		WithAccountType("corporate").
		WithRows(1000, 950, 900, 50, 0).
		WithStatus("completed").
		Build()

	assert.Equal(t, "test_import.csv", session.FileName)
	assert.Equal(t, "corporate", session.AccountType)
	assert.Equal(t, 1000, session.TotalRows)
	assert.Equal(t, 950, session.ProcessedRows)
	assert.Equal(t, 900, session.SuccessRows)
	assert.Equal(t, 50, session.ErrorRows)
	assert.Equal(t, 0, session.DuplicateRows)
	assert.Equal(t, "completed", session.Status)
}

func TestScenarios_SuccessfulImport(t *testing.T) {
	factory := NewTestFactory(200)
	scenarios := factory.Scenarios()

	session, records := scenarios.SuccessfulImport()

	assert.Equal(t, "completed", session.Status)
	assert.Equal(t, 100, session.TotalRows)
	assert.Equal(t, 95, session.SuccessRows)
	assert.Len(t, records, 95)
}

func TestScenarios_FailedImport(t *testing.T) {
	factory := NewTestFactory()
	scenarios := factory.Scenarios()

	session, records := scenarios.FailedImport()

	assert.Equal(t, "failed", session.Status)
	assert.Equal(t, 0, session.SuccessRows)
	assert.Nil(t, records)
}

func TestScenarios_MappingConflict(t *testing.T) {
	factory := NewTestFactory()
	scenarios := factory.Scenarios()

	records, mappings := scenarios.MappingConflict()

	assert.Len(t, records, 3)
	assert.Len(t, mappings, 3)

	// Verify all records have the same dtako_row_id
	dtakoID := records[0].DtakoRowID
	require.NotNil(t, dtakoID)
	for _, record := range records {
		assert.Equal(t, *dtakoID, *record.DtakoRowID)
	}

	// Verify mappings point to the same entity
	for _, mapping := range mappings {
		assert.Equal(t, *dtakoID, mapping.MappedEntityID)
	}
}

func TestScenarios_DuplicateRecords(t *testing.T) {
	factory := NewTestFactory(300)
	scenarios := factory.Scenarios()

	duplicates := scenarios.DuplicateRecords()

	assert.Len(t, duplicates, 3)

	// Verify key fields are identical
	first := duplicates[0]
	for _, dup := range duplicates[1:] {
		assert.Equal(t, first.Date, dup.Date)
		assert.Equal(t, first.Time, dup.Time)
		assert.Equal(t, first.EntranceIC, dup.EntranceIC)
		assert.Equal(t, first.ExitIC, dup.ExitIC)
		assert.Equal(t, first.CarNumber, dup.CarNumber)
		assert.Equal(t, first.TollAmount, dup.TollAmount)
		assert.Equal(t, first.Hash, dup.Hash)
	}
}

func TestScenarios_LargeDataset(t *testing.T) {
	factory := NewTestFactory()
	scenarios := factory.Scenarios()

	records, mappings := scenarios.LargeDataset(1000)

	assert.Len(t, records, 1000)
	assert.Len(t, mappings, 1000)

	// Verify each record has a corresponding mapping
	recordIDMap := make(map[int64]bool)
	for _, record := range records {
		recordIDMap[record.ID] = true
	}

	for _, mapping := range mappings {
		assert.True(t, recordIDMap[mapping.ETCRecordID])
	}
}

func TestScenarios_HighConfidenceMatches(t *testing.T) {
	factory := NewTestFactory()
	scenarios := factory.Scenarios()

	records, mappings := scenarios.HighConfidenceMatches()

	assert.Len(t, records, 10)
	assert.Len(t, mappings, 10)

	// Verify all mappings have high confidence
	for _, mapping := range mappings {
		assert.GreaterOrEqual(t, mapping.Confidence, float32(0.95))
		assert.Equal(t, "active", mapping.Status)
	}
}

func TestScenarios_MixedStatuses(t *testing.T) {
	factory := NewTestFactory()
	scenarios := factory.Scenarios()

	sessions := scenarios.MixedStatuses()

	expectedStatuses := []string{"pending", "processing", "completed", "failed", "cancelled"}
	assert.Len(t, sessions, len(expectedStatuses))

	// Verify we have all expected statuses
	statusMap := make(map[string]bool)
	for _, session := range sessions {
		statusMap[session.Status] = true
	}

	for _, status := range expectedStatuses {
		assert.True(t, statusMap[status], "Missing status: %s", status)
	}
}

// Benchmark tests
func BenchmarkFactory_CreateSingleRecord(b *testing.B) {
	factory := NewTestFactory(42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = factory.CreateETCMeisaiRecord()
	}
}

func BenchmarkFactory_CreateBatch100(b *testing.B) {
	factory := NewTestFactory(42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = factory.CreateETCMeisaiRecordBatch(100)
	}
}

func BenchmarkFactory_LargeDataset(b *testing.B) {
	factory := NewTestFactory(42)
	scenarios := factory.Scenarios()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = scenarios.LargeDataset(100)
	}
}