package main

import (
	"log"
	"os"
	"traffilk/api"
	"traffilk/db"
	"traffilk/scheduler"
)

func main() {
	// Initialize database
	os.MkdirAll("./data", 0755)
	db.InitDB("./data/data.db")

	// Start the background scheduler
	scheduler.Start()

	// Setup and start HTTP server
	r := api.SetupRouter()
	log.Println("Server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
