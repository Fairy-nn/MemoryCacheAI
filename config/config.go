package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port    string
	GinMode string

	// Upstash Redis
	UpstashRedisURL   string
	UpstashRedisToken string

	// Upstash Vector
	UpstashVectorURL   string
	UpstashVectorToken string

	// Upstash QStash
	QStashURL   string
	QStashToken string

	// Embedding Services
	EmbeddingProvider string // "jina" or "openai"

	// Jina AI
	JinaAPIKey string

	// OpenAI
	OpenAIAPIKey         string
	OpenAIEmbeddingModel string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		UpstashRedisURL:   getEnv("UPSTASH_REDIS_URL", ""),
		UpstashRedisToken: getEnv("UPSTASH_REDIS_TOKEN", ""),

		UpstashVectorURL:   getEnv("UPSTASH_VECTOR_URL", ""),
		UpstashVectorToken: getEnv("UPSTASH_VECTOR_TOKEN", ""),

		QStashURL:   getEnv("QSTASH_URL", "https://qstash.upstash.io"),
		QStashToken: getEnv("QSTASH_TOKEN", ""),

		EmbeddingProvider: getEnv("EMBEDDING_PROVIDER", "jina"),

		JinaAPIKey: getEnv("JINA_API_KEY", ""),

		OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
		OpenAIEmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
	}

	// Validate required configs
	if AppConfig.UpstashRedisURL == "" || AppConfig.UpstashRedisToken == "" {
		log.Fatal("Upstash Redis configuration is required")
	}
	if AppConfig.UpstashVectorURL == "" || AppConfig.UpstashVectorToken == "" {
		log.Fatal("Upstash Vector configuration is required")
	}

	// Validate embedding provider configuration
	switch AppConfig.EmbeddingProvider {
	case "jina":
		if AppConfig.JinaAPIKey == "" {
			log.Fatal("Jina API key is required when using Jina provider")
		}
	case "openai":
		if AppConfig.OpenAIAPIKey == "" {
			log.Fatal("OpenAI API key is required when using OpenAI provider")
		}
	default:
		log.Fatal("Invalid embedding provider. Must be 'jina' or 'openai'")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEmbeddingDimensions returns the expected dimensions for the current embedding provider
func GetEmbeddingDimensions() int {
	switch AppConfig.EmbeddingProvider {
	case "jina":
		return 1024 // Jina v3 dimensions
	case "openai":
		switch AppConfig.OpenAIEmbeddingModel {
		case "text-embedding-3-small":
			return 1536
		case "text-embedding-3-large":
			return 3072
		case "text-embedding-ada-002":
			return 1536
		default:
			return 1536 // default for OpenAI
		}
	default:
		return 1024 // default fallback
	}
}
