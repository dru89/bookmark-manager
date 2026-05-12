# bk

A bookmark manager for people with too many internal tools. Store bookmarks in a markdown file, find them fast with fuzzy search, open them in your browser. Also ships an MCP server so your coding agent can look up and manage bookmarks for you.

## Install

```bash
go install github.com/dru89/bookmark-manager/cmd/bk@latest
```

Or clone and build:

```bash
git clone https://github.com/dru89/bookmark-manager.git
cd bookmark-manager
make install
```

## Configure

The config file lives in your OS config directory:

| OS | Path |
|----|------|
| macOS / Linux | `~/.config/bk/config.toml` |
| Windows | `%APPDATA%\bk\config.toml` |

Contents:

```toml
file = "~/path/to/Bookmarks.md"
```

This points to the markdown file where bookmarks are stored. Supports `~` and environment variables (`$HOME`, `%USERPROFILE%`).

You can also set `BK_FILE` as an environment variable or pass `--file` on any command.

## Usage

### Find and open a bookmark

```bash
bk              # opens interactive picker
bk jenkins      # picker pre-filtered with "jenkins"
bk jenkins deploy # narrows to entries matching both words
```

Select an entry and press Enter to open it in your browser. Escape or Ctrl-C to cancel.

### Add a bookmark

```bash
bk add "https://example.com/long-url" -n "Short name" -t "keyword1 keyword2 keyword3"
```

Tags are space-separated single words. Pick words you'd reach for when trying to remember this link later.

### List all bookmarks

```bash
bk list
```

### Remove a bookmark

```bash
bk remove "Short name"
```

Name matching is case-insensitive.

## MCP server

Start the MCP server for agent integration:

```bash
bk mcp --file ~/path/to/Bookmarks.md
```

This exposes five tools over stdio: `bookmark_search`, `bookmark_list`, `bookmark_add`, `bookmark_update`, `bookmark_remove`.

### OpenCode config

In `~/.config/opencode/opencode.json`:

```json
{
  "mcp": {
    "bookmarks": {
      "type": "local",
      "command": ["bk", "mcp", "--file", "/absolute/path/to/Bookmarks.md"],
      "enabled": true
    }
  }
}
```

### Claude Code config

In `~/.claude.json` under `mcpServers`:

```json
{
  "bookmarks": {
    "command": "bk",
    "args": ["mcp", "--file", "/absolute/path/to/Bookmarks.md"]
  }
}
```

## How search works

Each query word must match somewhere in an entry for it to appear. Results are ranked by match quality:

- Exact tag matches rank highest ("jenkins" matches the tag `jenkins`)
- Tag prefixes next ("data" matches the tag `databricks`)
- Name and URL substrings in the middle
- Fuzzy subsequence matching on the name as a last resort ("dms" finds "Device Management Service")

Multiple words narrow results (AND logic) and their scores add up, so more specific queries surface better results.

## File format

The bookmark file is a standard markdown table. You can edit it by hand or through any markdown editor:

```markdown
| Name | Tags | URL |
|------|------|-----|
| Jenkins | ci cd deploy pipeline | https://jenkins.internal.example.com/ |
| Expense reports | expenses receipts reimbursement finance | https://expenses.example.com/submit |
```
