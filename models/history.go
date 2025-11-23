package models

import "time"

// ServiceHistory stores historical check results for a service
type ServiceHistory struct {
	ServiceID string              `json:"service_id"`
	Checks    []HealthCheckRecord `json:"checks"`
	MaxChecks int                 `json:"max_checks"` // Maximum number of checks to keep
}

// HealthCheckRecord represents a single health check result with timestamp
type HealthCheckRecord struct {
	Timestamp    time.Time     `json:"timestamp"`
	Status       ServiceStatus `json:"status"`
	ResponseTime int64         `json:"response_time"` // in milliseconds
	ErrorMessage string        `json:"error_message,omitempty"`
}

// NewServiceHistory creates a new service history
func NewServiceHistory(serviceID string, maxChecks int) *ServiceHistory {
	if maxChecks == 0 {
		maxChecks = 100 // default to keep last 100 checks
	}
	return &ServiceHistory{
		ServiceID: serviceID,
		Checks:    make([]HealthCheckRecord, 0),
		MaxChecks: maxChecks,
	}
}

// AddCheck adds a new check result to the history
func (h *ServiceHistory) AddCheck(record HealthCheckRecord) {
	h.Checks = append(h.Checks, record)

	// Keep only the last MaxChecks entries
	if len(h.Checks) > h.MaxChecks {
		h.Checks = h.Checks[len(h.Checks)-h.MaxChecks:]
	}
}

// GetStatistics calculates statistics from the history
func (h *ServiceHistory) GetStatistics() *ServiceStatistics {
	stats := &ServiceStatistics{
		ServiceID:  h.ServiceID,
		TotalChecks: len(h.Checks),
	}

	if len(h.Checks) == 0 {
		return stats
	}

	var totalResponseTime int64
	var upCount, downCount int
	var uptimeStart *time.Time
	var downtimeStart *time.Time
	var currentUptime, currentDowntime time.Duration
	var totalUptime, totalDowntime time.Duration

	for i, check := range h.Checks {
		if check.Status == StatusUp {
			upCount++
			totalResponseTime += check.ResponseTime

			if uptimeStart == nil {
				uptimeStart = &check.Timestamp
			}
		} else if check.Status == StatusDown {
			downCount++

			if downtimeStart == nil {
				downtimeStart = &check.Timestamp
			}
		}

		// Track uptime/downtime periods
		if i > 0 {
			prevCheck := h.Checks[i-1]
			duration := check.Timestamp.Sub(prevCheck.Timestamp)

			if prevCheck.Status == StatusUp {
				totalUptime += duration
			} else if prevCheck.Status == StatusDown {
				totalDowntime += duration
			}
		}
	}

	// Calculate current uptime/downtime
	lastCheck := h.Checks[len(h.Checks)-1]
	if lastCheck.Status == StatusUp && uptimeStart != nil {
		currentUptime = time.Since(*uptimeStart)
	} else if lastCheck.Status == StatusDown && downtimeStart != nil {
		currentDowntime = time.Since(*downtimeStart)
	}

	stats.UpCount = upCount
	stats.DownCount = downCount

	if upCount > 0 {
		stats.UptimePercentage = float64(upCount) / float64(len(h.Checks)) * 100
		stats.AverageResponseTime = totalResponseTime / int64(upCount)
	}

	stats.CurrentUptime = int64(currentUptime.Seconds())
	stats.CurrentDowntime = int64(currentDowntime.Seconds())
	stats.TotalUptime = int64(totalUptime.Seconds())
	stats.TotalDowntime = int64(totalDowntime.Seconds())

	// Get last 24 hours of checks
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	for _, check := range h.Checks {
		if check.Timestamp.After(oneDayAgo) {
			stats.Last24Hours = append(stats.Last24Hours, check)
		}
	}

	return stats
}

// ServiceStatistics represents aggregated statistics for a service
type ServiceStatistics struct {
	ServiceID           string              `json:"service_id"`
	TotalChecks         int                 `json:"total_checks"`
	UpCount             int                 `json:"up_count"`
	DownCount           int                 `json:"down_count"`
	UptimePercentage    float64             `json:"uptime_percentage"`
	AverageResponseTime int64               `json:"average_response_time"` // in milliseconds
	CurrentUptime       int64               `json:"current_uptime"`        // in seconds
	CurrentDowntime     int64               `json:"current_downtime"`      // in seconds
	TotalUptime         int64               `json:"total_uptime"`          // in seconds
	TotalDowntime       int64               `json:"total_downtime"`        // in seconds
	Last24Hours         []HealthCheckRecord `json:"last_24_hours"`
}
