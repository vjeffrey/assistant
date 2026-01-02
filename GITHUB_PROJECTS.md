# GitHub Projects Integration

This document explains how to use the GitHub Projects v2 integration to filter issues by project board and identify stale issues.

## Features

1. **Project Filtering**: View only issues assigned to you that are on a specific project board
2. **Stale Issue Detection**: Identify issues that have been on a project board for longer than a threshold (e.g., 3 weeks)
3. **Time Tracking**: See how long each issue has been on the board

## Setup

### Step 1: Add Project Scope to Your Token

The GitHub Projects v2 API requires the `project` scope. If you're already authenticated with `gh`, you can add this scope:

```bash
# Check current authentication
gh auth status

# Add the project scope
gh auth refresh -s project
```

Alternatively, create a new personal access token at [github.com/settings/tokens](https://github.com/settings/tokens) with these scopes:
- `repo` (for private repositories)
- `read:org` (for organization access)
- `project` (for GitHub Projects v2)

Then set the token as an environment variable:

```bash
export GITHUB_TOKEN_ASSISTANT_MONDOOHQ="ghp_your_token_here"

# Add to your shell profile for persistence
echo 'export GITHUB_TOKEN_ASSISTANT_MONDOOHQ="ghp_your_token_here"' >> ~/.zshrc
source ~/.zshrc
```

### Step 2: Get the Project Node ID

GitHub Projects v2 uses GraphQL node IDs (format: `PVT_...`). To get the ID for your project:

#### Option A: Use the helper script

```bash
./get-project-id.sh mondoohq 14
```

This will output:
```
Project ID: PVT_kwDOABpK8s4ApRn8
Title: Engineering Sprint Board
URL: https://github.com/orgs/mondoohq/projects/14
```

#### Option B: Query manually with gh CLI

```bash
GH_TOKEN="${GITHUB_TOKEN_ASSISTANT_MONDOOHQ}" gh api graphql -f query='
{
  organization(login: "mondoohq") {
    projectV2(number: 14) {
      id
      title
      url
    }
  }
}
' | jq -r '.data.organization.projectV2 | "Project ID: \(.id)\nTitle: \(.title)\nURL: \(.url)"'
```

#### Option C: Get it from the GitHub UI

1. Go to the project in your browser: https://github.com/orgs/mondoohq/projects/14
2. Open browser developer tools (F12)
3. Look for GraphQL requests in the Network tab
4. Find requests containing `projectV2` and look for the `id` field starting with `PVT_`

## Usage

### View Issues on a Project Board

```bash
./assistant --github-project PVT_kwDOABpK8s4ApRn8
```

Output:
```
🔍 Fetching GitHub assignments...

📊 Fetching issues from project board...

📋 Issues assigned to vjeffrey on project board:
================================================================================

#1234 - Implement user authentication
  Repository: mondoohq/server
  URL: https://github.com/mondoohq/server/issues/1234
  State: OPEN
  Updated: 2025-12-20 10:30
  Time on board: 11 days

#5678 - Fix database migration
  Repository: mondoohq/console
  URL: https://github.com/mondoohq/console/issues/5678
  State: OPEN
  Updated: 2025-12-15 14:20
  Time on board: 16 days

Total: 2 issue(s)
```

### Find Stale Issues (>3 weeks)

```bash
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --github-stale 3
```

Output:
```
🔍 Fetching GitHub assignments...

📊 Fetching issues from project board...

📋 Issues assigned to vjeffrey on project board:
================================================================================

#1234 - Implement user authentication
  Repository: mondoohq/server
  URL: https://github.com/mondoohq/server/issues/1234
  State: OPEN
  Updated: 2025-12-20 10:30
  Time on board: 11 days

#5678 - Fix database migration
  Repository: mondoohq/console
  URL: https://github.com/mondoohq/console/issues/5678
  State: OPEN
  Updated: 2025-12-15 14:20
  Time on board: 16 days

Total: 2 issue(s)

⏰ Issues on board for more than 3 weeks:
================================================================================

#9012 - Refactor authentication module
  Repository: mondoohq/server
  URL: https://github.com/mondoohq/server/issues/9012
  State: OPEN
  Updated: 2025-11-25 09:15
  Time on board: 36 days

Total: 1 stale issue(s)
```

## How It Works

### Data Retrieved

For each issue on the project board, the tool fetches:
- Issue number, title, URL, state
- Repository name
- Created and updated timestamps
- When the issue was added to the project board (used to calculate "time on board")
- Assignees (filtered to show only issues assigned to you)
- Custom field values (dates, text, single-select fields)

### Filtering Logic

1. **User Filtering**: Only shows issues where you are an assignee
2. **Type Filtering**: Only shows Issues (not PRs or draft issues)
3. **Stale Filtering**: When `--github-stale N` is used, shows issues where:
   - `time_now - added_to_project_at > N weeks`

### Pagination

The tool automatically handles pagination to retrieve all items from the project board (up to thousands of items).

## Common Use Cases

### Daily Standup Preparation

```bash
# Quick view of your current work
./assistant --github-project PVT_kwDOABpK8s4ApRn8
```

### Sprint Planning

```bash
# Identify issues that might need attention
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --github-stale 2
```

### Project Health Check

```bash
# Find issues that have been stale for a month
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --github-stale 4
```

## Troubleshooting

### Error: "Resource not accessible by personal access token"

Your token doesn't have the `project` scope. Solutions:

1. Add the scope: `gh auth refresh -s project`
2. Or create a new token with the `project` scope at github.com/settings/tokens

### Error: "GraphQL error: Could not resolve to a node with the global id"

The project node ID is incorrect. Verify:

1. The project exists and you have access to it
2. You're using the correct node ID (format: `PVT_...`)
3. Try getting the ID again using `./get-project-id.sh`

### No issues shown

Possible reasons:

1. No issues on the board are assigned to you
2. The username in the code (`vjeffrey`) doesn't match your GitHub username
3. All issues on the board are PRs or draft issues (only Issues are shown)

To debug, check the raw data:

```bash
GH_TOKEN="${GITHUB_TOKEN_ASSISTANT_MONDOOHQ}" gh api graphql -f query='{
  node(id: "PVT_kwDOABpK8s4ApRn8") {
    ... on ProjectV2 {
      items(first: 5) {
        nodes {
          content {
            __typename
            ... on Issue {
              number
              title
              assignees(first: 10) {
                nodes {
                  login
                }
              }
            }
          }
        }
      }
    }
  }
}'
```

## Code Architecture

The implementation consists of:

### Data Structures ([github.go](github.go))

- `ProjectV2`: Represents a GitHub Project v2 board
- `ProjectItem`: An item on the project board (issue, PR, or draft)
- `ProjectItemContent`: The actual issue/PR data
- `ProjectFieldValue`: Custom field values on the project

### Functions

- `GetProjectItems()`: Fetches all items from a project board (with pagination)
- `GetProjectIssuesForUser()`: Filters project items to issues assigned to a user
- `GetStaleProjectIssues()`: Filters to issues older than a threshold
- `FormatIssues()`: Displays issues with "time on board" information

### CLI Integration ([cli.go](cli.go), [main.go](main.go))

- `ShowGitHubAssignmentsWithOptions()`: Main entry point with project filtering
- Command-line flags: `--github-project`, `--github-stale`

## Limitations

1. **Server-side filtering**: The GitHub API doesn't support filtering by assignee at the query level, so we fetch all items and filter client-side
2. **Time on board**: Calculated from when the item was added to the project, not when the issue was created
3. **Pagination**: Fetches 100 items per page; very large boards (>1000 items) may be slow
4. **Single user**: Hardcoded to user `vjeffrey`; would need modification for other users

## Future Enhancements

Potential improvements:

- Support for filtering by custom field values (Status, Priority, etc.)
- Support for multiple projects in one command
- Export to CSV or JSON
- Integration with the daemon to send notifications for stale issues
- Support for viewing PR assignments on project boards
- Configurable username (currently hardcoded)
