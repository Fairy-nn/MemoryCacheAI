#!/bin/bash

# MemoryCacheAI Embedding Providers Test Script
# Test Jina AI and OpenAI embedding providers

BASE_URL="http://localhost:8080"
USER_ID="embedding_test_user"

echo "🧠 MemoryCacheAI Embedding Providers Test"
echo "=========================================="

# 1. Health Check
echo "1. Health Check..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# 2. Test current configured embedding provider
echo "2. Save test memories (using current configured provider)..."

# Save some test data
test_memories=(
    "I love walking in the park on weekends, especially when cherry blossoms are blooming"
    "My job is software engineer, mainly using Python and Go languages"
    "I have a golden retriever named Lucky, he is smart and friendly"
    "My favorite movie genre is sci-fi, especially Interstellar"
    "I'm learning machine learning, particularly interested in natural language processing"
)

session_id="embedding_test_session_$(date +%s)"

for i in "${!test_memories[@]}"; do
    memory="${test_memories[$i]}"
    echo "Saving memory $((i+1)): $memory"
    
    curl -s -X POST "$BASE_URL/memory/save" \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"$USER_ID\",
        \"session_id\": \"$session_id\",
        \"content\": \"$memory\",
        \"role\": \"user\"
      }" | jq '.message'
    
    # Brief delay to ensure processing
    sleep 1
done

echo ""
echo "⏳ Waiting for embedding processing..."
sleep 3

# 3. Test semantic search
echo "3. Test semantic search..."

queries=(
    "Tell me about my pet"
    "What is my profession?"
    "What kind of entertainment activities do I like?"
    "What technical field am I interested in?"
)

for query in "${queries[@]}"; do
    echo ""
    echo "🔍 Query: $query"
    
    result=$(curl -s -X POST "$BASE_URL/memory/query" \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"$USER_ID\",
        \"query\": \"$query\",
        \"limit\": 3,
        \"min_score\": 0.3
      }")
    
    echo "Results count: $(echo "$result" | jq '.total')"
    echo "Most relevant memory:"
    echo "$result" | jq -r '.results[0].content // "No relevant memory found"'
    echo "Similarity score: $(echo "$result" | jq -r '.results[0].score // "N/A"')"
done

echo ""
echo "4. Get memory statistics..."
curl -s "$BASE_URL/memory/stats" | jq '.'

echo ""
echo "5. Test different similarity threshold effects..."

thresholds=(0.1 0.3 0.5 0.7 0.9)
test_query="my hobbies and interests"

for threshold in "${thresholds[@]}"; do
    echo ""
    echo "🎯 Threshold: $threshold"
    
    result=$(curl -s -X POST "$BASE_URL/memory/query" \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"$USER_ID\",
        \"query\": \"$test_query\",
        \"limit\": 5,
        \"min_score\": $threshold
      }")
    
    total=$(echo "$result" | jq '.total')
    echo "Found $total memories"
    
    if [ "$total" -gt 0 ]; then
        echo "Highest score: $(echo "$result" | jq -r '.results[0].score')"
    fi
done

echo ""
echo "6. Batch query test..."

batch_queries=(
    "programming languages"
    "animals pets"
    "movies entertainment"
    "outdoor activities"
    "learning research"
)

for query in "${batch_queries[@]}"; do
    echo ""
    echo "📊 Batch query: $query"
    
    result=$(curl -s -X POST "$BASE_URL/memory/query" \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"$USER_ID\",
        \"query\": \"$query\",
        \"limit\": 2,
        \"min_score\": 0.4
      }")
    
    total=$(echo "$result" | jq '.total')
    echo "Related memories count: $total"
    
    if [ "$total" -gt 0 ]; then
        echo "$result" | jq -r '.results[] | "- \(.content) (score: \(.score))"'
    fi
done

echo ""
echo "✅ Embedding provider test complete!"
echo ""
echo "💡 Tips:"
echo "- Current embedding provider can be configured via EMBEDDING_PROVIDER environment variable"
echo "- Supported providers: jina, openai"
echo "- Different providers may have different embedding dimensions, requiring Vector database recreation"
echo "- You can control retrieval precision by adjusting the min_score parameter"

echo ""
echo "🔧 Provider switching guide:"
echo "1. Stop the service"
echo "2. Modify EMBEDDING_PROVIDER in .env file"
echo "3. Configure corresponding API keys"
echo "4. Restart the service"
echo "5. Note: After switching providers, all embeddings need to be regenerated" 