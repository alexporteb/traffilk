package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type Node struct {
	ID                      int    `json:"id"`
	Name                    string `json:"name"`
	URL                     string `json:"url"`
	Status                  string `json:"status"`
	TrafficUsedBytes        int64  `json:"trafficUsedBytes"`
	TrafficLimitBytes       int64  `json:"trafficLimitBytes"`
	IsTrafficTrackingActive bool   `json:"isTrafficTrackingActive"`
	TrafficResetDay         int    `json:"trafficResetDay"`
	RxBytesPerSec           int64  `json:"rxBytesPerSec"`
	TxBytesPerSec           int64  `json:"txBytesPerSec"`
}

type TrafficLog struct {
	ID        int       `json:"id"`
	NodeID    int       `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
	RxBytes   int64     `json:"rx_bytes"`
	TxBytes   int64     `json:"tx_bytes"`
}

type APIToken struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TokenHash string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
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

	// Safe migration: add status column if it doesn't exist
	DB.Exec("ALTER TABLE nodes ADD COLUMN status TEXT DEFAULT 'unknown'")
	DB.Exec("ALTER TABLE nodes ADD COLUMN traffic_used_bytes INTEGER DEFAULT 0")
	DB.Exec("ALTER TABLE nodes ADD COLUMN traffic_limit_bytes INTEGER DEFAULT 0")
	DB.Exec("ALTER TABLE nodes ADD COLUMN is_traffic_tracking_active BOOLEAN DEFAULT 0")
	DB.Exec("ALTER TABLE nodes ADD COLUMN traffic_reset_day INTEGER DEFAULT 1")
	DB.Exec("ALTER TABLE nodes ADD COLUMN rx_bytes_per_sec INTEGER DEFAULT 0")
	DB.Exec("ALTER TABLE nodes ADD COLUMN tx_bytes_per_sec INTEGER DEFAULT 0")

	_, err = DB.Exec(trafficTable)
	if err != nil {
		log.Fatal("Failed to create traffic_logs table:", err)
	}

	apiTokensTable := `
	CREATE TABLE IF NOT EXISTS api_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		token_hash TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(apiTokensTable)
	if err != nil {
		log.Fatal("Failed to create api_tokens table:", err)
	}

	// Performance indexes
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_traffic_node_timestamp ON traffic_logs(node_id, timestamp)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status)")
}

// GetNodes returns all nodes
func GetNodes() ([]Node, error) {
	rows, err := DB.Query("SELECT id, name, url, status, traffic_used_bytes, traffic_limit_bytes, is_traffic_tracking_active, traffic_reset_day, rx_bytes_per_sec, tx_bytes_per_sec FROM nodes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		err := rows.Scan(&n.ID, &n.Name, &n.URL, &n.Status, &n.TrafficUsedBytes, &n.TrafficLimitBytes, &n.IsTrafficTrackingActive, &n.TrafficResetDay, &n.RxBytesPerSec, &n.TxBytesPerSec)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

// AddNode adds a new node
func AddNode(n Node) error {
	if n.TrafficResetDay == 0 {
		n.TrafficResetDay = 1
	}
	_, err := DB.Exec("INSERT INTO nodes (name, url, traffic_limit_bytes, is_traffic_tracking_active, traffic_reset_day) VALUES (?, ?, ?, ?, ?)",
		n.Name, n.URL, n.TrafficLimitBytes, n.IsTrafficTrackingActive, n.TrafficResetDay)
	return err
}

// UpdateNode updates an existing node
func UpdateNode(id int, n Node) error {
	if n.TrafficResetDay == 0 {
		n.TrafficResetDay = 1
	}
	_, err := DB.Exec("UPDATE nodes SET name = ?, url = ?, traffic_limit_bytes = ?, is_traffic_tracking_active = ?, traffic_reset_day = ? WHERE id = ?",
		n.Name, n.URL, n.TrafficLimitBytes, n.IsTrafficTrackingActive, n.TrafficResetDay, id)
	return err
}

// UpdateNodeStatus updates the online/offline status of a node
func UpdateNodeStatus(id int, status string) error {
	_, err := DB.Exec("UPDATE nodes SET status = ? WHERE id = ?", status, id)
	return err
}

// UpdateNodeTrafficStats updates live traffic and speed for a node
func UpdateNodeTrafficStats(id int, addUsedBytes, rxSpeed, txSpeed int64) error {
	_, err := DB.Exec("UPDATE nodes SET traffic_used_bytes = traffic_used_bytes + ?, rx_bytes_per_sec = ?, tx_bytes_per_sec = ? WHERE id = ?", addUsedBytes, rxSpeed, txSpeed, id)
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

// AddAPIToken adds a new token hash
func AddAPIToken(name, tokenHash string) error {
	_, err := DB.Exec("INSERT INTO api_tokens (name, token_hash) VALUES (?, ?)", name, tokenHash)
	return err
}

// GetAPITokens returns all tokens without their hashes
func GetAPITokens() ([]APIToken, error) {
	rows, err := DB.Query("SELECT id, name, created_at FROM api_tokens")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []APIToken
	for rows.Next() {
		var t APIToken
		err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// DeleteAPIToken deletes a token by ID
func DeleteAPIToken(id int) error {
	_, err := DB.Exec("DELETE FROM api_tokens WHERE id = ?", id)
	return err
}

// ValidateAPIToken checks if a token hash exists
func ValidateAPIToken(tokenHash string) bool {
	var id int
	err := DB.QueryRow("SELECT id FROM api_tokens WHERE token_hash = ?", tokenHash).Scan(&id)
	return err == nil
}

