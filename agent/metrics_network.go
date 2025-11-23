package main

import (
	"os"
	"strconv"
	"strings"
)

type NetworkMetrics struct {
	RxBytes uint64                `json:"rx_bytes"`
	TxBytes uint64                `json:"tx_bytes"`
	RxMB    float64               `json:"rx_mb"`
	TxMB    float64               `json:"tx_mb"`
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
}

type NetworkInterface struct {
	Name    string  `json:"name"`
	RxBytes uint64  `json:"rx_bytes"`
	TxBytes uint64  `json:"tx_bytes"`
	RxMB    float64 `json:"rx_mb"`
	TxMB    float64 `json:"tx_mb"`
}

func CollectNetworkMetrics() NetworkMetrics {
	// Read /proc/net/dev (Linux)
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return NetworkMetrics{}
	}

	var totalRx, totalTx uint64
	var interfaces []NetworkInterface

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		// Skip header lines
		if i < 2 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		ifaceName := strings.TrimSuffix(fields[0], ":")

		// Skip loopback
		if ifaceName == "lo" {
			continue
		}

		rxBytes, _ := strconv.ParseUint(fields[1], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[9], 10, 64)

		totalRx += rxBytes
		totalTx += txBytes

		// Skip virtual interfaces with no traffic
		if rxBytes == 0 && txBytes == 0 {
			continue
		}

		interfaces = append(interfaces, NetworkInterface{
			Name:    ifaceName,
			RxBytes: rxBytes,
			TxBytes: txBytes,
			RxMB:    roundFloat(float64(rxBytes)/1024/1024, 2),
			TxMB:    roundFloat(float64(txBytes)/1024/1024, 2),
		})
	}

	return NetworkMetrics{
		RxBytes:    totalRx,
		TxBytes:    totalTx,
		RxMB:       roundFloat(float64(totalRx)/1024/1024, 2),
		TxMB:       roundFloat(float64(totalTx)/1024/1024, 2),
		Interfaces: interfaces,
	}
}
