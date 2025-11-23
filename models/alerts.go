package models

// SystemAlertConfig holds configuration for system resource alerts
type SystemAlertConfig struct {
	DiskSpaceThreshold float64 `json:"disk_space_threshold"` // Alert when disk usage > this % (e.g., 80.0)
	CPUThreshold       float64 `json:"cpu_threshold"`        // Alert when CPU usage > this % (e.g., 90.0)
	MemoryThreshold    float64 `json:"memory_threshold"`     // Alert when memory usage > this % (e.g., 90.0)
	Enabled            bool    `json:"enabled"`
}

// AlertState tracks which alerts have been sent to prevent spam
type AlertState struct {
	DiskAlertSent   bool `json:"disk_alert_sent"`
	CPUAlertSent    bool `json:"cpu_alert_sent"`
	MemoryAlertSent bool `json:"memory_alert_sent"`
}

// SystemAlert represents a system resource alert
type SystemAlert struct {
	Type        string  `json:"type"`        // "disk", "cpu", "memory"
	Message     string  `json:"message"`
	CurrentValue float64 `json:"current_value"`
	Threshold   float64 `json:"threshold"`
	Device      string  `json:"device,omitempty"` // For disk alerts
}
