package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "equity-calculation-api",
			"features": "postflop-only, adaptive-precision",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Equity calculation endpoint
		v1.POST("/equity", CalculateEquity)
		
		// Stud game endpoints
		stud := v1.Group("/stud")
		{
			stud.POST("/equity", CalculateStudEquity)
			stud.POST("/range-equity", CalculateStudRangeEquity)
		}
	}

	// CORS middleware for browser requests
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}