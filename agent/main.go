package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	port      = flag.Int("port", 9100, "Port to listen on")
	authToken = flag.String("token", "", "Authentication token (optional)")
	interval  = flag.Int("interval", 10, "Metrics collection interval in seconds")
)

// MetricsResponse represents the JSON response structure
type MetricsResponse struct {
	Hostname  string           `json:"hostname"`
	Timestamp string           `json:"timestamp"`
	Status    string           `json:"status"`
	CPU       CPUMetrics       `json:"cpu"`
	Memory    MemoryMetrics    `json:"memory"`
	Disk      []DiskMetrics    `json:"disk"`
	Network   NetworkMetrics   `json:"network"`
	Services  []ServiceMetrics `json:"services,omitempty"`
	Ports     []PortMetrics    `json:"ports,omitempty"`
}

// Global metrics cache
var cachedMetrics *MetricsResponse
var lastUpdate time.Time

func main() {
	flag.Parse()

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	log.Printf("Starting Monitoring Agent on port %d", *port)
	log.Printf("Hostname: %s", hostname)
	if *authToken != "" {
		log.Printf("Authentication enabled")
	}

	// Start metrics collector
	go metricsCollector(hostname)

	// HTTP handlers
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", rootHandler)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Agent listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func metricsCollector(hostname string) {
	for {
		metrics := collectAllMetrics(hostname)
		cachedMetrics = &metrics
		lastUpdate = time.Now()
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

func collectAllMetrics(hostname string) MetricsResponse {
	return MetricsResponse{
		Hostname:  hostname,
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    "healthy",
		CPU:       CollectCPUMetrics(),
		Memory:    CollectMemoryMetrics(),
		Disk:      CollectDiskMetrics(),
		Network:   CollectNetworkMetrics(),
		Services:  CollectServiceMetrics(),
		Ports:     CollectPortMetrics(),
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication if token is set
	if *authToken != "" {
		token := r.Header.Get("Authorization")
		if token != "Bearer "+*authToken && r.URL.Query().Get("token") != *authToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if cachedMetrics == nil {
		// First request before cache is ready
		hostname, _ := os.Hostname()
		metrics := collectAllMetrics(hostname)
		json.NewEncoder(w).Encode(metrics)
		return
	}

	json.NewEncoder(w).Encode(cachedMetrics)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"last_update": lastUpdate.Format(time.RFC3339),
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Monitoring Agent</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #3498db; }
        .info { background: #ecf0f1; padding: 15px; border-radius: 5px; margin: 10px 0; }
        code { background: #2c3e50; color: #ecf0f1; padding: 2px 6px; border-radius: 3px; }
        a { color: #3498db; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>📊 Monitoring Agent</h1>
    <div class="info">
        <strong>Hostname:</strong> %s<br>
        <strong>Version:</strong> 1.0.0<br>
        <strong>Status:</strong> Running<br>
        <strong>Last Update:</strong> %s
    </div>
    <h2>Endpoints</h2>
    <ul>
        <li><a href="/metrics">/metrics</a> - Get all system metrics (JSON)</li>
        <li><a href="/health">/health</a> - Health check endpoint</li>
    </ul>
    <h2>Usage</h2>
    <p>To fetch metrics:</p>
    <code>curl http://localhost:%d/metrics</code>
</body>
</html>`, hostname, lastUpdate.Format(time.RFC3339), *port)
}
