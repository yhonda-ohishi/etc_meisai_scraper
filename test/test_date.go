package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Printf("Today: %s\n", now.Format("2006-01-02"))

	// Calculate last month
	lastMonth := now.AddDate(0, -1, 0)
	fmt.Printf("Last month (AddDate): %s\n", lastMonth.Format("2006-01-02"))

	// First day of last month
	firstDay := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	fmt.Printf("First day of last month: %s\n", firstDay.Format("2006-01-02"))

	// Last day of last month
	lastDay := time.Date(lastMonth.Year(), lastMonth.Month()+1, 0, 0, 0, 0, 0, time.Local)
	fmt.Printf("Last day of last month: %s\n", lastDay.Format("2006-01-02"))

	// Format for ETC site (YYYYMMDD)
	fmt.Printf("\nETC Format:\n")
	fmt.Printf("From: %s\n", firstDay.Format("20060102"))
	fmt.Printf("To: %s\n", lastDay.Format("20060102"))
}