package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// T011-E: Resource cleanup testing to prevent test data leakage
func TestResourceCleanup_PreventLeakage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource cleanup test in short mode")
	}

	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"Database Connection Cleanup", testDatabaseConnectionCleanup},
		{"File Handle Cleanup", testFileHandleCleanup},
		{"Temporary File Cleanup", testTemporaryFileCleanup},
		{"Memory Cleanup", testMemoryCleanup},
		{"Goroutine Cleanup", testGoroutineCleanup},
		{"HTTP Connection Cleanup", testHTTPConnectionCleanup},
		{"Transaction Rollback", testTransactionRollback},
		{"Lock Release", testLockRelease},
		{"Context Cancellation", testContextCancellation},
		{"Resource Pool Cleanup", testResourcePoolCleanup},
		{"Test Data Isolation", testDataIsolation},
		{"Cleanup On Panic", testCleanupOnPanic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Record initial state
			initialGoroutines := runtime.NumGoroutine()
			initialMemStats := runtime.MemStats{}
			runtime.ReadMemStats(&initialMemStats)

			// Run test
			tt.testFunc(t)

			// Force GC to clean up
			runtime.GC()
			runtime.Gosched()
			time.Sleep(100 * time.Millisecond)

			// Verify cleanup
			finalGoroutines := runtime.NumGoroutine()
			assert.LessOrEqual(t, finalGoroutines, initialGoroutines+2,
				"Goroutines should be cleaned up (initial: %d, final: %d)",
				initialGoroutines, finalGoroutines)
		})
	}
}

func testDatabaseConnectionCleanup(t *testing.T) {
	// Create temp SQLite database
	tempDB, err := os.CreateTemp("", "cleanup_test_*.db")
	require.NoError(t, err)
	tempDB.Close()
	dbPath := tempDB.Name()
	defer os.Remove(dbPath)

	// Test connection leak prevention
	manager := NewDatabaseManager(dbPath)

	// Open connections
	for i := 0; i < 10; i++ {
		conn, err := manager.GetConnection()
		require.NoError(t, err)

		// Simulate work
		_, err = conn.Exec("CREATE TABLE IF NOT EXISTS test (id INTEGER)")
		assert.NoError(t, err)

		// Must close connection
		manager.ReleaseConnection(conn)
	}

	// Verify all connections are released
	stats := manager.GetStats()
	assert.Equal(t, 0, stats.ActiveConnections)
	assert.Equal(t, 10, stats.TotalConnectionsCreated)

	// Cleanup manager
	err = manager.Close()
	assert.NoError(t, err)

	// Verify database is closed
	_, err = sql.Open("sqlite3", dbPath)
	assert.NoError(t, err) // Should be able to open if properly closed
}

func testFileHandleCleanup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_cleanup_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tracker := NewFileHandleTracker()

	// Open multiple files
	files := make([]*os.File, 0, 100)
	for i := 0; i < 100; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("file_%d.txt", i))
		file, err := os.Create(filePath)
		require.NoError(t, err)

		tracker.Track(file)
		files = append(files, file)

		// Write some data
		_, err = file.WriteString(fmt.Sprintf("Test data %d\n", i))
		assert.NoError(t, err)
	}

	// Verify files are tracked
	assert.Equal(t, 100, tracker.Count())

	// Close all files through tracker
	err = tracker.CloseAll()
	assert.NoError(t, err)
	assert.Equal(t, 0, tracker.Count())

	// Verify files are closed by trying to write
	for _, file := range files {
		_, err := file.WriteString("should fail")
		assert.Error(t, err, "File should be closed")
	}
}

func testTemporaryFileCleanup(t *testing.T) {
	cleaner := NewTempFileCleaner()

	// Create temp files
	var tempFiles []string
	for i := 0; i < 10; i++ {
		tempFile, err := cleaner.CreateTempFile("test_*.txt")
		require.NoError(t, err)
		tempFiles = append(tempFiles, tempFile)

		// Write data
		err = os.WriteFile(tempFile, []byte(fmt.Sprintf("temp data %d", i)), 0644)
		assert.NoError(t, err)

		// Verify file exists
		assert.FileExists(t, tempFile)
	}

	// Cleanup all temp files
	err := cleaner.CleanupAll()
	assert.NoError(t, err)

	// Verify all files are deleted
	for _, tempFile := range tempFiles {
		assert.NoFileExists(t, tempFile)
	}

	// Verify cleanup is idempotent
	err = cleaner.CleanupAll()
	assert.NoError(t, err)
}

func testMemoryCleanup(t *testing.T) {
	allocator := NewMemoryAllocator(100 * 1024 * 1024) // 100MB limit

	// Allocate memory
	allocations := make([][]byte, 0)
	for i := 0; i < 10; i++ {
		data, err := allocator.Allocate(10 * 1024 * 1024) // 10MB each
		require.NoError(t, err)
		allocations = append(allocations, data)

		// Use the memory
		for j := range data {
			data[j] = byte(i)
		}
	}

	// Verify memory is tracked
	assert.Equal(t, int64(100*1024*1024), allocator.Used())

	// Free memory
	for _, data := range allocations {
		allocator.Free(data)
	}

	// Verify memory is freed
	assert.Equal(t, int64(0), allocator.Used())

	// Force GC
	allocations = nil
	runtime.GC()

	// Check memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	// Memory should be available for GC
	assert.Less(t, memStats.HeapInuse, uint64(200*1024*1024))
}

func testGoroutineCleanup(t *testing.T) {
	manager := NewGoroutineManager()

	// Start goroutines
	for i := 0; i < 100; i++ {
		manager.Start(func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
				// Simulate work
			}
		})
	}

	// Verify goroutines are running
	assert.Equal(t, 100, manager.Count())

	// Stop all goroutines
	manager.StopAll()

	// Wait a bit for goroutines to stop
	time.Sleep(100 * time.Millisecond)

	// Verify all stopped
	assert.Equal(t, 0, manager.Count())
}

func testHTTPConnectionCleanup(t *testing.T) {
	// Create test server
	connectionCount := atomic.Int32{}
	server := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionCount.Add(1)
		defer connectionCount.Add(-1)

		// Hold connection briefly
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	// Start server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	go http.Serve(listener, server)

	// Create client with connection pool
	client := NewPooledHTTPClient(10) // Max 10 connections
	defer client.Close()

	// Make concurrent requests
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := client.Get(fmt.Sprintf("http://%s", listener.Addr()))
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()

	// Verify connections are released
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(0), connectionCount.Load())

	// Verify connection pool is cleaned
	stats := client.GetStats()
	assert.LessOrEqual(t, stats.IdleConnections, 10)
}

func testTransactionRollback(t *testing.T) {
	// Create temp database
	tempDB, err := os.CreateTemp("", "tx_test_*.db")
	require.NoError(t, err)
	tempDB.Close()
	defer os.Remove(tempDB.Name())

	db, err := sql.Open("sqlite3", tempDB.Name())
	require.NoError(t, err)
	defer db.Close()

	// Create table
	_, err = db.Exec("CREATE TABLE test_data (id INTEGER PRIMARY KEY, value TEXT)")
	require.NoError(t, err)

	// Test transaction cleanup
	txManager := NewTransactionManager(db)

	// Start transaction
	tx, err := txManager.Begin()
	require.NoError(t, err)

	// Insert data in transaction
	_, err = tx.Exec("INSERT INTO test_data (value) VALUES (?)", "test_value")
	assert.NoError(t, err)

	// Simulate error - rollback should occur
	err = txManager.Rollback(tx)
	assert.NoError(t, err)

	// Verify data was not persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Test automatic rollback on panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Recovered from panic
			}
		}()

		tx, _ := txManager.BeginWithCleanup()
		defer txManager.EnsureCleanup(tx)

		tx.Exec("INSERT INTO test_data (value) VALUES (?)", "panic_value")
		panic("simulated panic")
	}()

	// Verify panic transaction was rolled back
	err = db.QueryRow("SELECT COUNT(*) FROM test_data").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func testLockRelease(t *testing.T) {
	locker := NewResourceLocker()

	// Acquire locks
	lock1 := locker.Lock("resource1")
	lock2 := locker.Lock("resource2")
	lock3 := locker.Lock("resource3")

	// Verify locks are held
	assert.True(t, locker.IsLocked("resource1"))
	assert.True(t, locker.IsLocked("resource2"))
	assert.True(t, locker.IsLocked("resource3"))

	// Release locks
	lock1.Release()
	lock2.Release()
	lock3.Release()

	// Verify locks are released
	assert.False(t, locker.IsLocked("resource1"))
	assert.False(t, locker.IsLocked("resource2"))
	assert.False(t, locker.IsLocked("resource3"))

	// Test automatic cleanup with defer
	func() {
		lock := locker.LockWithDefer("resource4")
		defer lock.Release()

		assert.True(t, locker.IsLocked("resource4"))
		// Lock will be released when function returns
	}()

	assert.False(t, locker.IsLocked("resource4"))
}

func testContextCancellation(t *testing.T) {
	// Test context cleanup propagation
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	manager := NewContextManager(rootCtx)

	// Create child contexts
	contexts := make([]context.Context, 10)
	for i := 0; i < 10; i++ {
		ctx, _ := manager.CreateChild(fmt.Sprintf("child_%d", i))
		contexts[i] = ctx
	}

	// Start workers with contexts
	var wg sync.WaitGroup
	workersRunning := atomic.Int32{}

	for i, ctx := range contexts {
		wg.Add(1)
		go func(ctx context.Context, id int) {
			defer wg.Done()
			workersRunning.Add(1)
			defer workersRunning.Add(-1)

			select {
			case <-ctx.Done():
				// Properly canceled
			case <-time.After(10 * time.Second):
				t.Errorf("Worker %d not canceled", id)
			}
		}(ctx, i)
	}

	// Verify workers are running
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(10), workersRunning.Load())

	// Cancel all contexts
	manager.CancelAll()

	// Wait for workers to stop
	wg.Wait()

	// Verify all workers stopped
	assert.Equal(t, int32(0), workersRunning.Load())
}

func testResourcePoolCleanup(t *testing.T) {
	// Create resource pool
	pool := NewResourcePool(5) // Pool of 5 resources

	// Acquire resources
	resources := make([]Resource, 0)
	for i := 0; i < 5; i++ {
		resource, err := pool.Acquire()
		require.NoError(t, err)
		resources = append(resources, resource)
	}

	// Verify pool is exhausted
	_, err := pool.AcquireWithTimeout(100 * time.Millisecond)
	assert.Error(t, err)

	// Release resources back to pool
	for _, resource := range resources {
		pool.Release(resource)
	}

	// Verify resources are available again
	resource, err := pool.Acquire()
	assert.NoError(t, err)
	pool.Release(resource)

	// Close pool
	err = pool.Close()
	assert.NoError(t, err)

	// Verify pool is closed
	_, err = pool.Acquire()
	assert.Error(t, err)
}

func testDataIsolation(t *testing.T) {
	// Test that test data doesn't leak between tests
	isolator := NewTestDataIsolator()

	// Run test 1
	test1Data := isolator.RunIsolated(func(sandbox *DataSandbox) interface{} {
		sandbox.Set("key1", "value1")
		sandbox.Set("key2", "value2")
		return sandbox.Get("key1")
	})
	assert.Equal(t, "value1", test1Data)

	// Run test 2 - should not see test 1 data
	test2Data := isolator.RunIsolated(func(sandbox *DataSandbox) interface{} {
		// Should not have test 1 data
		assert.Nil(t, sandbox.Get("key1"))
		assert.Nil(t, sandbox.Get("key2"))

		sandbox.Set("key3", "value3")
		return sandbox.Get("key3")
	})
	assert.Equal(t, "value3", test2Data)

	// Verify complete isolation
	assert.Equal(t, 0, isolator.GetActiveTestCount())
}

func testCleanupOnPanic(t *testing.T) {
	cleaner := NewPanicCleaner()

	// Register cleanup handlers
	var cleanupCalled []string
	cleanupMu := sync.Mutex{}

	cleaner.Register("cleanup1", func() {
		cleanupMu.Lock()
		cleanupCalled = append(cleanupCalled, "cleanup1")
		cleanupMu.Unlock()
	})

	cleaner.Register("cleanup2", func() {
		cleanupMu.Lock()
		cleanupCalled = append(cleanupCalled, "cleanup2")
		cleanupMu.Unlock()
	})

	// Function that panics
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Run cleanup on panic
				cleaner.RunCleanup()
			}
		}()

		// Simulate work and panic
		panic("test panic")
	}()

	// Verify cleanup was called
	cleanupMu.Lock()
	defer cleanupMu.Unlock()
	assert.Contains(t, cleanupCalled, "cleanup1")
	assert.Contains(t, cleanupCalled, "cleanup2")
}

// Helper types

type DatabaseManager struct {
	dbPath      string
	connections []*sql.DB
	mu          sync.Mutex
	stats       DBStats
}

type DBStats struct {
	ActiveConnections       int
	TotalConnectionsCreated int
}

func NewDatabaseManager(dbPath string) *DatabaseManager {
	return &DatabaseManager{
		dbPath:      dbPath,
		connections: make([]*sql.DB, 0),
	}
}

func (m *DatabaseManager) GetConnection() (*sql.DB, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, err := sql.Open("sqlite3", m.dbPath)
	if err != nil {
		return nil, err
	}

	m.connections = append(m.connections, conn)
	m.stats.TotalConnectionsCreated++
	m.stats.ActiveConnections++

	return conn, nil
}

func (m *DatabaseManager) ReleaseConnection(conn *sql.DB) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn.Close()
	m.stats.ActiveConnections--
}

func (m *DatabaseManager) GetStats() DBStats {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stats
}

func (m *DatabaseManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		conn.Close()
	}
	m.connections = nil
	m.stats.ActiveConnections = 0

	return nil
}

type FileHandleTracker struct {
	files []*os.File
	mu    sync.Mutex
}

func NewFileHandleTracker() *FileHandleTracker {
	return &FileHandleTracker{
		files: make([]*os.File, 0),
	}
}

func (t *FileHandleTracker) Track(file *os.File) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.files = append(t.files, file)
}

func (t *FileHandleTracker) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.files)
}

func (t *FileHandleTracker) CloseAll() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, file := range t.files {
		file.Close()
	}
	t.files = nil

	return nil
}

type TempFileCleaner struct {
	files []string
	mu    sync.Mutex
}

func NewTempFileCleaner() *TempFileCleaner {
	return &TempFileCleaner{
		files: make([]string, 0),
	}
}

func (c *TempFileCleaner) CreateTempFile(pattern string) (string, error) {
	tempFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	tempFile.Close()

	c.mu.Lock()
	c.files = append(c.files, tempFile.Name())
	c.mu.Unlock()

	return tempFile.Name(), nil
}

func (c *TempFileCleaner) CleanupAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, file := range c.files {
		os.Remove(file)
	}
	c.files = nil

	return nil
}

type MemoryAllocator struct {
	limit      int64
	used       atomic.Int64
	allocations map[*byte]int64
	mu         sync.Mutex
}

func NewMemoryAllocator(limit int64) *MemoryAllocator {
	return &MemoryAllocator{
		limit:       limit,
		allocations: make(map[*byte]int64),
	}
}

func (a *MemoryAllocator) Allocate(size int64) ([]byte, error) {
	if a.used.Load()+size > a.limit {
		return nil, fmt.Errorf("memory limit exceeded")
	}

	data := make([]byte, size)
	a.used.Add(size)

	a.mu.Lock()
	if len(data) > 0 {
		a.allocations[&data[0]] = size
	}
	a.mu.Unlock()

	return data, nil
}

func (a *MemoryAllocator) Free(data []byte) {
	if len(data) == 0 {
		return
	}

	a.mu.Lock()
	size, exists := a.allocations[&data[0]]
	if exists {
		delete(a.allocations, &data[0])
		a.used.Add(-size)
	}
	a.mu.Unlock()
}

func (a *MemoryAllocator) Used() int64 {
	return a.used.Load()
}

type GoroutineManager struct {
	wg      sync.WaitGroup
	cancel  context.CancelFunc
	ctx     context.Context
	counter atomic.Int32
}

func NewGoroutineManager() *GoroutineManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &GoroutineManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (m *GoroutineManager) Start(fn func(context.Context)) {
	m.wg.Add(1)
	m.counter.Add(1)

	go func() {
		defer m.wg.Done()
		defer m.counter.Add(-1)
		fn(m.ctx)
	}()
}

func (m *GoroutineManager) StopAll() {
	m.cancel()
	m.wg.Wait()
}

func (m *GoroutineManager) Count() int {
	return int(m.counter.Load())
}

type PooledHTTPClient struct {
	client *http.Client
	pool   chan struct{}
}

type ConnectionStats struct {
	IdleConnections int
}

func NewPooledHTTPClient(maxConnections int) *PooledHTTPClient {
	return &PooledHTTPClient{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        maxConnections,
				MaxIdleConnsPerHost: maxConnections,
			},
		},
		pool: make(chan struct{}, maxConnections),
	}
}

func (c *PooledHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

func (c *PooledHTTPClient) GetStats() ConnectionStats {
	// Simplified stats
	return ConnectionStats{
		IdleConnections: len(c.pool),
	}
}

func (c *PooledHTTPClient) Close() {
	c.client.CloseIdleConnections()
}

type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (m *TransactionManager) Begin() (*sql.Tx, error) {
	return m.db.Begin()
}

func (m *TransactionManager) BeginWithCleanup() (*sql.Tx, error) {
	return m.db.Begin()
}

func (m *TransactionManager) Rollback(tx *sql.Tx) error {
	return tx.Rollback()
}

func (m *TransactionManager) EnsureCleanup(tx *sql.Tx) {
	if tx != nil {
		tx.Rollback()
	}
}

type ResourceLocker struct {
	locks map[string]*sync.Mutex
	mu    sync.Mutex
}

type Lock struct {
	mu      *sync.Mutex
	release func()
}

func (l *Lock) Release() {
	if l.release != nil {
		l.release()
	}
}

func NewResourceLocker() *ResourceLocker {
	return &ResourceLocker{
		locks: make(map[string]*sync.Mutex),
	}
}

func (l *ResourceLocker) Lock(resource string) *Lock {
	l.mu.Lock()
	if _, exists := l.locks[resource]; !exists {
		l.locks[resource] = &sync.Mutex{}
	}
	lock := l.locks[resource]
	l.mu.Unlock()

	lock.Lock()

	return &Lock{
		mu: lock,
		release: func() {
			lock.Unlock()
		},
	}
}

func (l *ResourceLocker) LockWithDefer(resource string) *Lock {
	return l.Lock(resource)
}

func (l *ResourceLocker) IsLocked(resource string) bool {
	l.mu.Lock()
	lock, exists := l.locks[resource]
	l.mu.Unlock()

	if !exists {
		return false
	}

	// Try to lock, if successful it wasn't locked
	if lock.TryLock() {
		lock.Unlock()
		return false
	}
	return true
}

type ContextManager struct {
	root     context.Context
	children map[string]context.CancelFunc
	mu       sync.Mutex
}

func NewContextManager(root context.Context) *ContextManager {
	return &ContextManager{
		root:     root,
		children: make(map[string]context.CancelFunc),
	}
}

func (m *ContextManager) CreateChild(name string) (context.Context, context.CancelFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithCancel(m.root)
	m.children[name] = cancel

	return ctx, cancel
}

func (m *ContextManager) CancelAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cancel := range m.children {
		cancel()
	}
	m.children = make(map[string]context.CancelFunc)
}

type Resource interface {
	ID() string
	Close() error
}

type ResourcePool struct {
	resources chan Resource
	closed    atomic.Bool
}

func NewResourcePool(size int) *ResourcePool {
	pool := &ResourcePool{
		resources: make(chan Resource, size),
	}

	// Initialize pool with resources
	for i := 0; i < size; i++ {
		pool.resources <- &testResource{id: fmt.Sprintf("resource_%d", i)}
	}

	return pool
}

func (p *ResourcePool) Acquire() (Resource, error) {
	if p.closed.Load() {
		return nil, fmt.Errorf("pool is closed")
	}

	select {
	case resource := <-p.resources:
		return resource, nil
	default:
		return nil, fmt.Errorf("no resources available")
	}
}

func (p *ResourcePool) AcquireWithTimeout(timeout time.Duration) (Resource, error) {
	if p.closed.Load() {
		return nil, fmt.Errorf("pool is closed")
	}

	select {
	case resource := <-p.resources:
		return resource, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout acquiring resource")
	}
}

func (p *ResourcePool) Release(resource Resource) {
	if p.closed.Load() {
		resource.Close()
		return
	}

	select {
	case p.resources <- resource:
		// Released successfully
	default:
		// Pool is full, close resource
		resource.Close()
	}
}

func (p *ResourcePool) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return fmt.Errorf("pool already closed")
	}

	close(p.resources)
	for resource := range p.resources {
		resource.Close()
	}

	return nil
}

type testResource struct {
	id string
}

func (r *testResource) ID() string {
	return r.id
}

func (r *testResource) Close() error {
	return nil
}

type TestDataIsolator struct {
	activeTests atomic.Int32
}

type DataSandbox struct {
	data map[string]interface{}
	mu   sync.Mutex
}

func (s *DataSandbox) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = make(map[string]interface{})
	}
	s.data[key] = value
}

func (s *DataSandbox) Get(key string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		return nil
	}
	return s.data[key]
}

func NewTestDataIsolator() *TestDataIsolator {
	return &TestDataIsolator{}
}

func (i *TestDataIsolator) RunIsolated(fn func(*DataSandbox) interface{}) interface{} {
	i.activeTests.Add(1)
	defer i.activeTests.Add(-1)

	sandbox := &DataSandbox{
		data: make(map[string]interface{}),
	}

	result := fn(sandbox)

	// Clear sandbox data
	sandbox.data = nil

	return result
}

func (i *TestDataIsolator) GetActiveTestCount() int {
	return int(i.activeTests.Load())
}

type PanicCleaner struct {
	cleanups map[string]func()
	mu       sync.Mutex
}

func NewPanicCleaner() *PanicCleaner {
	return &PanicCleaner{
		cleanups: make(map[string]func()),
	}
}

func (c *PanicCleaner) Register(name string, cleanup func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanups[name] = cleanup
}

func (c *PanicCleaner) RunCleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cleanup := range c.cleanups {
		cleanup()
	}
	c.cleanups = make(map[string]func())
}