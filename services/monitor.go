package services

import (
	"crypto/tls"
	"fmt"
	"monitoring/models"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MonitorService handles health checking of services
type MonitorService struct {
	store    *models.ServiceStore
	history  *models.HistoryStore
	telegram *TelegramService
}

// NewMonitorService creates a new monitor service
func NewMonitorService(store *models.ServiceStore, history *models.HistoryStore, telegram *TelegramService) *MonitorService {
	return &MonitorService{
		store:    store,
		history:  history,
		telegram: telegram,
	}
}

// CheckService performs a health check on a single service
func (m *MonitorService) CheckService(service *models.MonitoredService) *models.HealthCheckResult {
	result := &models.HealthCheckResult{
		ServiceID: service.ID,
		CheckedAt: time.Now(),
	}

	// Default to HTTP if not specified (for backward compatibility)
	checkType := service.CheckType
	if checkType == "" {
		checkType = models.CheckTypeHTTP
	}

	// Route to appropriate check method based on type
	switch checkType {
	case models.CheckTypeTCP:
		return m.checkTCPPort(service, result)
	case models.CheckTypeUDP:
		return m.checkUDPPort(service, result)
	default: // HTTP
		return m.checkHTTP(service, result)
	}
}

// checkHTTP performs an HTTP/HTTPS health check
func (m *MonitorService) checkHTTP(service *models.MonitoredService, result *models.HealthCheckResult) *models.HealthCheckResult {
	timeout := time.Duration(service.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second // default timeout
	}

	client := &http.Client{
		Timeout: timeout,
	}

	start := time.Now()
	resp, err := client.Get(service.URL)
	elapsed := time.Since(start)

	result.ResponseTime = elapsed.Milliseconds()

	if err != nil {
		result.Status = models.StatusDown
		result.ErrorMessage = err.Error()
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Status = models.StatusUp
	} else {
		result.Status = models.StatusDown
		result.ErrorMessage = fmt.Sprintf("HTTP status code: %d", resp.StatusCode)
	}

	// Check SSL certificate if it's an HTTPS URL
	if strings.HasPrefix(strings.ToLower(service.URL), "https://") {
		m.checkSSLCertificate(service.URL, result)
	}

	return result
}

// checkTCPPort performs a TCP port health check
func (m *MonitorService) checkTCPPort(service *models.MonitoredService, result *models.HealthCheckResult) *models.HealthCheckResult {
	timeout := time.Duration(service.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	address := fmt.Sprintf("%s:%d", service.Host, service.Port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	elapsed := time.Since(start)

	result.ResponseTime = elapsed.Milliseconds()

	if err != nil {
		result.Status = models.StatusDown
		result.ErrorMessage = fmt.Sprintf("TCP connection failed: %v", err)
		return result
	}
	defer conn.Close()

	result.Status = models.StatusUp
	return result
}

// checkUDPPort performs a UDP port health check
func (m *MonitorService) checkUDPPort(service *models.MonitoredService, result *models.HealthCheckResult) *models.HealthCheckResult {
	timeout := time.Duration(service.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	address := fmt.Sprintf("%s:%d", service.Host, service.Port)
	start := time.Now()

	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		result.Status = models.StatusDown
		result.ErrorMessage = fmt.Sprintf("UDP address resolution failed: %v", err)
		return result
	}

	// Dial UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		elapsed := time.Since(start)
		result.ResponseTime = elapsed.Milliseconds()
		result.Status = models.StatusDown
		result.ErrorMessage = fmt.Sprintf("UDP connection failed: %v", err)
		return result
	}
	defer conn.Close()

	// Set deadline for the connection
	conn.SetDeadline(time.Now().Add(timeout))

	// Send a simple probe packet
	_, err = conn.Write([]byte("ping"))
	if err != nil {
		elapsed := time.Since(start)
		result.ResponseTime = elapsed.Milliseconds()
		result.Status = models.StatusDown
		result.ErrorMessage = fmt.Sprintf("UDP write failed: %v", err)
		return result
	}

	// Try to read response (with timeout)
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buffer)

	elapsed := time.Since(start)
	result.ResponseTime = elapsed.Milliseconds()

	// For UDP, we consider it "up" if we can establish connection and send data
	// Not receiving a response doesn't necessarily mean the port is down
	// as many UDP services don't respond to random data
	if err != nil {
		// Check if it's a timeout (which is expected for many UDP services)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.Status = models.StatusUp
			result.ErrorMessage = "UDP port is open (no response received, which is normal)"
		} else {
			result.Status = models.StatusDown
			result.ErrorMessage = fmt.Sprintf("UDP check failed: %v", err)
		}
	} else {
		result.Status = models.StatusUp
	}

	return result
}

// checkSSLCertificate checks the SSL certificate of an HTTPS URL
func (m *MonitorService) checkSSLCertificate(serviceURL string, result *models.HealthCheckResult) {
	parsedURL, err := url.Parse(serviceURL)
	if err != nil {
		return
	}

	host := parsedURL.Host
	if parsedURL.Port() == "" {
		host += ":443"
	}

	conn, err := tls.Dial("tcp", host, &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		return
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) > 0 {
		cert := certs[0]
		result.SSLCertExpiry = cert.NotAfter
		result.SSLCertIssuer = cert.Issuer.CommonName
		result.SSLDaysLeft = int(time.Until(cert.NotAfter).Hours() / 24)
	}
}

// UpdateServiceStatus updates the service based on health check result
func (m *MonitorService) UpdateServiceStatus(result *models.HealthCheckResult) error {
	service, err := m.store.Get(result.ServiceID)
	if err != nil {
		return err
	}

	// Track previous status for notification logic
	previousStatus := service.Status

	service.Status = result.Status
	service.LastCheck = result.CheckedAt
	service.ResponseTime = result.ResponseTime
	service.ErrorMessage = result.ErrorMessage

	// Update SSL certificate info if available
	if !result.SSLCertExpiry.IsZero() {
		service.SSLCertExpiry = result.SSLCertExpiry
		service.SSLCertIssuer = result.SSLCertIssuer
		service.SSLDaysLeft = result.SSLDaysLeft

		// Send SSL expiry alert if certificate expires in 30 days or less
		if result.SSLDaysLeft <= 30 && result.SSLDaysLeft > 0 && !service.SSLAlertSent {
			go func() {
				if err := m.telegram.SendSSLExpiryAlert(service); err != nil {
					fmt.Printf("Failed to send SSL expiry alert for %s: %v\n", service.Name, err)
				}
			}()
			service.SSLAlertSent = true
		} else if result.SSLDaysLeft > 30 {
			// Reset alert flag when SSL has more than 30 days
			service.SSLAlertSent = false
		}
	}

	// Log the check result to history
	m.history.AddCheckResult(result.ServiceID, models.HealthCheckRecord{
		Timestamp:    result.CheckedAt,
		Status:       result.Status,
		ResponseTime: result.ResponseTime,
		ErrorMessage: result.ErrorMessage,
	})

	if result.Status == models.StatusUp {
		service.LastUptime = result.CheckedAt
		// Send recovery notification if service was previously down
		if previousStatus == models.StatusDown {
			go func() {
				if err := m.telegram.SendServiceUpAlert(service); err != nil {
					fmt.Printf("Failed to send Telegram up alert for %s: %v\n", service.Name, err)
				}
			}()
		}
	} else if result.Status == models.StatusDown {
		service.LastDowntime = result.CheckedAt
		// Send down notification if service was previously up or unknown
		if previousStatus == models.StatusUp || previousStatus == models.StatusUnknown {
			go func() {
				if err := m.telegram.SendServiceDownAlert(service); err != nil {
					fmt.Printf("Failed to send Telegram down alert for %s: %v\n", service.Name, err)
				}
			}()
		}
	}

	return m.store.Update(service)
}

// CheckAll performs health checks on all monitored services
func (m *MonitorService) CheckAll() {
	services := m.store.GetAll()

	for _, service := range services {
		result := m.CheckService(service)
		if err := m.UpdateServiceStatus(result); err != nil {
			fmt.Printf("Error updating service %s: %v\n", service.ID, err)
		}
	}
}
