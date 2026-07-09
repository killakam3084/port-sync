#!/bin/bash
# Install VueTorrent theme from latest GitHub release.
# Runs at container startup via /custom-cont-init.d/
set -e

CONFIG_DIR="/config/qBittorrent"
THEMES_DIR="$CONFIG_DIR/themes"
VUETORRENT_DIR="$THEMES_DIR/vuetorrent"

# GitHub release URL for VueTorrent
VUETORRENT_URL="https://github.com/VueTorrent/VueTorrent/releases/download/v2.34.0/vuetorrent.zip"
TEMP_DIR="/tmp/vuetorrent-install"

echo "[VueTorrent] Installing VueTorrent theme..."

# Create directories
mkdir -p "$THEMES_DIR"
mkdir -p "$TEMP_DIR"

# Download and extract
echo "[VueTorrent] Downloading from $VUETORRENT_URL..."
if ! curl -sL -o "$TEMP_DIR/vuetorrent.zip" "$VUETORRENT_URL"; then
    echo "[VueTorrent] Error: Failed to download VueTorrent"
    exit 1
fi

echo "[VueTorrent] Extracting..."
cd "$TEMP_DIR"
unzip -q vuetorrent.zip

# Copy to config, preserving ownership
echo "[VueTorrent] Installing to $VUETORRENT_DIR..."
rm -rf "$VUETORRENT_DIR"
cp -r dist "$VUETORRENT_DIR"
chown -R abc:abc "$VUETORRENT_DIR"

# Cleanup
rm -rf "$TEMP_DIR"

echo "[VueTorrent] Installation complete"
