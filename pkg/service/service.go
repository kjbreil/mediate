package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job represents a scheduled job that runs at a specified interval
type Job struct {
	Name     string
	Interval time.Duration
	Enabled  bool
	Fn       func() error
	logger   *slog.Logger
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

// Service manages the execution of multiple jobs
type Service struct {
	jobs   []*Job
	logger *slog.Logger
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewService creates a new service with the given logger
func NewService(logger *slog.Logger) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		jobs:   make([]*Job, 0),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddJob adds a new job to the service
func (s *Service) AddJob(name string, interval time.Duration, fn func() error) *Job {
	job := &Job{
		Name:     name,
		Interval: interval,
		Enabled:  true,
		Fn:       fn,
		logger:   s.logger,
		wg:       &s.wg,
	}
	s.jobs = append(s.jobs, job)
	return job
}

// Start starts all enabled jobs
func (s *Service) Start() {
	for _, job := range s.jobs {
		if job.Enabled {
			s.wg.Add(1)
			go s.runJob(job)
		}
	}
}

// runJob runs a job at the specified interval
func (s *Service) runJob(job *Job) {
	defer s.wg.Done()

	// Run the job immediately
	s.executeJob(job)

	// Set up a ticker to run the job at the specified interval
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	jobCtx, cancel := context.WithCancel(s.ctx)
	job.cancel = cancel

	for {
		select {
		case <-ticker.C:
			s.executeJob(job)
		case <-jobCtx.Done():
			s.logger.Info("Job stopped", "name", job.Name)
			return
		}
	}
}

// executeJob executes a job and logs any errors
func (s *Service) executeJob(job *Job) {
	s.logger.Info("Running job", "name", job.Name)
	start := time.Now()

	err := job.Fn()
	if err != nil {
		s.logger.Error("Job failed", "name", job.Name, "error", err, "duration", time.Since(start))
		return
	}

	s.logger.Info("Job completed", "name", job.Name, "duration", time.Since(start))
}

// Stop stops all running jobs and waits for them to complete
func (s *Service) Stop() {
	s.logger.Info("Stopping all jobs")
	s.cancel()
	s.wg.Wait()
	s.logger.Info("All jobs stopped")
}

// EnableJob enables a job by name
func (s *Service) EnableJob(name string) bool {
	for _, job := range s.jobs {
		if job.Name == name {
			job.Enabled = true
			return true
		}
	}
	return false
}

// DisableJob disables a job by name
func (s *Service) DisableJob(name string) bool {
	for _, job := range s.jobs {
		if job.Name == name {
			job.Enabled = false
			if job.cancel != nil {
				job.cancel()
			}
			return true
		}
	}
	return false
}

// GetJobs returns all jobs
func (s *Service) GetJobs() []*Job {
	return s.jobs
}
