package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// Serve static files for debug interface
	router.Static("/static", "./static")
	router.StaticFile("/debug", "./static/debug.html")
	
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

}