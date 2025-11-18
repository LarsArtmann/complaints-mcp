#!/bin/bash

# Test complete cache stats workflow
echo "=== Testing Cache Statistics Tool ==="

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Test 1: File a complaint to generate cache data
echo "1. Filing a complaint to populate cache..."
cat << 'EOF' > /tmp/file_complaint.json
{
  "jsonrpc": "2.0", 
  "id": 1, 
  "method": "tools/call", 
  "params": {
    "name": "file_complaint", 
    "arguments": {
      "agent_name": "Test Agent",
      "task_description": "Test cache stats functionality",
      "severity": "medium",
      "session_name": "cache-test-session"
    }
  }
}
EOF

# Test 2: Get cache stats
echo "2. Getting cache statistics..."
cat << 'EOF' > /tmp/get_cache_stats.json
{
  "jsonrpc": "2.0", 
  "id": 2, 
  "method": "tools/call", 
  "params": {
    "name": "get_cache_stats", 
    "arguments": {}
  }
}
EOF

# Send requests (batch for testing)
echo "Sending batch request..."
(
  sleep 1
  cat /tmp/file_complaint.json
  sleep 1  
  cat /tmp/get_cache_stats.json
  sleep 1
) | timeout 15s "$PROJECT_ROOT/complaints-mcp" 2>&1 | grep -E "(Cache stats retrieved|get_cache_stats|cache_enabled|hit_rate|current_size)" || echo "No cache stats output found"

# Cleanup
rm -f /tmp/file_complaint.json /tmp/get_cache_stats.json

echo ""
echo "=== Testing with cache disabled ==="

# Test with cache disabled
echo "3. Testing with FileRepository (cache disabled)..."
cat << 'EOF' > /tmp/get_cache_stats_no_cache.json
{
  "jsonrpc": "2.0", 
  "id": 3, 
  "method": "tools/call", 
  "params": {
    "name": "get_cache_stats", 
    "arguments": {}
  }
}
EOF

timeout 10s "$PROJECT_ROOT/complaints-mcp" --cache-enabled=false < /tmp/get_cache_stats_no_cache.json 2>&1 | grep -E "(Cache stats retrieved|Cache disabled|cache_enabled)" || echo "No cache disabled output found"

rm -f /tmp/get_cache_stats_no_cache.json

echo ""
echo "=== Test completed ==="