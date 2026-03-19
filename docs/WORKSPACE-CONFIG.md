# Workspace Configuration

Share team defaults across your project using `.linear-workspace.yaml`.

## Overview

Workspace config allows teams to share default values for Linear commands:
- Default team for new issues
- Default labels applied to all issues
- Can be committed to version control

## Configuration File

Create `.linear-workspace.yaml` in your project root:

```yaml
defaults:
  team: Engineering
  labels:
    - triage
    - needs-review
```

## Usage

### Issue Creation

**Without workspace config:**
```bash
go-linear issue create --team=Engineering --label=triage --title="Fix bug"
```

**With workspace config:**
```bash
go-linear issue create --title="Fix bug"
# Automatically uses: team=Engineering, labels=[triage, needs-review]
```

**Override defaults:**
```bash
go-linear issue create --team=Platform --title="Fix bug"
# Uses team=Platform (overrides workspace default)
# Still applies labels=[triage, needs-review] from workspace
```

**Add to defaults:**
```bash
go-linear issue create --title="Fix bug" --label=urgent
# Uses: team=Engineering, labels=[triage, needs-review, urgent]
```

## Configuration Priority

Configs are merged in this order (highest to lowest priority):

1. **CLI flags** - Explicit `--team` or `--label` flags
2. **Workspace config** - `.linear-workspace.yaml` (current directory)
3. **User config** - `~/.config/linear/config.yaml`

## Example Workspace Setup

**`.linear-workspace.yaml`:**
```yaml
# Team defaults
defaults:
  team: Engineering
  labels:
    - triage

# Field display defaults
field_defaults:
  issue.list: "id,identifier,title,state.name,priority"
```

**Result:**
- All `issue create` commands default to Engineering team with "triage" label
- All `issue list` commands show 5 fields instead of 8

## Version Control

### Safe to Commit

```yaml
# .linear-workspace.yaml
defaults:
  team: Engineering
  labels: [bug, feature]
```

No secrets - safe to commit to Git.

### Gitignore Pattern

If you want to keep workspace config private:

```
.linear-workspace.yaml
```

## Supported Fields

Currently supports:
- `defaults.team` - Default team for issue creation
- `defaults.labels` - Default labels for issue creation
- `field_defaults` - Same as user config

## Future Extensions

- Saved filter queries
- Per-command defaults
- Workspace-specific MCP config
