#!/bin/bash
# Install script for BookMux (macOS and Linux)
set -e

REPO="MrZoidberg/bookmux"
OS_NAME=$(uname -s)
ARCH=$(uname -m)

# Map architecture names to match GoReleaser
case "$ARCH" in
    x86_64)  ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Error: Unsupported architecture $ARCH"; exit 1 ;;
esac

# Normalize OS Name for URL
# GoReleaser title cases the OS (Darwin, Linux)
case "$OS_NAME" in
    Darwin)  OS_NAME="Darwin" ;;
    Linux)   OS_NAME="Linux" ;;
    *) echo "Error: Unsupported OS $OS_NAME"; exit 1 ;;
esac

# Determine destination directory
BIN_DIR="$HOME/.local/bin"
mkdir -p "$BIN_DIR"

# Get latest release version
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: Could not determine the latest version from GitHub."
    exit 1
fi

# Construction of URL based on GoReleaser template: {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}.tar.gz
VERSION_NUM=${VERSION#v}
FILE_NAME="bookmux_${VERSION_NUM}_${OS_NAME}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$FILE_NAME"

echo "Downloading BookMux $VERSION for $OS_NAME $ARCH..."
TMP_DIR=$(mktemp -d)
curl -sSfL "$URL" -o "$TMP_DIR/bookmux.tar.gz"

# Extract and install
tar -xzf "$TMP_DIR/bookmux.tar.gz" -C "$TMP_DIR"
mv "$TMP_DIR/bookmux" "$BIN_DIR/bookmux"
chmod +x "$BIN_DIR/bookmux"

# Cleanup
rm -rf "$TMP_DIR"

echo "--------------------------------------------------------"
echo "BookMux $VERSION installed successfully!"
echo "Location: $BIN_DIR/bookmux"
echo ""
echo "If '$BIN_DIR' is not in your PATH, add it by running:"
echo "  export PATH=\$PATH:$BIN_DIR"
echo "Or for fish shell:"
echo "  fish_add_path $BIN_DIR"
echo "--------------------------------------------------------"
