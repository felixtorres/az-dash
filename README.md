# az-dash

A rich terminal UI for Azure DevOps — pull requests, work items, and pipelines at a glance.

Inspired by [gh-dash](https://github.com/dlvhdr/gh-dash). Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **3 dashboard views** — Pull Requests, Work Items, Pipelines
- **Configurable sections** — define your own filters per view via YAML
- **Vim-style navigation** — j/k, g/G, Tab/Shift+Tab, Ctrl+d/u
- **Preview pane** — toggleable sidebar with details, reviewers, stages
- **Inline actions** — approve PRs, change work item state, cancel builds
- **Multi-org support** — per-section organization/project overrides
- **Auto-refresh** — configurable polling interval
- **Theming** — full color customization via YAML config

## Install

### From source

```bash
go install github.com/felixtorres/az-dash@latest
```

### From release

Download the binary for your platform from the [Releases](https://github.com/felixtorres/az-dash/releases) page.

## Setup

### Authentication

**Option 1: Azure CLI** (recommended)
```bash
az login
az-dash
```

**Option 2: Personal Access Token**
```bash
export AZ_DEVOPS_PAT=your-token-here
az-dash
```

Or set it in the config file:
```yaml
auth:
  method: pat
  pat: your-token-here
```

PAT scopes needed: Code (Read/Write), Work Items (Read/Write), Build (Read/Execute).

### Configuration

On first run, az-dash creates `~/.az-dash.yml` with defaults. Edit it:

```yaml
organization: my-org
project: my-project

prSections:
  - title: Mine
    filters:
      creatorId: "@me"
      status: active
  - title: Reviewing
    filters:
      reviewerId: "@me"
      status: active
  - title: All Active
    filters:
      status: active

workItemSections:
  - title: My Tasks
    wiql: >
      SELECT [System.Id] FROM WorkItems
      WHERE [System.AssignedTo] = @me
      AND [System.State] <> 'Closed'
      ORDER BY [System.ChangedDate] DESC

pipelineSections:
  - title: My Runs
    filters:
      requestedFor: "@me"
  - title: Failed
    filters:
      resultFilter: failed
  - title: All Recent
```

## Keyboard Shortcuts

### Global
| Key | Action |
|-----|--------|
| `1` / `2` / `3` | Switch view (PRs / Work Items / Pipelines) |
| `j` / `k` | Move down / up |
| `g` / `G` | First / last item |
| `Tab` / `Shift+Tab` | Next / previous section |
| `p` | Toggle preview pane |
| `r` / `R` | Refresh section / all |
| `o` | Open in browser |
| `y` | Copy URL |
| `#` | Copy ID |
| `q` | Quit |

### Pull Requests
| Key | Action |
|-----|--------|
| `v` | Approve |
| `m` | Complete (merge) |
| `x` / `X` | Abandon / reactivate |
| `d` | View diff (via delta or less) |

### Work Items
| Key | Action |
|-----|--------|
| `a` / `A` | Assign to me / unassign |
| `S` | Cycle state |

### Pipelines
| Key | Action |
|-----|--------|
| `x` | Cancel build |
| `l` | View logs |

## License

MIT
