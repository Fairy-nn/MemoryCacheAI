#!/bin/bash

# MemoryCacheAI API Test Script
# Ensure the service is running at http://localhost:8080

BASE_URL="http://localhost:8080"
USER_ID="test_user_123"
SESSION_ID="test_session_456"

echo "🧠 MemoryCacheAI API Test"
echo "========================="

# 1. Health Check
echo "1. Health Check..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# 2. Get API Information
echo "2. API Information..."
curl -s "$BASE_URL/" | jq '.endpoints'
echo ""

# 3. Save Memory - User Message
echo "3. Save User Memory..."
curl -s -X POST "$BASE_URL/memory/save" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"session_id\": \"$SESSION_ID\",
    \"content\": \"I have an orange cat named Orange who loves sunbathing and eating dried fish\",
    \"role\": \"user\"
  }" | jq '.'
echo ""

# 4. Save Memory - Assistant Reply
echo "4. Save Assistant Memory..."
curl -s -X POST "$BASE_URL/memory/save" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"session_id\": \"$SESSION_ID\",
    \"content\": \"Orange sounds like a lovely orange cat! Orange cats are usually gentle in nature, and loving sunbaths is a natural behavior for cats.\",
    \"role\": \"assistant\"
  }" | jq '.'
echo ""

# 5. Save More Memories
echo "5. Save More Memories..."
curl -s -X POST "$BASE_URL/memory/save" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"session_id\": \"$SESSION_ID\",
    \"content\": \"I also love watching sci-fi movies, recently watching Interstellar\",
    \"role\": \"user\"
  }" | jq '.'
echo ""

# Wait for embedding processing
echo "⏳ Waiting for embedding processing..."
sleep 3

# 6. Query Memory - About Cat
echo "6. Query Memory About Cat..."
curl -s -X POST "$BASE_URL/memory/query" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"query\": \"Tell me about my pet cat\",
    \"limit\": 5,
    \"min_score\": 0.3
  }" | jq '.'
echo ""

# 7. Query Memory - About Movies
echo "7. Query Memory About Movies..."
curl -s -X POST "$BASE_URL/memory/query" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"$USER_ID\",
    \"query\": \"What kind of movies do I like?\",
    \"limit\": 5,
    \"min_score\": 0.3
  }" | jq '.'
echo ""

# 8. Get Session Information
echo "8. Get Session Information..."
curl -s "$BASE_URL/session/$SESSION_ID" | jq '.'
echo ""

# 9. Set Session Context
echo "9. Set Session Context..."
curl -s -X PUT "$BASE_URL/session/$SESSION_ID/context" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_name\": \"John Doe\",
    \"preferences\": {
      \"language\": \"en-US\",
      \"topic_interests\": [\"pets\", \"movies\", \"technology\"]
    }
  }" | jq '.'
echo ""

# 10. Get User Sessions List
echo "10. Get User Sessions List..."
curl -s "$BASE_URL/user/$USER_ID/sessions" | jq '.'
echo ""

# 11. Get Recent Memories
echo "11. Get Recent Memories..."
curl -s "$BASE_URL/user/$USER_ID/memories/recent?limit=10" | jq '.'
echo ""

# 12. Search Memories
echo "12. Search Memories..."
curl -s "$BASE_URL/user/$USER_ID/memories/search?q=cat&limit=5" | jq '.'
echo ""

# 13. Get Memory Statistics
echo "13. Get Memory Statistics..."
curl -s "$BASE_URL/memory/stats" | jq '.'
echo ""

# 14. Test Webhook
echo "14. Test Webhook..."
curl -s -X POST "$BASE_URL/webhook/test" \
  -H "Content-Type: application/json" \
  -d "{
    \"test_message\": \"Hello from test script\",
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
  }" | jq '.'
echo ""

# 15. Get Webhook Information
echo "15. Webhook Information..."
curl -s "$BASE_URL/webhook/info" | jq '.'
echo ""

echo "✅ Test Complete!"
echo ""
echo "💡 Tips:"
echo "- If some queries return empty results, it might be due to insufficient embedding similarity"
echo "- You can adjust the min_score parameter to get more results"
echo "- Make sure all API keys are properly configured" 