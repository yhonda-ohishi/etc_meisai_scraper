package models

import "time"

// DownloadParams contains parameters for downloading ETC meisai data
type DownloadParams struct {
	AccountType string    `json:"account_type"`
	AccountID   string    `json:"account_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	SessionID   string    `json:"session_id"`
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	Status     string   `json:"status"`
	TotalFiles int      `json:"total_files"`
	Downloaded int      `json:"downloaded"`
	Failed     int      `json:"failed"`
	FilePaths  []string `json:"file_paths"`
	Errors     []string `json:"errors,omitempty"`
}

// DownloadStatus represents the current status of a download session
type DownloadStatus struct {
	SessionID    string    `json:"session_id"`
	Status       string    `json:"status"`
	Progress     int       `json:"progress"`
	TotalFiles   int       `json:"total_files"`
	Processed    int       `json:"processed"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// ProcessResult represents the result of processing downloaded files
type ProcessResult struct {
	TotalRecords int      `json:"total_records"`
	Imported     int      `json:"imported"`
	Duplicates   int      `json:"duplicates"`
	Errors       int      `json:"errors"`
	ErrorDetails []string `json:"error_details,omitempty"`
}