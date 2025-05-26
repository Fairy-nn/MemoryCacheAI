# MemoryCacheAI

🧠 一个基于 Go + Upstash 的 AI 助理记忆缓存服务，提供短期和长期记忆管理功能。

**Language**: **Chinese** | [English](README.en.md)

## 🎯 项目目标

构建一个智能的记忆管理系统，为 AI 助理提供：
- **短期记忆**：使用 Redis 存储当前会话状态和上下文
- **长期记忆**：使用 Vector DB 存储语义 embedding，支持语义检索
- **异步清理**：使用 QStash 实现定时清理过期记忆

## 🛠 技术栈

- **Go 1.20+** - 后端服务
- **Gin** - HTTP 框架
- **Upstash Redis** - 短期记忆存储 (HTTP API)
- **Upstash Vector** - 长期记忆存储 (HTTP API)
- **Upstash QStash** - 异步任务队列 (Webhooks)
- **Jina AI / OpenAI Embeddings** - 向量生成服务（支持多种提供商）

## 🚀 快速开始

### 1. 环境配置

复制环境变量模板：
```bash
cp env.example .env
```

配置你的 API 密钥：
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

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## 📚 API 文档

### 记忆管理

#### 保存记忆
```http
POST /memory/save
Content-Type: application/json

{
  "user_id": "user123",
  "session_id": "session456",
  "content": "我家猫叫小橘子",
  "role": "user"
}
```

#### 查询记忆
```http
POST /memory/query
Content-Type: application/json

{
  "user_id": "user123",
  "query": "你还记得我的猫吗？",
  "limit": 10,
  "min_score": 0.7
}
```

#### 获取记忆统计
```http
GET /memory/stats
```

#### 获取 Embedding 提供商信息
```http
GET /memory/embedding-info
```

### 会话管理

#### 获取会话
```http
GET /session/{session_id}
```

#### 删除会话
```http
DELETE /session/{session_id}?delete_memories=true
```

#### 设置会话上下文
```http
PUT /session/{session_id}/context
Content-Type: application/json

{
  "user_name": "张三",
  "preferences": {
    "language": "zh-CN"
  }
}
```

### 用户管理

#### 获取用户会话列表
```http
GET /user/{user_id}/sessions
```

#### 获取最近记忆
```http
GET /user/{user_id}/memories/recent?limit=20
```

#### 搜索记忆
```http
GET /user/{user_id}/memories/search?q=猫&limit=10
```

#### 清理用户记忆
```http
DELETE /user/{user_id}/memories
```

### Webhook 端点

#### 处理清理任务
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

#### 调度定期清理
```http
POST /webhook/schedule-cleanup
Content-Type: application/json

{
  "callback_url": "https://your-domain.com/webhook/cleanup"
}
```

## 🧩 示例使用流程

### 1. 保存对话记忆
```bash
curl -X POST http://localhost:8080/memory/save \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "session_id": "session456",
    "content": "我家猫叫小橘子，它很喜欢晒太阳",
    "role": "user"
  }'
```

### 2. 查询相关记忆
```bash
curl -X POST http://localhost:8080/memory/query \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "query": "告诉我关于我的宠物的信息",
    "limit": 5
  }'
```

### 3. 获取会话信息
```bash
curl http://localhost:8080/session/session456
```

## 🏗 项目结构

```
github.com/Fairy-nn/MemoryCacheAI/
├── clients/          # 外部服务客户端
│   ├── embedding.go # Embedding 客户端 (Jina AI & OpenAI)
│   ├── redis.go     # Upstash Redis 客户端
│   ├── vector.go    # Upstash Vector 客户端
│   └── qstash.go    # Upstash QStash 客户端
├── config/          # 配置管理
│   └── config.go
├── handlers/        # HTTP 处理器
│   ├── memory.go    # 记忆相关接口
│   └── webhook.go   # Webhook 处理器
├── models/          # 数据模型
│   └── memory.go
├── services/        # 业务逻辑
│   └── memory.go    # 记忆服务
├── examples/        # 测试脚本
│   ├── test_api.sh  # API 测试脚本
│   └── test_embedding_providers.sh # Embedding 提供商测试
├── main.go          # 主程序入口
├── go.mod           # Go 模块文件
├── env.example      # 环境变量模板
├── README.md        # 项目文档 (中文)
└── README.en.md     # 项目文档 (英文)
```

## 🔧 配置说明

### Upstash 服务配置

1. **Redis**: 用于存储会话数据和短期记忆
   - 创建 Redis 数据库：https://console.upstash.com/redis
   - 获取 URL 和 Token

2. **Vector**: 用于存储和检索语义向量
   - 创建 Vector 数据库：https://console.upstash.com/vector
   - 选择合适的维度（Jina 默认 1024，OpenAI 根据模型而定）

3. **QStash**: 用于异步任务处理
   - 获取 QStash Token：https://console.upstash.com/qstash

### Embedding 服务配置

#### Jina AI 配置

1. 注册 Jina AI 账户：https://jina.ai/
2. 获取 API Key
3. 查看使用限制和定价

#### OpenAI 配置

1. 注册 OpenAI 账户：https://platform.openai.com/
2. 获取 API Key
3. 选择合适的 embedding 模型：
   - `text-embedding-3-small` (1536 维，性价比高)
   - `text-embedding-3-large` (3072 维，质量更高)
   - `text-embedding-ada-002` (1536 维，经典模型)

#### 切换 Embedding 提供商

1. 修改 `.env` 文件中的 `EMBEDDING_PROVIDER`
2. 配置相应的 API 密钥
3. 重启服务
4. **注意**：切换提供商后需要重新生成所有 embeddings，因为不同提供商的向量维度和特征可能不同

## 🧪 测试

### 健康检查
```bash
curl http://localhost:8080/health
```

### API 信息
```bash
curl http://localhost:8080/
```

### Webhook 测试
```bash
curl -X POST http://localhost:8080/webhook/test \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

### Embedding 提供商信息
```bash
curl http://localhost:8080/memory/embedding-info
```

### 完整功能测试
```bash
# 运行完整的 API 测试
bash examples/test_api.sh

# 运行 embedding 提供商测试
bash examples/test_embedding_providers.sh
```

## 🔄 数据流程

1. **保存记忆**：
   - 用户发送消息 → Redis (会话) + Jina (embedding) → Vector DB (长期记忆)

2. **查询记忆**：
   - 用户查询 → Jina (query embedding) → Vector DB (语义搜索) → 返回相关记忆

3. **清理记忆**：
   - QStash 定时任务 → Webhook → 清理过期数据

## 🚨 注意事项

1. **API 限制**：注意各服务的 API 调用限制
2. **数据安全**：生产环境请配置适当的认证和授权
3. **错误处理**：监控日志，及时处理异常情况
4. **成本控制**：合理设置 TTL，避免存储过多数据

## 📈 扩展功能

- [ ] 添加用户认证和授权
- [ ] 实现记忆重要性评分
- [ ] 支持多模态记忆（图片、音频）
- [ ] 添加记忆分类和标签
- [ ] 实现记忆压缩和摘要
- [ ] 支持记忆导出和备份

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License 

