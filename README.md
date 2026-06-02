# Traffilk V2

A lightweight, beautiful, and secure traffic monitoring dashboard for Prometheus Node Exporter.
Traffilk connects to your existing Prometheus metrics endpoints to fetch and aggregate daily incoming/outgoing network traffic. 

V2 features a completely redesigned user interface (inspired by Uptime Kuma) and built-in JWT authentication.

## Features

- **Beautiful Uptime Kuma-inspired UI**: Dark theme, rounded components, responsive design.
- **Secure Access**: Built-in login screen using JWT cookies.
- **Fast Updates**: Polls nodes every minute to provide real-time traffic deltas.
- **Multi-Language Support**: Fully translated in English (EN) and Russian (RU).
- **Easy Management**: Add, Edit, and Delete nodes seamlessly.

## Prerequisites

Each node you want to monitor must be running `node_exporter` (specifically exposing `node_network_receive_bytes_total` and `node_network_transmit_bytes_total`).

## Quick Start (Docker)

1. Clone the repository:
   ```bash
   git clone https://github.com/alexporteb/traefikk.git
   cd traefikk
   ```

2. Open `docker-compose.yml` and modify the environment variables to set your admin credentials:
   ```yaml
   environment:
     - ADMIN_USER=admin
     - ADMIN_PASS=your_secure_password
     - JWT_SECRET=change-this-secret
   ```

3. Start the container:
   ```bash
   docker compose up -d
   ```

4. Access the dashboard:
   Open your browser and navigate to `http://localhost:8080/ui/` (or your reverse proxy domain).

## Adding Nodes

1. Log in using your `ADMIN_USER` and `ADMIN_PASS`.
2. Click **Add New Monitor** in the sidebar.
3. Enter a **Friendly Name** (e.g., `Web Server`).
4. Enter the **Prometheus URL** (e.g., `https://your-node.com/metrics`).
5. Click **Save**.

*Note: Traffilk will instantly poll the node upon adding or editing it. If this is a new node, the chart will show 0 bytes for the current day because it needs at least two data points (1 minute apart) to calculate the traffic delta.*

## Technical Details

- **Backend**: Go (Gin framework, SQLite)
- **Frontend**: HTML5, Vue 3, TailwindCSS, Chart.js
- **Metrics Parser**: Custom lightweight HTTP scanner (no heavy Prometheus libraries required)
- **Database**: SQLite (stored in `./data/data.db`)
