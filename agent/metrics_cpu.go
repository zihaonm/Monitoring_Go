package main

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
	LoadAverage  string  `json:"load_average,omitempty"`
}

var lastCPUStats cpuStats

type cpuStats struct {
	user   uint64
	system uint64
	idle   uint64
	total  uint64
}

func CollectCPUMetrics() CPUMetrics {
	metrics := CPUMetrics{
		Cores:        runtime.NumCPU(),
		UsagePercent: getCPUUsage(),
	}

	// Try to get load average (Linux/Unix only)
	if loadavg := getLoadAverage(); loadavg != "" {
		metrics.LoadAverage = loadavg
	}

	return metrics
}

func getCPUUsage() float64 {
	stats := readCPUStats()
	if stats == nil {
		return 0.0
	}

	if lastCPUStats.total == 0 {
		lastCPUStats = *stats
		time.Sleep(100 * time.Millisecond)
		stats = readCPUStats()
		if stats == nil {
			return 0.0
		}
	}

	totalDelta := stats.total - lastCPUStats.total
	idleDelta := stats.idle - lastCPUStats.idle

	lastCPUStats = *stats

	if totalDelta == 0 {
		return 0.0
	}

	usage := 100.0 * (float64(totalDelta-idleDelta) / float64(totalDelta))
	return roundFloat(usage, 2)
}

func readCPUStats() *cpuStats {
	// Read /proc/stat (Linux)
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		// Fallback for non-Linux systems
		return &cpuStats{total: 1, idle: 0}
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return nil
	}

	// First line contains aggregate CPU stats
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return nil
	}

	user, _ := strconv.ParseUint(fields[1], 10, 64)
	nice, _ := strconv.ParseUint(fields[2], 10, 64)
	system, _ := strconv.ParseUint(fields[3], 10, 64)
	idle, _ := strconv.ParseUint(fields[4], 10, 64)
	iowait, _ := strconv.ParseUint(fields[5], 10, 64)

	total := user + nice + system + idle + iowait
	if len(fields) > 6 {
		irq, _ := strconv.ParseUint(fields[6], 10, 64)
		total += irq
	}
	if len(fields) > 7 {
		softirq, _ := strconv.ParseUint(fields[7], 10, 64)
		total += softirq
	}

	return &cpuStats{
		user:   user,
		system: system,
		idle:   idle,
		total:  total,
	}
}

func getLoadAverage() string {
	// Read /proc/loadavg (Linux)
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		return fields[0] + " " + fields[1] + " " + fields[2]
	}

	return ""
}
