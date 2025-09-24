package models

import "time"

// JobStatus represents the status of an async job
type JobStatus struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Message     string    `json:"message,omitempty"`
	Error       string    `json:"error,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Result      interface{} `json:"result,omitempty"`
}