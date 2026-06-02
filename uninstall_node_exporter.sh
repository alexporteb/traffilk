#!/bin/bash

echo "Starting node_exporter uninstallation..."

# Stop and disable service
sudo systemctl stop node_exporter &>/dev/null || true
sudo systemctl disable node_exporter &>/dev/null || true

# Remove systemd service file
sudo rm -f /etc/systemd/system/node_exporter.service
sudo systemctl daemon-reload

# Remove binary
sudo rm -f /usr/local/bin/node_exporter

# Remove user
sudo userdel node_exporter &>/dev/null || true

echo "node_exporter successfully uninstalled and completely removed from the system!"
