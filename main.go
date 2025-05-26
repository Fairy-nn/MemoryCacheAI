package main

import (
	"log"
	"net/http"

	"aimemohub/config"
	"aimemohub/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Set Gin mode
	gin.SetMode(config.AppConfig.GinMode)

	// Create Gin router
	router := gin.Default()

	// Add CORS middleware
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

	// Initialize handlers
	memoryHandler := handlers.NewMemoryHandler()
	webhookHandler := handlers.NewWebhookHandler()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "aimemohub",
			"version": "1.0.0",
		})
	})

	// API info endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     "MemoryCacheAI",
			"description": "AI Assistant Memory Cache Service",
			"version":     "1.0.0",
			"endpoints": map[string]interface{}{
				"memory": map[string]string{
					"save":           "POST /memory/save",
					"query":          "POST /memory/query",
					"stats":          "GET /memory/stats",
					"embedding_info": "GET /memory/embedding-info",
				},
				"sessions": map[string]string{
					"get":     "GET /session/:id",
					"delete":  "DELETE /session/:id",
					"context": "PUT /session/:id/context",
				},
				"users": map[string]string{
					"sessions":        "GET /user/:id/sessions",
					"recent_memories": "GET /user/:id/memories/recent",
					"search_memories": "GET /user/:id/memories/search?q=keyword",
					"cleanup":         "DELETE /user/:id/memories",
				},
				"webhooks": map[string]string{
					"cleanup":               "POST /webhook/cleanup",
					"schedule_cleanup":      "POST /webhook/schedule-cleanup",
					"schedule_user_cleanup": "POST /webhook/schedule-user-cleanup",
					"test":                  "POST /webhook/test",
					"info":                  "GET /webhook/info",
				},
			},
		})
	})

	// Memory routes
	memoryRoutes := router.Group("/memory")
	{
		memoryRoutes.POST("/save", memoryHandler.SaveMemory)
		memoryRoutes.POST("/query", memoryHandler.QueryMemory)
		memoryRoutes.GET("/stats", memoryHandler.GetMemoryStats)
		memoryRoutes.GET("/embedding-info", memoryHandler.GetEmbeddingInfo)
	}

	// Session routes
	sessionRoutes := router.Group("/session")
	{
		sessionRoutes.GET("/:id", memoryHandler.GetSession)
		sessionRoutes.DELETE("/:id", memoryHandler.DeleteSession)
		sessionRoutes.PUT("/:id/context", memoryHandler.SetSessionContext)
	}

	// User routes
	userRoutes := router.Group("/user")
	{
		userRoutes.GET("/:id/sessions", memoryHandler.GetUserSessions)
		userRoutes.GET("/:id/memories/recent", memoryHandler.GetRecentMemories)
		userRoutes.GET("/:id/memories/search", memoryHandler.SearchMemories)
		userRoutes.DELETE("/:id/memories", memoryHandler.CleanupUserMemories)
	}

	// Webhook routes
	webhookRoutes := router.Group("/webhook")
	{
		webhookRoutes.POST("/cleanup", webhookHandler.HandleCleanupWebhook)
		webhookRoutes.POST("/schedule-cleanup", webhookHandler.ScheduleCleanup)
		webhookRoutes.POST("/schedule-user-cleanup", webhookHandler.ScheduleUserCleanup)
		webhookRoutes.POST("/test", webhookHandler.TestWebhook)
		webhookRoutes.GET("/info", webhookHandler.GetWebhookInfo)
		webhookRoutes.GET("/validate", webhookHandler.ValidateWebhook)
	}

	// Start server
	port := ":" + config.AppConfig.Port
	log.Printf("🚀 MemoryCacheAI starting on port %s", config.AppConfig.Port)
	log.Printf("📚 Memory endpoints: /memory/save, /memory/query")
	log.Printf("🔗 Session endpoints: /session/:id")
	log.Printf("👤 User endpoints: /user/:id/sessions, /user/:id/memories/*")
	log.Printf("🪝 Webhook endpoints: /webhook/*")
	log.Printf("🏥 Health check: /health")

	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
