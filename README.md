# MemoryCacheAI

<img width="968" alt="poster" src="https://github.com/user-attachments/assets/31da8491-b7be-4bce-b396-3fa4f4106ecc" />

🧠 An AI assistant memory cache service based on Go + Upstash, providing short-term and long-term memory management.

**Language**: [中文](README.zh.md) | **English**

## 🎯 Project Goals

Build an intelligent memory management system for AI assistants, providing:
- **Short-term Memory**: Use Redis to store current session state and context
- **Long-term Memory**: Use Vector DB to store semantic embeddings, supporting semantic retrieval
- **Asynchronous Cleanup**: Use QStash to implement scheduled cleanup of expired memories

## 🛠 Tech Stack

- **Go 1.20+** - Backend service
- **Gin** - HTTP framework
- **Upstash Redis** - Short-term memory storage (HTTP API)
- **Upstash Vector** - Long-term memory storage (HTTP API)
- **Upstash QStash** - Asynchronous task queue (Webhooks)
- **Jina AI / OpenAI Embeddings** - Vector generation service (supports multiple providers)

## 🚀 Quick Start

### 1. Environment Configuration

Copy environment variable template:
```bash
cp env.example .env
```

Configure your API keys:
```env
# Upstash Redis
UPSTASH_REDIS_URL=https://your-redis-url.upstash.io
UPSTASH_REDIS_TOKEN=your-redis-token

# Upstash Vector
UPSTASH_VECTOR_URL=https://your-vector-url.upstash.io
UPSTASH_VECTOR_TOKEN=your-vector-token

# Upstash QStash
QSTASH_URL=https://qstash.upstash.io
QSTASH_TOKEN=your-qstash-token

# Embedding Provider (jina or openai)
EMBEDDING_PROVIDER=jina

# Jina AI Embeddings
JINA_API_KEY=your-jina-api-key

# OpenAI Embeddings
OPENAI_API_KEY=your-openai-api-key
OPENAI_EMBEDDING_MODEL=text-embedding-3-small

# Server
PORT=8080
GIN_MODE=debug
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Start Service

```bash
go run main.go
```

The service will start at `http://localhost:8080`.

## 📚 API Documentation

### Memory Management

#### Save Memory
```http
POST /memory/save
Content-Type: application/json

{
  "user_id": "user123",
  "session_id": "session456",
  "content": "I have a cat named Orange",
  "role": "user"
}
```

#### Query Memory
```http
POST /memory/query
Content-Type: application/json

{
  "user_id": "user123",
  "query": "Do you remember my cat?",
  "limit": 10,
  "min_score": 0.7
}
```

#### Get Memory Statistics
```http
GET /memory/stats
```

#### Get Embedding Provider Information
```http
GET /memory/embedding-info
```

### Session Management

#### Get Session
```http
GET /session/{session_id}
```

#### Delete Session
```http
DELETE /session/{session_id}?delete_memories=true
```

#### Set Session Context
```http
PUT /session/{session_id}/context
Content-Type: application/json

{
  "user_name": "John Doe",
  "preferences": {
    "language": "en-US"
  }
}
```

### User Management

#### Get User Session List
```http
GET /user/{user_id}/sessions
```

#### Get Recent Memories
```http
GET /user/{user_id}/memories/recent?limit=20
```

#### Search Memories
```http
GET /user/{user_id}/memories/search?q=cat&limit=10
```

#### Cleanup User Memories
```http
DELETE /user/{user_id}/memories
```

### Webhook Endpoints

#### Handle Cleanup Tasks
```http
POST /webhook/cleanup
Content-Type: application/json

{
  "task_type": "cleanup_expired_memories",
  "user_id": "user123",
  "timestamp": "2024-01-01T00:00:00Z",
  "ttl": 3600
}
```

#### Schedule Periodic Cleanup
```http
POST /webhook/schedule-cleanup
Content-Type: application/json

{
  "callback_url": "https://your-domain.com/webhook/cleanup"
}
```

## 🧩 Example Usage Flow

### 1. Save Conversation Memory
```bash
curl -X POST http://localhost:8080/memory/save \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "session_id": "session456",
    "content": "I have a cat named Orange who loves sunbathing",
    "role": "user"
  }'
```

### 2. Query Related Memories
```bash
curl -X POST http://localhost:8080/memory/query \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "query": "Tell me about my pet",
    "limit": 5
  }'
```

### 3. Get Session Information
```bash
curl http://localhost:8080/session/session456
```

## 🏗 Project Structure

```
github.com/Fairy-nn/MemoryCacheAI/
├── clients/          # External service clients
│   ├── embedding.go  # Embedding clients (Jina AI & OpenAI)
│   ├── redis.go      # Upstash Redis client
│   ├── vector.go     # Upstash Vector client
│   └── qstash.go     # Upstash QStash client
├── config/           # Configuration management
│   └── config.go
├── handlers/         # HTTP handlers
│   ├── memory.go     # Memory-related endpoints
│   └── webhook.go    # Webhook handlers
├── models/           # Data models
│   └── memory.go
├── services/         # Business logic
│   └── memory.go     # Memory service
├── main.go           # Main program entry
├── go.mod            # Go module file
├── env.example       # Environment variable template
└── README.md         # Project documentation
```

## 🔧 Configuration Guide

### Upstash Service Configuration

1. **Redis**: For storing session data and short-term memory
   - Create Redis database: https://console.upstash.com/redis
   - Get URL and Token

2. **Vector**: For storing and retrieving semantic vectors
   - Create Vector database: https://console.upstash.com/vector
   - Choose appropriate dimensions (Jina default 1024, OpenAI varies by model)

3. **QStash**: For asynchronous task processing
   - Get QStash Token: https://console.upstash.com/qstash

### Embedding Service Configuration

#### Jina AI Configuration

1. Register Jina AI account: https://jina.ai/
2. Get API Key
3. Check usage limits and pricing

#### OpenAI Configuration

1. Register OpenAI account: https://platform.openai.com/
2. Get API Key
3. Choose appropriate embedding model:
   - `text-embedding-3-small` (1536 dimensions, cost-effective)
   - `text-embedding-3-large` (3072 dimensions, higher quality)
   - `text-embedding-ada-002` (1536 dimensions, classic model)

#### Switching Embedding Providers

1. Modify `EMBEDDING_PROVIDER` in `.env` file
2. Configure corresponding API keys
3. Restart service
4. **Note**: After switching providers, all embeddings need to be regenerated as different providers may have different vector dimensions and features

## 🧪 Testing

### Health Check
```bash
curl http://localhost:8080/health
```

### API Information
```bash
curl http://localhost:8080/
```

### Webhook Test
```bash
curl -X POST http://localhost:8080/webhook/test \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

### Embedding Provider Information
```bash
curl http://localhost:8080/memory/embedding-info
```

### Complete Functionality Test
```bash
# Run complete API test
bash examples/test_api.sh

# Run embedding provider test
bash examples/test_embedding_providers.sh
```

## 🔄 Data Flow

1. **Save Memory**:
   - User sends message → Redis (session) + Embedding service → Vector DB (long-term memory)

2. **Query Memory**:
   - User query → Embedding service (query embedding) → Vector DB (semantic search) → Return related memories

3. **Cleanup Memory**:
   - QStash scheduled task → Webhook → Clean expired data

## 🚨 Important Notes

1. **API Limits**: Pay attention to API call limits for each service
2. **Data Security**: Configure appropriate authentication and authorization for production
3. **Error Handling**: Monitor logs and handle exceptions promptly
4. **Cost Control**: Set reasonable TTL to avoid storing too much data

## 📈 Extended Features

- [ ] Add user authentication and authorization
- [ ] Implement memory importance scoring
- [ ] Support multimodal memory (images, audio)
- [ ] Add memory classification and tagging
- [ ] Implement memory compression and summarization
- [ ] Support memory export and backup

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

MIT License 
