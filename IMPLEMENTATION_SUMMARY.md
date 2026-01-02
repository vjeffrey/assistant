# GitHub Projects Integration - Implementation Summary

## Overview

I've successfully integrated GitHub Projects v2 functionality into your assistant tool. This allows you to:

1. Filter issues to only show those on a specific project board
2. Identify issues that have been on the board for more than a threshold (e.g., 3 weeks)
3. Track how long each issue has been on the board

## What Was Added

### 1. New Data Structures ([github.go](github.go))

Added several structs to handle GitHub Projects v2 data:

- `ProjectV2`: Represents a GitHub Project board
- `ProjectItem`: An item on the project (issue, PR, or draft)
- `ProjectItemContent`: The content data (issue/PR details)
- `ProjectFieldValue`: Custom field values on the project
- Enhanced `GitHubIssue` with `AddedToProjectAt` field to track time on board

### 2. New Functions ([github.go](github.go))

**`GetProjectItems(projectNodeID, token)`**
- Fetches all items from a GitHub Project v2
- Handles pagination automatically
- Uses GraphQL API through `gh api graphql`
- Returns issues, PRs, assignees, and custom fields

**`GetProjectIssuesForUser(projectNodeID, token)`**
- Filters project items to only Issues (not PRs)
- Filters to only issues assigned to the user
- Returns formatted GitHubIssue objects with time-on-board data

**`GetStaleProjectIssues(projectNodeID, token, staleDuration)`**
- Takes a duration threshold (e.g., 3 weeks)
- Returns issues that have been on the board longer than the threshold
- Useful for identifying work that may be stuck

### 3. CLI Updates ([cli.go](cli.go))

**`ShowGitHubAssignmentsWithOptions(projectNodeID, staleThresholdWeeks)`**
- New function that handles both standard and project-filtered views
- When `projectNodeID` is provided: shows project-filtered issues
- When `staleThresholdWeeks > 0`: also shows stale issues
- When `projectNodeID` is empty: shows standard view (all orgs)

**Updated `FormatIssues()`**
- Now displays "Time on board: X days" when available
- Helps visualize how long issues have been in the backlog

### 4. Command-Line Interface ([main.go](main.go))

Added two new flags:

```bash
--github-project <node_id>     # Filter to a specific project board
--github-stale <weeks>          # Show issues older than N weeks
```

Updated help text to document the new features.

### 5. Helper Scripts

**`get-project-id.sh`**
- Helper script to retrieve the GraphQL node ID for a project
- Usage: `./get-project-id.sh mondoohq 14`
- Outputs the project ID needed for the `--github-project` flag

### 6. Documentation

**Updated [README.md](README.md)**
- Added examples of project filtering
- Documented the new flags
- Explained token scope requirements

**New [GITHUB_PROJECTS.md](GITHUB_PROJECTS.md)**
- Comprehensive guide to the Projects integration
- Setup instructions
- Usage examples
- Troubleshooting guide
- Architecture explanation

## Usage Examples

### Basic Project Filtering

```bash
./assistant --github-project PVT_kwDOABpK8s4ApRn8
```

Shows only issues assigned to you on that specific project board.

### Find Stale Issues

```bash
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --github-stale 3
```

Shows all your issues on the board, plus a separate section for issues older than 3 weeks.

## Setup Required

To use this feature, you need to:

1. **Add the `project` scope to your GitHub token:**

   ```bash
   gh auth refresh -s project
   ```

2. **Get the project node ID:**

   ```bash
   ./get-project-id.sh mondoohq 14
   ```

   This will output something like: `PVT_kwDOABpK8s4ApRn8`

3. **Use the node ID with the new flags:**

   ```bash
   ./assistant --github-project PVT_kwDOABpK8s4ApRn8
   ```

## Technical Details

### API Approach

Uses GitHub's GraphQL API (v4) via the `gh` CLI tool:
- Query endpoint: `gh api graphql`
- Node-based querying: `node(id: "PVT_...")`
- Pagination support: Uses `pageInfo` and cursors

### Key Design Decisions

1. **Client-side filtering**: GitHub's API doesn't support server-side filtering by assignee on project items, so we fetch all items and filter locally

2. **Time calculation**: "Time on board" is calculated from when the item was added to the project (`item.createdAt`), not when the issue was originally created

3. **Issue-only**: Currently only shows Issues, not PRs or draft issues. This can be extended if needed.

4. **Pagination**: Fetches 100 items per page to handle large project boards efficiently

### GraphQL Query Structure

The query fetches:
- Item metadata (ID, timestamps)
- Content type (Issue, PullRequest, DraftIssue)
- Issue/PR details (number, title, URL, state, timestamps)
- Assignees (up to 10 per issue)
- Custom field values (dates, text, single-select)

## Next Steps

To start using this feature:

1. Run `gh auth refresh -s project` to add the required scope
2. Run `./get-project-id.sh mondoohq 14` to get the project ID
3. Test with: `./assistant --github-project <project_id>`
4. Try stale detection: `./assistant --github-project <project_id> --github-stale 3`

## Troubleshooting

If you encounter issues:

1. **"Resource not accessible"**: Your token needs the `project` scope
2. **"Could not resolve to a node"**: Verify the project ID is correct
3. **No issues shown**: Check that issues are assigned to `vjeffrey` and exist on the board

See [GITHUB_PROJECTS.md](GITHUB_PROJECTS.md) for detailed troubleshooting.

## Files Modified

- [github.go](github.go): +207 lines (new structs and functions)
- [cli.go](cli.go): +48 lines (project filtering logic)
- [main.go](main.go): +5 lines (new command-line flags)
- [README.md](README.md): Updated with project examples and setup
- [get-project-id.sh](get-project-id.sh): New helper script
- [GITHUB_PROJECTS.md](GITHUB_PROJECTS.md): New comprehensive documentation

## Build and Test

```bash
# Build
CGO_ENABLED=1 go build

# Test without project filtering (should work now)
./assistant --github

# Test with project filtering (requires project scope on token)
./assistant --github-project <your_project_id>

# Test stale issue detection
./assistant --github-project <your_project_id> --github-stale 3
```

## Resources

- [GitHub Projects v2 API Documentation](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects)
- [GitHub GraphQL API](https://docs.github.com/en/graphql)
- [gh CLI Documentation](https://cli.github.com/manual/)
