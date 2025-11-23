package main

import (
	"os"
	"strings"
	"syscall"
)

type DiskMetrics struct {
	Mount        string  `json:"mount"`
	TotalGB      float64 `json:"total_gb"`
	UsedGB       float64 `json:"used_gb"`
	AvailableGB  float64 `json:"available_gb"`
	UsagePercent float64 `json:"usage_percent"`
	Filesystem   string  `json:"filesystem,omitempty"`
}

func CollectDiskMetrics() []DiskMetrics {
	var disks []DiskMetrics

	// Read /proc/mounts to get mounted filesystems (Linux)
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		// Fallback: just check root
		if metrics := getDiskUsage("/"); metrics != nil {
			disks = append(disks, *metrics)
		}
		return disks
	}

	lines := strings.Split(string(data), "\n")
	seen := make(map[string]bool)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		// Skip non-disk filesystems
		if strings.HasPrefix(device, "/dev/loop") ||
			strings.HasPrefix(device, "tmpfs") ||
			strings.HasPrefix(device, "devtmpfs") ||
			strings.HasPrefix(device, "sysfs") ||
			strings.HasPrefix(device, "proc") ||
			strings.HasPrefix(device, "cgroup") ||
			strings.HasPrefix(device, "securityfs") ||
			fsType == "squashfs" ||
			fsType == "overlay" {
			continue
		}

		// Skip if we've already seen this mount point
		if seen[mountPoint] {
			continue
		}
		seen[mountPoint] = true

		metrics := getDiskUsage(mountPoint)
		if metrics != nil {
			metrics.Filesystem = device
			disks = append(disks, *metrics)
		}
	}

	// If no disks found, add root as fallback
	if len(disks) == 0 {
		if metrics := getDiskUsage("/"); metrics != nil {
			disks = append(disks, *metrics)
		}
	}

	return disks
}

func getDiskUsage(path string) *DiskMetrics {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil
	}

	// Calculate sizes
	total := float64(stat.Blocks * uint64(stat.Bsize))
	available := float64(stat.Bavail * uint64(stat.Bsize))
	used := total - float64(stat.Bfree*uint64(stat.Bsize))

	usagePercent := 0.0
	if total > 0 {
		usagePercent = 100.0 * used / total
	}

	return &DiskMetrics{
		Mount:        path,
		TotalGB:      roundFloat(total/1024/1024/1024, 2),
		UsedGB:       roundFloat(used/1024/1024/1024, 2),
		AvailableGB:  roundFloat(available/1024/1024/1024, 2),
		UsagePercent: roundFloat(usagePercent, 2),
	}
}
