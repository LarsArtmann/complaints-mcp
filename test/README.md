# Test Directory

This directory contains all test-related files for the complaints-mcp project.

## Structure

```
test/
├── integration/          # Integration and E2E test scripts
│   ├── test.sh          # Basic server integration test
│   ├── test_list_tools.sh    # MCP tools listing test
│   ├── test_cache_stats.sh   # Cache stats tool test
│   └── test_cache_complete.sh # Complete cache workflow test
└── config/              # Test configuration files
    └── test-config.yaml # Test configuration
```

## Running Tests

### Using Just (Recommended)

```bash
# Run all unit tests
just test

# Run BDD tests
just test-bdd

# Run integration tests
just test-integration

# Run cache-specific integration tests
just test-cache

# Run full CI pipeline
just ci
```

### Direct Execution

```bash
# Run integration tests directly
cd test/integration
./test.sh
./test_cache_complete.sh
```

## Test Configuration

The `test/config/test-config.yaml` file contains configuration for testing:

- Storage directories for test data
- Cache settings optimized for testing
- Debug logging enabled

## Integration Tests

Integration tests test the MCP server as a whole, including:

- Server startup and shutdown
- MCP protocol communication
- Tool functionality (file_complaint, get_cache_stats, etc.)
- Cache behavior and performance

These tests use JSON-RPC requests over stdio to simulate real MCP client interactions.