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

// ReadPrometheusMetrics connects via HTTP(S) and parses Prometheus metrics to get total RX and TX bytes
func ReadPrometheusMetrics(url string) (rx int64, tx int64, err error) {
	// Disable TLS verification to allow self-signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var totalRx, totalTx float64

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "node_network_receive_bytes_total") {
			val, _ := parseMetricLine(line)
			totalRx += val
		} else if strings.HasPrefix(line, "node_network_transmit_bytes_total") {
			val, _ := parseMetricLine(line)
			totalTx += val
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("error reading metrics: %v", err)
	}

	return int64(totalRx), int64(totalTx), nil
}

func parseMetricLine(line string) (float64, string) {
	// node_network_receive_bytes_total{device="eth0"} 1.23456789e+08
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

	// Ignore loopback and virtual interfaces
	if device == "lo" || strings.HasPrefix(device, "veth") || strings.HasPrefix(device, "docker") || strings.HasPrefix(device, "br-") {
		return 0, device
	}

	valStr := parts[len(parts)-1]
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, device
	}

	return val, device
}
