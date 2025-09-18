package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SimpleAuth provides basic API key authentication
func SimpleAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For demo purposes, use a simple API key
		// In production, use proper JWT or OAuth
		apiKey := c.GetHeader("X-API-Key")
		
		// Allow requests without API key for demo
		// In production, make this mandatory
		if apiKey == "" {
			c.Next()
			return
		}
		
		// Simple validation - in production use proper key management
		if apiKey != "demo-api-key-12345" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// CORS middleware for frontend integration
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allow localhost for development
		if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}