package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Fairy-nn/MemoryCacheAI/config"
)

// EmbeddingProvider represents the embedding service provider
type EmbeddingProvider string

const (
	ProviderJina   EmbeddingProvider = "jina"
	ProviderOpenAI EmbeddingProvider = "openai"
)

// EmbeddingClient interface for different embedding providers
type EmbeddingClient interface {
	GenerateEmbedding(text string) ([]float64, error)
	GenerateEmbeddings(texts []string) ([]float64, error)
	GenerateBatchEmbeddings(texts []string) ([][]float64, error)
	GetProvider() EmbeddingProvider
	GetDimensions() int
}

// UnifiedEmbeddingClient wraps different embedding providers
type UnifiedEmbeddingClient struct {
	provider EmbeddingProvider
	client   EmbeddingClient
}

// JinaClient for Jina AI embeddings
type JinaClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// OpenAIClient for OpenAI embeddings
type OpenAIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// Jina AI request/response structures
type JinaEmbeddingRequest struct {
	Input         []string `json:"input"`
	Normalized    bool     `json:"normalized"`
	EmbeddingType string   `json:"embedding_type"`
}

type JinaEmbeddingResponse struct {
	Model  string `json:"model"`
	Object string `json:"object"`
	Usage  struct {
		TotalTokens  int `json:"total_tokens"`
		PromptTokens int `json:"prompt_tokens"`
	} `json:"usage"`
	Data []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// OpenAI request/response structures
type OpenAIEmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewEmbeddingClient creates a new embedding client based on configuration
func NewEmbeddingClient() EmbeddingClient {
	provider := strings.ToLower(config.AppConfig.EmbeddingProvider)

	switch provider {
	case "openai":
		return NewOpenAIClient()
	case "jina", "":
		// Default to Jina if not specified
		return NewJinaClient()
	default:
		// Fallback to Jina
		return NewJinaClient()
	}
}

// NewUnifiedEmbeddingClient creates a unified client that can switch providers
func NewUnifiedEmbeddingClient() *UnifiedEmbeddingClient {
	client := NewEmbeddingClient()
	return &UnifiedEmbeddingClient{
		provider: client.GetProvider(),
		client:   client,
	}
}

// Jina AI Client Implementation

func NewJinaClient() *JinaClient {
	return &JinaClient{
		apiKey:  config.AppConfig.JinaAPIKey,
		baseURL: "https://api.jina.ai/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (j *JinaClient) GetProvider() EmbeddingProvider {
	return ProviderJina
}

func (j *JinaClient) GetDimensions() int {
	return 1024 // Jina v3 default dimensions
}

func (j *JinaClient) GenerateEmbedding(text string) ([]float64, error) {
	embeddings, err := j.GenerateEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	return embeddings, nil
}

func (j *JinaClient) GenerateEmbeddings(texts []string) ([]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	reqBody := JinaEmbeddingRequest{
		Input:         texts,
		Normalized:    true,
		EmbeddingType: "float",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", j.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+j.apiKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jina API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response JinaEmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Return the first embedding (for single text input)
	return response.Data[0].Embedding, nil
}

func (j *JinaClient) GenerateBatchEmbeddings(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	reqBody := JinaEmbeddingRequest{
		Input:         texts,
		Normalized:    true,
		EmbeddingType: "float",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", j.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+j.apiKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jina API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response JinaEmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	embeddings := make([][]float64, len(response.Data))
	for i, data := range response.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

// OpenAI Client Implementation

func NewOpenAIClient() *OpenAIClient {
	model := config.AppConfig.OpenAIEmbeddingModel
	if model == "" {
		model = "text-embedding-3-small" // Default model
	}

	return &OpenAIClient{
		apiKey:  config.AppConfig.OpenAIAPIKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (o *OpenAIClient) GetProvider() EmbeddingProvider {
	return ProviderOpenAI
}

func (o *OpenAIClient) GetDimensions() int {
	// Return dimensions based on model
	switch o.model {
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	case "text-embedding-ada-002":
		return 1536
	default:
		return 1536 // Default
	}
}

func (o *OpenAIClient) GenerateEmbedding(text string) ([]float64, error) {
	embeddings, err := o.GenerateEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	return embeddings, nil
}

func (o *OpenAIClient) GenerateEmbeddings(texts []string) ([]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// For single text, pass as string; for multiple, pass as array
	var input interface{}
	if len(texts) == 1 {
		input = texts[0]
	} else {
		input = texts
	}

	reqBody := OpenAIEmbeddingRequest{
		Input:          input,
		Model:          o.model,
		EncodingFormat: "float",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", o.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIEmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Return the first embedding (for single text input)
	return response.Data[0].Embedding, nil
}

func (o *OpenAIClient) GenerateBatchEmbeddings(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	reqBody := OpenAIEmbeddingRequest{
		Input:          texts,
		Model:          o.model,
		EncodingFormat: "float",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", o.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIEmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	embeddings := make([][]float64, len(response.Data))
	for i, data := range response.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

// Unified Client Methods

func (u *UnifiedEmbeddingClient) GenerateEmbedding(text string) ([]float64, error) {
	return u.client.GenerateEmbedding(text)
}

func (u *UnifiedEmbeddingClient) GenerateEmbeddings(texts []string) ([]float64, error) {
	return u.client.GenerateEmbeddings(texts)
}

func (u *UnifiedEmbeddingClient) GenerateBatchEmbeddings(texts []string) ([][]float64, error) {
	return u.client.GenerateBatchEmbeddings(texts)
}

func (u *UnifiedEmbeddingClient) GetProvider() EmbeddingProvider {
	return u.provider
}

func (u *UnifiedEmbeddingClient) GetDimensions() int {
	return u.client.GetDimensions()
}

func (u *UnifiedEmbeddingClient) SwitchProvider(provider EmbeddingProvider) error {
	switch provider {
	case ProviderJina:
		u.client = NewJinaClient()
	case ProviderOpenAI:
		u.client = NewOpenAIClient()
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
	u.provider = provider
	return nil
}

// Helper functions for backward compatibility

// NewJinaClientLegacy creates a Jina client (for backward compatibility)
func NewJinaClientLegacy() *JinaClient {
	return NewJinaClient()
}
