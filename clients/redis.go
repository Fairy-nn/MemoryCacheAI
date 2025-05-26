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

type RedisClient struct {
	url    string
	token  string
	client *http.Client
}

type RedisCommand []interface{}

type RedisResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

func NewRedisClient() *RedisClient {
	return &RedisClient{
		url:   config.AppConfig.UpstashRedisURL,
		token: config.AppConfig.UpstashRedisToken,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *RedisClient) executeCommand(cmd RedisCommand) (*RedisResponse, error) {
	jsonData, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	// Ensure URL has the correct path for Upstash Redis REST API
	url := r.url
	if url[len(url)-1] != '/' {
		url += "/"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.token)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Redis request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response RedisResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("Redis error: %s", response.Error)
	}

	return &response, nil
}

func (r *RedisClient) SaveSession(sessionData *models.SessionData) error {
	key := fmt.Sprintf("session:%s", sessionData.SessionID)

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Set with TTL of 24 hours
	cmd := RedisCommand{"SETEX", key, 86400, string(jsonData)}

	_, err = r.executeCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Also save user session mapping
	userKey := fmt.Sprintf("user_sessions:%s", sessionData.UserID)
	cmd = RedisCommand{"SADD", userKey, sessionData.SessionID}

	_, err = r.executeCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to save user session mapping: %w", err)
	}

	// Set TTL for user sessions set
	cmd = RedisCommand{"EXPIRE", userKey, 86400}

	_, err = r.executeCommand(cmd)
	return err
}

func (r *RedisClient) GetSession(sessionID string) (*models.SessionData, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	cmd := RedisCommand{"GET", key}

	resp, err := r.executeCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if resp.Result == nil {
		return nil, fmt.Errorf("session not found")
	}

	jsonStr, ok := resp.Result.(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data format")
	}

	var sessionData models.SessionData
	if err := json.Unmarshal([]byte(jsonStr), &sessionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &sessionData, nil
}

func (r *RedisClient) GetUserSessions(userID string) ([]string, error) {
	key := fmt.Sprintf("user_sessions:%s", userID)

	cmd := RedisCommand{"SMEMBERS", key}

	resp, err := r.executeCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	if resp.Result == nil {
		return []string{}, nil
	}

	// Convert interface{} slice to string slice
	resultSlice, ok := resp.Result.([]interface{})
	if !ok {
		return []string{}, nil
	}

	sessions := make([]string, len(resultSlice))
	for i, v := range resultSlice {
		if str, ok := v.(string); ok {
			sessions[i] = str
		}
	}

	return sessions, nil
}

func (r *RedisClient) DeleteSession(sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)

	cmd := RedisCommand{"DEL", key}

	_, err := r.executeCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (r *RedisClient) UpdateSessionActivity(sessionID string) error {
	// Get current session
	session, err := r.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Update last activity
	session.LastActivity = time.Now()

	// Save back
	return r.SaveSession(session)
}

func (r *RedisClient) AddMessageToSession(sessionID string, message models.Message) error {
	session, err := r.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.Messages = append(session.Messages, message)
	session.LastActivity = time.Now()

	return r.SaveSession(session)
}

func (r *RedisClient) SetSessionContext(sessionID string, context map[string]interface{}) error {
	session, err := r.GetSession(sessionID)
	if err != nil {
		return err
	}

	if session.Context == nil {
		session.Context = make(map[string]interface{})
	}

	for k, v := range context {
		session.Context[k] = v
	}

	session.LastActivity = time.Now()

	return r.SaveSession(session)
}
