package main

import (
	"os"
	"strconv"
	"strings"
)

type MemoryMetrics struct {
	TotalMB       uint64  `json:"total_mb"`
	UsedMB        uint64  `json:"used_mb"`
	AvailableMB   uint64  `json:"available_mb"`
	UsagePercent  float64 `json:"usage_percent"`
	SwapTotalMB   uint64  `json:"swap_total_mb,omitempty"`
	SwapUsedMB    uint64  `json:"swap_used_mb,omitempty"`
	SwapPercent   float64 `json:"swap_percent,omitempty"`
}

func CollectMemoryMetrics() MemoryMetrics {
	// Read /proc/meminfo (Linux)
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return MemoryMetrics{}
	}

	memInfo := parseMemInfo(string(data))

	total := memInfo["MemTotal"]
	available := memInfo["MemAvailable"]
	if available == 0 {
		// Fallback calculation for older kernels
		free := memInfo["MemFree"]
		buffers := memInfo["Buffers"]
		cached := memInfo["Cached"]
		available = free + buffers + cached
	}

	used := total - available
	usagePercent := 0.0
	if total > 0 {
		usagePercent = 100.0 * float64(used) / float64(total)
	}

	swapTotal := memInfo["SwapTotal"]
	swapFree := memInfo["SwapFree"]
	swapUsed := swapTotal - swapFree
	swapPercent := 0.0
	if swapTotal > 0 {
		swapPercent = 100.0 * float64(swapUsed) / float64(swapTotal)
	}

	return MemoryMetrics{
		TotalMB:      total / 1024,
		UsedMB:       used / 1024,
		AvailableMB:  available / 1024,
		UsagePercent: roundFloat(usagePercent, 2),
		SwapTotalMB:  swapTotal / 1024,
		SwapUsedMB:   swapUsed / 1024,
		SwapPercent:  roundFloat(swapPercent, 2),
	}
}

func parseMemInfo(data string) map[string]uint64 {
	result := make(map[string]uint64)
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		result[key] = value
	}

	return result
}
