package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type GitHubRepository struct {
	NameWithOwner string `json:"nameWithOwner"`
}

type GitHubIssue struct {
	Number           int              `json:"number"`
	Title            string           `json:"title"`
	Repository       GitHubRepository `json:"repository"`
	URL              string           `json:"url"`
	State            string           `json:"state"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	RepoName         string           // Extracted repository name
	AddedToProjectAt time.Time        // When the issue was added to a project
	ProjectStatus    string           // Status field from project board
}

type GitHubPR struct {
	Number     int              `json:"number"`
	Title      string           `json:"title"`
	Repository GitHubRepository `json:"repository"`
	URL        string           `json:"url"`
	State      string           `json:"state"`
	MergedAt   time.Time        `json:"mergedAt"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
	RepoName   string           // Extracted repository name
}

type GitHubManager struct {
	username string
}

// ProjectItemContent represents the content of a project item (issue or PR)
type ProjectItemContent struct {
	Type      string    `json:"__typename"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	State     string    `json:"state"`
	Assignees struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"assignees"`
}

// ProjectItem represents an item in a GitHub Project
type ProjectItem struct {
	ID          string             `json:"id"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	Content     ProjectItemContent `json:"content"`
	FieldValues struct {
		Nodes []ProjectFieldValue `json:"nodes"`
	} `json:"fieldValues"`
}

// ProjectFieldValue represents a custom field value in a project
type ProjectFieldValue struct {
	Type  string `json:"__typename"`
	Date  string `json:"date,omitempty"` // Date string in YYYY-MM-DD format
	Text  string `json:"text,omitempty"`
	Name  string `json:"name,omitempty"`
	Field struct {
		Name string `json:"name"`
	} `json:"field"`
}

// ProjectV2 represents a GitHub Project v2
type ProjectV2 struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Items struct {
		Nodes    []ProjectItem `json:"nodes"`
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
	} `json:"items"`
}

func NewGitHubManager(username string) *GitHubManager {
	return &GitHubManager{username: username}
}

// SetupGitHubToken sets up the token as an environment variable for gh commands
func (g *GitHubManager) SetupGitHubToken(token string, cmdEnv []string) []string {
	if token == "" {
		return cmdEnv
	}
	// gh CLI recognizes GH_TOKEN environment variable
	if cmdEnv == nil {
		cmdEnv = os.Environ()
	}
	return append(cmdEnv, "GH_TOKEN="+token)
}

// GetProjectItems fetches all items from a GitHub Project v2
// projectNodeID is the GraphQL node ID of the project (format: PVT_...)
// This requires the token to have 'project' scope
func (g *GitHubManager) GetProjectItems(projectNodeID string, token string) ([]ProjectItem, error) {
	var allItems []ProjectItem
	cursor := ""
	hasNextPage := true

	for hasNextPage {
		// Build the GraphQL query
		afterClause := ""
		if cursor != "" {
			afterClause = fmt.Sprintf(`, after: "%s"`, cursor)
		}

		query := fmt.Sprintf(`{
  node(id: "%s") {
    ... on ProjectV2 {
      items(first: 100%s) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          id
          createdAt
          updatedAt
          content {
            __typename
            ... on Issue {
              number
              title
              url
              state
              createdAt
              updatedAt
              assignees(first: 10) {
                nodes {
                  login
                }
              }
            }
            ... on PullRequest {
              number
              title
              url
              state
              createdAt
              updatedAt
              assignees(first: 10) {
                nodes {
                  login
                }
              }
            }
          }
          fieldValues(first: 20) {
            nodes {
              __typename
              ... on ProjectV2ItemFieldDateValue {
                date
                field {
                  ... on ProjectV2Field {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldTextValue {
                text
                field {
                  ... on ProjectV2Field {
                    name
                  }
                }
              }
              ... on ProjectV2ItemFieldSingleSelectValue {
                name
                field {
                  ... on ProjectV2Field {
                    name
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`, projectNodeID, afterClause)

		cmd := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", query))
		cmd.Env = g.SetupGitHubToken(token, nil)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to query project: %w (output: %s)", err, string(output))
		}

		// Parse the response
		var response struct {
			Data struct {
				Node ProjectV2 `json:"node"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}

		if err := json.Unmarshal(output, &response); err != nil {
			return nil, fmt.Errorf("failed to parse project response: %w", err)
		}

		if len(response.Errors) > 0 {
			return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
		}

		allItems = append(allItems, response.Data.Node.Items.Nodes...)
		hasNextPage = response.Data.Node.Items.PageInfo.HasNextPage
		cursor = response.Data.Node.Items.PageInfo.EndCursor
	}

	return allItems, nil
}

// hasExcludedStatus checks if an item has a status that should be excluded
func hasExcludedStatus(item ProjectItem) bool {
	excludedStatuses := []string{
		"set for development",
		"needs code review",
		"waiting for customer feedback",
		"customer reported",
	}

	// Look for Status field in field values
	for _, fieldValue := range item.FieldValues.Nodes {
		if fieldValue.Type == "ProjectV2ItemFieldSingleSelectValue" {
			statusValue := strings.ToLower(fieldValue.Name)
			return slices.Contains(excludedStatuses, statusValue)
		}
	}
	return false
}

// hasMatchingStatus checks if an item has a status that matches the filter
// Returns true if the item has a status field matching the provided filterStatus (case-insensitive)
// If filterStatus is empty, always returns true (no filtering)
func hasMatchingStatus(item ProjectItem, filterStatus string) bool {
	if filterStatus == "" {
		return true
	}

	filterStatusLower := strings.ToLower(filterStatus)

	// Look for Status field in field values
	for _, fieldValue := range item.FieldValues.Nodes {
		if fieldValue.Type == "ProjectV2ItemFieldSingleSelectValue" {
			statusValue := strings.ToLower(fieldValue.Name)
			if statusValue == filterStatusLower {
				return true
			}
		}
	}
	return false
}

// GetProjectIssuesForUser filters project items to only include issues assigned to the specified user
func (g *GitHubManager) GetProjectIssuesForUser(projectNodeID string, token string, filterStatus string) ([]GitHubIssue, error) {
	items, err := g.GetProjectItems(projectNodeID, token)
	if err != nil {
		return nil, err
	}

	var issues []GitHubIssue
	for _, item := range items {
		// Skip if not an Issue
		if item.Content.Type != "Issue" {
			continue
		}

		// Skip closed issues
		if strings.ToUpper(item.Content.State) == "CLOSED" {
			continue
		}

		// Skip issues with excluded statuses
		if hasExcludedStatus(item) {
			continue
		}

		// Filter by status if specified
		if !hasMatchingStatus(item, filterStatus) {
			continue
		}

		// Check if user is assigned
		isAssigned := false
		for _, assignee := range item.Content.Assignees.Nodes {
			if assignee.Login == g.username {
				isAssigned = true
				break
			}
		}

		if !isAssigned {
			continue
		}

		// Extract repository name from URL
		repoName := ""
		parts := strings.Split(item.Content.URL, "/")
		if len(parts) >= 5 {
			repoName = parts[3] + "/" + parts[4]
		}

		// Extract project status
		projectStatus := ""
		for _, fieldValue := range item.FieldValues.Nodes {
			if fieldValue.Type == "ProjectV2ItemFieldSingleSelectValue" && fieldValue.Field.Name == "Status" {
				projectStatus = fieldValue.Name
				break
			}
		}

		issue := GitHubIssue{
			Number:           item.Content.Number,
			Title:            item.Content.Title,
			URL:              item.Content.URL,
			State:            item.Content.State,
			CreatedAt:        item.Content.CreatedAt,
			UpdatedAt:        item.Content.UpdatedAt,
			RepoName:         repoName,
			AddedToProjectAt: item.CreatedAt, // When the item was added to the project
			ProjectStatus:    projectStatus,
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// GetStaleProjectIssues returns issues that have been on the project board for more than the specified duration
// This includes ALL issues (not just those assigned to the user)
func (g *GitHubManager) GetStaleProjectIssues(projectNodeID string, token string, staleDuration time.Duration, filterStatus string) ([]GitHubIssue, error) {
	items, err := g.GetProjectItems(projectNodeID, token)
	if err != nil {
		return nil, err
	}

	var staleIssues []GitHubIssue
	now := time.Now()

	for _, item := range items {
		// Skip if not an Issue
		if item.Content.Type != "Issue" {
			continue
		}

		// Skip closed issues
		if strings.ToUpper(item.Content.State) == "CLOSED" {
			continue
		}

		// Skip issues with excluded statuses
		if hasExcludedStatus(item) {
			continue
		}

		// Filter by status if specified
		if !hasMatchingStatus(item, filterStatus) {
			continue
		}

		// Calculate how long the issue has been on the board
		timeOnBoard := now.Sub(item.CreatedAt)

		// Only include stale issues
		if timeOnBoard <= staleDuration {
			continue
		}

		// Extract repository name from URL
		repoName := ""
		parts := strings.Split(item.Content.URL, "/")
		if len(parts) >= 5 {
			repoName = parts[3] + "/" + parts[4]
		}

		// Extract project status
		projectStatus := ""
		for _, fieldValue := range item.FieldValues.Nodes {
			if fieldValue.Type == "ProjectV2ItemFieldSingleSelectValue" && fieldValue.Field.Name == "Status" {
				projectStatus = fieldValue.Name
				break
			}
		}

		issue := GitHubIssue{
			Number:           item.Content.Number,
			Title:            item.Content.Title,
			URL:              item.Content.URL,
			State:            item.Content.State,
			CreatedAt:        item.Content.CreatedAt,
			UpdatedAt:        item.Content.UpdatedAt,
			RepoName:         repoName,
			AddedToProjectAt: item.CreatedAt, // When the item was added to the project
			ProjectStatus:    projectStatus,
		}

		staleIssues = append(staleIssues, issue)
	}

	return staleIssues, nil
}

// GetAssignedIssues fetches all issues assigned to the user from specified organizations
func (g *GitHubManager) GetAssignedIssues(org string, token string) ([]GitHubIssue, error) {
	var allIssues []GitHubIssue

	// Use gh search to find issues - this includes the repo URL
	// Pass each part of the query as separate arguments to avoid shell quoting issues
	orgQuery := fmt.Sprintf("org:%s", org)
	assigneeQuery := fmt.Sprintf("assignee:%s", g.username)
	cmd := exec.Command("gh", "search", "issues", orgQuery, assigneeQuery, "is:issue", "is:open", "--json", "number,title,url,state,createdAt,updatedAt", "--limit", "100")
	cmd.Env = g.SetupGitHubToken(token, nil)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If search fails, return empty results with the error
		return allIssues, fmt.Errorf("failed to search issues in %s: %w (output: %s)", org, err, string(output))
	}

	var issues []GitHubIssue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues from %s: %w", org, err)
	}

	// Extract repository name from URL
	for i := range issues {
		// URL format: https://github.com/org/repo/issues/123
		parts := strings.Split(issues[i].URL, "/")
		if len(parts) >= 5 {
			issues[i].RepoName = parts[3] + "/" + parts[4]
		}
	}

	allIssues = append(allIssues, issues...)

	return allIssues, nil
}

// GetMentionedIssuesAndPRs fetches all issues and PRs where the user is mentioned
func (g *GitHubManager) GetMentionedIssuesAndPRs(org string, token string) ([]GitHubIssue, []GitHubPR, error) {
	var allIssues []GitHubIssue
	var allPRs []GitHubPR

	// Search for issues mentioning the user
	orgQuery := fmt.Sprintf("org:%s", org)
	mentionsQuery := fmt.Sprintf("mentions:%s", g.username)
	issueCmd := exec.Command("gh", "search", "issues", orgQuery, mentionsQuery, "is:issue", "--json", "number,title,url,state,createdAt,updatedAt", "--limit", "100")
	issueCmd.Env = g.SetupGitHubToken(token, nil)

	issueOutput, err := issueCmd.Output()
	if err == nil {
		var issues []GitHubIssue
		if err := json.Unmarshal(issueOutput, &issues); err != nil {
			return nil, nil, fmt.Errorf("failed to parse mentioned issues from %s: %w", org, err)
		}

		// Extract repository name from URL
		for i := range issues {
			parts := strings.Split(issues[i].URL, "/")
			if len(parts) >= 5 {
				issues[i].RepoName = parts[3] + "/" + parts[4]
			}
		}

		allIssues = append(allIssues, issues...)
	}

	// Search for PRs mentioning the user
	prCmd := exec.Command("gh", "search", "prs", orgQuery, mentionsQuery, "is:pr", "--json", "number,title,url,state,createdAt,updatedAt", "--limit", "100")
	prCmd.Env = g.SetupGitHubToken(token, nil)

	prOutput, err := prCmd.Output()
	if err == nil {
		var prs []GitHubPR
		if err := json.Unmarshal(prOutput, &prs); err != nil {
			return nil, nil, fmt.Errorf("failed to parse mentioned PRs from %s: %w", org, err)
		}

		// Extract repository name from URL
		for i := range prs {
			parts := strings.Split(prs[i].URL, "/")
			if len(parts) >= 5 {
				prs[i].RepoName = parts[3] + "/" + parts[4]
			}
		}

		allPRs = append(allPRs, prs...)
	}

	return allIssues, allPRs, nil
}

// GetRecentMergedPRs fetches PRs merged in the last 12 hours from specified repositories
func (g *GitHubManager) GetRecentMergedPRs(repos []string, hoursAgo int, token string) ([]GitHubPR, error) {
	var allPRs []GitHubPR

	for _, repo := range repos {
		// Use gh CLI to list recently merged PRs
		cmd := exec.Command("gh", "pr", "list",
			"--repo", repo,
			"--state", "merged",
			"--json", "number,title,url,state,mergedAt,createdAt,updatedAt",
			"--limit", "100",
		)
		cmd.Env = g.SetupGitHubToken(token, nil)

		output, err := cmd.Output()
		if err != nil {
			// If repo doesn't exist or access denied, skip it
			continue
		}

		var prs []GitHubPR
		if err := json.Unmarshal(output, &prs); err != nil {
			return nil, fmt.Errorf("failed to parse merged PRs from %s: %w", repo, err)
		}

		// Extract repository name and filter by merge time
		for i := range prs {
			prs[i].RepoName = repo

			// Only include if merged within the time window
			if !prs[i].MergedAt.IsZero() {
				timeSinceMerge := time.Since(prs[i].MergedAt)
				if timeSinceMerge.Hours() <= float64(hoursAgo) {
					allPRs = append(allPRs, prs[i])
				}
			}
		}
	}

	return allPRs, nil
}

// FormatIssues formats issues for display
func FormatIssues(issues []GitHubIssue) string {
	if len(issues) == 0 {
		return "No issues found."
	}

	var output strings.Builder
	for _, issue := range issues {
		output.WriteString(fmt.Sprintf("\n#%d - %s\n", issue.Number, issue.Title))
		output.WriteString(fmt.Sprintf("  Repository: %s\n", issue.RepoName))
		output.WriteString(fmt.Sprintf("  URL: %s\n", issue.URL))
		output.WriteString(fmt.Sprintf("  State: %s\n", issue.State))

		// Show project status if available
		if issue.ProjectStatus != "" {
			output.WriteString(fmt.Sprintf("  Project Status: %s\n", issue.ProjectStatus))
		}

		output.WriteString(fmt.Sprintf("  Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04")))

		// Show time on board if available
		if !issue.AddedToProjectAt.IsZero() {
			daysOnBoard := int(time.Since(issue.AddedToProjectAt).Hours() / 24)
			output.WriteString(fmt.Sprintf("  Time on board: %d days\n", daysOnBoard))
		}
	}

	return output.String()
}

// FormatPRs formats pull requests for display
func FormatPRs(prs []GitHubPR) string {
	if len(prs) == 0 {
		return "No pull requests found."
	}

	var output strings.Builder
	for _, pr := range prs {
		output.WriteString(fmt.Sprintf("\n#%d - %s\n", pr.Number, pr.Title))
		output.WriteString(fmt.Sprintf("  Repository: %s\n", pr.RepoName))
		output.WriteString(fmt.Sprintf("  URL: %s\n", pr.URL))
		output.WriteString(fmt.Sprintf("  State: %s\n", pr.State))
		if !pr.MergedAt.IsZero() {
			output.WriteString(fmt.Sprintf("  Merged: %s\n", pr.MergedAt.Format("2006-01-02 15:04")))
		} else {
			output.WriteString(fmt.Sprintf("  Updated: %s\n", pr.UpdatedAt.Format("2006-01-02 15:04")))
		}
	}

	return output.String()
}
