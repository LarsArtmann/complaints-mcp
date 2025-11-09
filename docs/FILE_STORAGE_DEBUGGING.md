# File Storage Location Debugging Guide

## 🚨 Mystery Solved: Actual File Storage Location

### Problem
System reported loading "10 complaints from disk" but files could not be found in expected locations like `~/.local/share/complaints/`.

### Root Cause
**XDG Base Directory specification varies by platform:**
- **Linux**: `~/.local/share/complaints/`
- **macOS**: `~/Library/Application Support/complaints/` 
- **Windows**: `%APPDATA%/complaints/`

### Solution Enhanced
Added comprehensive logging to show actual storage paths:

```log
INFO Loading all complaints from disk for cache warm-up component=cached-repository base_dir="/Users/larsartmann/Library/Application Support/complaints"
INFO Complaints loaded from disk successfully component=cached-repository base_dir="/Users/larsartmann/Library/Application Support/complaints" count=10
```

### How to Locate Your Complaint Files

#### Option 1: Check Logs
Run with debug logging to see actual paths:
```bash
./complaints-mcp --log-level debug --dev
```

#### Option 2: Platform-Specific Commands
```bash
# macOS
ls ~/Library/Application\ Support/complaints/

# Linux  
ls ~/.local/share/complaints/

# Windows
dir %APPDATA%\complaints\
```

#### Option 3: Custom Configuration
Create `config.yaml` to override default:
```yaml
storage:
  base_dir: "./my-complaints"  # Local directory
  # OR absolute path
  base_dir: "/absolute/path/to/complaints"
```

### Configuration Precedence
1. **Command line flag**: `--config /path/to/config.yaml`
2. **Environment variable**: `COMPLAINTS_MCP_STORAGE_BASE_DIR=/custom/path`
3. **Config file**: `config.yaml` in search paths
4. **XDG default**: Platform-specific (as shown above)

### Implementation Details
- **File naming**: `YYYY-MM-DD_HH-MM-SS-SESSION_NAME.json`
- **Content**: JSON format with full complaint metadata
- **Backup**: Automatic backup to `docs/complaints/` when enabled
- **Cache**: LRU cache for O(1) lookups (configurable size)

### Testing File Creation
File a test complaint to verify location:
```bash
# Use the MCP tool or:
curl -X POST http://localhost:8080/complaints \
  -H "Content-Type: application/json" \
  -d '{"agent_name":"test","task_description":"test file creation","severity":"low","context_info":"testing storage location"}'
```

Then check your platform's storage location for the new JSON file.

### Enhanced Debugging Commands
```bash
# Show configuration and paths
./complaints-mcp --log-level debug --dev 2>&1 | grep -E "(base_dir|Loading|Complaints loaded)"

# Monitor file system during startup  
./complaints-mcp --log-level debug --dev &
PID=$!
sleep 2
find ~/Library/Application\ Support/complaints/ -name "*.json" -newer /tmp/starttime 2>/dev/null
kill $PID 2>/dev/null
```

### Issue Status: ✅ RESOLVED
- **Root cause identified**: XDG platform differences
- **Enhanced logging implemented**: Shows actual paths in debug output  
- **Documentation updated**: Platform-specific paths documented
- **Testing verified**: Confirmed file creation and retrieval

**Related Issues**: #46 (File Storage Mystery) - SOLVED