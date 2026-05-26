# gitlab-activity-cli

A simple, agent-friendly CLI tool for retrieving GitLab user activity. Designed for automation and agent consumption with structured output and minimal dependencies.

## Features

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
# Basic usage
./gitlab-activity-cli me --instance https://gitlab.com --token ~/.gitlab/token

# Specific date range
./gitlab-activity-cli me --days 2 --instance https://git.example.com --token ~/.gitlab/token

# Filter by project, JSON output
./gitlab-activity-cli me --project myproject --json --instance https://gitlab.com --token ~/.gitlab/token
```

### Get specific user's activity

```bash
./gitlab-activity-cli user username --days 3 --instance https://gitlab.com --token ~/.gitlab/token
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

The CLI auto-detects tokens from `~/.gitlab/` directory. For custom instances, specify both `-instance` and `-token` flags.

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

## SuperCLI Integration

This CLI is available as a [SuperCLI plugin](https://github.com/javimosch/supercli). SuperCLI is a unified plugin manager for CLI tools, making it easy to discover, install, and manage command-line utilities.

```bash
# Install via SuperCLI (once available)
sc plugins install gitlab-activity
```