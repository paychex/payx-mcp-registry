# Publishing MCP Servers to Internal Registry

This guide covers how to publish your MCP server to Paychex's internal MCP registry using GitHub Actions.

## Overview

The internal registry uses GitHub OIDC for authentication, allowing any repository under the `paychex` GitHub organization to publish servers automatically without managing credentials.

**Key benefits:**
- ✅ No secrets to manage
- ✅ Automatic publishing on git tags
- ✅ Org-wide access for all Paychex repos

## Prerequisites

- Repository under `paychex` GitHub organization
- MCP server implementation
- GitHub Actions enabled in your repository

## Setup

### 1. Create `server.json`

Add a `server.json` file to your repository root with minimal metadata:

```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-10-17/server.schema.json",
  "name": "io.github.paychex/YOUR-SERVER-NAME",
  "version": "0.1.0",
  "description": "Brief description of your MCP server"
}
```

**⚠️ Important:** Server name MUST start with `io.github.paychex/` to match OIDC permissions.

### 2. Create Publishing Workflow

Create `.github/workflows/publish.yml`:

```yaml
name: Publish MCP Server

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:

jobs:
  publish:
    permissions:
      contents: read
      id-token: write
    uses: paychex/payx-mcp-registry/.github/workflows/internal-mcp-publish.yml@64e4ed4f0138f2b4f59f40cf1b2d571a28101558
    with:
      server-json: 'server.json'
      set-version-from-tag: ${{ github.ref_type == 'tag' }}
      mcp-publisher-version: 'latest'
    secrets: inherit
```

### 3. Configure Registry URL

Set the internal registry URL as a repository secret:

```bash
gh secret set MCP_REGISTRY_URL \
  --body "https://conmcpregb001.livelysmoke-0f2aa8bb.eastus.azurecontainerapps.io"
```

Or configure it at the organization level for all repos to inherit.

## Publishing

### Manual Trigger

Test the workflow manually:

```bash
gh workflow run publish.yml
gh run watch
```

### Automatic via Tags

Create and push a version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow automatically:
1. Extracts version from tag (`v0.1.0` → `0.1.0`)
2. Authenticates via GitHub OIDC
3. Publishes to internal registry

## Verification

Check if your server was published:

```bash
curl -fsSL "https://conmcpregb001.livelysmoke-0f2aa8bb.eastus.azurecontainerapps.io/v0.1/servers" | \
  jq '.servers[] | select(.name | contains("YOUR-SERVER"))'
```

## Troubleshooting

### "Failed to connect to localhost"

**Cause:** Both a variable and secret named `MCP_REGISTRY_URL` exist. Variables take precedence.

**Fix:**
```bash
gh variable delete MCP_REGISTRY_URL
```

Keep only the secret.

### "Permission denied" (403)

**Cause:** Server name doesn't match OIDC permissions pattern.

**Fix:** Ensure `server.json` name starts with `io.github.paychex/`:

```diff
{
-  "name": "paychex/my-server",
+  "name": "io.github.paychex/my-server",
   ...
}
```

### "Validation failed" (422)

**Cause:** Registry validates against minimal schema. Extra fields are rejected.

**Fix:** Use only these fields in `server.json`:
- `$schema`
- `name`
- `version`
- `description`

Remove: `icon`, `homepage`, `tools`, `resources`, `config`, etc.

### Workflow not triggering on tags

**Checklist:**
- [ ] Tag pattern matches (`v*`)
- [ ] Tag pushed to remote: `git push origin v0.1.0`
- [ ] Workflow file exists on default branch
- [ ] GitHub Actions enabled in repo settings

## How It Works

```
Your Repo → GitHub OIDC → Registry JWT → Publish
            (audience:       (grants:
             mcp-registry)    io.github.paychex/*)
```

1. **Workflow triggers** on tag or manual dispatch
2. **GitHub OIDC** issues token with repository context
3. **Registry validates** token against GitHub's JWKS
4. **Permission granted** for `io.github.paychex/*` namespace
5. **Server published** to internal registry

## Reference

- **Registry URL:** `https://conmcpregb001.livelysmoke-0f2aa8bb.eastus.azurecontainerapps.io`
- **Reusable Workflow:** `paychex/payx-mcp-registry/.github/workflows/internal-mcp-publish.yml`
- **Working Example:** [payx-demo-calculator-mcp](https://github.com/paychex/payx-demo-calculator-mcp)

## Support

- **Slack:** #mcp-registry-support
- **Email:** mcp-registry-team@paychex.com
- **Issues:** https://github.com/paychex/payx-mcp-registry/issues
