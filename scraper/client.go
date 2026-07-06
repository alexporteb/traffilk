package scraper

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// NodeMetrics holds all parsed metrics from a node_exporter endpoint
type NodeMetrics struct {
	RxBytes         int64
	TxBytes         int64
	CpuIdleTotal    float64 // sum of node_cpu_seconds_total{mode="idle"} across all cores
	CpuTotal        float64 // sum of node_cpu_seconds_total across all modes and cores
	LoadAvg1        float64 // node_load1
	LoadAvg5        float64 // node_load5
	LoadAvg15       float64 // node_load15
	MemTotalBytes   int64   // node_memory_MemTotal_bytes
	MemAvailBytes   int64   // node_memory_MemAvailable_bytes
	BootTime        float64 // node_boot_time_seconds
	NodeTime        float64 // node_time_seconds
	NetDropsRx      int64   // sum of node_network_receive_drop_total (physical interfaces)
	NetDropsTx      int64   // sum of node_network_transmit_drop_total (physical interfaces)
	FileDescriptors int64   // node_filefd_allocated
}

// UptimeSeconds returns the computed uptime in seconds
func (m *NodeMetrics) UptimeSeconds() int64 {
	if m.NodeTime > 0 && m.BootTime > 0 {
		return int64(m.NodeTime - m.BootTime)
	}
	return 0
}

// ReadAllMetrics connects to a node_exporter /metrics endpoint and parses
// all relevant system metrics in a single pass
func ReadAllMetrics(url string) (*NodeMetrics, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	if !strings.HasSuffix(url, "/metrics") {
		// If URL is just http://ip:port or http://ip:port/, append /metrics
		if strings.Count(strings.TrimPrefix(url, "http://"), "/") == 0 ||
			strings.Count(strings.TrimPrefix(url, "https://"), "/") == 0 ||
			strings.HasSuffix(url, "/") {
			url = strings.TrimRight(url, "/") + "/metrics"
		}
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	metrics := &NodeMetrics{}
	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer for large metrics pages
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Network traffic (physical interfaces only)
		if strings.HasPrefix(line, "node_network_receive_bytes_total") {
			val, device := parseMetricLine(line)
			if !isVirtualDevice(device) {
				metrics.RxBytes += int64(val)
			}
		} else if strings.HasPrefix(line, "node_network_transmit_bytes_total") {
			val, device := parseMetricLine(line)
			if !isVirtualDevice(device) {
				metrics.TxBytes += int64(val)
			}
		} else if strings.HasPrefix(line, "node_network_receive_drop_total") {
			val, device := parseMetricLine(line)
			if !isVirtualDevice(device) {
				metrics.NetDropsRx += int64(val)
			}
		} else if strings.HasPrefix(line, "node_network_transmit_drop_total") {
			val, device := parseMetricLine(line)
			if !isVirtualDevice(device) {
				metrics.NetDropsTx += int64(val)
			}
		} else if strings.HasPrefix(line, "node_cpu_seconds_total") {
			// CPU: accumulate total and idle separately
			val, _ := parseMetricValue(line)
			metrics.CpuTotal += val
			if strings.Contains(line, `mode="idle"`) {
				metrics.CpuIdleTotal += val
			}
		} else if strings.HasPrefix(line, "node_load1 ") {
			metrics.LoadAvg1, _ = parseMetricValue(line)
		} else if strings.HasPrefix(line, "node_load5 ") {
			metrics.LoadAvg5, _ = parseMetricValue(line)
		} else if strings.HasPrefix(line, "node_load15 ") {
			metrics.LoadAvg15, _ = parseMetricValue(line)
		} else if strings.HasPrefix(line, "node_memory_MemTotal_bytes ") {
			val, _ := parseMetricValue(line)
			metrics.MemTotalBytes = int64(val)
		} else if strings.HasPrefix(line, "node_memory_MemAvailable_bytes ") {
			val, _ := parseMetricValue(line)
			metrics.MemAvailBytes = int64(val)
		} else if strings.HasPrefix(line, "node_boot_time_seconds ") {
			metrics.BootTime, _ = parseMetricValue(line)
		} else if strings.HasPrefix(line, "node_time_seconds ") {
			metrics.NodeTime, _ = parseMetricValue(line)
		} else if strings.HasPrefix(line, "node_filefd_allocated ") {
			val, _ := parseMetricValue(line)
			metrics.FileDescriptors = int64(val)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading metrics: %v", err)
	}

	if metrics.CpuTotal == 0 && metrics.MemTotalBytes == 0 {
		return nil, fmt.Errorf("no metrics found at endpoint")
	}

	return metrics, nil
}

// parseMetricLine extracts the float value and device label from a Prometheus metric line
// e.g. node_network_receive_bytes_total{device="eth0"} 1.23456789e+08
func parseMetricLine(line string) (float64, string) {
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return 0, ""
	}

	// Extract device name from labels if present
	device := ""
	startIdx := strings.Index(parts[0], `device="`)
	if startIdx != -1 {
		endIdx := strings.Index(parts[0][startIdx+8:], `"`)
		if endIdx != -1 {
			device = parts[0][startIdx+8 : startIdx+8+endIdx]
		}
	}

	valStr := parts[len(parts)-1]
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, device
	}

	return val, device
}

// parseMetricValue extracts just the float value from a simple metric line
// e.g. node_load1 0.52
func parseMetricValue(line string) (float64, error) {
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid metric line")
	}
	return strconv.ParseFloat(parts[len(parts)-1], 64)
}

// isVirtualDevice returns true for loopback, docker, veth, and bridge interfaces
func isVirtualDevice(device string) bool {
	if device == "" || device == "lo" {
		return true
	}
	return strings.HasPrefix(device, "veth") ||
		strings.HasPrefix(device, "docker") ||
		strings.HasPrefix(device, "br-") ||
		strings.HasPrefix(device, "virbr")
}
