// T013-D: Enhanced test data factory pattern implementation
// This package provides comprehensive factories for consistent test data creation

package fixtures

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/pb"
	"gorm.io/datatypes"
)

// TestFactory provides factory methods for creating test data with deterministic generation
type TestFactory struct {
	seed    int64
	rand    *rand.Rand
	mu      sync.Mutex
	counter map[string]int
}

// NewTestFactory creates a new test factory with optional seed
func NewTestFactory(seed ...int64) *TestFactory {
	var s int64 = time.Now().UnixNano()
	if len(seed) > 0 {
		s = seed[0]
	}
	return &TestFactory{
		seed:    s,
		rand:    rand.New(rand.NewSource(s)),
		counter: make(map[string]int),
	}
}

// WithSeed sets a specific seed for reproducible test data
func (f *TestFactory) WithSeed(seed int64) *TestFactory {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seed = seed
	f.rand = rand.New(rand.NewSource(seed))
	return f
}

// NextID generates sequential IDs for a given entity type
func (f *TestFactory) NextID(entity string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.counter[entity]++
	return f.counter[entity]
}

// RandomString generates a random string of specified length
func (f *TestFactory) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[f.rand.Intn(len(charset))]
	}
	return string(b)
}

// RandomInt generates a random integer within range
func (f *TestFactory) RandomInt(min, max int) int {
	return min + f.rand.Intn(max-min+1)
}

// RandomAmount generates a random amount in the typical range for tolls
func (f *TestFactory) RandomAmount() int {
	return f.RandomInt(300, 5000)
}

// RandomIC generates a random IC name
func (f *TestFactory) RandomIC() string {
	ics := []string{
		"東京IC", "横浜IC", "名古屋IC", "大阪IC", "神戸IC",
		"京都IC", "川崎IC", "千葉IC", "さいたまIC", "浦和IC",
	}
	return ics[f.rand.Intn(len(ics))]
}

// RandomCarNumber generates a random Japanese car number
func (f *TestFactory) RandomCarNumber() string {
	locations := []string{"品川", "横浜", "名古屋", "大阪", "神戸"}
	types := []string{"300", "500", "100"}
	hiragana := []string{"あ", "い", "う", "え", "お"}

	location := locations[f.rand.Intn(len(locations))]
	carType := types[f.rand.Intn(len(types))]
	hira := hiragana[f.rand.Intn(len(hiragana))]
	number := fmt.Sprintf("%04d", f.rand.Intn(10000))

	return fmt.Sprintf("%s%s%s%s", location, carType, hira, number)
}

// RandomETCNumber generates a random ETC card number
func (f *TestFactory) RandomETCNumber() string {
	return fmt.Sprintf("%04d-%04d-%04d-%04d",
		f.rand.Intn(10000), f.rand.Intn(10000),
		f.rand.Intn(10000), f.rand.Intn(10000))
}

// CreateETCMeisai creates a test ETCMeisai with default or random values
func (f *TestFactory) CreateETCMeisai(opts ...func(*models.ETCMeisai)) *models.ETCMeisai {
	id := f.NextID("etc_meisai")
	now := time.Now()

	etc := &models.ETCMeisai{
		ID:        int64(id),
		UseDate:   now.AddDate(0, 0, -f.rand.Intn(30)),
		UseTime:   fmt.Sprintf("%02d:%02d", f.rand.Intn(24), f.rand.Intn(60)),
		EntryIC:   f.RandomIC(),
		ExitIC:    f.RandomIC(),
		Amount:    int32(f.RandomAmount()),
		CarNumber: f.RandomCarNumber(),
		ETCNumber: f.RandomETCNumber(),
		Hash:      fmt.Sprintf("hash_%s", f.RandomString(10)),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Apply options
	for _, opt := range opts {
		opt(etc)
	}

	return etc
}

// CreateETCMeisaiBatch creates multiple ETCMeisai records
func (f *TestFactory) CreateETCMeisaiBatch(count int, opts ...func(*models.ETCMeisai)) []*models.ETCMeisai {
	records := make([]*models.ETCMeisai, count)
	for i := 0; i < count; i++ {
		records[i] = f.CreateETCMeisai(opts...)
	}
	return records
}

// CreateETCMeisaiRecord creates a test ETCMeisaiRecord model with random values
func (f *TestFactory) CreateETCMeisaiRecord(opts ...func(*models.ETCMeisaiRecord)) *models.ETCMeisaiRecord {
	id := f.NextID("etc_record")
	now := time.Now()
	date := now.AddDate(0, 0, -f.rand.Intn(30))

	record := &models.ETCMeisaiRecord{
		ID:            int64(id),
		Date:          date,
		Time:          fmt.Sprintf("%02d:%02d", f.rand.Intn(24), f.rand.Intn(60)),
		EntranceIC:    f.RandomIC(),
		ExitIC:        f.RandomIC(),
		TollAmount:    f.RandomAmount(),
		CarNumber:     f.RandomCarNumber(),
		ETCCardNumber: f.RandomETCNumber(),
		ETCNum:        ptr(fmt.Sprintf("ETC%03d", f.rand.Intn(1000))),
		DtakoRowID:    ptrInt64(int64(f.rand.Intn(100000))),
		Hash:          fmt.Sprintf("hash_%s", f.RandomString(10)),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Apply options
	for _, opt := range opts {
		opt(record)
	}

	return record
}

// CreateETCMeisaiRecordBatch creates multiple ETCMeisaiRecord records
func (f *TestFactory) CreateETCMeisaiRecordBatch(count int, opts ...func(*models.ETCMeisaiRecord)) []*models.ETCMeisaiRecord {
	records := make([]*models.ETCMeisaiRecord, count)
	for i := 0; i < count; i++ {
		records[i] = f.CreateETCMeisaiRecord(opts...)
	}
	return records
}

// CreateETCMeisaiMapping creates a test mapping with random values
func (f *TestFactory) CreateETCMeisaiMapping(opts ...func(*models.ETCMeisaiMapping)) *models.ETCMeisaiMapping {
	id := f.NextID("mapping")
	now := time.Now()
	mappingTypes := []string{"auto", "manual", "hybrid"}

	mapping := &models.ETCMeisaiMapping{
		ID:          int64(id),
		ETCMeisaiID: int64(f.RandomInt(1, 100)),
		DTakoRowID:  fmt.Sprintf("DTAKO%05d", f.rand.Intn(100000)),
		MappingType: mappingTypes[f.rand.Intn(len(mappingTypes))],
		Confidence:  float32(0.5 + f.rand.Float64()*0.5), // 0.5 to 1.0
		Notes:       fmt.Sprintf("Test mapping %d", id),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Apply options
	for _, opt := range opts {
		opt(mapping)
	}

	return mapping
}

// CreateETCMapping creates a test ETCMapping model with random values
func (f *TestFactory) CreateETCMapping(opts ...func(*models.ETCMapping)) *models.ETCMapping {
	id := f.NextID("etc_mapping")
	now := time.Now()
	mappingTypes := []string{"dtako", "manual", "auto"}
	statuses := []string{"active", "inactive", "pending"}
	entityTypes := []string{"dtako_record", "manual_entry", "import_record"}

	metadata := fmt.Sprintf(`{"source": "%s", "version": %d}`,
		mappingTypes[f.rand.Intn(len(mappingTypes))],
		f.rand.Intn(10))

	mapping := &models.ETCMapping{
		ID:               int64(id),
		ETCRecordID:      int64(f.RandomInt(1, 100)),
		MappingType:      mappingTypes[f.rand.Intn(len(mappingTypes))],
		MappedEntityID:   int64(f.RandomInt(100, 1000)),
		MappedEntityType: entityTypes[f.rand.Intn(len(entityTypes))],
		Confidence:       float32(0.5 + f.rand.Float64()*0.5),
		Status:           statuses[f.rand.Intn(len(statuses))],
		Metadata:         datatypes.JSON(metadata),
		CreatedBy:        fmt.Sprintf("user_%d", f.rand.Intn(10)),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Apply options
	for _, opt := range opts {
		opt(mapping)
	}

	return mapping
}

// CreateImportBatch creates a test import batch with random values
func (f *TestFactory) CreateImportBatch(opts ...func(*models.ETCImportBatch)) *models.ETCImportBatch {
	id := f.NextID("import_batch")
	now := time.Now()
	importTypes := []string{"manual", "auto", "scheduled"}
	statuses := []string{"pending", "processing", "completed", "failed"}

	totalRows := f.RandomInt(50, 500)
	processedRows := f.RandomInt(0, totalRows)
	errorCount := f.RandomInt(0, processedRows/10)
	successCount := processedRows - errorCount

	batch := &models.ETCImportBatch{
		ID:            int64(id),
		FileName:      fmt.Sprintf("import_%d_%s.csv", id, f.RandomString(6)),
		FileSize:      int64(f.RandomInt(512, 10240)),
		AccountID:     fmt.Sprintf("ACC%03d", f.rand.Intn(1000)),
		ImportType:    importTypes[f.rand.Intn(len(importTypes))],
		Status:        statuses[f.rand.Intn(len(statuses))],
		TotalRows:     int64(totalRows),
		ProcessedRows: int64(processedRows),
		SuccessCount:  int64(successCount),
		ErrorCount:    int64(errorCount),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Apply options
	for _, opt := range opts {
		opt(batch)
	}

	return batch
}

// CreateImportSession creates a test import session with random values
func (f *TestFactory) CreateImportSession(opts ...func(*models.ImportSession)) *models.ImportSession {
	id := f.NextID("session")
	now := time.Now()
	accountTypes := []string{"corporate", "personal", "fleet"}
	statuses := []string{"pending", "processing", "completed", "failed", "cancelled"}

	totalRows := f.RandomInt(50, 500)
	processedRows := f.RandomInt(0, totalRows)
	errorRows := f.RandomInt(0, processedRows/10)
	duplicateRows := f.RandomInt(0, processedRows/20)
	successRows := processedRows - errorRows - duplicateRows

	duration := time.Duration(f.RandomInt(1, 30)) * time.Minute
	startedAt := now.Add(-duration)

	session := &models.ImportSession{
		ID:            fmt.Sprintf("session-%03d-%s", id, f.RandomString(6)),
		FileName:      fmt.Sprintf("import_%d_%s.csv", id, f.RandomString(6)),
		FileSize:      int64(f.RandomInt(512, 10240)),
		AccountType:   accountTypes[f.rand.Intn(len(accountTypes))],
		AccountID:     fmt.Sprintf("ACC%03d", f.rand.Intn(1000)),
		Status:        statuses[f.rand.Intn(len(statuses))],
		TotalRows:     totalRows,
		ProcessedRows: processedRows,
		SuccessRows:   successRows,
		ErrorRows:     errorRows,
		DuplicateRows: duplicateRows,
		CreatedBy:     fmt.Sprintf("user_%d", f.rand.Intn(10)),
		StartedAt:     startedAt,
		CompletedAt:   ptrTime(now),
		CreatedAt:     startedAt,
	}

	// Apply options
	for _, opt := range opts {
		opt(session)
	}

	return session
}

// CreateProtoETCMeisaiRecord creates a test proto ETCMeisaiRecord with random values
func (f *TestFactory) CreateProtoETCMeisaiRecord(opts ...func(*pb.ETCMeisaiRecord)) *pb.ETCMeisaiRecord {
	id := int64(f.NextID("proto_record"))
	date := time.Now().AddDate(0, 0, -f.rand.Intn(30))

	record := &pb.ETCMeisaiRecord{
		Id:            id,
		Date:          date.Format("2006-01-02"),
		Time:          fmt.Sprintf("%02d:%02d", f.rand.Intn(24), f.rand.Intn(60)),
		EntranceIc:    f.RandomIC(),
		ExitIc:        f.RandomIC(),
		TollAmount:    int32(f.RandomAmount()),
		CarNumber:     f.RandomCarNumber(),
		EtcCardNumber: f.RandomETCNumber(),
		EtcNum:        ptr(fmt.Sprintf("ETC%03d", f.rand.Intn(1000))),
		DtakoRowId:    ptrInt64(int64(f.rand.Intn(100000))),
		Hash:          fmt.Sprintf("hash_%s", f.RandomString(10)),
	}

	// Apply options
	for _, opt := range opts {
		opt(record)
	}

	return record
}

// CreateValidationError creates a test validation error
func (f *TestFactory) CreateValidationError(field, message string) models.ValidationError {
	return models.ValidationError{
		Field:   field,
		Message: message,
	}
}

// CreateETCSummary creates a test summary with random values
func (f *TestFactory) CreateETCSummary(opts ...func(*models.ETCSummary)) *models.ETCSummary {
	totalCount := f.RandomInt(50, 500)
	totalAmount := totalCount * f.RandomInt(500, 2000)
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	summary := &models.ETCSummary{
		TotalCount:  int64(totalCount),
		TotalAmount: int64(totalAmount),
		StartDate:   startDate,
		EndDate:     endDate,
		ByETCNumber: make(map[string]*models.ETCNumberSummary),
		ByMonth:     make(map[string]*models.ETCMonthlySummary),
	}

	// Add some sample ETC number summaries
	for i := 0; i < 3; i++ {
		etcNum := f.RandomETCNumber()
		count := f.RandomInt(10, 50)
		summary.ByETCNumber[etcNum] = &models.ETCNumberSummary{
			ETCNumber:   etcNum,
			TotalCount:  int64(count),
			TotalAmount: int64(count * f.RandomInt(500, 2000)),
		}
	}

	// Add monthly summaries
	for i := 0; i < 3; i++ {
		date := startDate.AddDate(0, i, 0)
		monthKey := date.Format("2006-01")
		count := f.RandomInt(10, 50)
		summary.ByMonth[monthKey] = &models.ETCMonthlySummary{
			Year:        date.Year(),
			Month:       int(date.Month()),
			TotalCount:  int64(count),
			TotalAmount: int64(count * f.RandomInt(500, 2000)),
		}
	}

	// Apply options
	for _, opt := range opts {
		opt(summary)
	}

	return summary
}

// Helper function to create pointer to value
func ptr[T any](v T) *T {
	return &v
}

// Helper function to create pointer to int64
func ptrInt64(v int64) *int64 {
	return &v
}

// Helper function to create pointer to time
func ptrTime(v time.Time) *time.Time {
	return &v
}

// T013-D: Builder pattern for complex object construction
// ETCMeisaiBuilder provides fluent interface for building ETCMeisai
type ETCMeisaiBuilder struct {
	factory *TestFactory
	record  *models.ETCMeisai
}

// NewETCMeisaiBuilder creates a new builder
func (f *TestFactory) NewETCMeisaiBuilder() *ETCMeisaiBuilder {
	return &ETCMeisaiBuilder{
		factory: f,
		record:  f.CreateETCMeisai(),
	}
}

// WithDate sets the use date
func (b *ETCMeisaiBuilder) WithDate(date time.Time) *ETCMeisaiBuilder {
	b.record.UseDate = date
	return b
}

// WithRoute sets entry and exit ICs
func (b *ETCMeisaiBuilder) WithRoute(entryIC, exitIC string) *ETCMeisaiBuilder {
	b.record.EntryIC = entryIC
	b.record.ExitIC = exitIC
	return b
}

// WithAmount sets the toll amount
func (b *ETCMeisaiBuilder) WithAmount(amount int) *ETCMeisaiBuilder {
	b.record.Amount = int32(amount)
	return b
}

// WithCarNumber sets the car number
func (b *ETCMeisaiBuilder) WithCarNumber(carNumber string) *ETCMeisaiBuilder {
	b.record.CarNumber = carNumber
	return b
}

// WithETCNumber sets the ETC card number
func (b *ETCMeisaiBuilder) WithETCNumber(etcNumber string) *ETCMeisaiBuilder {
	b.record.ETCNumber = etcNumber
	return b
}

// Build returns the constructed ETCMeisai
func (b *ETCMeisaiBuilder) Build() *models.ETCMeisai {
	return b.record
}

// ImportSessionBuilder provides fluent interface for building ImportSession
type ImportSessionBuilder struct {
	factory *TestFactory
	session *models.ImportSession
}

// NewImportSessionBuilder creates a new builder
func (f *TestFactory) NewImportSessionBuilder() *ImportSessionBuilder {
	return &ImportSessionBuilder{
		factory: f,
		session: f.CreateImportSession(),
	}
}

// WithFileName sets the file name
func (b *ImportSessionBuilder) WithFileName(fileName string) *ImportSessionBuilder {
	b.session.FileName = fileName
	return b
}

// WithAccountType sets the account type
func (b *ImportSessionBuilder) WithAccountType(accountType string) *ImportSessionBuilder {
	b.session.AccountType = accountType
	return b
}

// WithRows sets row counts
func (b *ImportSessionBuilder) WithRows(total, processed, success, error, duplicate int) *ImportSessionBuilder {
	b.session.TotalRows = total
	b.session.ProcessedRows = processed
	b.session.SuccessRows = success
	b.session.ErrorRows = error
	b.session.DuplicateRows = duplicate
	return b
}

// WithStatus sets the status
func (b *ImportSessionBuilder) WithStatus(status string) *ImportSessionBuilder {
	b.session.Status = status
	return b
}

// Build returns the constructed ImportSession
func (b *ImportSessionBuilder) Build() *models.ImportSession {
	return b.session
}

// T013-D: Scenario-based factories for common test cases
// CreateScenario provides pre-configured test scenarios
type CreateScenario struct {
	factory *TestFactory
}

// Scenarios returns scenario factory
func (f *TestFactory) Scenarios() *CreateScenario {
	return &CreateScenario{factory: f}
}

// SuccessfulImport creates a successful import scenario
func (s *CreateScenario) SuccessfulImport() (*models.ImportSession, []*models.ETCMeisaiRecord) {
	session := s.factory.NewImportSessionBuilder().
		WithStatus("completed").
		WithRows(100, 100, 95, 5, 0).
		Build()

	records := s.factory.CreateETCMeisaiRecordBatch(95)
	return session, records
}

// FailedImport creates a failed import scenario
func (s *CreateScenario) FailedImport() (*models.ImportSession, []*models.ETCMeisaiRecord) {
	session := s.factory.NewImportSessionBuilder().
		WithStatus("failed").
		WithRows(100, 20, 0, 20, 0).
		Build()

	return session, nil
}

// MappingConflict creates a mapping conflict scenario
func (s *CreateScenario) MappingConflict() ([]*models.ETCMeisaiRecord, []*models.ETCMapping) {
	// Create records with same dtako_row_id
	dtakoID := int64(12345)
	records := make([]*models.ETCMeisaiRecord, 3)
	mappings := make([]*models.ETCMapping, 3)

	for i := 0; i < 3; i++ {
		records[i] = s.factory.CreateETCMeisaiRecord(func(r *models.ETCMeisaiRecord) {
			r.DtakoRowID = &dtakoID
		})
		mappings[i] = s.factory.CreateETCMapping(func(m *models.ETCMapping) {
			m.ETCRecordID = records[i].ID
			m.MappedEntityID = dtakoID
		})
	}

	return records, mappings
}

// DuplicateRecords creates duplicate ETC records scenario
func (s *CreateScenario) DuplicateRecords() []*models.ETCMeisaiRecord {
	baseRecord := s.factory.CreateETCMeisaiRecord()
	duplicates := make([]*models.ETCMeisaiRecord, 3)

	for i := 0; i < 3; i++ {
		duplicates[i] = s.factory.CreateETCMeisaiRecord(func(r *models.ETCMeisaiRecord) {
			r.Date = baseRecord.Date
			r.Time = baseRecord.Time
			r.EntranceIC = baseRecord.EntranceIC
			r.ExitIC = baseRecord.ExitIC
			r.CarNumber = baseRecord.CarNumber
			r.TollAmount = baseRecord.TollAmount
			r.Hash = baseRecord.Hash
		})
	}

	return duplicates
}

// LargeDataset creates a large dataset for performance testing
func (s *CreateScenario) LargeDataset(recordCount int) ([]*models.ETCMeisaiRecord, []*models.ETCMapping) {
	records := s.factory.CreateETCMeisaiRecordBatch(recordCount)
	mappings := make([]*models.ETCMapping, recordCount)

	for i, record := range records {
		mappings[i] = s.factory.CreateETCMapping(func(m *models.ETCMapping) {
			m.ETCRecordID = record.ID
		})
	}

	return records, mappings
}

// HighConfidenceMatches creates records with high confidence mappings
func (s *CreateScenario) HighConfidenceMatches() ([]*models.ETCMeisaiRecord, []*models.ETCMapping) {
	records := s.factory.CreateETCMeisaiRecordBatch(10)
	mappings := make([]*models.ETCMapping, 10)

	for i, record := range records {
		mappings[i] = s.factory.CreateETCMapping(func(m *models.ETCMapping) {
			m.ETCRecordID = record.ID
			m.Confidence = float32(0.95 + s.factory.rand.Float64()*0.05) // 0.95 to 1.0
			m.Status = "active"
		})
	}

	return records, mappings
}

// MixedStatuses creates records with various statuses
func (s *CreateScenario) MixedStatuses() []*models.ImportSession {
	statuses := []string{"pending", "processing", "completed", "failed", "cancelled"}
	sessions := make([]*models.ImportSession, len(statuses))

	for i, status := range statuses {
		sessions[i] = s.factory.CreateImportSession(func(session *models.ImportSession) {
			session.Status = status
		})
	}

	return sessions
}