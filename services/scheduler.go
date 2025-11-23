package services

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler manages periodic health checks
type Scheduler struct {
	cron    *cron.Cron
	monitor *MonitorService
}

// NewScheduler creates a new scheduler
func NewScheduler(monitor *MonitorService) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		monitor: monitor,
	}
}

// Start begins the scheduled health checks
func (s *Scheduler) Start() {
	// Run health checks every 30 seconds
	_, err := s.cron.AddFunc("*/30 * * * * *", func() {
		fmt.Printf("[%s] Running scheduled health checks\n", time.Now().Format("2006-01-02 15:04:05"))
		s.monitor.CheckAll()
	})

	if err != nil {
		fmt.Printf("Error adding cron job: %v\n", err)
		return
	}

	s.cron.Start()
	fmt.Println("Scheduler started - health checks will run every 30 seconds")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	fmt.Println("Scheduler stopped")
}
