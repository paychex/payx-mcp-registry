# Paychex Internal Registry Metadata

## Namespace

**Important**: The official MCP server.json schema only allows `io.modelcontextprotocol.registry/publisher-provided` as a top-level `_meta` namespace. Therefore, all Paychex-specific metadata is **nested within** the `publisher-provided` namespace as a `paychex` object:

```
_meta.io.modelcontextprotocol.registry/publisher-provided.paychex
```

This approach:
- ✅ Complies with the official MCP schema
- ✅ Allows us to add governance metadata in our fork
- ✅ Keeps Paychex fields isolated within a dedicated sub-object

## Metadata Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `published_by` | string | Yes | Publisher username or email |
| `publish_date` | string | Yes | Publishing date (ISO 8601 format: YYYY-MM-DD) |
| `ref_ticket` | string | Yes | Tracking ticket reference (JIRA/SNOW) |
| `server_source` | string | Yes | Server origin - `internal` for Paychex-developed, `external` for third-party |
| `external_source` | string | No | For external servers, specify source (e.g., `anthropic`, `langchain`, `github/user/repo`) |
| `line_of_business` | string | Yes | Paychex line of business the server is published by|

### Field Details

**server_source**: Indicates whether the server is developed internally or comes from external sources
- `internal` - Developed and maintained by Paychex
- `external` - Third-party server from public registries

**external_source**: When `server_source` is `external`, this field must specify the origin:
- Public registry source (e.g., anthropic, microsoft)
- Repository location (e.g., `github/langchain-ai/langchain`)

**Size Limit**: The entire metadata namespace is limited to 4KB.

## Examples

### Example 1: Internal Paychex Server
```json
{
  "name": "com.paychex.internal/payroll-assistant",
  "description": "Internal MCP server for payroll processing",
  "version": "2.1.0",
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "paychex": {
        "published_by": "jane.smith@paychex.com",
        "publish_date": "2026-01-15",
        "ref_ticket": "AIA-2001",
        "server_source": "internal",
        "line_of_business": "Payroll Services"
      }
    }
  }
}
```

### Example 2: External Registry Server (From Anthropic)
```json
{
  "name": "ai.anthropic/claude-toolkit",
  "description": "Anthropic's official Claude MCP toolkit",
  "version": "1.5.0",
  "_meta": {
    "io.modelcontextprotocol.registry/publisher-provided": {
      "paychex": {
        "published_by": "registry-admin@paychex.com",
        "publish_date": "2026-01-20",
        "ref_ticket": "AIA-2050",
        "server_source": "external",
        "external_source": "anthropic",
        "line_of_business": "AI & Automation"
      }
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
    "io.modelcontextprotocol.registry/publisher-provided": {
      "paychex": {
        "published_by": "dev-team@paychex.com",
        "publish_date": "2024-12-01",
        "ref_ticket": "AIA-1503",
        "server_source": "external",
        "external_source": "github/langchain-ai/langchain",
        "line_of_business": "IS"
      }
    }
  }
}
```

## Best Practices

1. **Use the correct nesting**: Always place Paychex fields under `_meta["io.modelcontextprotocol.registry/publisher-provided"]["paychex"]`
2. **Required for all servers**: All servers in the Paychex registry MUST include the paychex metadata object
3. **Stay under 4KB**: Keep the entire `publisher-provided` object (including paychex) under 4KB
4. **Use specific external registry names**: For `external_source`, use specific names like `anthropic`, `microsoft`
5. **Maintain consistency**: Use defined _meta fields for tracking and governance

## Comparison with Official Registry

| Aspect | Official Registry | Paychex Fork |
|--------|------------------|--------------||
| Top-level namespace | `io.modelcontextprotocol.registry/publisher-provided` | Same (schema-compliant) |
| Paychex-specific fields | N/A | Nested under `publisher-provided.paychex` |
| Data Type | `map[string]interface{}` | `map[string]interface{}` |
| Required | No | **Yes** (all servers must have paychex metadata) |
| Size Limit | 4KB | 4KB |
| Purpose | Build/CI metadata | Governance & tracking |

See [examples/paychex-server-example.json](../examples/paychex-server-example.json) for a complete working example.