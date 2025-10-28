#!/usr/bin/env bash
set -e

APP_NAME="MyWailsApp"
OUTPUT_DIR="build/bin"

echo "🚀 Building Wails app for the current platform..."

# Check for Wails CLI
if ! command -v wails >/dev/null 2>&1; then
  echo "❌ Wails CLI not found."
  echo "Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
  exit 1
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac

# Normalize macOS name
if [[ "$OS" == "darwin" ]]; then
  OS="darwin"
elif [[ "$OS" == "linux" ]]; then
  OS="linux"
elif [[ "$OS" =~ mingw|msys|cygwin ]]; then
  OS="windows"
fi

echo "Detected platform: $OS/$ARCH"

# Run build
wails build -platform "${OS}/${ARCH}"

echo "✅ Build complete!"
echo "📦 Output: ${OUTPUT_DIR}/${OS}-${ARCH}/"
