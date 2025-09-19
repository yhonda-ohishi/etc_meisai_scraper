package parser

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// EncodingType represents the detected encoding
type EncodingType int

const (
	EncodingUTF8 EncodingType = iota
	EncodingShiftJIS
	EncodingUnknown
)

// String returns the string representation of the encoding type
func (e EncodingType) String() string {
	switch e {
	case EncodingUTF8:
		return "UTF-8"
	case EncodingShiftJIS:
		return "Shift-JIS"
	default:
		return "Unknown"
	}
}

// EncodingDetector handles automatic encoding detection
type EncodingDetector struct{}

// NewEncodingDetector creates a new encoding detector
func NewEncodingDetector() *EncodingDetector {
	return &EncodingDetector{}
}

// DetectFileEncoding detects the encoding of a file
func (d *EncodingDetector) DetectFileEncoding(filePath string) (EncodingType, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return EncodingUnknown, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the first few bytes to detect encoding
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return EncodingUnknown, fmt.Errorf("failed to read file: %w", err)
	}

	return d.DetectEncoding(buffer[:n]), nil
}

// DetectEncoding detects encoding from byte slice
func (d *EncodingDetector) DetectEncoding(data []byte) EncodingType {
	// Check for BOM
	if len(data) >= 3 && bytes.Equal(data[:3], []byte{0xEF, 0xBB, 0xBF}) {
		return EncodingUTF8
	}

	// Check if valid UTF-8
	if d.isValidUTF8(data) {
		return EncodingUTF8
	}

	// Try to decode as Shift-JIS and see if it produces valid results
	if d.canDecodeAsShiftJIS(data) {
		return EncodingShiftJIS
	}

	return EncodingUnknown
}

// OpenFileWithDetectedEncoding opens a file and returns a reader with the correct encoding
func (d *EncodingDetector) OpenFileWithDetectedEncoding(filePath string) (io.Reader, EncodingType, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, EncodingUnknown, fmt.Errorf("failed to open file: %w", err)
	}

	// First, detect the encoding
	encoding, err := d.DetectFileEncoding(filePath)
	if err != nil {
		file.Close()
		return nil, EncodingUnknown, err
	}

	// Reopen the file for reading
	file.Close()
	file, err = os.Open(filePath)
	if err != nil {
		return nil, EncodingUnknown, fmt.Errorf("failed to reopen file: %w", err)
	}

	switch encoding {
	case EncodingShiftJIS:
		// Convert Shift-JIS to UTF-8
		utf8Reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
		return utf8Reader, encoding, nil
	case EncodingUTF8:
		// Skip BOM if present
		if d.hasBOM(file) {
			file.Seek(3, io.SeekStart)
		}
		return file, encoding, nil
	default:
		// Default to UTF-8
		return file, encoding, nil
	}
}

// isValidUTF8 checks if the data is valid UTF-8
func (d *EncodingDetector) isValidUTF8(data []byte) bool {
	// Simple heuristic: if we can successfully convert to string and it contains
	// no replacement characters, it's likely UTF-8
	str := string(data)

	// Check for replacement characters
	for _, r := range str {
		if r == '\uFFFD' { // Unicode replacement character
			return false
		}
	}

	// Additional check: look for common CSV characters in reasonable positions
	return d.looksLikeCSV(str)
}

// canDecodeAsShiftJIS attempts to decode as Shift-JIS and checks if result is reasonable
func (d *EncodingDetector) canDecodeAsShiftJIS(data []byte) bool {
	decoder := japanese.ShiftJIS.NewDecoder()
	result, err := decoder.Bytes(data)
	if err != nil {
		return false
	}

	// Check if the decoded result looks like reasonable text
	str := string(result)
	return d.looksLikeCSV(str)
}

// looksLikeCSV performs heuristic checks to see if the string looks like CSV data
func (d *EncodingDetector) looksLikeCSV(str string) bool {
	// Look for common CSV indicators
	hasCommas := bytes.Count([]byte(str), []byte(",")) > 0
	hasNewlines := bytes.Count([]byte(str), []byte("\n")) > 0

	// Look for Japanese characters that would indicate ETC data
	hasJapanese := false
	for _, r := range str {
		if (r >= 0x3040 && r <= 0x309F) || // Hiragana
		   (r >= 0x30A0 && r <= 0x30FF) || // Katakana
		   (r >= 0x4E00 && r <= 0x9FAF) {  // CJK Unified Ideographs
			hasJapanese = true
			break
		}
	}

	return hasCommas && hasNewlines && hasJapanese
}

// hasBOM checks if the file starts with a UTF-8 BOM
func (d *EncodingDetector) hasBOM(file *os.File) bool {
	file.Seek(0, io.SeekStart)
	buffer := make([]byte, 3)
	n, err := file.Read(buffer)
	file.Seek(0, io.SeekStart) // Reset position

	if err != nil || n < 3 {
		return false
	}

	return bytes.Equal(buffer, []byte{0xEF, 0xBB, 0xBF})
}