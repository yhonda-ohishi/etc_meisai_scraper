package parser_test

import (
	"os"
	"testing"

	"github.com/yhonda-ohishi/etc_meisai/src/parser"
	"github.com/yhonda-ohishi/etc_meisai/tests/helpers"
)

func TestEncodingDetector_DetectEncoding(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected parser.EncodingType
	}{
		{
			name:     "UTF-8 CSV content",
			content:  []byte("利用年月日,利用時刻,入口IC,出口IC\n2025/01/15,09:30,東京IC,大阪IC"),
			expected: parser.EncodingUTF8,
		},
		{
			name:     "ASCII content",
			content:  []byte("date,time,entry,exit"),
			expected: parser.EncodingUnknown, // ASCII without Japanese characters doesn't look like CSV
		},
		{
			name:     "binary content",
			content:  []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE},
			expected: parser.EncodingUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := parser.NewEncodingDetector()
			encoding := detector.DetectEncoding(tt.content)

			helpers.AssertEqual(t, tt.expected, encoding)
		})
	}
}

func TestEncodingDetector_DetectFileEncoding(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test_encoding_*.csv")
	helpers.AssertNoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := "利用年月日,利用時刻,入口IC,出口IC\n2025/01/15,09:30,東京IC,大阪IC"
	_, err = tmpfile.WriteString(content)
	helpers.AssertNoError(t, err)
	tmpfile.Close()

	detector := parser.NewEncodingDetector()
	encoding, err := detector.DetectFileEncoding(tmpfile.Name())

	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, parser.EncodingUTF8, encoding)
}

func TestEncodingDetector_OpenFileWithDetectedEncoding(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test_encoding_*.csv")
	helpers.AssertNoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := "利用年月日,利用時刻,入口IC,出口IC\n2025/01/15,09:30,東京IC,大阪IC"
	_, err = tmpfile.WriteString(content)
	helpers.AssertNoError(t, err)
	tmpfile.Close()

	detector := parser.NewEncodingDetector()
	reader, encoding, err := detector.OpenFileWithDetectedEncoding(tmpfile.Name())

	helpers.AssertNoError(t, err)
	helpers.AssertNotNil(t, reader)
	helpers.AssertEqual(t, parser.EncodingUTF8, encoding)

	// Ensure we can close the reader if needed
	if closer, ok := reader.(interface{ Close() error }); ok {
		err = closer.Close()
		helpers.AssertNoError(t, err)
	}
}