package scheduler

import (
	"log"
	"time"
	"traffilk/db"
	"traffilk/scraper"

	"github.com/robfig/cron/v3"
)

var c *cron.Cron

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
		rx, tx, err := scraper.ReadPrometheusMetrics(node.URL)
		if err != nil {
			log.Printf("Error polling node %s (%s): %v\n", node.Name, node.URL, err)
			db.UpdateNodeStatus(node.ID, "down")
			db.UpdateNodeTrafficStats(node.ID, 0, 0, 0)
			continue
		}

		prevLog, _ := db.GetLatestTrafficLog(node.ID)

		err = db.AddTrafficLog(node.ID, rx, tx)
		if err != nil {
			log.Printf("Error saving traffic log for node %s: %v\n", node.Name, err)
		} else {
			db.UpdateNodeStatus(node.ID, "up")
			log.Printf("Successfully polled node %s: RX %d, TX %d\n", node.Name, rx, tx)

			var rxSpeed, txSpeed, addUsedBytes int64
			if prevLog != nil {
				deltaRx := rx - prevLog.RxBytes
				deltaTx := tx - prevLog.TxBytes
				if deltaRx < 0 {
					deltaRx = rx
				}
				if deltaTx < 0 {
					deltaTx = tx
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
	}
	log.Println("Polling complete.")
}
