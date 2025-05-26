package services

import (
	"fmt"
	"time"

	"github.com/Fairy-nn/MemoryCacheAI/clients"
	"github.com/Fairy-nn/MemoryCacheAI/config"
	"github.com/Fairy-nn/MemoryCacheAI/models"

	"github.com/google/uuid"
)

type MemoryService struct {
	redisClient     *clients.RedisClient
	vectorClient    *clients.VectorClient
	embeddingClient clients.EmbeddingClient
	qstashClient    *clients.QStashClient
}

func NewMemoryService() *MemoryService {
	return &MemoryService{
		redisClient:     clients.NewRedisClient(),
		vectorClient:    clients.NewVectorClient(),
		embeddingClient: clients.NewEmbeddingClient(),
		qstashClient:    clients.NewQStashClient(),
	}
}

// SaveMemory saves both short-term (Redis) and long-term (Vector) memory
func (m *MemoryService) SaveMemory(req models.SaveMemoryRequest) error {
	now := time.Now()
	messageID := uuid.New().String()

	// Create message for session
	message := models.Message{
		ID:        messageID,
		Role:      req.Role,
		Content:   req.Content,
		Timestamp: now,
	}

	// Save to Redis (short-term memory)
	session, err := m.redisClient.GetSession(req.SessionID)
	if err != nil {
		// Create new session if not exists
		session = &models.SessionData{
			UserID:       req.UserID,
			SessionID:    req.SessionID,
			Messages:     []models.Message{},
			Context:      make(map[string]interface{}),
			LastActivity: now,
			CreatedAt:    now,
		}
	}

	// Add message to session
	session.Messages = append(session.Messages, message)
	session.LastActivity = now

	if err := m.redisClient.SaveSession(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Generate embedding for long-term memory
	embedding, err := m.embeddingClient.GenerateEmbedding(req.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create memory entry for vector storage
	memoryEntry := &models.MemoryEntry{
		ID:        messageID,
		UserID:    req.UserID,
		Content:   req.Content,
		Embedding: embedding,
		Metadata: map[string]interface{}{
			"session_id": req.SessionID,
			"role":       req.Role,
		},
		Timestamp: now,
		TTL:       30 * 24 * 60 * 60, // 30 days TTL
	}

	// Save to Vector DB (long-term memory)
	if err := m.vectorClient.UpsertMemory(memoryEntry); err != nil {
		return fmt.Errorf("failed to save vector memory: %w", err)
	}

	return nil
}

// QueryMemory searches for relevant memories using semantic similarity
func (m *MemoryService) QueryMemory(req models.QueryMemoryRequest) (*models.QueryMemoryResponse, error) {
	fmt.Printf("🔍 QueryMemory: UserID=%s, Query=%s, Limit=%d, MinScore=%f\n", req.UserID, req.Query, req.Limit, req.MinScore)

	// Generate embedding for query
	queryEmbedding, err := m.embeddingClient.GenerateEmbedding(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	fmt.Printf("📊 Generated embedding with %d dimensions\n", len(queryEmbedding))

	// Set default values
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	minScore := req.MinScore
	if minScore <= 0 {
		minScore = 0.5 // Lower default similarity threshold for better recall
	}
	fmt.Printf("⚙️ Using limit=%d, minScore=%f\n", limit, minScore)

	// Query vector database
	results, err := m.vectorClient.QueryMemories(req.UserID, queryEmbedding, limit, minScore)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	fmt.Printf("📋 Vector query returned %d results\n", len(results))

	response := &models.QueryMemoryResponse{
		Results: results,
		Total:   len(results),
	}

	return response, nil
}

// GetSession retrieves current session data
func (m *MemoryService) GetSession(sessionID string) (*models.SessionData, error) {
	session, err := m.redisClient.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Update last activity
	if err := m.redisClient.UpdateSessionActivity(sessionID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to update session activity: %v\n", err)
	}

	return session, nil
}

// GetUserSessions retrieves all sessions for a user
func (m *MemoryService) GetUserSessions(userID string) ([]string, error) {
	return m.redisClient.GetUserSessions(userID)
}

// DeleteSession removes a session and optionally its memories
func (m *MemoryService) DeleteSession(sessionID string, deleteMemories bool) error {
	// Get session first to get user ID (if needed for memory deletion)
	if deleteMemories {
		_, err := m.redisClient.GetSession(sessionID)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		// This is a simplified approach - in production, you might want to
		// query by session_id metadata and delete specific memories
		fmt.Printf("Note: Memory deletion by session not implemented in this example\n")
	}

	// Delete from Redis
	if err := m.redisClient.DeleteSession(sessionID); err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	return nil
}

// SetSessionContext updates session context
func (m *MemoryService) SetSessionContext(sessionID string, context map[string]interface{}) error {
	return m.redisClient.SetSessionContext(sessionID, context)
}

// GetMemoryStats returns statistics about stored memories
func (m *MemoryService) GetMemoryStats() (map[string]interface{}, error) {
	vectorStats, err := m.vectorClient.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get vector stats: %w", err)
	}

	stats := map[string]interface{}{
		"vector_db": vectorStats,
		"timestamp": time.Now(),
	}

	return stats, nil
}

// CleanupExpiredMemories removes expired memories from vector database
func (m *MemoryService) CleanupExpiredMemories() error {
	return m.vectorClient.DeleteExpiredMemories()
}

// CleanupUserMemories removes all memories for a specific user
func (m *MemoryService) CleanupUserMemories(userID string) error {
	// Delete from vector database
	if err := m.vectorClient.DeleteUserMemories(userID); err != nil {
		return fmt.Errorf("failed to delete user memories from vector DB: %w", err)
	}

	// Delete user sessions from Redis
	sessions, err := m.redisClient.GetUserSessions(userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, sessionID := range sessions {
		if err := m.redisClient.DeleteSession(sessionID); err != nil {
			fmt.Printf("Warning: failed to delete session %s: %v\n", sessionID, err)
		}
	}

	return nil
}

// ScheduleCleanup schedules periodic cleanup tasks
func (m *MemoryService) ScheduleCleanup(callbackURL string) (string, error) {
	// Schedule daily cleanup at 2 AM
	cronExpression := "0 2 * * *"

	scheduleID, err := m.qstashClient.ScheduleCleanupTask(callbackURL, cronExpression)
	if err != nil {
		return "", fmt.Errorf("failed to schedule cleanup: %w", err)
	}

	return scheduleID, nil
}

// ScheduleDelayedUserCleanup schedules cleanup for a specific user after delay
func (m *MemoryService) ScheduleDelayedUserCleanup(callbackURL string, userID string, delaySeconds int) (string, error) {
	messageID, err := m.qstashClient.PublishDelayedMemoryCleanup(callbackURL, userID, delaySeconds)
	if err != nil {
		return "", fmt.Errorf("failed to schedule user cleanup: %w", err)
	}

	return messageID, nil
}

// GetRecentMemories retrieves recent memories for a user
func (m *MemoryService) GetRecentMemories(userID string, limit int) ([]models.MemoryResult, error) {
	if limit <= 0 {
		limit = 20
	}

	// Use a generic query to get recent memories
	// This is a simplified approach - you might want to implement time-based filtering
	queryReq := models.QueryMemoryRequest{
		UserID:   userID,
		Query:    "recent conversation", // Generic query
		Limit:    limit,
		MinScore: 0.1, // Lower threshold for recent memories
	}

	response, err := m.QueryMemory(queryReq)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

// SearchMemoriesByKeyword searches memories using keyword matching
func (m *MemoryService) SearchMemoriesByKeyword(userID string, keyword string, limit int) ([]models.MemoryResult, error) {
	queryReq := models.QueryMemoryRequest{
		UserID:   userID,
		Query:    keyword,
		Limit:    limit,
		MinScore: 0.6, // Higher threshold for keyword search
	}

	response, err := m.QueryMemory(queryReq)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

// GetEmbeddingInfo returns information about the current embedding provider
func (m *MemoryService) GetEmbeddingInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{
		"provider":   string(m.embeddingClient.GetProvider()),
		"dimensions": m.embeddingClient.GetDimensions(),
		"timestamp":  time.Now(),
	}

	// Add provider-specific information
	switch m.embeddingClient.GetProvider() {
	case "jina":
		info["api_url"] = "https://api.jina.ai/v1"
		info["model"] = "jina-embeddings-v3"
		info["features"] = []string{"multilingual", "high-performance", "normalized"}
	case "openai":
		info["api_url"] = "https://api.openai.com/v1"
		info["model"] = config.AppConfig.OpenAIEmbeddingModel
		info["features"] = []string{"high-quality", "widely-supported", "english-optimized"}
	}

	return info, nil
}

// DeleteMemory removes a specific memory by ID for a user
func (m *MemoryService) DeleteMemory(memoryID string, userID string) error {
	// First verify that the memory belongs to the specified user
	// We'll use the QueryMemory method which handles embedding generation

	// Query the memory to verify ownership
	request := models.QueryMemoryRequest{
		UserID:   userID,
		Query:    "verify memory ownership", // Just a placeholder
		Limit:    100,
		MinScore: 0.0, // Get all memories regardless of score
	}

	response, err := m.QueryMemory(request)
	if err != nil {
		return fmt.Errorf("failed to verify memory ownership: %w", err)
	}

	// Check if the memory belongs to the user
	memoryFound := false
	for _, result := range response.Results {
		if id, ok := result.Metadata["id"].(string); ok && id == memoryID {
			memoryFound = true
			break
		}
	}

	if !memoryFound {
		return fmt.Errorf("memory not found or does not belong to the specified user")
	}

	// Delete the memory
	if err := m.vectorClient.DeleteMemory(memoryID); err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	return nil
}
