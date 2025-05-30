package models

import "time"

// SessionData represents short-term memory stored in Redis
type SessionData struct {
	UserID       string                 `json:"user_id"`
	SessionID    string                 `json:"session_id"`
	Messages     []Message              `json:"messages"`
	Context      map[string]interface{} `json:"context"`
	LastActivity time.Time              `json:"last_activity"`
	CreatedAt    time.Time              `json:"created_at"`
}

// Message represents a single conversation message
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// MemoryEntry represents long-term memory stored in Vector DB
type MemoryEntry struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Content   string                 `json:"content"`
	Embedding []float64              `json:"embedding"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	TTL       int64                  `json:"ttl"` // Time to live in seconds
}

// VectorMetadata represents metadata stored with vector embeddings
type VectorMetadata struct {
	UserID    string    `json:"user_id"`
	Source    string    `json:"source"` // "user" or "assistant"
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
	TTL       int64     `json:"ttl"`
}

// SaveMemoryRequest represents the request to save memory
type SaveMemoryRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

// QueryMemoryRequest represents the request to query memory
type QueryMemoryRequest struct {
	UserID   string  `json:"user_id" binding:"required"`
	Query    string  `json:"query" binding:"required"`
	Limit    int     `json:"limit,omitempty"`
	MinScore float64 `json:"min_score,omitempty"`
}

// QueryMemoryResponse represents the response from memory query
type QueryMemoryResponse struct {
	Results []MemoryResult `json:"results"`
	Total   int            `json:"total"`
}

// MemoryResult represents a single memory search result
type MemoryResult struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Score     float64                `json:"score"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// CleanupTask represents a cleanup task for QStash
type CleanupTask struct {
	TaskType  string    `json:"task_type"`
	UserID    string    `json:"user_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	TTL       int64     `json:"ttl"`
}
