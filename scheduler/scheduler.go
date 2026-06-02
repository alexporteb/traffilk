package scheduler

import (
	"log"
	"traffilk/db"
	"traffilk/scraper"

	"github.com/robfig/cron/v3"
)

var c *cron.Cron

func Start() {
	c = cron.New()
	
	// Run every hour
	_, err := c.AddFunc("@hourly", func() {
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
			continue
		}

		err = db.AddTrafficLog(node.ID, rx, tx)
		if err != nil {
			log.Printf("Error saving traffic log for node %s: %v\n", node.Name, err)
		} else {
			log.Printf("Successfully polled node %s: RX %d, TX %d\n", node.Name, rx, tx)
		}
	}
	log.Println("Polling complete.")
}
