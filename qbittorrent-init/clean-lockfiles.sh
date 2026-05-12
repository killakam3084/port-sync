#!/bin/bash
# /custom-cont-init.d/clean-lockfiles.sh
#
# Runs before s6 starts qbittorrent-nox (via LSIO's custom-cont-init.d hook).
# Removes stale lockfile and ipc-socket left behind by unclean shutdowns —
# without this, qBittorrent detects the lockfile, assumes another instance is
# already running, and exits 0 silently, causing s6 to restart it endlessly.

QBIT_CONF_DIR="${XDG_CONFIG_HOME:-/config}/qBittorrent"

for f in lockfile ipc-socket; do
    path="${QBIT_CONF_DIR}/${f}"
    if [[ -e "${path}" ]]; then
        echo "[custom-init] Removing stale ${f}: ${path}"
        rm -f "${path}"
    fi
done
