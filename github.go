package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GitHubRepository struct {
	NameWithOwner string `json:"nameWithOwner"`
}

type GitHubIssue struct {
	Number     int              `json:"number"`
	Title      string           `json:"title"`
	Repository GitHubRepository `json:"repository"`
	URL        string           `json:"url"`
	State      string           `json:"state"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
	RepoName   string           // Extracted repository name
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
		output.WriteString(fmt.Sprintf("  Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04")))
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
