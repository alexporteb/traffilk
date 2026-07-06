package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
	"traffilk/db"
	"traffilk/scheduler"

	"github.com/gin-gonic/gin"
)

// loginAttempt tracks brute-force protection state
type loginAttempt struct {
	count    int
	lastTime time.Time
}

var (
	loginAttempts = make(map[string]*loginAttempt)
	loginMu       sync.Mutex
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Security headers middleware
	r.Use(securityHeaders())

	// Unprotected routes
	r.POST("/api/traffilk/login", rateLimitLogin(), LoginHandler)
	r.POST("/api/traffilk/logout", LogoutHandler)

	// API routes (protected)
	api := r.Group("/api/traffilk")
	api.Use(AuthMiddleware())
	{
		api.GET("/nodes", getNodes)
		api.POST("/nodes", addNode)
		api.PUT("/nodes/:id", updateNode)
		api.DELETE("/nodes/:id", deleteNode)
		api.GET("/nodes/:id/traffic", getNodeTraffic)
		api.POST("/nodes/:id/poll", pollNode)

		api.GET("/tokens", getTokens)
		api.POST("/tokens", createToken)
		api.DELETE("/tokens/:id", deleteToken)
	}

	// Serve React SPA
	setupSPA(r)

	return r
}

// setupSPA serves the React frontend build output
func setupSPA(r *gin.Engine) {
	// Determine the frontend dist path
	distPath := "./frontend/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		// Fallback: try legacy ui/ directory
		distPath = "./ui"
	}

	// Serve static assets (js, css, images)
	r.Static("/assets", distPath+"/assets")
	r.StaticFile("/favicon.svg", distPath+"/favicon.svg")

	// SPA fallback: serve index.html for all non-API, non-asset routes
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Don't serve HTML for API routes
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
		c.File(distPath + "/index.html")
	})
}

// securityHeaders adds security-related HTTP headers to all responses
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// CSP: allow self, inline styles (Mantine needs them)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; font-src 'self'; img-src 'self' data:; connect-src 'self'")
		c.Next()
	}
}

// rateLimitLogin prevents brute-force login attacks (max 5 attempts per IP per minute)
func rateLimitLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		loginMu.Lock()
		attempt, exists := loginAttempts[ip]
		now := time.Now()

		if !exists {
			loginAttempts[ip] = &loginAttempt{count: 1, lastTime: now}
			loginMu.Unlock()
			c.Next()
			return
		}

		// Reset counter if more than 1 minute has passed
		if now.Sub(attempt.lastTime) > time.Minute {
			attempt.count = 1
			attempt.lastTime = now
			loginMu.Unlock()
			c.Next()
			return
		}

		attempt.count++
		attempt.lastTime = now

		if attempt.count > 5 {
			loginMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Try again later."})
			c.Abort()
			return
		}

		loginMu.Unlock()
		c.Next()
	}
}

func getNodes(c *gin.Context) {
	nodes, err := db.GetNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if nodes == nil {
		nodes = []db.Node{} // Return empty array instead of null
	}
	c.JSON(http.StatusOK, nodes)
}

func addNode(c *gin.Context) {
	var node db.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidNodeURL(node.URL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL or blocked internal IP (SSRF Protection)"})
		return
	}

	err := db.AddNode(node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger an immediate poll so the user doesn't have to wait for the next cron job
	go scheduler.PollAllNodes()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func deleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = db.DeleteNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func updateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var node db.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidNodeURL(node.URL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL or blocked internal IP (SSRF Protection)"})
		return
	}

	err = db.UpdateNode(id, node)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trigger immediate poll if URL changed
	go scheduler.PollAllNodes()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getNodeTraffic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	daily, err := db.GetDailyTraffic(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if daily == nil {
		daily = []db.DailyTraffic{}
	}
	c.JSON(http.StatusOK, daily)
}

func pollNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	nodes, err := db.GetNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var targetNode *db.Node
	for _, n := range nodes {
		if n.ID == id {
			targetNode = &n
			break
		}
	}

	if targetNode == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Run in background to not block the request
	go scheduler.PollNode(*targetNode)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// isValidNodeURL checks if the URL is valid and mitigates basic SSRF targeting cloud metadata
func isValidNodeURL(u string) bool {
	parsed, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	// Block AWS/GCP/Azure cloud metadata IPs
	if parsed.Hostname() == "169.254.169.254" || parsed.Hostname() == "169.254.169.253" || parsed.Hostname() == "metadata.google.internal" {
		return false
	}
	return true
}

func generateToken() (string, string) {
	b := make([]byte, 16)
	rand.Read(b)
	token := "trfk_" + hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(token))
	hashString := hex.EncodeToString(hash[:])
	return token, hashString
}

func getTokens(c *gin.Context) {
	tokens, err := db.GetAPITokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tokens == nil {
		tokens = []db.APIToken{}
	}
	c.JSON(http.StatusOK, tokens)
}

func createToken(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	token, hash := generateToken()
	if err := db.AddAPIToken(req.Name, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Return the raw token only once
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func deleteToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := db.DeleteAPIToken(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
