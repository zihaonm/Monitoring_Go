# Complete Deployment Guide: Multi-Server Monitoring

This guide shows you how to monitor multiple servers without installing anything complex on each server.

## Architecture Overview

```
┌─────────────────────────────────────┐
│   Your Monitoring Dashboard         │
│   (Main Server)                     │
│   - Web Interface                   │
│   - Service Monitoring              │
│   - Telegram Alerts                 │
└──────────┬──────────────────────────┘
           │
           │ HTTP Requests (every 60s)
           │
     ┌─────┴──────┬───────────┬─────────────┐
     │            │           │             │
     ▼            ▼           ▼             ▼
┌─────────┐  ┌─────────┐ ┌─────────┐  ┌─────────┐
│Server 1 │  │Server 2 │ │Server 3 │  │Server N │
│         │  │         │ │         │  │         │
│ Agent   │  │ Agent   │ │ Agent   │  │ Agent   │
│ :9100   │  │ :9100   │ │ :9100   │  │ :9100   │
└─────────┘  └─────────┘ └─────────┘  └─────────┘
```

## What Gets Monitored

### System Metrics (Automatic)
- ✅ CPU usage percentage
- ✅ Memory usage (total, used, available)
- ✅ Disk usage per partition
- ✅ Network traffic (RX/TX bytes)
- ✅ Load average (Linux)

### Services (Automatic Detection)
- ✅ nginx / apache / httpd
- ✅ mysql / mariadb / postgresql
- ✅ redis / mongodb
- ✅ docker
- ✅ Any custom process

### Ports (Automatic Detection)
- ✅ SSH (22)
- ✅ HTTP/HTTPS (80, 443)
- ✅ MySQL (3306)
- ✅ PostgreSQL (5432)
- ✅ Redis (6379)
- ✅ Custom ports

## Step-by-Step Deployment

### Step 1: Build the Agent

On your development machine (or main monitoring server):

```bash
cd agent/

# Build for Linux servers (most common)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o monitoring-agent-linux-amd64

# Or build for all platforms
chmod +x build.sh
./build.sh
```

This creates a single binary file (~5-8MB).

### Step 2: Deploy Agent to Servers

**Option A: Quick Deploy (Single Server)**

```bash
# Copy agent and install script
scp monitoring-agent-linux-amd64 install.sh user@192.168.1.10:/tmp/

# SSH and install
ssh user@192.168.1.10
cd /tmp
sudo mv monitoring-agent-linux-amd64 monitoring-agent
sudo chmod +x install.sh monitoring-agent
sudo ./install.sh
```

**Option B: Bulk Deploy (Multiple Servers)**

Create a deployment script:

```bash
#!/bin/bash
# deploy-all.sh

SERVERS=(
  "user@192.168.1.10"
  "user@192.168.1.11"
  "user@192.168.1.12"
)

for SERVER in "${SERVERS[@]}"; do
  echo "==== Deploying to $SERVER ===="

  # Copy files
  scp monitoring-agent-linux-amd64 install.sh $SERVER:/tmp/

  # Install
  ssh $SERVER "cd /tmp && sudo mv monitoring-agent-linux-amd64 monitoring-agent && sudo chmod +x install.sh monitoring-agent && sudo ./install.sh"

  # Verify
  ssh $SERVER "systemctl status monitoring-agent --no-pager"

  echo "✅ $SERVER completed"
  echo ""
done

echo "All servers deployed!"
```

Run it:
```bash
chmod +x deploy-all.sh
./deploy-all.sh
```

### Step 3: Verify Agents are Running

Test each server:

```bash
# Check if agent responds
curl http://192.168.1.10:9100/health
curl http://192.168.1.11:9100/health
curl http://192.168.1.12:9100/health

# Get full metrics
curl http://192.168.1.10:9100/metrics | jq
```

### Step 4: Add Servers to Monitoring Dashboard

For each server, add a monitoring service:

1. **Open your monitoring dashboard** (http://your-main-server:8080)

2. **Click "➕ Add New Service"**

3. **Fill in the form:**
   ```
   Check Type: HTTP/HTTPS
   Service Name: Web Server 01 - System Metrics
   URL: http://192.168.1.10:9100/metrics
   Check Interval: 60 seconds
   Timeout: 10 seconds
   ```

4. **(Optional) Add Telegram overrides** for specific servers

5. **Click "Save"**

6. **Repeat for all servers**

### Step 5: Configure Alerts

The dashboard will automatically detect when:

- Server becomes unreachable
- Response time is slow
- Service goes down

You'll receive Telegram alerts based on your configuration.

## Advanced Configuration

### Use Authentication

**On remote servers:**
```bash
# Install with auth token
sudo AGENT_PORT=9100 AUTH_TOKEN="your-secret-token" ./install.sh
```

**In dashboard:**
When adding service, use:
```
URL: http://192.168.1.10:9100/metrics?token=your-secret-token
```

### Custom Port

If port 9100 is already in use:

```bash
# Install on different port
sudo AGENT_PORT=9200 ./install.sh
```

### Firewall Configuration

**Restrict access to monitoring server only:**

```bash
# On remote servers
sudo ufw allow from 192.168.1.5 to any port 9100  # Replace with your monitoring server IP
```

### Monitor Specific Metrics

To monitor specific values like CPU > 80%, you'll need to enhance your dashboard to parse JSON. Here's the structure:

```json
{
  "cpu": {"usage_percent": 45.2},      // Alert if > 80
  "memory": {"usage_percent": 55.2},    // Alert if > 90
  "disk": [
    {
      "mount": "/",
      "usage_percent": 65.5           // Alert if > 85
    }
  ],
  "services": [
    {
      "name": "nginx",
      "status": "running"             // Alert if != "running"
    }
  ]
}
```

## Monitoring Scenarios

### Scenario 1: Web Application Stack

**Servers:**
- 2x Web servers (nginx + app)
- 1x Database server (MySQL)
- 1x Redis cache

**Setup:**
```bash
# Deploy agent to all 4 servers
./deploy-all.sh

# Add to dashboard:
# - Web Server 1: http://web1:9100/metrics
# - Web Server 2: http://web2:9100/metrics
# - Database: http://db:9100/metrics
# - Redis: http://redis:9100/metrics
```

**What you'll monitor:**
- Web servers: nginx running, CPU < 80%, disk space
- Database: mysql running, memory usage, disk I/O
- Redis: redis-server running, memory usage

### Scenario 2: Microservices

**Servers:**
- 5x Application servers
- 2x Load balancers
- 1x Database

**Setup:**
Deploy agent to all 8 servers and add them individually to the dashboard. Each service can have unique Telegram alert settings.

### Scenario 3: Mixed Environment

**Servers:**
- Cloud servers (AWS, DigitalOcean)
- On-premise servers
- Raspberry Pi devices

**Setup:**
Build for each architecture:
```bash
# Cloud (usually AMD64)
GOOS=linux GOARCH=amd64 go build -o agent-linux-amd64

# Raspberry Pi (ARM64)
GOOS=linux GOARCH=arm64 go build -o agent-linux-arm64
```

Deploy appropriate binary to each server.

## Troubleshooting

### Agent won't start

```bash
# Check logs
journalctl -u monitoring-agent -n 50

# Check if port is available
netstat -tuln | grep 9100

# Try manual start to see errors
/usr/local/bin/monitoring-agent -port 9100
```

### Can't connect from dashboard

```bash
# Test from dashboard server
curl http://remote-server:9100/health

# Check firewall on remote server
sudo ufw status

# Allow your monitoring server
sudo ufw allow from YOUR_DASHBOARD_IP to any port 9100
```

### High resource usage

```bash
# Reduce collection frequency (default: 10s)
# Edit /etc/systemd/system/monitoring-agent.service
ExecStart=/usr/local/bin/monitoring-agent -port 9100 -interval 30

sudo systemctl daemon-reload
sudo systemctl restart monitoring-agent
```

### Metrics not updating

```bash
# Check last update time
curl http://server:9100/health

# Restart agent
sudo systemctl restart monitoring-agent

# Check if metrics are being collected
curl http://server:9100/metrics | jq '.timestamp'
```

## Maintenance

### Update Agent

```bash
# Build new version
go build -o monitoring-agent-linux-amd64

# Deploy to servers
for SERVER in server1 server2 server3; do
  scp monitoring-agent-linux-amd64 $SERVER:/tmp/
  ssh $SERVER "sudo systemctl stop monitoring-agent && \
               sudo mv /tmp/monitoring-agent-linux-amd64 /usr/local/bin/monitoring-agent && \
               sudo chmod +x /usr/local/bin/monitoring-agent && \
               sudo systemctl start monitoring-agent"
done
```

### Check Agent Version

```bash
curl http://server:9100/ | grep Version
```

### Uninstall Agent

```bash
sudo systemctl stop monitoring-agent
sudo systemctl disable monitoring-agent
sudo rm /usr/local/bin/monitoring-agent
sudo rm /etc/systemd/system/monitoring-agent.service
sudo systemctl daemon-reload
```

## Security Best Practices

1. **Use Authentication**
   - Always set `-token` in production
   - Use different tokens per environment

2. **Firewall Rules**
   - Only allow monitoring server IP
   - Block external access

3. **Network Segmentation**
   - Keep agents on private network
   - Use VPN if monitoring across internet

4. **HTTPS (Optional)**
   - Put agent behind nginx reverse proxy
   - Use SSL certificates

5. **Regular Updates**
   - Keep agent updated
   - Monitor for security issues

## Performance Impact

| Metric | Value |
|--------|-------|
| CPU Usage (idle) | ~0.1% |
| CPU Usage (collecting) | ~1-2% |
| Memory Usage | ~10-20 MB |
| Disk Space | ~5-8 MB |
| Network (per request) | ~2-5 KB |

**Recommendation:** With 60-second check intervals, impact is negligible even on busy servers.

## Cost Analysis

| Solution | Per Server Cost | Complexity | Setup Time |
|----------|----------------|------------|------------|
| **This Agent** | $0 | Low | 5 min |
| Datadog | $15-31/month | Medium | 15 min |
| New Relic | $25-100/month | Medium | 20 min |
| Prometheus + Grafana | $0 | High | 2-3 hours |

**For 10 servers:** Save $150-1000/month vs commercial solutions!

## Next Steps

1. ✅ Deploy agents to all servers
2. ✅ Add all servers to monitoring dashboard
3. ✅ Configure Telegram alerts
4. ⏭️ Set up specific metric thresholds (enhance dashboard)
5. ⏭️ Create custom alerts for critical services
6. ⏭️ Set up dashboards for visualization (optional)

---

**Congratulations!** You now have a complete multi-server monitoring system running with minimal overhead and zero ongoing costs. 🎉

For questions or issues, refer to:
- `agent/README.md` - Full agent documentation
- `agent/QUICKSTART.md` - Quick deployment guide
- GitHub Issues - Report problems
