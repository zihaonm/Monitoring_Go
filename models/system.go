package models

// SystemInfo represents system information and metrics
type SystemInfo struct {
	Hostname     string       `json:"hostname"`
	Platform     string       `json:"platform"`
	OS           string       `json:"os"`
	Uptime       uint64       `json:"uptime"` // in seconds
	CPUInfo      CPUInfo      `json:"cpu"`
	MemoryInfo   MemoryInfo   `json:"memory"`
	DiskInfo     []DiskInfo   `json:"disks"`
}

// CPUInfo represents CPU usage information
type CPUInfo struct {
	Cores       int     `json:"cores"`
	UsagePercent float64 `json:"usage_percent"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Total       uint64  `json:"total"`        // in bytes
	Used        uint64  `json:"used"`         // in bytes
	Available   uint64  `json:"available"`    // in bytes
	UsedPercent float64 `json:"used_percent"`
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Total       uint64  `json:"total"`        // in bytes
	Used        uint64  `json:"used"`         // in bytes
	Free        uint64  `json:"free"`         // in bytes
	UsedPercent float64 `json:"used_percent"`
}
