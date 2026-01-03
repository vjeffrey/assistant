# Development Notes

## GitHub CLI Integration

### Authentication Issues and Solutions

When integrating the GitHub CLI (`gh`) with Go's `exec.Command`, there are several important considerations:

#### Problem 1: Environment Variables

**Issue**: The code was setting `GITHUB_TOKEN` as an environment variable, but the `gh` CLI doesn't recognize it.

**Solution**: Use `GH_TOKEN` instead. The `gh` CLI specifically looks for the `GH_TOKEN` environment variable for authentication.

```go
// Correct approach
cmd.Env = append(os.Environ(), "GH_TOKEN="+token)
```

#### Problem 2: Command Argument Quoting

**Issue**: When passing a search query as a single concatenated string to `exec.Command`, the shell quotes it incorrectly:

```go
// WRONG - causes shell quoting issues
searchQuery := fmt.Sprintf("org:%s assignee:%s is:issue is:open", org, username)
cmd := exec.Command("gh", "search", "issues", searchQuery, "--json", "...")
// Results in: gh search issues "org:mondoohq assignee:vjeffrey is:issue is:open"
// GitHub interprets this as an invalid quoted query
```

**Solution**: Pass each part of the query as separate arguments:

```go
// CORRECT - no shell quoting issues
orgQuery := fmt.Sprintf("org:%s", org)
assigneeQuery := fmt.Sprintf("assignee:%s", username)
cmd := exec.Command("gh", "search", "issues", orgQuery, assigneeQuery, "is:issue", "is:open", "--json", "...")
// Results in: gh search issues org:mondoohq assignee:vjeffrey is:issue is:open
```

#### Problem 3: Proper Environment Setup

**Issue**: When setting `cmd.Env`, you need to include the existing environment variables, not just the new one.

**Solution**: Always start with `os.Environ()`:

```go
func SetupGitHubToken(token string, cmdEnv []string) []string {
    if token == "" {
        return cmdEnv
    }
    // Start with existing environment or os.Environ()
    if cmdEnv == nil {
        cmdEnv = os.Environ()
    }
    return append(cmdEnv, "GH_TOKEN="+token)
}
```

### Testing Authentication

To test if authentication is working:

```bash
# Set the token and test a simple query
GH_TOKEN="your-token" gh search issues org:yourorg assignee:username is:issue is:open --json number,title --limit 5
```

If this works in the terminal but not in your Go code, check:
1. Are you using `GH_TOKEN` (not `GITHUB_TOKEN`)?
2. Are you passing query parts as separate arguments?
3. Are you including the existing environment with `os.Environ()`?

### Build Requirements

This project uses SQLite and requires CGO:

```bash
CGO_ENABLED=1 go build
```

**Important Notes:**
- Without CGO enabled, you'll get an error: "Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work."
- On macOS, XProtect may aggressively quarantine or delete CGO-enabled binaries built in the current directory
- **Recommended workaround**: Use `go install` instead:
  ```bash
  CGO_ENABLED=1 go install
  # Binary will be at ~/go/bin/assistant
  ```
- The `build.sh` script is provided for local builds, but binaries may be removed by macOS security on execution

## GitHub Projects v2 Integration

### Overview

The tool now supports filtering GitHub issues by project board and identifying stale issues using the GitHub Projects v2 GraphQL API.

### Key Implementation Details

**Authentication & Scopes**
- Requires `project` scope in addition to `repo` and `read:org`
- Add scope: `gh auth refresh -s project`
- Uses `GH_TOKEN` environment variable (same pattern as existing GitHub integration)

**Project Node IDs**
- Projects are identified by GraphQL node IDs (format: `PVT_...`)
- Cannot be queried directly from organization without `project` scope
- Helper script provided: `./get-project-id.sh <org> <project-number>`

**GraphQL API Structure**
```graphql
{
  node(id: "PVT_...") {
    ... on ProjectV2 {
      items(first: 100) {
        nodes {
          content {
            ... on Issue {
              number, title, assignees { nodes { login } }
            }
          }
          createdAt  # When added to project (used for "time on board")
        }
      }
    }
  }
}
```

**Filtering Approach**
- Server-side filtering by assignee not supported by GitHub API
- Solution: Fetch all project items, filter client-side by assignee
- Only shows Issues (not PRs or drafts) assigned to the user
- Time on board calculated from `ProjectItem.createdAt` (when added to project)

**Pagination**
- Uses cursor-based pagination (`pageInfo.hasNextPage`, `pageInfo.endCursor`)
- Fetches 100 items per page
- Continues until all items retrieved

### Usage

The project node ID is read from the `GITHUB_PROJECT` environment variable:

```bash
# View issues on project board (CLI)
GITHUB_PROJECT=PVT_kwDOABpK8s4ApRn8 ./assistant --github

# View stale issues (>3 weeks on board)
GITHUB_PROJECT=PVT_kwDOABpK8s4ApRn8 ./assistant --github --github-stale 3

# Filter by status field
GITHUB_PROJECT=PVT_kwDOABpK8s4ApRn8 ./assistant --github --filter-status '5-9 january'

# View recently merged PRs only (works with or without GITHUB_PROJECT set)
./assistant --merged
GITHUB_PROJECT=PVT_kwDOABpK8s4ApRn8 ./assistant --merged

# Start web UI (opens on http://localhost:8080)
./assistant --web 8080
GITHUB_PROJECT=PVT_kwDOABpK8s4ApRn8 ./assistant --web 8080

# Or set it in your environment
export GITHUB_PROJECT=PVT_kwDOABpK8s4ApRn8
./assistant --github
./assistant --github --github-stale 3
./assistant --merged
./assistant --web 8080
```

### Common Errors

**"Resource not accessible by personal access token"**
- Missing `project` scope
- Solution: `gh auth refresh -s project`

**"Could not resolve to a node with the global id"**
- Invalid project node ID
- Solution: Use `./get-project-id.sh` to get correct ID

### Environment Variables

- `GITHUB_PROJECT`: Project GraphQL node ID (format: `PVT_...`)
  - When set, the `--github` flag will filter issues to this project board
  - Replaces the previous `--github-project` command-line flag

### Web UI

A lightweight web interface is available for easier viewing of GitHub data:

**Features:**
- Clickable issue/PR links that open in new tabs
- Dark theme matching GitHub's UI
- Real-time refresh capability
- Displays project status for board issues
- Shows time on board for project items
- Responsive layout with proper sectioning

**Usage:**
```bash
./assistant --web 8080
# Opens at http://localhost:8080

# Development mode with fake data (for styling/development work)
./assistant --web-dev 8080
# Opens at http://localhost:8080 with mock data - no GitHub API calls
```

The web UI automatically displays:
- Project board issues (if `GITHUB_PROJECT` is set)
- Stale issues (>3 weeks on board)
- All assigned issues from configured orgs
- Recently merged PRs (last 12 hours)

**Development Mode:**
The `--web-dev` flag starts the web server with realistic fake data instead of making GitHub API calls. This is useful for:
- Working on CSS styling and layout changes
- Testing UI components without API rate limits
- Developing without network connectivity
- Quick iteration on visual changes

The mock data includes sample issues, PRs, labels, assignees, and project statuses that match the real data structure.

### Output Enhancements

**Project Status Display:**
- When viewing project board issues, the status field is now displayed in both CLI and web UI
- Status is extracted from the Project's "Status" field (single-select field type)
- Shows alongside other issue metadata like repository, state, and time on board

### Files Added/Modified

- `github.go`: Added `ProjectV2`, `ProjectItem` structs, query functions, and `ProjectStatus` field
- `cli.go`: Added `ShowGitHubAssignmentsWithOptions()` for project filtering with `--merged` flag
- `main.go`: Changed from `--github-project` flag to `GITHUB_PROJECT` env var, added `--web` and `--merged` flags
- `web.go`: New file - lightweight web UI with HTML templates and HTTP server
- `get-project-id.sh`: Helper script to retrieve project node IDs
- `GITHUB_PROJECTS.md`: Comprehensive documentation
