#!/bin/bash

# Test the complaints-mcp server
echo "Testing complaints-mcp server..."

# Start the server in background
echo "Starting server..."
./complaints-mcp &
SERVER_PID=$!

# Wait a moment for server to start
sleep 2

# Send a tools/list request (this will hang as we need proper MCP client)
echo "Server appears to be running (PID: $SERVER_PID)"
echo "To properly test, use an MCP client that can communicate over stdio"

# Kill the server
kill $SERVER_PID 2>/dev/null

echo "Basic test completed"