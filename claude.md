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

Without CGO enabled, you'll get an error: "Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work."
