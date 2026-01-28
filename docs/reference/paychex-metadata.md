# Paychex Internal Registry Metadata

## Namespace

All Paychex-specific metadata is stored under `io.github.paychex.payx-mcp-registry/internal` following the official MCP registry namespacing pattern.

## Metadata Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `published_by` | string | Yes | Publisher username or email |
| `publish_date` | string | Yes | Publishing date (ISO 8601 format: YYYY-MM-DD) |
| `ref_ticket` | string | Yes | Tracking ticket reference (JIRA/SNOW) |
| `server_source` | string | Yes | Server origin - `internal` for Paychex-developed, `external` for third-party |
| `external_source` | string | No | For external servers: specify source (e.g., `anthropic`, `langchain`, `github/user/repo`) |
| `line_of_business` | string | Yes | Paychex line of business |

### Field Details

**server_source**: Indicates whether the server is developed internally or comes from external sources
- `internal` - Developed and maintained by Paychex
- `external` - Third-party server from public registries or vendors

**external_source**: When `server_source` is `external`, this field specifies the origin:
- Public vendor name (e.g., `anthropic`, `microsoft`)
- Public registry source (e.g., official MCP registry)
- Repository location (e.g., `github/langchain-ai/langchain`)
- Omit if not from a known public registry or vendor

**Size Limit**: The entire metadata namespace is limited to 4KB.

## Examples

### Example 1: Internal Paychex Server
```json
{
  "name": "com.paychex.internal/payroll-assistant",
  "description": "Internal MCP server for payroll processing",
  "version": "2.1.0",
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

### Example 2: External Vendor Server (From Anthropic)
```json
{
  "name": "ai.anthropic/claude-toolkit",
  "description": "Anthropic's official Claude MCP toolkit",
  "version": "1.5.0",
  "_meta": {
    "io.github.paychex.payx-mcp-registry/internal": {
      "published_by": "registry-admin@paychex.com",
      "publish_date": "2026-01-20",
      "ref_ticket": "AIA-2050",
      "server_source": "external",
      "external_source": "anthropic",
      "line_of_business": "AI & Automation"
    }
  }
}
```

### Example 3: External Third-Party Server (Not in Public Registry)
```json
{
  "name": "com.langchain.mcp/docs",
  "description": "LangChain documentation MCP server",
  "version": "1.0.0",
  "_meta": {
    "io.github.paychex.payx-mcp-registry/internal": {
      "published_by": "dev-team@paychex.com",
      "publish_date": "2024-12-01",
      "ref_ticket": "AIA-1503",
      "server_source": "external",
      "external_source": "github/langchain-ai/langchain",
      "line_of_business": "IS"
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
