package handlers

import (
	"net/http"

	"github.com/Fairy-nn/MemoryCacheAI/models"
	"github.com/Fairy-nn/MemoryCacheAI/services"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	memoryService *services.MemoryService
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		memoryService: services.NewMemoryService(),
	}
}

// HandleCleanupWebhook handles QStash cleanup webhooks
func (h *WebhookHandler) HandleCleanupWebhook(c *gin.Context) {
	// Parse the cleanup task from request body
	var task models.CleanupTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid task format",
			"details": err.Error(),
		})
		return
	}

	// Process the cleanup task based on type
	switch task.TaskType {
	case "cleanup_expired_memories":
		if err := h.memoryService.CleanupExpiredMemories(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to cleanup expired memories",
				"details": err.Error(),
			})
			return
		}

	case "cleanup_user_memories":
		if task.UserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User ID is required for user memory cleanup",
			})
			return
		}

		if err := h.memoryService.CleanupUserMemories(task.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to cleanup user memories",
				"details": err.Error(),
			})
			return
		}

	case "cleanup_session":
		if task.UserID == "" { // UserID field is reused for session ID
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Session ID is required for session cleanup",
			})
			return
		}

		if err := h.memoryService.DeleteSession(task.UserID, false); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to cleanup session",
				"details": err.Error(),
			})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unknown task type: " + task.TaskType,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cleanup task completed successfully",
		"task_type": task.TaskType,
		"timestamp": task.Timestamp,
	})
}

// ScheduleCleanup handles POST /webhook/schedule-cleanup
func (h *WebhookHandler) ScheduleCleanup(c *gin.Context) {
	type ScheduleRequest struct {
		CallbackURL string `json:"callback_url" binding:"required"`
	}

	var req ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	scheduleID, err := h.memoryService.ScheduleCleanup(req.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to schedule cleanup",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Cleanup scheduled successfully",
		"schedule_id":  scheduleID,
		"callback_url": req.CallbackURL,
	})
}

// ScheduleUserCleanup handles POST /webhook/schedule-user-cleanup
func (h *WebhookHandler) ScheduleUserCleanup(c *gin.Context) {
	type ScheduleUserCleanupRequest struct {
		CallbackURL  string `json:"callback_url" binding:"required"`
		UserID       string `json:"user_id" binding:"required"`
		DelaySeconds int    `json:"delay_seconds"`
	}

	var req ScheduleUserCleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Default delay of 1 hour if not specified
	if req.DelaySeconds <= 0 {
		req.DelaySeconds = 3600
	}

	messageID, err := h.memoryService.ScheduleDelayedUserCleanup(req.CallbackURL, req.UserID, req.DelaySeconds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to schedule user cleanup",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "User cleanup scheduled successfully",
		"message_id":    messageID,
		"user_id":       req.UserID,
		"delay_seconds": req.DelaySeconds,
	})
}

// GetWebhookInfo handles GET /webhook/info
func (h *WebhookHandler) GetWebhookInfo(c *gin.Context) {
	info := gin.H{
		"endpoints": map[string]string{
			"cleanup":               "POST /webhook/cleanup - Handle cleanup tasks from QStash",
			"schedule_cleanup":      "POST /webhook/schedule-cleanup - Schedule periodic cleanup",
			"schedule_user_cleanup": "POST /webhook/schedule-user-cleanup - Schedule user-specific cleanup",
		},
		"supported_tasks": []string{
			"cleanup_expired_memories",
			"cleanup_user_memories",
			"cleanup_session",
		},
		"example_payload": models.CleanupTask{
			TaskType: "cleanup_expired_memories",
			UserID:   "user123",
			TTL:      3600,
		},
	}

	c.JSON(http.StatusOK, info)
}

// ValidateWebhook handles webhook signature validation (if needed)
func (h *WebhookHandler) ValidateWebhook(c *gin.Context) {
	// Get QStash signature from headers
	signature := c.GetHeader("Upstash-Signature")

	// In a production environment, you would validate the signature here
	// For now, we'll just log it
	if signature != "" {
		// Log the signature for debugging
		// In production, implement proper signature validation
	}

	// For this example, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook validation endpoint",
		"headers": map[string]string{
			"Upstash-Signature": signature,
		},
	})
}

// TestWebhook handles POST /webhook/test - for testing webhook functionality
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	// Parse any JSON payload
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		// If no JSON, that's fine for a test endpoint
		payload = map[string]interface{}{
			"message": "Test webhook called without JSON payload",
		}
	}

	response := gin.H{
		"message":   "Test webhook received successfully",
		"timestamp": "2024-01-01T00:00:00Z", // You might want to use time.Now()
		"payload":   payload,
		"headers": map[string]string{
			"Content-Type":      c.GetHeader("Content-Type"),
			"User-Agent":        c.GetHeader("User-Agent"),
			"Upstash-Signature": c.GetHeader("Upstash-Signature"),
		},
	}

	c.JSON(http.StatusOK, response)
}
