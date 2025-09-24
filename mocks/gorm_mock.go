package mocks

import (
	"context"
	"database/sql"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of gorm.DB
type MockDB struct {
	mock.Mock
}

// AutoMigrate mocks the AutoMigrate method
func (m *MockDB) AutoMigrate(dst ...interface{}) error {
	args := m.Called(dst)
	return args.Error(0)
}

// Create mocks the Create method
func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// Save mocks the Save method
func (m *MockDB) Save(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

// First mocks the First method
func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

// Find mocks the Find method
func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

// Where mocks the Where method
func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*gorm.DB)
}

// Delete mocks the Delete method
func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(value, conds)
	return args.Get(0).(*gorm.DB)
}

// Update mocks the Update method
func (m *MockDB) Update(column string, value interface{}) *gorm.DB {
	args := m.Called(column, value)
	return args.Get(0).(*gorm.DB)
}

// Updates mocks the Updates method
func (m *MockDB) Updates(values interface{}) *gorm.DB {
	args := m.Called(values)
	return args.Get(0).(*gorm.DB)
}

// Begin mocks the Begin method
func (m *MockDB) Begin(opts ...*sql.TxOptions) *gorm.DB {
	args := m.Called(opts)
	return args.Get(0).(*gorm.DB)
}

// Commit mocks the Commit method
func (m *MockDB) Commit() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// Rollback mocks the Rollback method
func (m *MockDB) Rollback() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// Error returns the error from the last operation
func (m *MockDB) Error() error {
	args := m.Called()
	return args.Error(0)
}

// MockGormDB is a complete mock implementation that implements the gorm.DB interface
type MockGormDB struct {
	*gorm.DB
	mock.Mock
}

// NewMockGormDB creates a new mock GORM DB
func NewMockGormDB() *MockGormDB {
	// Create a mock DB that satisfies the gorm.DB interface
	mockDB := &MockGormDB{}

	// Initialize with a minimal GORM DB structure
	mockDB.DB = &gorm.DB{
		Config: &gorm.Config{},
		Statement: &gorm.Statement{
			Context: context.Background(),
		},
	}

	return mockDB
}

// Override common methods to return the mock itself for chaining
func (m *MockGormDB) WithContext(ctx context.Context) *gorm.DB {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Model(value interface{}) *gorm.DB {
	args := m.Called(value)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Table(name string, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(name, args)
	if mockArgs.Get(0) == nil {
		return m.DB
	}
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Select(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return m.DB
	}
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Omit(columns ...string) *gorm.DB {
	args := m.Called(columns)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Distinct(args ...interface{}) *gorm.DB {
	mockArgs := m.Called(args)
	if mockArgs.Get(0) == nil {
		return m.DB
	}
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Order(value interface{}) *gorm.DB {
	args := m.Called(value)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Limit(limit int) *gorm.DB {
	args := m.Called(limit)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Offset(offset int) *gorm.DB {
	args := m.Called(offset)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Group(name string) *gorm.DB {
	args := m.Called(name)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Having(query interface{}, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return m.DB
	}
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Joins(query string, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return m.DB
	}
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Preload(query string, args ...interface{}) *gorm.DB {
	mockArgs := m.Called(query, args)
	if mockArgs.Get(0) == nil {
		return m.DB
	}
	return mockArgs.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Raw(sql string, values ...interface{}) *gorm.DB {
	args := m.Called(sql, values)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Exec(sql string, values ...interface{}) *gorm.DB {
	args := m.Called(sql, values)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Scan(dest interface{}) *gorm.DB {
	args := m.Called(dest)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) ScanRows(rows *sql.Rows, dest interface{}) error {
	args := m.Called(rows, dest)
	return args.Error(0)
}

func (m *MockGormDB) Row() *sql.Row {
	args := m.Called()
	return args.Get(0).(*sql.Row)
}

func (m *MockGormDB) Rows() (*sql.Rows, error) {
	args := m.Called()
	return args.Get(0).(*sql.Rows), args.Error(1)
}

func (m *MockGormDB) Count(count *int64) *gorm.DB {
	args := m.Called(count)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Pluck(column string, dest interface{}) *gorm.DB {
	args := m.Called(column, dest)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	args := m.Called(fc, opts)
	return args.Error(0)
}

func (m *MockGormDB) Debug() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockGormDB) Session(config *gorm.Session) *gorm.DB {
	args := m.Called(config)
	if args.Get(0) == nil {
		return m.DB
	}
	return args.Get(0).(*gorm.DB)
}

// MockTx represents a database transaction mock
type MockTx struct {
	*MockGormDB
}

// NewMockTx creates a new mock transaction
func NewMockTx() *MockTx {
	return &MockTx{
		MockGormDB: NewMockGormDB(),
	}
}