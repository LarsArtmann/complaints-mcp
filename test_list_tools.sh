#!/bin/bash

# Test by listing available tools
echo "=== Testing MCP Server Tools ==="

# List tools request
cat << 'EOF' > /tmp/list_tools.json
{
  "jsonrpc": "2.0", 
  "id": 1, 
  "method": "tools/list", 
  "params": {}
}
EOF

echo "Sending tools/list request..."
timeout 5s ./complaints-mcp < /tmp/list_tools.json 2>&1 | head -30

rm -f /tmp/list_tools.json

echo ""
echo "=== Checking server logs for tool registration ==="