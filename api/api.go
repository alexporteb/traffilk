package api

import (
	"net/http"
	"net/url"
	"strconv"
	"traffilk/db"
	"traffilk/scheduler"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Unprotected routes
	r.StaticFile("/login", "./ui/login.html")
	r.StaticFile("/favicon.ico", "./ui/logo.png")
	r.POST("/api/traffilk/login", LoginHandler)
	r.POST("/api/traffilk/logout", LogoutHandler)

	// UI protection middleware
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/ui" || (len(path) >= 4 && path[:4] == "/ui/") {
			// Allow public access to static assets required for login page
			if (len(path) >= 11 && path[:11] == "/ui/assets/") || path == "/ui/logo.png" {
				c.Next()
				return
			}
			
			_, err := c.Cookie("token")
			if err != nil {
				c.Header("Location", "../login")
				c.AbortWithStatus(http.StatusFound)
				return
			}
		}
		c.Next()
	})

	// Serve static files from UI folder
	r.Static("/ui", "./ui")
	r.Any("/", func(c *gin.Context) {
		c.Header("Location", "ui/")
		c.AbortWithStatus(http.StatusFound)
	})

	api := r.Group("/api/traffilk")
	api.Use(AuthMiddleware())
	{
		api.GET("/nodes", getNodes)
		api.POST("/nodes", addNode)
		api.PUT("/nodes/:id", updateNode)
		api.DELETE("/nodes/:id", deleteNode)
		api.GET("/nodes/:id/traffic", getNodeTraffic)
	}

	return r
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

	err = db.UpdateNode(id, node.Name, node.URL)
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
