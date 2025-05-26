# Troubleshooting Guide

## 🔧 Common Issues and Solutions

### 1. Redis Connection Error: `ERR failed to parse command`

**Root Cause**: Incorrect Redis URL format

**Solution**:
- Ensure Redis URL ends with `/`
- Correct format: `https://your-redis-url.upstash.io/`
- Incorrect format: `https://your-redis-url.upstash.io`

### 2. Jina API Error: `you must provide a model parameter`

**Root Cause**: Missing model parameter in Jina API request

**Solution**:
- Fixed: Added `"model": "jina-embeddings-v3"` to the code
- Make sure you're using the latest code version

### 3. Vector Dimension Mismatch

**Root Cause**: Vector database dimensions don't match embedding model

**Solution**:
- **Jina v3**: Use `1024` dimensions
- **OpenAI text-embedding-3-small**: Use `1536` dimensions
- **OpenAI text-embedding-3-large**: Use `3072` dimensions

**Important**: Vector database dimensions cannot be modified after creation - you need to recreate!

### 4. Environment Variables Checklist

```bash
# Check if environment variables are correctly set
echo "Redis URL: $UPSTASH_REDIS_URL"
echo "Vector URL: $UPSTASH_VECTOR_URL"
echo "Embedding Provider: $EMBEDDING_PROVIDER"
echo "Jina API Key: $JINA_API_KEY"
```

### 5. Testing Steps

1. **Rebuild the project**:
```bash
go mod tidy
go build
```

2. **Restart the service**:
```bash
./MemoryCacheAI
```

3. **Run tests**:
```bash
bash examples/test_api.sh
```

### 6. Vector Database Reconfiguration

If dimensions don't match, you need to:

1. Delete the existing vector database
2. Create a new vector database based on your embedding provider:
   - **Using Jina**: Set dimensions to `1024`
   - **Using OpenAI**: Set dimensions to `1536` or `3072`
3. Update the URL and Token in environment variables

### 7. Configuration Verification

Run the following commands to verify configuration:

```bash
# Check embedding information
curl http://localhost:8080/memory/embedding-info

# Check vector database statistics
curl http://localhost:8080/memory/stats
```

### 8. Common Error Codes

- **400**: Bad request format, check API parameters
- **401**: Authentication failed, check API keys
- **404**: Resource not found, check URL configuration
- **500**: Internal server error, check logs

## 📞 Getting Help

If issues persist:
1. Check service logs
2. Verify all API keys
3. Confirm network connectivity
4. Check Upstash console status 

Or create an issue on [GitHub](https://github.com/Fairy-nn/MemoryCacheAI/issues)