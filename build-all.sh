#!/bin/bash

# Master Build Script - Builds both Agent and Dashboard for Linux (Ubuntu)
# This script builds all components of the monitoring system

set -e  # Exit on error

echo "╔════════════════════════════════════════════════════════════╗"
echo "║   Monitoring System - Build All Components (Linux)        ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Get version from git or use default
VERSION=$(git describe --tags --always 2>/dev/null || echo "v1.0.0")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo "📦 Build Information:"
echo "   Version: ${VERSION}"
echo "   Build Time: ${BUILD_TIME}"
echo "   Target: Linux AMD64 (Ubuntu)"
echo ""

# Create main dist directory
mkdir -p dist

# ============================================================
# Build 1: Monitoring Agent
# ============================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  Building Monitoring Agent..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

cd agent
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" \
    -o ../dist/monitoring-agent-linux-amd64 \
    .
chmod +x ../dist/monitoring-agent-linux-amd64
cd ..

AGENT_SIZE=$(ls -lh dist/monitoring-agent-linux-amd64 | awk '{print $5}')
echo "✅ Agent built: dist/monitoring-agent-linux-amd64 (${AGENT_SIZE})"
echo ""

# ============================================================
# Build 2: Monitoring Dashboard
# ============================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  Building Monitoring Dashboard..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" \
    -o dist/monitoring-dashboard-linux-amd64 \
    .
chmod +x dist/monitoring-dashboard-linux-amd64

DASHBOARD_SIZE=$(ls -lh dist/monitoring-dashboard-linux-amd64 | awk '{print $5}')
echo "✅ Dashboard built: dist/monitoring-dashboard-linux-amd64 (${DASHBOARD_SIZE})"
echo ""

# ============================================================
# Create deployment package
# ============================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  Creating deployment package..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Copy templates for dashboard
cp -r templates dist/

# Create deployment instructions
cat > dist/DEPLOY.md << 'DEPLOY_EOF'
# Deployment Instructions for Ubuntu Server

## 1. Deploy Monitoring Agent (on servers to monitor)

```bash
# Upload agent
scp monitoring-agent-linux-amd64 user@server:/tmp/

# Install agent
ssh user@server << 'SSH_EOF'
sudo mv /tmp/monitoring-agent-linux-amd64 /usr/local/bin/monitoring-agent
sudo chmod +x /usr/local/bin/monitoring-agent
monitoring-agent -port 9100 -token YOUR_SECRET_TOKEN
SSH_EOF
```

## 2. Deploy Monitoring Dashboard (on central server)

```bash
# Create directory
ssh user@server 'sudo mkdir -p /opt/monitoring && sudo chown $USER /opt/monitoring'

# Upload files
scp monitoring-dashboard-linux-amd64 user@server:/opt/monitoring/monitoring
scp -r templates user@server:/opt/monitoring/

# Run dashboard
ssh user@server 'cd /opt/monitoring && ./monitoring'
```

## 3. Access

- Dashboard: http://server-ip:8080
- Agent API: http://server-ip:9100/metrics

DEPLOY_EOF

echo "✅ Deployment package created in dist/"
echo ""

# ============================================================
# Summary
# ============================================================
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                    BUILD COMPLETE! ✅                      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "📦 Build artifacts in dist/:"
ls -lh dist/ | grep -v total | awk '{print "   - " $9 " (" $5 ")"}'
echo ""
echo "📖 Deployment instructions: dist/DEPLOY.md"
echo ""
echo "🚀 Quick Deploy Commands:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Agent:     scp dist/monitoring-agent-linux-amd64 user@server:/usr/local/bin/monitoring-agent"
echo "Dashboard: scp dist/monitoring-dashboard-linux-amd64 user@server:/opt/monitoring/monitoring"
echo ""
