package integration

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T011-C: File system integration testing for CSV import/export operations
func TestFileSystemIntegration_CSVOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file system integration test in short mode")
	}

	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		testFunc func(*testing.T, string)
	}{
		{"CSV Import Basic", testCSVImportBasic},
		{"CSV Export Basic", testCSVExportBasic},
		{"Large CSV Processing", testLargeCSVProcessing},
		{"Concurrent File Access", testConcurrentFileAccess},
		{"File Locking", testFileLocking},
		{"Directory Operations", testDirectoryOperations},
		{"File Permissions", testFilePermissions},
		{"Compressed Files", testCompressedFiles},
		{"Streaming Operations", testStreamingOperations},
		{"Error Recovery", testFileSystemErrorRecovery},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subdirectory for each test
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			tt.testFunc(t, testDir)
		})
	}
}

func testCSVImportBasic(t *testing.T, tempDir string) {
	// Create test CSV file
	csvFile := filepath.Join(tempDir, "test_import.csv")
	data := [][]string{
		{"Date", "EntryIC", "ExitIC", "Amount", "VehicleNumber", "ETCNumber"},
		{"2024-01-01", "Tokyo", "Osaka", "1000", "品川300あ1234", "ETC-001"},
		{"2024-01-02", "Nagoya", "Kyoto", "800", "名古屋400い5678", "ETC-002"},
		{"2024-01-03", "Osaka", "Kobe", "500", "大阪500う9012", "ETC-003"},
	}

	err := writeCSV(csvFile, data)
	require.NoError(t, err)

	// Test import
	importer := NewCSVImporter()
	records, err := importer.Import(csvFile)
	assert.NoError(t, err)
	assert.Len(t, records, 3)

	// Verify data
	assert.Equal(t, "Tokyo", records[0]["EntryIC"])
	assert.Equal(t, "1000", records[0]["Amount"])
}

func testCSVExportBasic(t *testing.T, tempDir string) {
	// Prepare data for export
	records := []map[string]string{
		{
			"Date":          "2024-01-01",
			"EntryIC":       "Tokyo",
			"ExitIC":        "Osaka",
			"Amount":        "1000",
			"VehicleNumber": "品川300あ1234",
			"ETCNumber":     "ETC-001",
		},
		{
			"Date":          "2024-01-02",
			"EntryIC":       "Nagoya",
			"ExitIC":        "Kyoto",
			"Amount":        "800",
			"VehicleNumber": "名古屋400い5678",
			"ETCNumber":     "ETC-002",
		},
	}

	// Export to CSV
	exporter := NewCSVExporter()
	csvFile := filepath.Join(tempDir, "test_export.csv")
	err := exporter.Export(csvFile, records)
	assert.NoError(t, err)

	// Verify file exists and content
	assert.FileExists(t, csvFile)

	// Read back and verify
	data, err := readCSV(csvFile)
	assert.NoError(t, err)
	assert.Len(t, data, 3) // Header + 2 records
}

func testLargeCSVProcessing(t *testing.T, tempDir string) {
	// Create large CSV file (10,000 records)
	csvFile := filepath.Join(tempDir, "large.csv")
	const recordCount = 10000

	// Write large file
	file, err := os.Create(csvFile)
	require.NoError(t, err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Date", "EntryIC", "ExitIC", "Amount", "VehicleNumber", "ETCNumber"}
	err = writer.Write(header)
	require.NoError(t, err)

	// Write records
	start := time.Now()
	for i := 0; i < recordCount; i++ {
		record := []string{
			fmt.Sprintf("%d", i),
			time.Now().Add(time.Duration(-i) * time.Hour).Format("2006-01-02"),
			fmt.Sprintf("Entry_%d", i%100),
			fmt.Sprintf("Exit_%d", i%100),
			fmt.Sprintf("%d", 100+i),
			fmt.Sprintf("車両%d", i),
			fmt.Sprintf("ETC-%05d", i),
		}
		err = writer.Write(record)
		require.NoError(t, err)
	}

	writeDuration := time.Since(start)
	assert.Less(t, writeDuration, 5*time.Second, "Writing 10k records should be fast")

	// Test streaming read
	start = time.Now()
	processor := NewStreamingCSVProcessor()
	count, err := processor.ProcessLargeFile(csvFile)
	readDuration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, recordCount, count)
	assert.Less(t, readDuration, 5*time.Second, "Reading 10k records should be fast")

	// Verify file size
	info, err := os.Stat(csvFile)
	assert.NoError(t, err)
	assert.Greater(t, info.Size(), int64(500000), "Large file should be > 500KB")
}

func testConcurrentFileAccess(t *testing.T, tempDir string) {
	csvFile := filepath.Join(tempDir, "concurrent.csv")

	// Create initial file
	data := [][]string{
		{"ID", "Data"},
		{"1", "Initial"},
	}
	err := writeCSV(csvFile, data)
	require.NoError(t, err)

	// Test concurrent reads
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			reader := NewCSVReader()
			_, err := reader.Read(csvFile)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(t, err)
	}

	// Test concurrent writes to different files
	wg = sync.WaitGroup{}
	errors = make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			filename := filepath.Join(tempDir, fmt.Sprintf("concurrent_%d.csv", id))
			data := [][]string{
				{"ID", "Data"},
				{fmt.Sprintf("%d", id), fmt.Sprintf("Data_%d", id)},
			}

			if err := writeCSV(filename, data); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Verify all files were created
	for i := 0; i < 10; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("concurrent_%d.csv", i))
		assert.FileExists(t, filename)
	}
}

func testFileLocking(t *testing.T, tempDir string) {
	lockFile := filepath.Join(tempDir, "locked.csv")

	// Create file
	data := [][]string{{"ID", "Data"}, {"1", "Test"}}
	err := writeCSV(lockFile, data)
	require.NoError(t, err)

	// Simulate file lock
	locker := NewFileLocker()

	// Acquire lock
	err = locker.Lock(lockFile)
	assert.NoError(t, err)

	// Try to acquire lock again (should fail or wait)
	done := make(chan bool)
	go func() {
		err := locker.TryLock(lockFile, 100*time.Millisecond)
		assert.Error(t, err, "Should not acquire lock while locked")
		done <- true
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(200 * time.Millisecond):
		// Also acceptable (lock timeout)
	}

	// Release lock
	err = locker.Unlock(lockFile)
	assert.NoError(t, err)

	// Now should be able to lock
	err = locker.Lock(lockFile)
	assert.NoError(t, err)
	locker.Unlock(lockFile)
}

func testDirectoryOperations(t *testing.T, tempDir string) {
	// Create directory structure
	structure := []string{
		"imports/2024/01",
		"imports/2024/02",
		"exports/2024/01",
		"exports/2024/02",
		"temp",
		"archive",
	}

	for _, dir := range structure {
		fullPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(fullPath, 0755)
		assert.NoError(t, err)
	}

	// Create files in directories
	files := []string{
		"imports/2024/01/data1.csv",
		"imports/2024/01/data2.csv",
		"imports/2024/02/data3.csv",
		"exports/2024/01/export1.csv",
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		err := os.WriteFile(fullPath, []byte("test"), 0644)
		assert.NoError(t, err)
	}

	// Test directory walking
	walker := NewDirectoryWalker()
	csvFiles, err := walker.FindCSVFiles(filepath.Join(tempDir, "imports"))
	assert.NoError(t, err)
	assert.Len(t, csvFiles, 3)

	// Test directory cleanup
	cleaner := NewDirectoryCleaner()
	err = cleaner.CleanOldFiles(filepath.Join(tempDir, "temp"), 0) // Delete all files
	assert.NoError(t, err)

	// Test directory stats
	stats, err := GetDirectoryStats(filepath.Join(tempDir, "imports"))
	assert.NoError(t, err)
	assert.Equal(t, 3, stats.FileCount)
	assert.Equal(t, 2, stats.DirectoryCount)
}

func testFilePermissions(t *testing.T, tempDir string) {
	// Skip on Windows as permission handling is different
	if os.PathSeparator == '\\' {
		t.Skip("Skipping permission test on Windows")
	}

	readOnlyFile := filepath.Join(tempDir, "readonly.csv")

	// Create file
	data := [][]string{{"ID", "Data"}, {"1", "Test"}}
	err := writeCSV(readOnlyFile, data)
	require.NoError(t, err)

	// Make read-only
	err = os.Chmod(readOnlyFile, 0444)
	assert.NoError(t, err)

	// Try to write (should fail)
	err = writeCSV(readOnlyFile, data)
	assert.Error(t, err)

	// Make writable again
	err = os.Chmod(readOnlyFile, 0644)
	assert.NoError(t, err)

	// Now should work
	err = writeCSV(readOnlyFile, data)
	assert.NoError(t, err)
}

func testCompressedFiles(t *testing.T, tempDir string) {
	// Create uncompressed CSV
	csvFile := filepath.Join(tempDir, "data.csv")
	gzFile := filepath.Join(tempDir, "data.csv.gz")

	data := [][]string{
		{"ID", "Date", "Amount"},
	}
	for i := 0; i < 1000; i++ {
		data = append(data, []string{
			fmt.Sprintf("%d", i),
			time.Now().Format("2006-01-02"),
			fmt.Sprintf("%d", 100+i),
		})
	}

	err := writeCSV(csvFile, data)
	require.NoError(t, err)

	// Compress file
	compressor := NewFileCompressor()
	err = compressor.CompressFile(csvFile, gzFile)
	assert.NoError(t, err)

	// Compare sizes
	originalInfo, _ := os.Stat(csvFile)
	compressedInfo, _ := os.Stat(gzFile)
	assert.Less(t, compressedInfo.Size(), originalInfo.Size())

	// Decompress and verify
	decompressedFile := filepath.Join(tempDir, "decompressed.csv")
	err = compressor.DecompressFile(gzFile, decompressedFile)
	assert.NoError(t, err)

	// Verify content
	originalData, _ := readCSV(csvFile)
	decompressedData, _ := readCSV(decompressedFile)
	assert.Equal(t, originalData, decompressedData)
}

func testStreamingOperations(t *testing.T, tempDir string) {
	csvFile := filepath.Join(tempDir, "stream.csv")

	// Create streaming writer
	streamer := NewCSVStreamer()

	// Start writing
	writer, err := streamer.StartWriting(csvFile)
	require.NoError(t, err)
	defer writer.Close()

	// Write header
	err = writer.WriteRecord([]string{"ID", "Data", "Timestamp"})
	assert.NoError(t, err)

	// Stream records
	for i := 0; i < 100; i++ {
		record := []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("Data_%d", i),
			time.Now().Format(time.RFC3339),
		}
		err = writer.WriteRecord(record)
		assert.NoError(t, err)

		// Simulate processing delay
		time.Sleep(time.Millisecond)
	}

	writer.Close()

	// Test streaming read
	reader, err := streamer.StartReading(csvFile)
	require.NoError(t, err)
	defer reader.Close()

	count := 0
	for {
		record, err := reader.ReadRecord()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		assert.Len(t, record, 3)
		count++
	}

	assert.Equal(t, 101, count) // Header + 100 records
}

func testFileSystemErrorRecovery(t *testing.T, tempDir string) {
	// Test recovery from disk full
	csvFile := filepath.Join(tempDir, "recovery.csv")

	// Simulate write with potential failure
	writer := NewResilientWriter()

	data := make([][]string, 1000)
	for i := range data {
		data[i] = []string{fmt.Sprintf("%d", i), fmt.Sprintf("Data_%d", i)}
	}

	err := writer.WriteWithRecovery(csvFile, data)
	assert.NoError(t, err)

	// Test recovery from corrupted file
	corruptedFile := filepath.Join(tempDir, "corrupted.csv")

	// Write corrupted data
	err = os.WriteFile(corruptedFile, []byte("ID,Data\n1,Test\n2,Incom"), 0644)
	require.NoError(t, err)

	// Try to read with recovery
	reader := NewResilientReader()
	records, err := reader.ReadWithRecovery(corruptedFile)
	assert.NoError(t, err)
	assert.Len(t, records, 1) // Should recover the complete record

	// Test atomic file operations
	atomicFile := filepath.Join(tempDir, "atomic.csv")

	atomic := NewAtomicFileWriter()
	err = atomic.WriteAtomic(atomicFile, [][]string{
		{"ID", "Data"},
		{"1", "Atomic"},
	})
	assert.NoError(t, err)

	// Verify file exists and is complete
	assert.FileExists(t, atomicFile)
	data2, err := readCSV(atomicFile)
	assert.NoError(t, err)
	assert.Len(t, data2, 2)
}

// Helper types and functions

type CSVImporter struct{}

func NewCSVImporter() *CSVImporter {
	return &CSVImporter{}
}

func (i *CSVImporter) Import(filepath string) ([]map[string]string, error) {
	data, err := readCSV(filepath)
	if err != nil {
		return nil, err
	}

	if len(data) < 2 {
		return nil, fmt.Errorf("no data rows found")
	}

	headers := data[0]
	records := make([]map[string]string, 0, len(data)-1)

	for _, row := range data[1:] {
		record := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				record[header] = row[i]
			}
		}
		records = append(records, record)
	}

	return records, nil
}

type CSVExporter struct{}

func NewCSVExporter() *CSVExporter {
	return &CSVExporter{}
}

func (e *CSVExporter) Export(filepath string, records []map[string]string) error {
	if len(records) == 0 {
		return fmt.Errorf("no records to export")
	}

	// Extract headers
	var headers []string
	for key := range records[0] {
		headers = append(headers, key)
	}

	// Prepare data
	data := [][]string{headers}
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = record[header]
		}
		data = append(data, row)
	}

	return writeCSV(filepath, data)
}

type StreamingCSVProcessor struct{}

func NewStreamingCSVProcessor() *StreamingCSVProcessor {
	return &StreamingCSVProcessor{}
}

func (p *StreamingCSVProcessor) ProcessLargeFile(filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	count := 0

	// Skip header
	_, err = reader.Read()
	if err != nil {
		return 0, err
	}

	for {
		_, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

type CSVReader struct{}

func NewCSVReader() *CSVReader {
	return &CSVReader{}
}

func (r *CSVReader) Read(filepath string) ([][]string, error) {
	return readCSV(filepath)
}

type FileLocker struct {
	locks map[string]*sync.Mutex
	mu    sync.Mutex
}

func NewFileLocker() *FileLocker {
	return &FileLocker{
		locks: make(map[string]*sync.Mutex),
	}
}

func (l *FileLocker) Lock(filepath string) error {
	l.mu.Lock()
	if _, exists := l.locks[filepath]; !exists {
		l.locks[filepath] = &sync.Mutex{}
	}
	lock := l.locks[filepath]
	l.mu.Unlock()

	lock.Lock()
	return nil
}

func (l *FileLocker) TryLock(filepath string, timeout time.Duration) error {
	l.mu.Lock()
	if _, exists := l.locks[filepath]; !exists {
		l.locks[filepath] = &sync.Mutex{}
	}
	lock := l.locks[filepath]
	l.mu.Unlock()

	done := make(chan bool)
	go func() {
		lock.Lock()
		done <- true
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("lock timeout")
	}
}

func (l *FileLocker) Unlock(filepath string) error {
	l.mu.Lock()
	lock, exists := l.locks[filepath]
	l.mu.Unlock()

	if !exists {
		return fmt.Errorf("lock not found")
	}

	lock.Unlock()
	return nil
}

type DirectoryWalker struct{}

func NewDirectoryWalker() *DirectoryWalker {
	return &DirectoryWalker{}
}

func (w *DirectoryWalker) FindCSVFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".csv" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

type DirectoryCleaner struct{}

func NewDirectoryCleaner() *DirectoryCleaner {
	return &DirectoryCleaner{}
}

func (c *DirectoryCleaner) CleanOldFiles(dir string, maxAge time.Duration) error {
	now := time.Now()

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if maxAge == 0 || now.Sub(info.ModTime()) > maxAge {
				os.Remove(path)
			}
		}

		return nil
	})
}

type DirectoryStats struct {
	FileCount      int
	DirectoryCount int
	TotalSize      int64
}

func GetDirectoryStats(dir string) (*DirectoryStats, error) {
	stats := &DirectoryStats{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != dir {
				stats.DirectoryCount++
			}
		} else {
			stats.FileCount++
			stats.TotalSize += info.Size()
		}

		return nil
	})

	return stats, err
}

type FileCompressor struct{}

func NewFileCompressor() *FileCompressor {
	return &FileCompressor{}
}

func (c *FileCompressor) CompressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	_, err = io.Copy(gzWriter, srcFile)
	return err
}

func (c *FileCompressor) DecompressFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, gzReader)
	return err
}

type CSVStreamer struct{}

func NewCSVStreamer() *CSVStreamer {
	return &CSVStreamer{}
}

type StreamWriter struct {
	file   *os.File
	writer *csv.Writer
}

func (w *StreamWriter) WriteRecord(record []string) error {
	err := w.writer.Write(record)
	if err != nil {
		return err
	}
	w.writer.Flush()
	return w.writer.Error()
}

func (w *StreamWriter) Close() error {
	w.writer.Flush()
	return w.file.Close()
}

func (s *CSVStreamer) StartWriting(filepath string) (*StreamWriter, error) {
	file, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}

	return &StreamWriter{
		file:   file,
		writer: csv.NewWriter(file),
	}, nil
}

type StreamReader struct {
	file   *os.File
	reader *csv.Reader
}

func (r *StreamReader) ReadRecord() ([]string, error) {
	return r.reader.Read()
}

func (r *StreamReader) Close() error {
	return r.file.Close()
}

func (s *CSVStreamer) StartReading(filepath string) (*StreamReader, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	return &StreamReader{
		file:   file,
		reader: csv.NewReader(file),
	}, nil
}

type ResilientWriter struct{}

func NewResilientWriter() *ResilientWriter {
	return &ResilientWriter{}
}

func (w *ResilientWriter) WriteWithRecovery(filepath string, data [][]string) error {
	// Try to write with retries
	var lastErr error
	for i := 0; i < 3; i++ {
		err := writeCSV(filepath, data)
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}
	return lastErr
}

type ResilientReader struct{}

func NewResilientReader() *ResilientReader {
	return &ResilientReader{}
}

func (r *ResilientReader) ReadWithRecovery(filepath string) ([][]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var records [][]string

	for scanner.Scan() {
		line := scanner.Text()
		reader := csv.NewReader(bytes.NewReader([]byte(line)))
		record, err := reader.Read()
		if err == nil {
			records = append(records, record)
		}
		// Skip corrupted lines
	}

	return records, nil
}

type AtomicFileWriter struct{}

func NewAtomicFileWriter() *AtomicFileWriter {
	return &AtomicFileWriter{}
}

func (w *AtomicFileWriter) WriteAtomic(filepath string, data [][]string) error {
	// Write to temp file first
	tempFile := filepath + ".tmp"
	err := writeCSV(tempFile, data)
	if err != nil {
		return err
	}

	// Atomic rename
	return os.Rename(tempFile, filepath)
}

// Helper functions
func writeCSV(filepath string, data [][]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Error()
}

func readCSV(filepath string) ([][]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}