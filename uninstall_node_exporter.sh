#!/bin/bash

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  SUDO="sudo"
fi

echo "Starting node_exporter uninstallation..."

# Stop and disable service
$SUDO systemctl stop node_exporter &>/dev/null || true
$SUDO systemctl disable node_exporter &>/dev/null || true

# Remove systemd service file
$SUDO rm -f /etc/systemd/system/node_exporter.service
$SUDO systemctl daemon-reload

# Remove binary
$SUDO rm -f /usr/local/bin/node_exporter

# Remove user
$SUDO userdel node_exporter &>/dev/null || true

echo "node_exporter successfully uninstalled and completely removed from the system!"
