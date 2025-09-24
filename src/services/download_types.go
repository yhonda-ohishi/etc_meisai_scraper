package services

import "time"

// DownloadJobStatus represents the status of a download job
type DownloadJobStatus struct {
	JobID         string        `json:"job_id"`
	Status        string        `json:"status"` // pending, running, completed, failed, cancelled
	Progress      float64       `json:"progress"` // 0.0 to 1.0
	TotalFiles    int           `json:"total_files"`
	CompletedFiles int          `json:"completed_files"`
	StartedAt     time.Time     `json:"started_at"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	EstimatedTime *time.Duration `json:"estimated_time,omitempty"`
	Message       string        `json:"message,omitempty"`
	Errors        []string      `json:"errors,omitempty"`
}

// DownloadServiceInterface defines the contract for download service operations
type DownloadServiceInterface interface {
	// Account management
	GetAllAccountIDs() []string

	// Job management
	ProcessAsync(jobID string, accounts []string, fromDate, toDate string)
	GetJobStatus(jobID string) (*DownloadJob, bool)
}