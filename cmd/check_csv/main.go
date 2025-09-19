package etc_meisai

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	file, err := os.Open("test_downloads/202509150949.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Create Shift-JIS decoder
	reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	csvReader := csv.NewReader(reader)

	fmt.Println("=== Latest CSV Content (All rows) ===")
	rowNum := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading CSV: %v\n", err)
			break
		}
		rowNum++
		if rowNum == 1 {
			fmt.Printf("Header: %v\n", record)
		} else {
			// Show date columns (first 2 columns)
			fmt.Printf("Row %d: Date=%s %s, IC=%s->%s, Amount=%s\n",
				rowNum, record[0], record[1], record[4], record[5], record[8])
		}
	}
}