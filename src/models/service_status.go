package models

import "time"

// ServiceStatus represents the health status of a service
type ServiceStatus struct {
	Healthy   bool      `json:"healthy"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}