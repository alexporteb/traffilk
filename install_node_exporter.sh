#!/bin/bash
set -e

echo "Starting node_exporter installation..."

# 1. Determine privilege escalation method
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
  if command -v sudo >/dev/null 2>&1; then
    SUDO="sudo"
  else
    echo "Error: This script must be run as root or with sudo privileges."
    exit 1
  fi
fi

# 2. Determine download tool (curl or wget)
if command -v curl >/dev/null 2>&1; then
  DOWNLOAD="curl -sL"
elif command -v wget >/dev/null 2>&1; then
  DOWNLOAD="wget -qO-"
else
  echo "Error: Neither curl nor wget is installed. Please install one of them."
  exit 1
fi

# 3. Create a safe temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

VERSION="1.8.2"
ARCHIVE="node_exporter-${VERSION}.linux-amd64.tar.gz"
URL="https://github.com/prometheus/node_exporter/releases/download/v${VERSION}/${ARCHIVE}"
CHECKSUM="6809dd0b3ec45fd6e992c19071d6b5253aed3ead7bf0686885a51d85c6643c66"

echo "Downloading node_exporter v${VERSION}..."
if ! $DOWNLOAD "$URL" > "$ARCHIVE"; then
    echo "Error: Failed to download node_exporter."
    cd /
    rm -rf "$TMP_DIR"
    exit 1
fi

echo "$CHECKSUM  $ARCHIVE" | sha256sum -c - || {
    echo "Error: Checksum verification failed!"
    cd /
    rm -rf "$TMP_DIR"
    exit 1
}

tar xvz -f "$ARCHIVE"

# 4. Secure installation
echo "Installing binary..."
$SUDO mv node_exporter-${VERSION}.linux-amd64/node_exporter /usr/local/bin/
$SUDO chown root:root /usr/local/bin/node_exporter
$SUDO chmod 755 /usr/local/bin/node_exporter

# Clean up
cd /
rm -rf "$TMP_DIR"

# 5. Create user if not exists
if ! id -u node_exporter &>/dev/null; then
    echo "Creating node_exporter user..."
    $SUDO useradd -rs /bin/false node_exporter
fi

# 6. Create systemd service securely
echo "Creating systemd service..."
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

# 7. Start service
echo "Starting service..."
$SUDO systemctl daemon-reload
$SUDO systemctl enable node_exporter
$SUDO systemctl restart node_exporter

echo "✅ node_exporter successfully installed and secured! Running on port 9100."
