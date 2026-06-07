# cs-mount

Mount GitHub Codespaces as Windows drive letters.

## Quick Start

```bash
# List codespaces
cs-mount list

# Mount (interactive)
cs-mount mount

# Unmount
cs-mount unmount

# Status
cs-mount status

# Open in IDE
cs-mount open
```

## How it works

1. **List** — Fetch codespaces from GitHub
2. **Mount** — Create SSH tunnel → setup rclone SFTP → mount as drive
3. **Unmount** — Kill processes, cleanup config, update state
4. **Status** — Check tunnel and mount health
5. **Open** — Launch VS Code, IntelliJ, or Zed with SSH remote

## Requirements

- `gh` — GitHub CLI (authenticated)
- `rclone` — Must be in PATH
- SSH key — `~/.ssh/codespaces.auto` (or custom via `--key-file`)

## Flags

- `--start-port N` — Starting SSH tunnel port (default: 2223)
- `--key-file PATH` — SSH private key path
- `--combine-remote NAME` — rclone combine remote name
- `-f, --force` — Skip confirmations

## Architecture

```
cmd/
  ├── list.go         List codespaces
  ├── mount.go        Mount orchestration
  ├── unmount.go      Cleanup
  ├── status.go       Health check
  └── open.go         Launch IDE

internal/
  ├── codespace/      GitHub interaction
  ├── tunnel/         SSH tunnel management
  ├── rclone/         rclone config & mount
  ├── state/          Persistent state (JSON)
  ├── ui/             Interactive prompts
  └── ide/            IDE detection
```

## License

MIT
