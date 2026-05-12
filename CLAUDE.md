# bookmark-manager

A fuzzy bookmark manager backed by a markdown table. CLI tool (`bk`) with an MCP server for agent integration.

## Architecture

```
cmd/bk/              CLI entrypoint (cobra commands, open helper)
internal/
  bookmarks/         Markdown table parser/writer, token-based search
  config/            TOML config loader (~, $HOME expansion)
  picker/            Bubbletea interactive fuzzy finder
  mcp/               MCP stdio server (5 tools)
```

### Data format

Bookmarks are stored as a markdown table in a single file:

```markdown
| Name | Tags | URL |
|------|------|-----|
| Jenkins | ci cd deploy pipeline | https://jenkins.internal.example.com/ |
```

- **Name**: short human-readable label
- **Tags**: space-separated single-word search keywords
- **URL**: the full URL (anchors, query params preserved)

The file path is configurable via `--file` flag, `~/.config/bk/config.toml`, or `BK_FILE` env var.

### Scoring (picker)

The bubbletea picker uses a tiered scoring model, not raw fzf-style subsequence matching. Each query word is scored independently against each entry, then scores are summed. All query words must match (AND semantics).

| Score | Match type |
|-------|-----------|
| 300 | Exact tag match (query token = a tag) |
| 200 | Tag prefix (query token is prefix of a tag) |
| 150 | Name substring (query token appears in name) |
| 100 | Tag substring (query token contained within a tag) |
| 50 | URL substring (query token appears in URL) |
| 1+ | Name fuzzy (subsequence match on name only, fallback) |

Fuzzy subsequence matching only applies to the name field and only as a last resort. Tags are never fuzzy-matched across boundaries — each tag is treated as an atomic token.

### MCP tools

- `bookmark_search(query)` — token-based substring search on name+tags. Returns matching entries.
- `bookmark_list()` — returns all entries. Fallback when search is too narrow.
- `bookmark_add(url, name, tags?)` — add a new bookmark. Agent infers name/tags from context.
- `bookmark_update(name, new_name?, new_tags?, new_url?)` — update by current name.
- `bookmark_remove(name)` — delete by name (case-insensitive).

The MCP search uses simple token substring matching (not the tiered scoring). Each query word must appear somewhere in name+tags. This is intentional — the agent reasons about relevance itself, so it just needs candidate filtering.

## Development

```bash
make build      # compile to ./bk
make install    # install to $GOBIN/bk
make test       # run tests
make deps       # tidy and download modules
```

## Config

`~/.config/bk/config.toml`:

```toml
file = "~/path/to/Bookmarks.md"
```

Supports `~`, `$HOME`, and `${VAR}` expansion. The `--file` flag overrides config, and `BK_FILE` env var is lowest precedence.

## CLI usage

```bash
bk                  # interactive picker, opens selected URL in browser
bk jenkins          # picker pre-filtered with "jenkins"
bk add <url> -n "Name" -t "tag1 tag2"
bk remove "Name"
bk list             # print all bookmarks
bk mcp              # start MCP server on stdio
```
