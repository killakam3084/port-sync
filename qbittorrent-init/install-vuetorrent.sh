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
EXTRACT_DIR="$TEMP_DIR/extracted"
rm -rf "$EXTRACT_DIR"
mkdir -p "$EXTRACT_DIR"
unzip -q "$TEMP_DIR/vuetorrent.zip" -d "$EXTRACT_DIR"

# Copy to config, preserving ownership. VueTorrent release zips may contain
# either a top-level "vuetorrent/" folder or legacy "dist/" output.
if [ -d "$EXTRACT_DIR/vuetorrent" ]; then
    SOURCE_DIR="$EXTRACT_DIR/vuetorrent"
elif [ -d "$EXTRACT_DIR/dist" ]; then
    SOURCE_DIR="$EXTRACT_DIR/dist"
else
    echo "[VueTorrent] ERROR: Unexpected archive layout (no vuetorrent/ or dist/)" >&2
    exit 1
fi

echo "[VueTorrent] Installing to $VUETORRENT_DIR from $SOURCE_DIR..." >&2
rm -rf "$VUETORRENT_DIR"
cp -r "$SOURCE_DIR" "$VUETORRENT_DIR"
chown -R abc:abc "$VUETORRENT_DIR"

# Cleanup
rm -rf "$TEMP_DIR"

echo "[VueTorrent] Installation complete" >&2
