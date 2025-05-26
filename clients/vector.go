package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Fairy-nn/MemoryCacheAI/config"
	"github.com/Fairy-nn/MemoryCacheAI/models"
)

type VectorClient struct {
	url        string
	token      string
	client     *http.Client
	dimensions int // cached dimensions
}

type UpsertRequest struct {
	ID       string                 `json:"id"`
	Vector   []float64              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type QueryRequest struct {
	Vector          []float64 `json:"vector"`
	TopK            int       `json:"topK"`
	IncludeMetadata bool      `json:"includeMetadata"`
	IncludeVectors  bool      `json:"includeVectors"`
	Filter          string    `json:"filter,omitempty"`
}

type QueryResponse struct {
	Result []QueryMatch `json:"result"`
}

type QueryMatch struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Vector   []float64              `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type DeleteRequest struct {
	ID string `json:"id"`
}

func NewVectorClient() *VectorClient {
	return &VectorClient{
		url:   config.AppConfig.UpstashVectorURL,
		token: config.AppConfig.UpstashVectorToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (v *VectorClient) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, v.url+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v.token)

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (v *VectorClient) UpsertMemory(memory *models.MemoryEntry) error {
	metadata := map[string]interface{}{
		"user_id":   memory.UserID,
		"content":   memory.Content,
		"timestamp": memory.Timestamp.Unix(),
		"ttl":       memory.TTL,
	}

	// Add custom metadata
	for k, val := range memory.Metadata {
		metadata[k] = val
	}

	request := UpsertRequest{
		ID:       memory.ID,
		Vector:   memory.Embedding,
		Metadata: metadata,
	}

	_, err := v.makeRequest("POST", "/upsert", request)
	if err != nil {
		return fmt.Errorf("failed to upsert memory: %w", err)
	}

	return nil
}

func (v *VectorClient) QueryMemories(userID string, queryVector []float64, limit int, minScore float64) ([]models.MemoryResult, error) {
	if limit <= 0 {
		limit = 10
	}

	request := QueryRequest{
		Vector:          queryVector,
		TopK:            limit,
		IncludeMetadata: true,
		IncludeVectors:  false,
		Filter:          fmt.Sprintf("user_id = '%s'", userID),
	}
	fmt.Printf("🔍 Vector query: UserID=%s, VectorDim=%d, TopK=%d, Filter=%s\n", userID, len(queryVector), limit, request.Filter)

	respBody, err := v.makeRequest("POST", "/query", request)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	fmt.Printf("📡 Vector API response: %s\n", string(respBody))

	var response QueryResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal query response: %w", err)
	}
	fmt.Printf("📊 Raw matches from vector DB: %d\n", len(response.Result))

	results := make([]models.MemoryResult, 0, len(response.Result))
	for i, match := range response.Result {
		fmt.Printf("  Match %d: ID=%s, Score=%f\n", i, match.ID, match.Score)
		if match.Score < minScore {
			fmt.Printf("    ❌ Filtered out (score %f < minScore %f)\n", match.Score, minScore)
			continue
		}

		result := models.MemoryResult{
			Score:    match.Score,
			Metadata: match.Metadata,
		}

		// Add memory ID to metadata
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["id"] = match.ID

		// Extract content from metadata
		if content, ok := match.Metadata["content"].(string); ok {
			result.Content = content
		}

		// Extract timestamp from metadata
		if timestampFloat, ok := match.Metadata["timestamp"].(float64); ok {
			result.Timestamp = time.Unix(int64(timestampFloat), 0)
		}

		results = append(results, result)
		fmt.Printf("    ✅ Added to results\n")
	}
	fmt.Printf("📋 Final filtered results: %d\n", len(results))

	return results, nil
}

func (v *VectorClient) DeleteMemory(id string) error {
	request := DeleteRequest{
		ID: id,
	}

	_, err := v.makeRequest("POST", "/delete", request)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	return nil
}

func (v *VectorClient) DeleteUserMemories(userID string) error {
	// Get vector dimensions dynamically
	dimensions, err := v.GetDimensions()
	if err != nil {
		// Fallback to configured dimensions if we can't get them from the database
		dimensions = config.GetEmbeddingDimensions()
		fmt.Printf("Warning: Could not get dimensions from database, using configured dimensions %d: %v\n", dimensions, err)
	}

	// First query all memories for the user
	queryRequest := QueryRequest{
		Vector:          make([]float64, dimensions), // Dynamic vector dimensions
		TopK:            1000,                        // Large number to get all
		IncludeMetadata: true,
		IncludeVectors:  false,
		Filter:          fmt.Sprintf("user_id = '%s'", userID),
	}

	respBody, err := v.makeRequest("POST", "/query", queryRequest)
	if err != nil {
		return fmt.Errorf("failed to query user memories for deletion: %w", err)
	}

	var response QueryResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal query response: %w", err)
	}

	// Delete each memory
	for _, match := range response.Result {
		if err := v.DeleteMemory(match.ID); err != nil {
			return fmt.Errorf("failed to delete memory %s: %w", match.ID, err)
		}
	}

	return nil
}

func (v *VectorClient) DeleteExpiredMemories() error {
	now := time.Now().Unix()

	// Get vector dimensions dynamically
	dimensions, err := v.GetDimensions()
	if err != nil {
		// Fallback to configured dimensions if we can't get them from the database
		dimensions = config.GetEmbeddingDimensions()
		fmt.Printf("Warning: Could not get dimensions from database, using configured dimensions %d: %v\n", dimensions, err)
	}

	// Query all memories (this is a simplified approach)
	queryRequest := QueryRequest{
		Vector:          make([]float64, dimensions), // Dynamic vector dimensions
		TopK:            10000,                       // Large number
		IncludeMetadata: true,
		IncludeVectors:  false,
	}

	respBody, err := v.makeRequest("POST", "/query", queryRequest)
	if err != nil {
		return fmt.Errorf("failed to query memories for cleanup: %w", err)
	}

	var response QueryResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal query response: %w", err)
	}

	// Check each memory for expiration
	for _, match := range response.Result {
		if timestampFloat, ok := match.Metadata["timestamp"].(float64); ok {
			if ttlFloat, ok := match.Metadata["ttl"].(float64); ok {
				expirationTime := int64(timestampFloat) + int64(ttlFloat)
				if now > expirationTime {
					if err := v.DeleteMemory(match.ID); err != nil {
						fmt.Printf("Failed to delete expired memory %s: %v\n", match.ID, err)
					}
				}
			}
		}
	}

	return nil
}

func (v *VectorClient) GetStats() (map[string]interface{}, error) {
	respBody, err := v.makeRequest("GET", "/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get vector stats: %w", err)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(respBody, &stats); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stats response: %w", err)
	}

	return stats, nil
}

// GetDimensions returns the vector dimensions from the database (with caching)
func (v *VectorClient) GetDimensions() (int, error) {
	// Return cached dimensions if available
	if v.dimensions > 0 {
		return v.dimensions, nil
	}

	// Fetch dimensions from database
	stats, err := v.GetStats()
	if err != nil {
		return 0, fmt.Errorf("failed to get vector stats: %w", err)
	}

	var dimensions int
	if result, ok := stats["result"].(map[string]interface{}); ok {
		if dim, ok := result["dimension"].(float64); ok {
			dimensions = int(dim)
		}
	}

	if dimensions == 0 {
		return 0, fmt.Errorf("could not determine vector dimensions from database")
	}

	// Cache the dimensions for future use
	v.dimensions = dimensions
	return dimensions, nil
}
