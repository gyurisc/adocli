# 🔧 ado — Azure DevOps in your terminal

Fast, script-friendly CLI for Azure DevOps. Built for developers and AI agents who live in the terminal.

Single binary, zero dependencies, JSON-first output. Like [`gh`](https://cli.github.com/) but for Azure DevOps.

## Features

- **Work Items** — query, create, update, search with WIQL
- **Repos** — list repos, manage pull requests, browse branches
- **Pipelines** — list, trigger, stream logs in real-time
- **Boards** — sprints, backlog, iterations
- **Multiple orgs** — switch between organizations seamlessly
- **On-prem support** — works with Azure DevOps Server (not just cloud)
- **JSON output** — pipe-friendly, automation-ready
- **Secure auth** — PAT tokens stored in OS keyring

## Install

### Homebrew (macOS/Linux)

```bash
brew install gyurisc/tap/adocli
```

### Windows (Scoop)

```bash
scoop bucket add adocli https://github.com/gyurisc/scoop-bucket
scoop install adocli
```

### From source

```bash
git clone https://github.com/gyurisc/adocli.git
cd adocli && make install
```

### Download binary

Grab the latest release for your platform from [Releases](https://github.com/gyurisc/adocli/releases).

## Quick Start

```bash
# Authenticate with a Personal Access Token
ado auth login --org https://dev.azure.com/myorg

# Check auth status
ado auth status
```

### Work Items

```bash
# List recent work items
ado workitem list --project MyProject

# Show a specific work item
ado workitem show 1234

# Create a user story
ado workitem create --type "User Story" --title "Add dark mode" --project MyProject

# Update a work item
ado workitem update 1234 --state "Active" --assign "me"

# Query with WIQL
ado workitem query "SELECT [Id], [Title] FROM WorkItems WHERE [State] = 'Active'"
```

### Repos & Pull Requests

```bash
# List repos
ado repos list --project MyProject

# List open pull requests
ado pr list --project MyProject

# Create a pull request
ado pr create --title "Fix login bug" --source feature/fix-login --target main

# Show PR details
ado pr show 42

# Approve a PR
ado pr approve 42
```

### Pipelines

```bash
# List pipelines
ado pipelines list --project MyProject

# Trigger a pipeline run
ado pipelines run --id 42 --branch main

# Stream pipeline logs
ado pipelines logs --run-id 123 --follow

# List recent runs
ado pipelines runs --id 42 --top 10
```

### Boards & Sprints

```bash
# List sprints
ado boards sprints --project MyProject --team "My Team"

# Show current sprint backlog
ado boards backlog --project MyProject --current
```

## Configuration

Config is stored in `~/.config/ado/config.json`. Set defaults to skip repetitive flags:

```bash
# Set default org and project
ado config set org https://dev.azure.com/myorg
ado config set project MyProject

# Now just:
ado workitem list
```

## Output Formats

```bash
# Table (default, human-friendly)
ado workitem list

# JSON (for scripting and piping)
ado workitem list --json

# Plain (minimal, one value per line)
ado workitem list --plain
```

## Why not `az devops`?

|  | `ado` | `az devops` |
|---|---|---|
| **Startup time** | ~10ms | ~2s |
| **Dependencies** | None (single binary) | Python + Azure CLI |
| **On-prem support** | ✅ | ❌ |
| **JSON-first** | ✅ | Partial |
| **Agent-friendly** | ✅ | ❌ |
| **Cross-platform binary** | ✅ | Requires Python |
| **Secure token storage** | OS keyring | File-based fallback |

## Azure DevOps API Coverage

| Area | Status |
|------|--------|
| Work Items (CRUD + WIQL) | ✅ |
| Repos | ✅ |
| Pull Requests | ✅ |
| Pipelines | ✅ |
| Boards & Sprints | ✅ |
| Test Plans | 🔜 |
| Artifacts | 🔜 |
| Wiki | 🔜 |

## Authentication

### Personal Access Token (recommended)

```bash
ado auth login --org https://dev.azure.com/myorg
# Paste your PAT when prompted
```

Create a PAT at: `https://dev.azure.com/{org}/_usersSettings/tokens`

### Azure DevOps Server (on-prem)

```bash
ado auth login --org https://devops.mycompany.com/tfs
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
# Development setup
git clone https://github.com/gyurisc/adocli.git
cd adocli
make build
make test
```

## License

[MIT](LICENSE) — use it however you want.

## Acknowledgements

Inspired by [gh](https://github.com/cli/cli) and [gogcli](https://github.com/steipete/gogcli).

---

Built with ❤️ by [Krisztian Gyuris](https://github.com/gyurisc)