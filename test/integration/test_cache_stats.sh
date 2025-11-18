#!/bin/bash

# Test get_cache_stats tool with proper MCP JSON-RPC
echo "Testing get_cache_stats tool..."

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Create a JSON-RPC request for get_cache_stats
cat << 'EOF' > /tmp/test_request.json
{
  "jsonrpc": "2.0", 
  "id": 1, 
  "method": "tools/call", 
  "params": {
    "name": "get_cache_stats", 
    "arguments": {}
  }
}
EOF

# Send request to server and capture response
echo "Sending request..."
timeout 10s "$PROJECT_ROOT/complaints-mcp" < /tmp/test_request.json 2>/dev/null | head -20

echo ""
echo "Testing with cache disabled..."
timeout 10s "$PROJECT_ROOT/complaints-mcp" --cache-enabled=false < /tmp/test_request.json 2>/dev/null | head -20

# Cleanup
rm -f /tmp/test_request.json