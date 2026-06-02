package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Node struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type TrafficLog struct {
	ID        int       `json:"id"`
	NodeID    int       `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
	RxBytes   int64     `json:"rx_bytes"`
	TxBytes   int64     `json:"tx_bytes"`
}

type DailyTraffic struct {
	Date    string `json:"date"`
	RxBytes int64  `json:"rx_bytes"`
	TxBytes int64  `json:"tx_bytes"`
}

func InitDB(filepath string) {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	createTables()
}

func createTables() {
	nodesTable := `
	CREATE TABLE IF NOT EXISTS nodes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL
	);`

	trafficTable := `
	CREATE TABLE IF NOT EXISTS traffic_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		node_id INTEGER,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		rx_bytes INTEGER,
		tx_bytes INTEGER,
		FOREIGN KEY(node_id) REFERENCES nodes(id)
	);`

	_, err := DB.Exec(nodesTable)
	if err != nil {
		log.Fatal("Failed to create nodes table:", err)
	}

	_, err = DB.Exec(trafficTable)
	if err != nil {
		log.Fatal("Failed to create traffic_logs table:", err)
	}
}

// GetNodes returns all nodes
func GetNodes() ([]Node, error) {
	rows, err := DB.Query("SELECT id, name, url FROM nodes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		err := rows.Scan(&n.ID, &n.Name, &n.URL)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// AddNode adds a new node
func AddNode(n Node) error {
	_, err := DB.Exec("INSERT INTO nodes (name, url) VALUES (?, ?)",
		n.Name, n.URL)
	return err
}

// DeleteNode removes a node
func DeleteNode(id int) error {
	_, err := DB.Exec("DELETE FROM nodes WHERE id = ?", id)
	if err == nil {
		DB.Exec("DELETE FROM traffic_logs WHERE node_id = ?", id)
	}
	return err
}

// AddTrafficLog adds a reading
func AddTrafficLog(nodeID int, rx, tx int64) error {
	_, err := DB.Exec("INSERT INTO traffic_logs (node_id, rx_bytes, tx_bytes) VALUES (?, ?, ?)",
		nodeID, rx, tx)
	return err
}

// GetLatestTrafficLog gets the latest raw counter reading to calculate deltas
func GetLatestTrafficLog(nodeID int) (*TrafficLog, error) {
	row := DB.QueryRow("SELECT id, node_id, timestamp, rx_bytes, tx_bytes FROM traffic_logs WHERE node_id = ? ORDER BY timestamp DESC LIMIT 1", nodeID)
	var t TrafficLog
	err := row.Scan(&t.ID, &t.NodeID, &t.Timestamp, &t.RxBytes, &t.TxBytes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetDailyTraffic returns aggregated daily traffic for a node
func GetDailyTraffic(nodeID int) ([]DailyTraffic, error) {
	// Simple query to get daily deltas. We need to sum up deltas correctly.
	// We'll calculate the daily deltas from the Go side or SQL side.
	// Doing it from SQL side with window functions (SQLite 3.25+):
	query := `
	WITH ranked AS (
		SELECT 
			date(timestamp) as date,
			rx_bytes,
			tx_bytes,
			LAG(rx_bytes) OVER (ORDER BY timestamp) as prev_rx,
			LAG(tx_bytes) OVER (ORDER BY timestamp) as prev_tx
		FROM traffic_logs
		WHERE node_id = ?
	)
	SELECT 
		date,
		SUM(CASE WHEN rx_bytes >= IFNULL(prev_rx, 0) THEN rx_bytes - IFNULL(prev_rx, rx_bytes) ELSE rx_bytes END) as daily_rx,
		SUM(CASE WHEN tx_bytes >= IFNULL(prev_tx, 0) THEN tx_bytes - IFNULL(prev_tx, tx_bytes) ELSE tx_bytes END) as daily_tx
	FROM ranked
	GROUP BY date
	ORDER BY date ASC
	`
	rows, err := DB.Query(query, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var daily []DailyTraffic
	for rows.Next() {
		var d DailyTraffic
		err := rows.Scan(&d.Date, &d.RxBytes, &d.TxBytes)
		if err != nil {
			return nil, err
		}
		daily = append(daily, d)
	}
	return daily, nil
}
