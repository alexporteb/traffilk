package main

import (
	"fmt"
	"strings"
	"strconv"
)

func parseMetricValue(line string) (float64, error) {
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid metric line")
	}
	return strconv.ParseFloat(parts[len(parts)-1], 64)
}

func main() {
	lines := []string{
		`node_cpu_seconds_total{cpu="0",mode="idle"} 1551.94`,
		`node_cpu_seconds_total{cpu="0",mode="system"} 18.23`,
	}
	var total, idle float64
	for _, line := range lines {
		if strings.HasPrefix(line, "node_cpu_seconds_total") {
			val, _ := parseMetricValue(line)
			total += val
			if strings.Contains(line, `mode="idle"`) {
				idle += val
			}
		}
	}
	fmt.Printf("Total: %f, Idle: %f\n", total, idle)
}
