# gitlab-activity-cli

A simple, agent-friendly CLI tool for retrieving GitLab user activity. Designed for automation and agent consumption with structured output and minimal dependencies.

## Features

- **Auto-detect GitLab instance** from token files
- **Flexible time filtering** (days, date ranges)
- **Project filtering** 
- **Multiple output formats** (structured text, JSON)
- **Agent-friendly design** (non-interactive, semantic exit codes)
- **Zero external dependencies** (Go stdlib only)

## Installation

```bash
cd ~/ai/supercli-clis/gitlab-activity-cli
go build -ldflags "-s -w" -o gitlab-activity-cli main.go
```

## Usage

### Get current user's activity

```bash
# Default: last 7 days, auto-detect instance
./gitlab-activity-cli me

# Specific date range on self-hosted
./gitlab-activity-cli me --days 2 --instance https://git.geored.fr

# Filter by project, JSON output
./gitlab-activity-cli me --project georedv3 --json
```

### Get specific user's activity

```bash
./gitlab-activity-cli user jarancibia --days 3
```

### Options

- `-days int` - Number of days to look back (default: 7)
- `-instance string` - GitLab instance URL (default: auto-detect)
- `-token string` - Path to token file (default: auto-detect)
- `-project string` - Filter by project name
- `-since string` - Start date (YYYY-MM-DD)
- `-until string` - End date (YYYY-MM-DD)
- `-json` - Output in JSON format

## Token Configuration

The CLI auto-detects tokens from `~/.gitlab/`:
- `jar-token` → gitlab.com
- `geored` → git.geored.fr

For custom instances, specify both `-instance` and `-token`.

## Output Formats

### Default (structured text)

```
Total events: 8

georedv3 (4 events):
  2026-05-22 16:46:02 - pushed to
    refactor (shared) Replace b-btn with button in TableButton.vue
  2026-05-22 08:40:15 - pushed to
    fix (submenu) Restore v-if feature-right guards on nav items
```

### JSON

```json
{
  "version": "1.0",
  "user": "jarancibia",
  "instance": "https://git.geored.fr",
  "period": {"days": 7},
  "total_events": 8,
  "projects": [...]
}
```

## Exit Codes

- `0` - Success
- `85` - Invalid arguments
- `92` - Resource not found
- `100` - API error
- `110` - Internal error

## Agent-Friendly Design

- **Non-interactive**: No prompts, all flags
- **Semantic exit codes**: For automated decision-making
- **Stable JSON schema**: Versioned output format
- **Structured output**: Line-based text for grep/awk, JSON for parsing

## License

MIT