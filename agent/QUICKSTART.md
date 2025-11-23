# Quick Start Guide

## 1. Deploy Agent to Remote Server

### Option A: Download and Run (Easiest)

```bash
# On remote server:
# 1. Copy the binary to the server
scp monitoring-agent-linux-amd64 user@server:/tmp/

# 2. SSH into the server
ssh user@server

# 3. Test the agent
chmod +x /tmp/monitoring-agent-linux-amd64
/tmp/monitoring-agent-linux-amd64 -port 9100

# 4. Test it works
curl http://localhost:9100/metrics
```

### Option B: Install as Service (Recommended)

```bash
# On remote server:
# 1. Copy files to server
scp monitoring-agent-linux-amd64 install.sh user@server:/tmp/

# 2. SSH into server
ssh user@server

# 3. Install
cd /tmp
chmod +x monitoring-agent-linux-amd64 install.sh
sudo mv monitoring-agent-linux-amd64 monitoring-agent
sudo ./install.sh

# Agent is now running!
curl http://localhost:9100/metrics
```

### Option C: One-Line Install

```bash
# On remote server (if you have the binary URL):
curl -O https://your-domain.com/monitoring-agent && \
chmod +x monitoring-agent && \
sudo mv monitoring-agent /usr/local/bin/ && \
sudo monitoring-agent -port 9100 &
```

## 2. Add to Monitoring Dashboard

### In Your Monitoring Dashboard:

1. Click **"➕ Add New Service"**

2. Fill in the form:
   - **Check Type:** HTTP/HTTPS
   - **Service Name:** Server 1 - System Metrics
   - **URL:** `http://192.168.1.10:9100/metrics`
   - **Check Interval:** 60 seconds
   - **Timeout:** 10 seconds

3. (Optional) Add Telegram overrides for this specific server

4. Click **"Save"**

5. The dashboard will now monitor:
   - ✅ If the server is reachable
   - ✅ System metrics (CPU, memory, disk)
   - ✅ Running services
   - ✅ Open ports

## 3. Multiple Servers Setup

### Deploy to Multiple Servers

```bash
# Create a simple deploy script
cat > deploy-agent.sh <<'EOF'
#!/bin/bash
SERVERS=(
  "192.168.1.10"
  "192.168.1.11"
  "192.168.1.12"
)

for SERVER in "${SERVERS[@]}"; do
  echo "Deploying to $SERVER..."
  scp monitoring-agent-linux-amd64 install.sh root@$SERVER:/tmp/
  ssh root@$SERVER "cd /tmp && mv monitoring-agent-linux-amd64 monitoring-agent && chmod +x install.sh monitoring-agent && ./install.sh"
  echo "✅ $SERVER done"
done
EOF

chmod +x deploy-agent.sh
./deploy-agent.sh
```

### Add All Servers to Dashboard

For each server, add a new service:

| Server Name | URL |
|-------------|-----|
| Web Server 1 | http://192.168.1.10:9100/metrics |
| Web Server 2 | http://192.168.1.11:9100/metrics |
| Database Server | http://192.168.1.12:9100/metrics |

## 4. Testing

### Test Agent Locally

```bash
# Start agent
./monitoring-agent -port 9100

# In another terminal, test endpoints:
curl http://localhost:9100/
curl http://localhost:9100/health
curl http://localhost:9100/metrics | jq
```

### Test from Monitoring Dashboard

```bash
# From your dashboard server:
curl http://192.168.1.10:9100/metrics
```

## 5. With Authentication

### Start Agent with Token

```bash
./monitoring-agent -port 9100 -token "my-secret-token"
```

### Test with Token

```bash
# Using header
curl -H "Authorization: Bearer my-secret-token" http://localhost:9100/metrics

# Using query parameter
curl http://localhost:9100/metrics?token=my-secret-token
```

### Add to Dashboard with Auth

When adding the service in your dashboard, you would need to add custom headers support (or include token in URL for now):

- **URL:** `http://192.168.1.10:9100/metrics?token=my-secret-token`

## 6. Monitoring Specific Metrics

The JSON response includes all these metrics:

```json
{
  "cpu": {
    "usage_percent": 45.2,  // Alert if > 80
    "cores": 4
  },
  "memory": {
    "usage_percent": 55.2,  // Alert if > 90
    "used_mb": 4521,
    "total_mb": 8192
  },
  "disk": [
    {
      "mount": "/",
      "usage_percent": 65.5,  // Alert if > 85
      "used_gb": 65.5,
      "total_gb": 100.0
    }
  ],
  "services": [
    {
      "name": "nginx",
      "status": "running",  // Alert if not "running"
      "pid": 1234
    }
  ]
}
```

You can parse these in your dashboard to create custom alerts.

## 7. Firewall Configuration

### Allow Port 9100

```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 9100/tcp

# CentOS/RHEL (firewalld)
sudo firewall-cmd --add-port=9100/tcp --permanent
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p tcp --dport 9100 -j ACCEPT
sudo service iptables save
```

### Restrict to Monitoring Server Only

```bash
# Only allow from your monitoring dashboard IP (e.g., 192.168.1.5)
sudo ufw allow from 192.168.1.5 to any port 9100
```

## 8. Troubleshooting

### Agent Not Starting

```bash
# Check if port is already in use
netstat -tuln | grep 9100

# Try different port
./monitoring-agent -port 9200
```

### Can't Connect from Dashboard

```bash
# Test connectivity
ping server-ip

# Test port
telnet server-ip 9100

# Check firewall
sudo ufw status
```

### High Resource Usage

```bash
# Increase collection interval (default: 10s)
./monitoring-agent -interval 30  # Collect every 30 seconds
```

## 9. Production Checklist

- [ ] Built binary for correct platform (linux-amd64, linux-arm64, etc.)
- [ ] Deployed to all servers
- [ ] Installed as systemd service (so it auto-starts on reboot)
- [ ] Configured authentication token
- [ ] Firewall rules configured (only allow from monitoring server)
- [ ] Added all servers to monitoring dashboard
- [ ] Tested all endpoints respond
- [ ] Set up alerts for CPU, memory, disk thresholds
- [ ] Verified services are being detected
- [ ] Verified ports are being detected

## 10. Next Steps

### Enhance Your Monitoring Dashboard

Consider adding features to parse the agent's JSON response:

1. **Metric Extraction** - Parse specific values (cpu.usage_percent, memory.usage_percent)
2. **Threshold Alerts** - Alert when metrics exceed thresholds
3. **Graphs** - Visualize trends over time
4. **Service Checks** - Alert if critical services stop running
5. **Port Monitoring** - Alert if expected ports are not listening

### Example Alert Rules

- 🔥 **CPU > 80%** → Send Telegram alert
- ⚠️ **Memory > 90%** → Send Telegram alert
- 💾 **Disk > 85%** → Send Telegram alert
- 🔴 **nginx not running** → Send Telegram alert
- 🔌 **Port 80 not listening** → Send Telegram alert

---

**That's it!** You now have lightweight agents running on all your servers, reporting system metrics to your central monitoring dashboard. 🎉
