package handlers

import (
	"net/http"
	"strconv"

	"github.com/Fairy-nn/MemoryCacheAI/models"
	"github.com/Fairy-nn/MemoryCacheAI/services"

	"github.com/gin-gonic/gin"
)

type MemoryHandler struct {
	memoryService *services.MemoryService
}

func NewMemoryHandler() *MemoryHandler {
	return &MemoryHandler{
		memoryService: services.NewMemoryService(),
	}
}

// SaveMemory handles POST /memory/save
func (h *MemoryHandler) SaveMemory(c *gin.Context) {
	var req models.SaveMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.memoryService.SaveMemory(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save memory",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Memory saved successfully",
		"user_id":    req.UserID,
		"session_id": req.SessionID,
	})
}

// QueryMemory handles POST /memory/query
func (h *MemoryHandler) QueryMemory(c *gin.Context) {
	var req models.QueryMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.memoryService.QueryMemory(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to query memory",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSession handles GET /session/:id
func (h *MemoryHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	session, err := h.memoryService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Session not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetUserSessions handles GET /user/:id/sessions
func (h *MemoryHandler) GetUserSessions(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	sessions, err := h.memoryService.GetUserSessions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user sessions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// DeleteSession handles DELETE /session/:id
func (h *MemoryHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	// Check if we should delete memories too
	deleteMemoriesStr := c.Query("delete_memories")
	deleteMemories := deleteMemoriesStr == "true"

	if err := h.memoryService.DeleteSession(sessionID, deleteMemories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Session deleted successfully",
		"session_id":       sessionID,
		"deleted_memories": deleteMemories,
	})
}

// SetSessionContext handles PUT /session/:id/context
func (h *MemoryHandler) SetSessionContext(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	var context map[string]interface{}
	if err := c.ShouldBindJSON(&context); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid context format",
			"details": err.Error(),
		})
		return
	}

	if err := h.memoryService.SetSessionContext(sessionID, context); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to set session context",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session context updated successfully",
		"session_id": sessionID,
	})
}

// GetMemoryStats handles GET /memory/stats
func (h *MemoryHandler) GetMemoryStats(c *gin.Context) {
	stats, err := h.memoryService.GetMemoryStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get memory stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentMemories handles GET /user/:id/memories/recent
func (h *MemoryHandler) GetRecentMemories(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	limitStr := c.Query("limit")
	limit := 20 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	memories, err := h.memoryService.GetRecentMemories(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get recent memories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"memories": memories,
		"total":    len(memories),
	})
}

// SearchMemories handles GET /user/:id/memories/search
func (h *MemoryHandler) SearchMemories(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	limitStr := c.Query("limit")
	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	memories, err := h.memoryService.SearchMemoriesByKeyword(userID, keyword, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to search memories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"query":    keyword,
		"memories": memories,
		"total":    len(memories),
	})
}

// CleanupUserMemories handles DELETE /user/:id/memories
func (h *MemoryHandler) CleanupUserMemories(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	if err := h.memoryService.CleanupUserMemories(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to cleanup user memories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User memories cleaned up successfully",
		"user_id": userID,
	})
}

// GetEmbeddingInfo handles GET /memory/embedding-info
func (h *MemoryHandler) GetEmbeddingInfo(c *gin.Context) {
	info, err := h.memoryService.GetEmbeddingInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get embedding info",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// DeleteMemory handles DELETE /memory/:id
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	memoryID := c.Param("id")
	if memoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Memory ID is required",
		})
		return
	}

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	if err := h.memoryService.DeleteMemory(memoryID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete memory",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Memory deleted successfully",
		"memory_id": memoryID,
		"user_id":   userID,
	})
}
