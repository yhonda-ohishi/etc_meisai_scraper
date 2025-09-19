package contract

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yhonda-ohishi/etc_meisai/src/models"
	"github.com/yhonda-ohishi/etc_meisai/src/repositories"
)

// ETCRepositoryContract defines the contract tests that any ETCRepository implementation must pass
type ETCRepositoryContract struct {
	NewRepository func() repositories.ETCRepository
	CleanupFunc   func()
}

// RunContractTests runs all contract tests for ETCRepository
func (c *ETCRepositoryContract) RunContractTests(t *testing.T) {
	t.Run("Create", c.TestCreate)
	t.Run("GetByID", c.TestGetByID)
	t.Run("Update", c.TestUpdate)
	t.Run("Delete", c.TestDelete)
	t.Run("List", c.TestList)
	t.Run("BulkInsert", c.TestBulkInsert)
	t.Run("GetByDateRange", c.TestGetByDateRange)
	t.Run("CheckDuplicatesByHash", c.TestCheckDuplicatesByHash)
	t.Run("GetByETCNumber", c.TestGetByETCNumber)
	t.Run("GetByCarNumber", c.TestGetByCarNumber)
	t.Run("CountByDateRange", c.TestCountByDateRange)
	t.Run("GetSummaryByDateRange", c.TestGetSummaryByDateRange)
}

func (c *ETCRepositoryContract) TestCreate(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Prepare test data
	etc := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		UseTime:   "14:30",
		EntryIC:   "東京IC",
		ExitIC:    "横浜IC",
		Amount:    1500,
		CarNumber: "品川300あ1234",
		ETCNumber: "1234567890123456",
	}
	etc.Hash = etc.GenerateHash()

	// Test Create
	err := repo.Create(etc)
	require.NoError(t, err, "Create should not return error")
	assert.NotZero(t, etc.ID, "ID should be set after creation")
	assert.NotZero(t, etc.CreatedAt, "CreatedAt should be set")
	assert.NotZero(t, etc.UpdatedAt, "UpdatedAt should be set")
}

func (c *ETCRepositoryContract) TestGetByID(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create a record first
	original := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		UseTime:   "15:00",
		EntryIC:   "新宿IC",
		ExitIC:    "渋谷IC",
		Amount:    800,
		CarNumber: "品川500か5678",
		ETCNumber: "9876543210987654",
	}
	original.Hash = original.GenerateHash()

	err := repo.Create(original)
	require.NoError(t, err)
	require.NotZero(t, original.ID)

	// Test GetByID
	retrieved, err := repo.GetByID(original.ID)
	require.NoError(t, err, "GetByID should not return error")
	require.NotNil(t, retrieved, "Retrieved record should not be nil")

	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Hash, retrieved.Hash)
	assert.Equal(t, original.ETCNumber, retrieved.ETCNumber)
	assert.Equal(t, original.Amount, retrieved.Amount)

	// Test GetByID with non-existent ID
	_, err = repo.GetByID(999999)
	assert.Error(t, err, "GetByID should return error for non-existent ID")
}

func (c *ETCRepositoryContract) TestUpdate(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create a record first
	etc := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		UseTime:   "16:00",
		EntryIC:   "品川IC",
		ExitIC:    "大井IC",
		Amount:    600,
		CarNumber: "品川100さ9999",
		ETCNumber: "1111222233334444",
	}
	etc.Hash = etc.GenerateHash()

	err := repo.Create(etc)
	require.NoError(t, err)
	originalID := etc.ID

	// Update the record
	etc.Amount = 700
	etc.ExitIC = "大田IC"
	etc.Hash = etc.GenerateHash()

	err = repo.Update(etc)
	require.NoError(t, err, "Update should not return error")

	// Verify update
	updated, err := repo.GetByID(originalID)
	require.NoError(t, err)
	assert.Equal(t, int32(700), updated.Amount)
	assert.Equal(t, "大田IC", updated.ExitIC)
}

func (c *ETCRepositoryContract) TestDelete(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create a record first
	etc := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		UseTime:   "17:00",
		EntryIC:   "池袋IC",
		ExitIC:    "練馬IC",
		Amount:    900,
		CarNumber: "練馬300た1111",
		ETCNumber: "5555666677778888",
	}
	etc.Hash = etc.GenerateHash()

	err := repo.Create(etc)
	require.NoError(t, err)

	// Delete the record
	err = repo.Delete(etc.ID)
	require.NoError(t, err, "Delete should not return error")

	// Verify deletion
	_, err = repo.GetByID(etc.ID)
	assert.Error(t, err, "GetByID should return error after deletion")
}

func (c *ETCRepositoryContract) TestList(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create multiple records
	for i := 0; i < 5; i++ {
		etc := &models.ETCMeisai{
			UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
			UseTime:   fmt.Sprintf("%02d:00", 10+i),
			EntryIC:   fmt.Sprintf("IC_%d", i),
			ExitIC:    fmt.Sprintf("IC_%d", i+1),
			Amount:    int32(1000 + i*100),
			CarNumber: fmt.Sprintf("TEST_%d", i),
			ETCNumber: fmt.Sprintf("123456789012345%d", i),
		}
		etc.Hash = etc.GenerateHash()
		err := repo.Create(etc)
		require.NoError(t, err)
	}

	// Test List with pagination
	params := &models.ETCListParams{
		Limit:  3,
		Offset: 0,
	}
	params.SetDefaults()

	records, total, err := repo.List(params)
	require.NoError(t, err, "List should not return error")
	assert.Len(t, records, 3, "Should return 3 records with limit=3")
	assert.GreaterOrEqual(t, total, int64(5), "Total should be at least 5")
}

func (c *ETCRepositoryContract) TestBulkInsert(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Prepare bulk records
	var records []*models.ETCMeisai
	for i := 0; i < 10; i++ {
		etc := &models.ETCMeisai{
			UseDate:   time.Date(2025, 1, 20+i, 0, 0, 0, 0, time.UTC),
			UseTime:   fmt.Sprintf("%02d:30", i),
			EntryIC:   fmt.Sprintf("Entry_%d", i),
			ExitIC:    fmt.Sprintf("Exit_%d", i),
			Amount:    int32(500 + i*50),
			CarNumber: fmt.Sprintf("BULK_%d", i),
			ETCNumber: fmt.Sprintf("987654321098765%d", i),
		}
		etc.Hash = etc.GenerateHash()
		records = append(records, etc)
	}

	// Test BulkInsert
	err := repo.BulkInsert(records)
	require.NoError(t, err, "BulkInsert should not return error")

	// Verify all records were inserted
	for _, record := range records {
		assert.NotZero(t, record.ID, "Each record should have ID after bulk insert")
	}
}

func (c *ETCRepositoryContract) TestGetByDateRange(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create records with different dates
	dates := []time.Time{
		time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 25, 0, 0, 0, 0, time.UTC),
	}

	for i, date := range dates {
		etc := &models.ETCMeisai{
			UseDate:   date,
			UseTime:   "12:00",
			EntryIC:   "DateTest",
			ExitIC:    "DateTest",
			Amount:    1000,
			CarNumber: fmt.Sprintf("DATE_%d", i),
			ETCNumber: "1234567890123456",
		}
		etc.Hash = etc.GenerateHash()
		err := repo.Create(etc)
		require.NoError(t, err)
	}

	// Test GetByDateRange
	from := time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)

	records, err := repo.GetByDateRange(from, to)
	require.NoError(t, err, "GetByDateRange should not return error")
	assert.Len(t, records, 2, "Should return 2 records in date range")
}

func (c *ETCRepositoryContract) TestCheckDuplicatesByHash(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create a record
	etc := &models.ETCMeisai{
		UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
		UseTime:   "18:00",
		EntryIC:   "HashTest",
		ExitIC:    "HashTest",
		Amount:    2000,
		CarNumber: "HASH_001",
		ETCNumber: "9999888877776666",
	}
	etc.Hash = etc.GenerateHash()
	err := repo.Create(etc)
	require.NoError(t, err)

	// Test CheckDuplicatesByHash
	hashes := []string{
		etc.Hash,
		"non_existent_hash_123",
		"another_non_existent_hash",
	}

	duplicates, err := repo.CheckDuplicatesByHash(hashes)
	require.NoError(t, err, "CheckDuplicatesByHash should not return error")
	assert.True(t, duplicates[etc.Hash], "Existing hash should be marked as duplicate")
	assert.False(t, duplicates["non_existent_hash_123"], "Non-existent hash should not be duplicate")
}

func (c *ETCRepositoryContract) TestGetByETCNumber(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	etcNumber := "1234567890UNIQUE"

	// Create records with same ETC number
	for i := 0; i < 3; i++ {
		etc := &models.ETCMeisai{
			UseDate:   time.Date(2025, 1, 20+i, 0, 0, 0, 0, time.UTC),
			UseTime:   fmt.Sprintf("%02d:00", i),
			EntryIC:   "ETCNumTest",
			ExitIC:    "ETCNumTest",
			Amount:    1000,
			CarNumber: fmt.Sprintf("ETC_%d", i),
			ETCNumber: etcNumber,
		}
		etc.Hash = etc.GenerateHash()
		err := repo.Create(etc)
		require.NoError(t, err)
	}

	// Test GetByETCNumber
	records, err := repo.GetByETCNumber(etcNumber, 10)
	require.NoError(t, err, "GetByETCNumber should not return error")
	assert.Len(t, records, 3, "Should return 3 records with same ETC number")
}

func (c *ETCRepositoryContract) TestGetByCarNumber(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	carNumber := "品川999せ9999"

	// Create records with same car number
	for i := 0; i < 2; i++ {
		etc := &models.ETCMeisai{
			UseDate:   time.Date(2025, 1, 21+i, 0, 0, 0, 0, time.UTC),
			UseTime:   fmt.Sprintf("%02d:30", i),
			EntryIC:   "CarNumTest",
			ExitIC:    "CarNumTest",
			Amount:    1500,
			CarNumber: carNumber,
			ETCNumber: fmt.Sprintf("CAR_%d", i),
		}
		etc.Hash = etc.GenerateHash()
		err := repo.Create(etc)
		require.NoError(t, err)
	}

	// Test GetByCarNumber
	records, err := repo.GetByCarNumber(carNumber, 10)
	require.NoError(t, err, "GetByCarNumber should not return error")
	assert.Len(t, records, 2, "Should return 2 records with same car number")
}

func (c *ETCRepositoryContract) TestCountByDateRange(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create records
	for i := 0; i < 5; i++ {
		etc := &models.ETCMeisai{
			UseDate:   time.Date(2025, 1, 15+i, 0, 0, 0, 0, time.UTC),
			UseTime:   "10:00",
			EntryIC:   "CountTest",
			ExitIC:    "CountTest",
			Amount:    1000,
			CarNumber: fmt.Sprintf("COUNT_%d", i),
			ETCNumber: "0000111122223333",
		}
		etc.Hash = etc.GenerateHash()
		err := repo.Create(etc)
		require.NoError(t, err)
	}

	// Test CountByDateRange
	from := time.Date(2025, 1, 14, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)

	count, err := repo.CountByDateRange(from, to)
	require.NoError(t, err, "CountByDateRange should not return error")
	assert.Equal(t, int64(3), count, "Should count 3 records in date range")
}

func (c *ETCRepositoryContract) TestGetSummaryByDateRange(t *testing.T) {
	repo := c.NewRepository()
	defer c.CleanupFunc()

	// Create records
	amounts := []int32{1000, 1500, 2000}
	for i, amount := range amounts {
		etc := &models.ETCMeisai{
			UseDate:   time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
			UseTime:   fmt.Sprintf("%02d:00", i),
			EntryIC:   "SummaryTest",
			ExitIC:    "SummaryTest",
			Amount:    amount,
			CarNumber: fmt.Sprintf("SUM_%d", i),
			ETCNumber: "4444555566667777",
		}
		etc.Hash = etc.GenerateHash()
		err := repo.Create(etc)
		require.NoError(t, err)
	}

	// Test GetSummaryByDateRange
	from := time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)

	summary, err := repo.GetSummaryByDateRange(from, to)
	require.NoError(t, err, "GetSummaryByDateRange should not return error")
	require.NotNil(t, summary, "Summary should not be nil")

	assert.Equal(t, int64(4500), summary.TotalAmount, "Total amount should be 4500")
	assert.Equal(t, int64(3), summary.TotalCount, "Total count should be 3")
}