package services

import (
	"monitoring/models"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemService handles system information and metrics
type SystemService struct {
	alertConfig *models.SystemAlertConfig
	alertState  *models.AlertState
	telegram    *TelegramService
}

// NewSystemService creates a new system service
func NewSystemService(telegram *TelegramService) *SystemService {
	return &SystemService{
		alertConfig: &models.SystemAlertConfig{
			DiskSpaceThreshold: 80.0, // Default: alert at 80% disk usage
			CPUThreshold:       90.0, // Default: alert at 90% CPU
			MemoryThreshold:    90.0, // Default: alert at 90% memory
			Enabled:            true,
		},
		alertState: &models.AlertState{},
		telegram:   telegram,
	}
}

// SetAlertConfig updates the alert configuration
func (s *SystemService) SetAlertConfig(config *models.SystemAlertConfig) {
	s.alertConfig = config
}

// GetAlertConfig returns the current alert configuration
func (s *SystemService) GetAlertConfig() *models.SystemAlertConfig {
	return s.alertConfig
}

// LoadAlertConfig loads alert configuration (for persistence)
func (s *SystemService) LoadAlertConfig(config *models.SystemAlertConfig) {
	if config != nil {
		s.alertConfig = config
	}
}

// GetSystemInfo retrieves current system information and metrics
func (s *SystemService) GetSystemInfo() (*models.SystemInfo, error) {
	info := &models.SystemInfo{}

	// Get host info
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}
	info.Hostname = hostInfo.Hostname
	info.Platform = hostInfo.Platform
	info.OS = hostInfo.OS
	info.Uptime = hostInfo.Uptime

	// Get CPU info
	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		info.CPUInfo.UsagePercent = cpuPercent[0]
	}
	info.CPUInfo.Cores = runtime.NumCPU()

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.MemoryInfo.Total = memInfo.Total
		info.MemoryInfo.Used = memInfo.Used
		info.MemoryInfo.Available = memInfo.Available
		info.MemoryInfo.UsedPercent = memInfo.UsedPercent
	}

	// Get disk info
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err == nil {
				info.DiskInfo = append(info.DiskInfo, models.DiskInfo{
					Device:      partition.Device,
					Mountpoint:  partition.Mountpoint,
					Total:       usage.Total,
					Used:        usage.Used,
					Free:        usage.Free,
					UsedPercent: usage.UsedPercent,
				})
			}
		}
	}

	// Check for alerts if enabled
	if s.alertConfig.Enabled && s.telegram.IsEnabled() {
		s.checkResourceAlerts(info)
	}

	return info, nil
}

// checkResourceAlerts checks system resources and sends alerts if thresholds are exceeded
func (s *SystemService) checkResourceAlerts(info *models.SystemInfo) {
	// Check disk space
	for _, diskInfo := range info.DiskInfo {
		// Only alert on main filesystems (skip special mounts)
		if diskInfo.Mountpoint == "/" || diskInfo.Mountpoint == "/home" {
			if diskInfo.UsedPercent > s.alertConfig.DiskSpaceThreshold && !s.alertState.DiskAlertSent {
				s.sendDiskAlert(diskInfo)
				s.alertState.DiskAlertSent = true
			} else if diskInfo.UsedPercent <= s.alertConfig.DiskSpaceThreshold && s.alertState.DiskAlertSent {
				// Reset alert state when disk usage drops below threshold
				s.alertState.DiskAlertSent = false
			}
		}
	}

	// Check CPU
	if info.CPUInfo.UsagePercent > s.alertConfig.CPUThreshold && !s.alertState.CPUAlertSent {
		s.sendCPUAlert(info.CPUInfo.UsagePercent)
		s.alertState.CPUAlertSent = true
	} else if info.CPUInfo.UsagePercent <= s.alertConfig.CPUThreshold && s.alertState.CPUAlertSent {
		s.alertState.CPUAlertSent = false
	}

	// Check Memory
	if info.MemoryInfo.UsedPercent > s.alertConfig.MemoryThreshold && !s.alertState.MemoryAlertSent {
		s.sendMemoryAlert(info.MemoryInfo.UsedPercent)
		s.alertState.MemoryAlertSent = true
	} else if info.MemoryInfo.UsedPercent <= s.alertConfig.MemoryThreshold && s.alertState.MemoryAlertSent {
		s.alertState.MemoryAlertSent = false
	}
}

// sendDiskAlert sends a disk space alert via Telegram
func (s *SystemService) sendDiskAlert(diskInfo models.DiskInfo) {
	go s.telegram.SendSystemAlert("disk", diskInfo.Mountpoint, diskInfo.UsedPercent, s.alertConfig.DiskSpaceThreshold)
}

// sendCPUAlert sends a CPU usage alert via Telegram
func (s *SystemService) sendCPUAlert(usage float64) {
	go s.telegram.SendSystemAlert("cpu", "", usage, s.alertConfig.CPUThreshold)
}

// sendMemoryAlert sends a memory usage alert via Telegram
func (s *SystemService) sendMemoryAlert(usage float64) {
	go s.telegram.SendSystemAlert("memory", "", usage, s.alertConfig.MemoryThreshold)
}
