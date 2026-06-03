#!/bin/bash

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  SUDO="sudo"
fi

echo "Starting node_exporter installation..."

# Download and extract
wget -qO- https://github.com/prometheus/node_exporter/releases/download/v1.6.1/node_exporter-1.6.1.linux-amd64.tar.gz | tar xvz

# Move binary
$SUDO mv node_exporter-1.6.1.linux-amd64/node_exporter /usr/local/bin/
rm -rf node_exporter-1.6.1.linux-amd64

# Create user if not exists
$SUDO id -u node_exporter &>/dev/null || $SUDO useradd -rs /bin/false node_exporter

# Create systemd service
cat << 'EOF' | $SUDO tee /etc/systemd/system/node_exporter.service > /dev/null
[Unit]
Description=Node Exporter
After=network.target

[Service]
User=node_exporter
Group=node_exporter
Type=simple
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=multi-user.target
EOF

# Start service
$SUDO systemctl daemon-reload
$SUDO systemctl enable --now node_exporter

echo "node_exporter successfully installed and started on port 9100!"
