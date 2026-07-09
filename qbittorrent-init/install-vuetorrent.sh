#!/bin/bash
# Install VueTorrent theme from latest GitHub release.
# Runs at container startup via /custom-cont-init.d/
# Output visible in `docker logs qbittorrent`
set -e

CONFIG_DIR="/config/qBittorrent"
THEMES_DIR="$CONFIG_DIR/themes"
VUETORRENT_DIR="$THEMES_DIR/vuetorrent"

# GitHub release URL for VueTorrent
VUETORRENT_URL="https://github.com/VueTorrent/VueTorrent/releases/download/v2.34.0/vuetorrent.zip"
TEMP_DIR="/tmp/vuetorrent-install"

echo "[VueTorrent] Installing VueTorrent theme..."
echo "[VueTorrent] URL: $VUETORRENT_URL" >&2

# Create directories
mkdir -p "$THEMES_DIR"
mkdir -p "$TEMP_DIR"

# Download and extract
echo "[VueTorrent] Downloading..." >&2
if ! curl -sL -o "$TEMP_DIR/vuetorrent.zip" "$VUETORRENT_URL"; then
    echo "[VueTorrent] ERROR: Failed to download VueTorrent" >&2
    exit 1
fi

echo "[VueTorrent] Extracting..." >&2
cd "$TEMP_DIR"
unzip -q vuetorrent.zip

# Copy to config, preserving ownership
echo "[VueTorrent] Installing to $VUETORRENT_DIR..." >&2
rm -rf "$VUETORRENT_DIR"
cp -r dist "$VUETORRENT_DIR"
chown -R abc:abc "$VUETORRENT_DIR"

# Cleanup
rm -rf "$TEMP_DIR"

echo "[VueTorrent] Installation complete" >&2
