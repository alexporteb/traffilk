package api

import (
	"crypto/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func initJwtKey() []byte {
	key := getEnv("JWT_SECRET", "")
	if key != "" {
		return []byte(key)
	}
	b := make([]byte, 32)
	rand.Read(b)
	return b
}

var jwtKey = initJwtKey()
var adminUser = getEnv("ADMIN_USER", "admin")
var adminPass = getEnv("ADMIN_PASS", "admin")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// LoginHandler handles POST /api/login
func LoginHandler(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if creds.Username != adminUser || creds.Password != adminPass {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// Set cookie (valid for 24 hours, path / so it applies to /traffilk/ as well if stripped)
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("token", tokenString, int(24*time.Hour.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "token": tokenString})
}

// LogoutHandler handles POST /api/logout
func LogoutHandler(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// AuthMiddleware protects routes by validating the JWT cookie or Bearer token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			// Fallback to Bearer token
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Proceed to next handler
		c.Next()
	}
}
