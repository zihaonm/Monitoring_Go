#!/bin/bash
# Monitoring Agent Installation Script

set -e

# Configuration
AGENT_PORT="${AGENT_PORT:-9100}"
AUTH_TOKEN="${AUTH_TOKEN:-}"
INSTALL_DIR="/usr/local/bin"
SERVICE_FILE="/etc/systemd/system/monitoring-agent.service"

echo "======================================"
echo "Monitoring Agent Installation"
echo "======================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="arm"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
echo "Detected: $OS $ARCH"
echo ""

# Check if binary exists in current directory
if [ -f "./monitoring-agent" ]; then
    echo "Using monitoring-agent binary from current directory"
    BINARY_PATH="./monitoring-agent"
elif [ -f "./monitoring-agent-$OS-$ARCH" ]; then
    echo "Using monitoring-agent-$OS-$ARCH binary from current directory"
    BINARY_PATH="./monitoring-agent-$OS-$ARCH"
else
    echo "Error: monitoring-agent binary not found"
    echo "Please build the agent first:"
    echo "  go build -o monitoring-agent"
    exit 1
fi

# Stop existing service if running
if systemctl is-active --quiet monitoring-agent; then
    echo "Stopping existing monitoring-agent service..."
    systemctl stop monitoring-agent
fi

# Copy binary
echo "Installing agent to $INSTALL_DIR..."
cp "$BINARY_PATH" "$INSTALL_DIR/monitoring-agent"
chmod +x "$INSTALL_DIR/monitoring-agent"

# Create systemd service
echo "Creating systemd service..."
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Monitoring Agent
Documentation=https://github.com/yourusername/monitoring-agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_DIR/monitoring-agent -port $AGENT_PORT ${AUTH_TOKEN:+-token $AUTH_TOKEN}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable and start service
echo "Enabling and starting monitoring-agent service..."
systemctl enable monitoring-agent
systemctl start monitoring-agent

# Wait a moment for service to start
sleep 2

# Check status
if systemctl is-active --quiet monitoring-agent; then
    echo ""
    echo "======================================"
    echo "✅ Installation successful!"
    echo "======================================"
    echo ""
    echo "Agent is running on port $AGENT_PORT"
    echo ""
    echo "Test with:"
    echo "  curl http://localhost:$AGENT_PORT/metrics"
    echo ""
    echo "View logs:"
    echo "  journalctl -u monitoring-agent -f"
    echo ""
    echo "Control service:"
    echo "  systemctl status monitoring-agent"
    echo "  systemctl stop monitoring-agent"
    echo "  systemctl restart monitoring-agent"
    echo ""
else
    echo ""
    echo "❌ Installation failed - service is not running"
    echo "Check logs with: journalctl -u monitoring-agent -n 50"
    exit 1
fi
