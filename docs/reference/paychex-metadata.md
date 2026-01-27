# Paychex Internal Registry Metadata Fields

This document describes the Paychex-specific `_meta` fields used in the internal MCP registry.

## Overview

The Paychex MCP registry follows the official MCP registry pattern by using **reverse DNS namespacing** for custom metadata. All Paychex-specific fields are stored under the namespace `io.github.paychex.payx-mcp-registry/internal` as a flexible JSON object.

## Namespace Convention

Following the official MCP registry pattern (which uses `io.modelcontextprotocol.registry/publisher-provided`), we use:

```json
{
  "_meta": {
    "io.github.paychex.payx-mcp-registry/internal": {
      // Your custom fields here
    }
  }
}
```

### Why Use Namespacing?

1. **Compatibility**: Follows the official MCP registry pattern
2. **No Conflicts**: Your fork won't conflict with official registry updates
3. **Flexibility**: Store any JSON structure (not limited to predefined fields)
4. **Future-Proof**: Easy to add/remove fields without schema changes

## Field Reference

All fields are stored as properties within `io.github.paychex.payx-mcp-registry/internal`:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `published_by` | string | Username or email of publisher | `"john.doe@paychex.com"` |
| `publish_date` | string | ISO 8601 date of publishing | `"2026-01-27"` |
| `ref_ticket` | string | Tracking ticket reference | `"AIA-1234"` |
| `server_source` | string | Origin: `internal`, `external`, or vendor name | `"external"`, `"anthropic"` |
| `line_of_business` | string | Line of business | `"AI & Automation"` |

### Size Limit

Like the official `publisher-provided` metadata, the `io.github.paychex.payx-mcp-registry/internal` namespace has a **4KB limit** to prevent abuse and ensure reasonable payload sizes.

## Complete Examples

### External Third-Party Server

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "com.figma.mcp/mcp",
  "description": "Official Figma MCP server for AI design workflows",
  "version": "1.0.3",
  "packages": [...],
  
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "tool": "npm-publisher",
      "version": "1.0.1"
    },
    "io.github.paychex.payx-mcp-registry/internal": {
      "published_by": "test-approver",
      "publish_date": "2024-12-01",
      "ref_ticket": "AIA-1503",
      "server_source": "external",
      "line_of_business": "AI & Automation"
    }
  }
}
```

### Internal Paychex Server

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "com.paychex.internal/payroll-assistant",
  "description": "Internal MCP server for payroll processing assistance",
  "version": "2.1.0",
  "packages": [...],
  
  "_meta": {
    "io.github.paychex.payx-mcp-registry/internal": {
      "published_by": "jane.smith@paychex.com",
      "publish_date": "2026-01-15",
      "ref_ticket": "AIA-2001",
      "server_source": "internal",
      "line_of_business": "Payroll Services"
    }
  }
}
```

### Vendor Server (Anthropic)

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "ai.anthropic.mcp/claude-toolkit",
  "description": "Anthropic's official Claude MCP toolkit",
  "version": "1.5.0",
  "packages": [...],
  
  "_meta": {
    "io.github.paychex.payx-mcp-registry/internal": {
      "published_by": "registry-admin@paychex.com",
      "publish_date": "2026-01-20",
      "ref_ticket": "AIA-2050",
      "server_source": "anthropic",
      "line_of_business": "AI & Automation"
    }
  }
}
```

## Publishing with Paychex Metadata

When publishing a server to the internal Paychex registry:

```bash
# Using the mcp-publisher CLI
./bin/mcp-publisher publish server.json
```

The registry validates:
- The `io.github.paychex.payx-mcp-registry/internal` object doesn't exceed 4KB
- JSON structure is valid
- All required server fields are present

## Extending the Metadata

Since the namespace stores a flexible JSON object, you can easily add new fields without code changes:

```json
{
  "_meta": {
    "io.github.paychex.payx-mcp-registry/internal": {
      "published_by": "...",
      "publish_date": "...",
      "ref_ticket": "...",
      "server_source": "...",
      "line_of_business": "...",
      // New fields can be added anytime
      "security_scan_date": "2026-01-20",
      "compliance_tags": ["SOX", "PCI"],
      "approval_chain": ["manager", "security", "compliance"]
    }
  }
}
```

## Best Practices

1. **Use the namespace**: Always place fields under `io.github.paychex.payx-mcp-registry/internal`
2. **Stay under 4KB**: Keep metadata concise and relevant
3. **Use specific vendor names**: For `server_source`, use vendor names like `anthropic`, `microsoft` instead of generic `external`
4. **Document new fields**: If you add fields beyond the 5 core ones, document them for your team
5. **Maintain compatibility**: Don't remove existing fields that might be in use

## Comparison with Official Registry

| Aspect | Official Registry | Paychex Fork |
|--------|------------------|--------------|
| Namespace | `io.modelcontextprotocol.registry/publisher-provided` | `io.github.paychex.payx-mcp-registry/internal` |
| Data Type | `map[string]interface{}` | `map[string]interface{}` |
| Size Limit | 4KB | 4KB |
| Purpose | Build/CI metadata | Governance & tracking |
| Preservation | Only this key preserved | Your fork, you control |

## Migration from approved-mcp-servers.json

Existing entries map to the new structure:

| Old Field | New Location |
|-----------|--------------|
| `approved_by` | `io.github.paychex.payx-mcp-registry/internal.published_by` |
| `approval_date` | `io.github.paychex.payx-mcp-registry/internal.publish_date` |
| `jira_ticket` | `io.github.paychex.payx-mcp-registry/internal.ref_ticket` |
| N/A | `io.github.paychex.payx-mcp-registry/internal.server_source` |
| N/A | `io.github.paychex.payx-mcp-registry/internal.line_of_business` |

See [examples/paychex-server-example.json](../examples/paychex-server-example.json) for a complete working example.

## Why This Approach is Correct

Based on the [official MCP registry documentation](https://github.com/modelcontextprotocol/registry):

1. ✅ **Follows reverse DNS convention**: Like `io.modelcontextprotocol.registry/publisher-provided`
2. ✅ **Uses flexible map structure**: `map[string]interface{}` allows any JSON
3. ✅ **Enforces size limits**: 4KB limit prevents abuse
4. ✅ **Namespace isolation**: No conflicts with official registry or future updates
5. ✅ **Easy to extend**: Add fields without code changes

This approach ensures your fork remains compatible with upstream changes while allowing custom internal metadata.
