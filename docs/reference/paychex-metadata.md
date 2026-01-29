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

### Example 2: External Registry Server (From Anthropic)
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
3. **Use specific external registry names**: For `external_source`, use specific names like `anthropic`, `microsoft`
4. **Maintain consistency**: Use defined _meta fields for tracking and governance

## Comparison with Official Registry

| Aspect | Official Registry | Paychex Fork |
|--------|------------------|--------------|
| Namespace | `io.modelcontextprotocol.registry/publisher-provided` | `io.github.paychex.payx-mcp-registry/internal` |
| Data Type | `map[string]interface{}` | `map[string]interface{}` |
| Size Limit | 4KB | 4KB |
| Purpose | Build/CI metadata | Governance & tracking |

See [examples/paychex-server-example.json](../examples/paychex-server-example.json) for a complete working example.