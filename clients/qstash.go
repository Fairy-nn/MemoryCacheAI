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

type QStashClient struct {
	url    string
	token  string
	client *http.Client
}

type PublishRequest struct {
	URL             string            `json:"url"`
	Body            string            `json:"body,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Delay           int               `json:"delay,omitempty"`           // Delay in seconds
	NotBefore       int64             `json:"notBefore,omitempty"`       // Unix timestamp
	Retries         int               `json:"retries,omitempty"`         // Number of retries
	Callback        string            `json:"callback,omitempty"`        // Callback URL
	FailureCallback string            `json:"failureCallback,omitempty"` // Failure callback URL
}

type PublishResponse struct {
	MessageID string `json:"messageId"`
}

type ScheduleRequest struct {
	Destination string            `json:"destination"`
	Body        string            `json:"body,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Cron        string            `json:"cron,omitempty"`    // Cron expression
	Delay       int               `json:"delay,omitempty"`   // Delay in seconds
	Retries     int               `json:"retries,omitempty"` // Number of retries
}

type ScheduleResponse struct {
	ScheduleID string `json:"scheduleId"`
}

func NewQStashClient() *QStashClient {
	return &QStashClient{
		url:   config.AppConfig.QStashURL,
		token: config.AppConfig.QStashToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (q *QStashClient) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, q.url+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+q.token)

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("QStash request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (q *QStashClient) PublishCleanupTask(callbackURL string, task models.CleanupTask, delay int) (string, error) {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cleanup task: %w", err)
	}

	request := PublishRequest{
		URL:  callbackURL,
		Body: string(taskJSON),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Delay:   delay,
		Retries: 3,
	}

	respBody, err := q.makeRequest("POST", "/v2/publish", request)
	if err != nil {
		return "", fmt.Errorf("failed to publish cleanup task: %w", err)
	}

	var response PublishResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal publish response: %w", err)
	}

	return response.MessageID, nil
}

func (q *QStashClient) ScheduleCleanupTask(callbackURL string, cronExpression string) (string, error) {
	task := models.CleanupTask{
		TaskType:  "cleanup_expired_memories",
		Timestamp: time.Now(),
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cleanup task: %w", err)
	}

	request := ScheduleRequest{
		Destination: callbackURL,
		Body:        string(taskJSON),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Cron:    cronExpression,
		Retries: 3,
	}

	respBody, err := q.makeRequest("POST", "/v2/schedules", request)
	if err != nil {
		return "", fmt.Errorf("failed to schedule cleanup task: %w", err)
	}

	var response ScheduleResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal schedule response: %w", err)
	}

	return response.ScheduleID, nil
}

func (q *QStashClient) PublishDelayedMemoryCleanup(callbackURL string, userID string, delaySeconds int) (string, error) {
	task := models.CleanupTask{
		TaskType:  "cleanup_user_memories",
		UserID:    userID,
		Timestamp: time.Now(),
		TTL:       int64(delaySeconds),
	}

	return q.PublishCleanupTask(callbackURL, task, delaySeconds)
}

func (q *QStashClient) PublishSessionCleanup(callbackURL string, sessionID string, delaySeconds int) (string, error) {
	task := models.CleanupTask{
		TaskType:  "cleanup_session",
		UserID:    sessionID, // Reusing UserID field for session ID
		Timestamp: time.Now(),
		TTL:       int64(delaySeconds),
	}

	return q.PublishCleanupTask(callbackURL, task, delaySeconds)
}

func (q *QStashClient) CancelSchedule(scheduleID string) error {
	_, err := q.makeRequest("DELETE", "/v2/schedules/"+scheduleID, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel schedule: %w", err)
	}

	return nil
}

func (q *QStashClient) GetSchedules() ([]map[string]interface{}, error) {
	respBody, err := q.makeRequest("GET", "/v2/schedules", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	var schedules []map[string]interface{}
	if err := json.Unmarshal(respBody, &schedules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schedules response: %w", err)
	}

	return schedules, nil
}

func (q *QStashClient) GetMessages() ([]map[string]interface{}, error) {
	respBody, err := q.makeRequest("GET", "/v2/messages", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	var messages []map[string]interface{}
	if err := json.Unmarshal(respBody, &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages response: %w", err)
	}

	return messages, nil
}
