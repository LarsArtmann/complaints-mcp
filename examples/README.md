# Configuration Examples

This directory contains example configuration files for the complaints-mcp system.

## Files

### `cache-config.toml`
Demonstrates various cache configuration scenarios:

- **Basic Setup**: File repository without caching
- **Development Cache**: Small 50-item cache for local development
- **Production Small**: 1,000 item cache for small production workloads
- **Production Large**: 50,000 item cache for high-traffic systems
- **FIFO Eviction**: Predictable cleanup behavior
- **No Eviction**: Manual cache management

## Cache Configuration Parameters

| Parameter | Type | Range | Default | Description |
|-----------|------|-------|---------|-------------|
| `cache_enabled` | bool | - | `false` | Enable/disable caching |
| `cache_max_size` | int | 1-100,000 | `1,000` | Maximum cache entries |
| `cache_eviction` | string | `lru`, `fifo`, `none` | `lru` | Eviction policy |

## Performance Guidelines

### Cache Size Selection

- **50-100**: Development environments
- **500-2,000**: Small production deployments (< 100 req/s)
- **5,000-20,000**: Medium deployments (100-500 req/s)
- **50,000+**: High-traffic deployments (500+ req/s)

### Eviction Policy Selection

- **LRU** (`"lru"`): Best for general workloads (default)
- **FIFO** (`"fifo"`): Predictable cleanup, memory-efficient
- **None** (`"none"`): Manual cache management, requires monitoring

## Type Safety Features

The configuration system provides compile-time guarantees:

✅ **Integer Overflow Protection**: Cache sizes limited to 1-100,000
✅ **Enum Validation**: Only valid eviction policies accepted  
✅ **Runtime Validation**: Invalid values rejected with clear errors
✅ **32-bit Safety**: No overflow risk on any architecture

## Usage

```bash
# Use specific config file
complaints-mcp --config examples/cache-config.toml

# Use environment variables
export COMPLAINTS_STORAGE_CACHE_ENABLED=true
export COMPLAINTS_STORAGE_CACHE_MAX_SIZE=5000
complaints-mcp
```

## Migration Guide

Upgrading from older configurations:

1. **Old Config** (unsafe):
   ```toml
   [storage]
   cache_max_size = 9223372036854775807  # Risky!
   cache_eviction = "invalid_policy"  # Crashes!
   ```

2. **New Config** (type-safe):
   ```toml
   [storage]
   cache_enabled = true
   cache_max_size = 50000  # Validated range
   cache_eviction = "lru"  # Enum-validated
   ```

The new system automatically validates configuration and provides helpful error messages for invalid values.