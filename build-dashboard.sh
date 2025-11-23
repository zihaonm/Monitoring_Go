#!/bin/bash

# Monitoring Dashboard Build Script for Linux (Ubuntu)
# This script builds the main monitoring dashboard for deployment on Linux servers

set -e  # Exit on error

echo "🔨 Building Monitoring Dashboard for Linux (Ubuntu)..."
echo ""

# Get version from git or use default
VERSION=$(git describe --tags --always 2>/dev/null || echo "v1.0.0")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# Build information
GOOS=linux
GOARCH=amd64
OUTPUT_DIR="dist"
OUTPUT_NAME="monitoring-dashboard-linux-amd64"

echo "📦 Build Information:"
echo "   OS/Arch: ${GOOS}/${GOARCH}"
echo "   Version: ${VERSION}"
echo "   Build Time: ${BUILD_TIME}"
echo ""

# Create dist directory
mkdir -p ${OUTPUT_DIR}

# Build the binary with optimizations
echo "⚙️  Compiling..."
CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
    -ldflags="-w -s -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" \
    -o ${OUTPUT_DIR}/${OUTPUT_NAME} \
    .

# Make it executable
chmod +x ${OUTPUT_DIR}/${OUTPUT_NAME}

# Get file size
FILE_SIZE=$(ls -lh ${OUTPUT_DIR}/${OUTPUT_NAME} | awk '{print $5}')

echo ""
echo "✅ Build completed successfully!"
echo "   Output: ${OUTPUT_DIR}/${OUTPUT_NAME}"
echo "   Size: ${FILE_SIZE}"
echo ""
echo "📂 Required files to deploy:"
echo "   - ${OUTPUT_DIR}/${OUTPUT_NAME} (binary)"
echo "   - templates/ (HTML templates)"
echo "   - monitoring_data.json (will be created on first run)"
echo ""
echo "📤 To deploy to Ubuntu server:"
echo "   # Create app directory"
echo "   ssh user@server 'sudo mkdir -p /opt/monitoring && sudo chown \$USER /opt/monitoring'"
echo ""
echo "   # Upload binary and templates"
echo "   scp ${OUTPUT_DIR}/${OUTPUT_NAME} user@server:/opt/monitoring/monitoring"
echo "   scp -r templates user@server:/opt/monitoring/"
echo ""
echo "   # Optional: Upload existing data"
echo "   scp monitoring_data.json user@server:/opt/monitoring/"
echo ""
echo "🚀 To run on server:"
echo "   ssh user@server 'cd /opt/monitoring && ./monitoring'"
echo ""
echo "🌐 Access dashboard:"
echo "   http://server-ip:8080"
echo ""
