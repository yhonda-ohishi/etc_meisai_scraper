package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCanceled  JobStatus = "canceled"
)

// Job represents an async job
type Job struct {
	ID          string
	Type        string
	Status      JobStatus
	Progress    int
	Total       int
	Result      interface{}
	Error       error
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Context     context.Context
	Cancel      context.CancelFunc
}

// JobService manages async jobs
type JobService struct {
	jobs       map[string]*Job
	mu         sync.RWMutex
	maxWorkers int
	queue      chan *Job
	wg         sync.WaitGroup
}

// NewJobService creates a new job service
func NewJobService(maxWorkers int) *JobService {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}

	js := &JobService{
		jobs:       make(map[string]*Job),
		maxWorkers: maxWorkers,
		queue:      make(chan *Job, 100),
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		go js.worker()
	}

	return js
}

// CreateJob creates a new job
func (s *JobService) CreateJob(jobType string) *Job {
	ctx, cancel := context.WithCancel(context.Background())

	job := &Job{
		ID:        fmt.Sprintf("%s_%d", jobType, time.Now().UnixNano()),
		Type:      jobType,
		Status:    JobStatusPending,
		Progress:  0,
		Total:     100,
		CreatedAt: time.Now(),
		Context:   ctx,
		Cancel:    cancel,
	}

	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	return job
}

// SubmitJob submits a job for execution
func (s *JobService) SubmitJob(job *Job, fn func(*Job) error) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		// Add to queue
		s.queue <- job

		// Wait for execution
		<-job.Context.Done()
	}()

	// Store the execution function
	go func() {
		// Wait for worker to pick up
		for {
			s.mu.RLock()
			if job.Status == JobStatusRunning {
				s.mu.RUnlock()
				break
			}
			s.mu.RUnlock()
			time.Sleep(10 * time.Millisecond)
		}

		// Execute job
		err := fn(job)

		// Update job status
		s.mu.Lock()
		if err != nil {
			job.Status = JobStatusFailed
			job.Error = err
		} else {
			job.Status = JobStatusCompleted
			job.Progress = job.Total
		}
		now := time.Now()
		job.CompletedAt = &now
		s.mu.Unlock()

		// Signal completion
		job.Cancel()
	}()
}

// worker processes jobs from the queue
func (s *JobService) worker() {
	for job := range s.queue {
		s.mu.Lock()
		job.Status = JobStatusRunning
		now := time.Now()
		job.StartedAt = &now
		s.mu.Unlock()

		// Job will be executed by the goroutine in SubmitJob
		// Worker just marks it as running
	}
}

// GetJob retrieves a job by ID
func (s *JobService) GetJob(jobID string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// UpdateProgress updates job progress
func (s *JobService) UpdateProgress(jobID string, progress int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	job.Progress = progress
	return nil
}

// CancelJob cancels a running job
func (s *JobService) CancelJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status == JobStatusRunning || job.Status == JobStatusPending {
		job.Status = JobStatusCanceled
		job.Cancel()
		now := time.Now()
		job.CompletedAt = &now
	}

	return nil
}

// GetAllJobs returns all jobs
func (s *JobService) GetAllJobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// GetJobsByStatus returns jobs filtered by status
func (s *JobService) GetJobsByStatus(status JobStatus) []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*Job
	for _, job := range s.jobs {
		if job.Status == status {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

// CleanupOldJobs removes completed jobs older than specified duration
func (s *JobService) CleanupOldJobs(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, job := range s.jobs {
		if job.Status == JobStatusCompleted || job.Status == JobStatusFailed || job.Status == JobStatusCanceled {
			if job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
				delete(s.jobs, id)
				removed++
			}
		}
	}

	return removed
}

// GetStats returns job statistics
func (s *JobService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]int{
		"total":     len(s.jobs),
		"pending":   0,
		"running":   0,
		"completed": 0,
		"failed":    0,
		"canceled":  0,
	}

	for _, job := range s.jobs {
		switch job.Status {
		case JobStatusPending:
			stats["pending"]++
		case JobStatusRunning:
			stats["running"]++
		case JobStatusCompleted:
			stats["completed"]++
		case JobStatusFailed:
			stats["failed"]++
		case JobStatusCanceled:
			stats["canceled"]++
		}
	}

	return map[string]interface{}{
		"stats":      stats,
		"maxWorkers": s.maxWorkers,
		"queueSize":  len(s.queue),
	}
}

// Shutdown gracefully shuts down the job service
func (s *JobService) Shutdown() {
	// Cancel all pending/running jobs
	s.mu.Lock()
	for _, job := range s.jobs {
		if job.Status == JobStatusPending || job.Status == JobStatusRunning {
			job.Cancel()
		}
	}
	s.mu.Unlock()

	// Close queue
	close(s.queue)

	// Wait for workers to finish
	s.wg.Wait()
}