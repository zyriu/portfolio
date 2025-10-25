#!/usr/bin/env bash
set -e

APP_NAME="Portfolio"
OUTPUT_DIR="build/bin"

echo "🧱 Starting Wails cross-platform build..."

# Check prerequisites
command -v wails >/dev/null 2>&1 || { echo "❌ Wails not found. Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest"; exit 1; }

# Detect host OS
HOST_OS="$(uname -s)"

# --- Linux host ---
if [[ "$HOST_OS" == "Linux" ]]; then
    echo "🐧 Building from Linux..."
    echo "Installing cross-compilers (requires sudo)..."
    sudo apt-get update && sudo apt-get install -y mingw-w64 gcc-aarch64-linux-gnu || true

    echo "➡️  Building Linux targets..."
    wails build -platform linux/amd64,linux/arm64

    echo "➡️  Building Windows targets..."
    CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc wails build -platform windows/amd64
    CGO_ENABLED=1 CC=aarch64-w64-mingw32-gcc wails build -platform windows/arm64 || true

# --- macOS host ---
elif [[ "$HOST_OS" == "Darwin" ]]; then
    echo "🍏 Building from macOS..."
    xcode-select --install 2>/dev/null || true

    echo "➡️  Building macOS universal binary..."
    wails build -platform darwin/universal

    echo "➡️  Building Linux and Windows (requires cross-toolchains)..."
    brew install mingw-w64 || true
    CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc wails build -platform windows/amd64 || true
    wails build -platform linux/amd64 || true

# --- Windows host ---
elif [[ "$HOST_OS" == "MINGW"* || "$HOST_OS" == "MSYS"* || "$HOST_OS" == "CYGWIN"* ]]; then
    echo "🪟 Building from Windows..."
    wails build -platform windows/amd64,windows/arm64

else
    echo "❌ Unsupported host OS: $HOST_OS"
    exit 1
fi

echo "✅ All builds complete!"
echo "📦 Output located in: $OUTPUT_DIR"
