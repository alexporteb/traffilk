package scheduler

import (
	"log"
	"sync"
	"time"
	"traffilk/db"
	"traffilk/scraper"

	"github.com/robfig/cron/v3"
)

var c *cron.Cron

// cpuState stores previous CPU counters for delta calculation
type cpuState struct {
	idleTotal float64
	cpuTotal  float64
}

var (
	prevCPU   = make(map[int]cpuState)
	prevCPUMu sync.Mutex
)

func Start() {
	c = cron.New()

	// Run every minute
	_, err := c.AddFunc("* * * * *", func() {
		PollAllNodes()
	})
	if err != nil {
		log.Fatalf("Error scheduling cron: %v", err)
	}

	c.Start()
	log.Println("Scheduler started")

	// Run once on startup
	go PollAllNodes()
}

func PollAllNodes() {
	log.Println("Starting polling for all nodes...")
	nodes, err := db.GetNodes()
	if err != nil {
		log.Println("Error fetching nodes:", err)
		return
	}

	for _, node := range nodes {
		PollNode(node)
	}
	log.Println("Polling complete.")
}

// PollNode fetches metrics for a single node and updates the database
func PollNode(node db.Node) {
	metrics, err := scraper.ReadAllMetrics(node.URL)
	if err != nil {
		log.Printf("Error polling node %s (%s): %v\n", node.Name, node.URL, err)
		db.UpdateNodeStatus(node.ID, "down")
		db.UpdateNodeTrafficStats(node.ID, 0, 0, 0)
		// Zero out system metrics when node is down
		db.UpdateNodeSystemMetrics(node.ID, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		return
	}

	prevLog, _ := db.GetLatestTrafficLog(node.ID)

	err = db.AddTrafficLog(node.ID, metrics.RxBytes, metrics.TxBytes)
	if err != nil {
		log.Printf("Error saving traffic log for node %s: %v\n", node.Name, err)
	} else {
		db.UpdateNodeStatus(node.ID, "up")
		log.Printf("Successfully polled node %s: RX %d, TX %d\n", node.Name, metrics.RxBytes, metrics.TxBytes)

		var rxSpeed, txSpeed, addUsedBytes int64
		if prevLog != nil {
			deltaRx := metrics.RxBytes - prevLog.RxBytes
			deltaTx := metrics.TxBytes - prevLog.TxBytes
			if deltaRx < 0 {
				deltaRx = metrics.RxBytes
			}
			if deltaTx < 0 {
				deltaTx = metrics.TxBytes
			}

			seconds := int64(time.Since(prevLog.Timestamp).Seconds())
			if seconds <= 0 {
				seconds = 1
			}

			rxSpeed = deltaRx / seconds
			txSpeed = deltaTx / seconds
			addUsedBytes = deltaRx + deltaTx
		}
		db.UpdateNodeTrafficStats(node.ID, addUsedBytes, rxSpeed, txSpeed)
	}

	// Calculate CPU percentage from deltas
	cpuPercent := calculateCPUPercent(node.ID, metrics.CpuIdleTotal, metrics.CpuTotal)

	// Calculate memory used
	memUsed := metrics.MemTotalBytes - metrics.MemAvailBytes
	if memUsed < 0 {
		memUsed = 0
	}

	// Update system metrics
	db.UpdateNodeSystemMetrics(
		node.ID,
		cpuPercent,
		metrics.LoadAvg1,
		metrics.LoadAvg5,
		metrics.LoadAvg15,
		metrics.MemTotalBytes,
		memUsed,
		metrics.UptimeSeconds(),
		metrics.NetDropsRx,
		metrics.NetDropsTx,
		metrics.FileDescriptors,
		metrics.TcpConnections,
	)
}

// calculateCPUPercent computes CPU usage percentage from cumulative counters.
// Uses the delta between the current and previous poll's idle/total CPU seconds.
func calculateCPUPercent(nodeID int, idleTotal, cpuTotal float64) float64 {
	prevCPUMu.Lock()
	defer prevCPUMu.Unlock()

	prev, exists := prevCPU[nodeID]
	prevCPU[nodeID] = cpuState{idleTotal: idleTotal, cpuTotal: cpuTotal}

	if !exists || cpuTotal <= prev.cpuTotal {
		return 0
	}

	deltaTotal := cpuTotal - prev.cpuTotal
	deltaIdle := idleTotal - prev.idleTotal

	if deltaTotal <= 0 {
		return 0
	}

	percent := 100.0 * (1.0 - deltaIdle/deltaTotal)
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	return percent
}
