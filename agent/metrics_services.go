package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ServiceMetrics struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "running", "stopped", "unknown"
	PID    int    `json:"pid,omitempty"`
}

type PortMetrics struct {
	Port   int    `json:"port"`
	Status string `json:"status"` // "listening", "closed"
	Process string `json:"process,omitempty"`
}

func CollectServiceMetrics() []ServiceMetrics {
	var services []ServiceMetrics

	// Common services to check
	serviceNames := []string{
		"nginx", "apache2", "httpd",
		"mysql", "mysqld", "mariadb",
		"postgresql", "postgres",
		"redis", "redis-server",
		"mongodb", "mongod",
		"docker", "dockerd",
	}

	for _, name := range serviceNames {
		if status := checkProcess(name); status != nil {
			services = append(services, *status)
		}
	}

	return services
}

func checkProcess(name string) *ServiceMetrics {
	// Try pgrep first (most reliable)
	cmd := exec.Command("pgrep", "-x", name)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		pidStr := strings.TrimSpace(string(output))
		pids := strings.Split(pidStr, "\n")
		pid, _ := strconv.Atoi(pids[0])
		return &ServiceMetrics{
			Name:   name,
			Status: "running",
			PID:    pid,
		}
	}

	// Fallback: check if process name exists in /proc
	entries, err := os.ReadDir("/proc")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Check if directory name is a number (PID)
			pid, err := strconv.Atoi(entry.Name())
			if err != nil {
				continue
			}

			// Read process name from /proc/[pid]/comm
			commPath := "/proc/" + entry.Name() + "/comm"
			commData, err := os.ReadFile(commPath)
			if err != nil {
				continue
			}

			processName := strings.TrimSpace(string(commData))
			if processName == name {
				return &ServiceMetrics{
					Name:   name,
					Status: "running",
					PID:    pid,
				}
			}
		}
	}

	return nil
}

func CollectPortMetrics() []PortMetrics {
	var ports []PortMetrics

	// Common ports to check
	commonPorts := []int{
		22,   // SSH
		80,   // HTTP
		443,  // HTTPS
		3000, // Common app port
		3306, // MySQL
		5432, // PostgreSQL
		6379, // Redis
		8080, // Alt HTTP
		9100, // This agent
	}

	for _, port := range commonPorts {
		if status := checkPort(port); status != nil {
			ports = append(ports, *status)
		}
	}

	return ports
}

func checkPort(port int) *PortMetrics {
	// Read /proc/net/tcp and /proc/net/tcp6
	listening := false

	// Check IPv4
	if checkPortInFile("/proc/net/tcp", port) {
		listening = true
	}

	// Check IPv6
	if !listening && checkPortInFile("/proc/net/tcp6", port) {
		listening = true
	}

	if listening {
		return &PortMetrics{
			Port:   port,
			Status: "listening",
		}
	}

	return nil
}

func checkPortInFile(filename string, port int) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}

	// Convert port to hex
	portHex := strconv.FormatInt(int64(port), 16)
	portHex = strings.ToUpper(portHex)
	if len(portHex) < 4 {
		portHex = strings.Repeat("0", 4-len(portHex)) + portHex
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Format: "local_address:port"
		localAddr := fields[1]
		parts := strings.Split(localAddr, ":")
		if len(parts) != 2 {
			continue
		}

		if parts[1] == portHex {
			// Check if it's in LISTEN state (0A in hex)
			state := fields[3]
			if state == "0A" {
				return true
			}
		}
	}

	return false
}
