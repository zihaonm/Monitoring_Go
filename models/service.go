package models

import "time"

// ServiceStatus represents the current status of a monitored service
type ServiceStatus string

const (
	StatusUp      ServiceStatus = "up"
	StatusDown    ServiceStatus = "down"
	StatusUnknown ServiceStatus = "unknown"
)

// CheckType represents the type of health check
type CheckType string

const (
	CheckTypeHTTP CheckType = "http"
	CheckTypeTCP  CheckType = "tcp"
	CheckTypeUDP  CheckType = "udp"
)

// MonitoredService represents a service or website to be monitored
type MonitoredService struct {
	ID              string        `json:"id"`
	Name            string        `json:"name" binding:"required"`
	CheckType       CheckType     `json:"check_type"`                // http, tcp, or udp
	URL             string        `json:"url"`                       // For HTTP checks
	Host            string        `json:"host"`                      // For TCP/UDP checks
	Port            int           `json:"port"`                      // For TCP/UDP checks
	CheckInterval   int           `json:"check_interval"`            // in seconds
	Timeout         int           `json:"timeout"`                   // in seconds
	Status          ServiceStatus `json:"status"`
	LastCheck       time.Time     `json:"last_check"`
	LastUptime      time.Time     `json:"last_uptime"`
	LastDowntime    time.Time     `json:"last_downtime"`
	ResponseTime    int64         `json:"response_time"`             // in milliseconds
	ErrorMessage    string        `json:"error_message,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	SSLCertExpiry   time.Time     `json:"ssl_cert_expiry,omitempty"` // SSL certificate expiry date
	SSLCertIssuer   string        `json:"ssl_cert_issuer,omitempty"` // SSL certificate issuer
	SSLDaysLeft     int           `json:"ssl_days_left,omitempty"`   // Days until SSL expires
	SSLAlertSent    bool          `json:"ssl_alert_sent,omitempty"`  // Track if SSL expiry alert was sent

	// Telegram alert overrides (optional, falls back to default if not set)
	TelegramBotToken string `json:"telegram_bot_token,omitempty"` // Override bot token for this service
	TelegramChatID   string `json:"telegram_chat_id,omitempty"`   // Override chat ID for this service
	TelegramEnabled  *bool  `json:"telegram_enabled,omitempty"`   // Override enabled status (nil = use default)
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	ServiceID     string
	Status        ServiceStatus
	ResponseTime  int64
	ErrorMessage  string
	CheckedAt     time.Time
	SSLCertExpiry time.Time
	SSLCertIssuer string
	SSLDaysLeft   int
}
