package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

type WebServer struct {
	db *Database
}

func NewWebServer(db *Database) *WebServer {
	return &WebServer{db: db}
}

type WebData struct {
	Title          string
	AssignedIssues []GitHubIssue
	ProjectIssues  []GitHubIssue
	StaleIssues    []GitHubIssue
	RecentPRs      []GitHubPR
	Error          string
	ProjectName    string
	LastUpdated    string
}

func (ws *WebServer) Start(port string) error {
	http.HandleFunc("/", ws.handleIndex)
	http.HandleFunc("/api/refresh", ws.handleRefresh)

	addr := ":" + port
	fmt.Printf("\n🌐 Web UI available at http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop the server")

	return http.ListenAndServe(addr, nil)
}

func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := ws.fetchGitHubData()

	funcMap := template.FuncMap{
		"daysOnBoard": func(t time.Time) int {
			if t.IsZero() {
				return 0
			}
			return int(time.Since(t).Hours() / 24)
		},
	}

	tmpl := template.Must(template.New("index").Funcs(funcMap).Parse(htmlTemplate))
	tmpl.Execute(w, data)
}

func (ws *WebServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	data := ws.fetchGitHubData()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (ws *WebServer) fetchGitHubData() WebData {
	gh := NewGitHubManager("vjeffrey")
	data := WebData{
		Title:       "GitHub Dashboard",
		LastUpdated: time.Now().Format("2006-01-02 15:04:05"),
	}

	mondoohqToken := os.Getenv("GITHUB_TOKEN_ASSISTANT_MONDOOHQ")
	if mondoohqToken == "" {
		data.Error = "GITHUB_TOKEN_ASSISTANT_MONDOOHQ not set"
		return data
	}

	githubProject := os.Getenv("GITHUB_PROJECT")

	// Fetch project issues if GITHUB_PROJECT is set
	if githubProject != "" {
		data.ProjectName = githubProject
		projectIssues, err := gh.GetProjectIssuesForUser(githubProject, mondoohqToken, "")
		if err != nil {
			data.Error = fmt.Sprintf("Failed to fetch project issues: %v", err)
		} else {
			data.ProjectIssues = projectIssues
		}

		// Fetch stale issues (>3 weeks)
		staleDuration := time.Duration(3) * 7 * 24 * time.Hour
		staleIssues, err := gh.GetStaleProjectIssues(githubProject, mondoohqToken, staleDuration, "")
		if err != nil {
			if data.Error != "" {
				data.Error += "; "
			}
			data.Error += fmt.Sprintf("Failed to fetch stale issues: %v", err)
		} else {
			data.StaleIssues = staleIssues
		}
	}

	// Fetch assigned issues from all orgs
	orgTokens := map[string]string{
		"mondoohq":         "GITHUB_TOKEN_ASSISTANT_MONDOOHQ",
		"mondoo-community": "GITHUB_TOKEN_ASSISTANT_MONDOO_COMMUNITY",
	}

	var allIssues []GitHubIssue
	for org, tokenEnv := range orgTokens {
		token := os.Getenv(tokenEnv)
		if token == "" {
			continue
		}

		issues, err := gh.GetAssignedIssues(org, token)
		if err != nil {
			if data.Error != "" {
				data.Error += "; "
			}
			data.Error += fmt.Sprintf("Failed to fetch issues from %s: %v", org, err)
		} else {
			allIssues = append(allIssues, issues...)
		}
	}
	data.AssignedIssues = allIssues

	// Fetch recent merged PRs
	repos := []string{"mondoohq/server", "mondoohq/console", "mondoohq/test-metrics-bigquery"}
	recentPRs, err := gh.GetRecentMergedPRs(repos, 12, mondoohqToken)
	if err != nil {
		if data.Error != "" {
			data.Error += "; "
		}
		data.Error += fmt.Sprintf("Failed to fetch recent PRs: %v", err)
	} else {
		data.RecentPRs = recentPRs
	}

	return data
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            background: #0d1117;
            color: #c9d1d9;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        header {
            background: #161b22;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            border: 1px solid #30363d;
        }
        h1 {
            color: #58a6ff;
            margin-bottom: 10px;
        }
        .last-updated {
            color: #8b949e;
            font-size: 0.9em;
        }
        .error {
            background: #f85149;
            color: white;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 20px;
        }
        .section {
            background: #161b22;
            padding: 20px;
            margin-bottom: 20px;
            border-radius: 8px;
            border: 1px solid #30363d;
        }
        .section h2 {
            color: #58a6ff;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 1px solid #30363d;
        }
        .badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.85em;
            margin-left: 10px;
        }
        .badge-count {
            background: #1f6feb;
            color: white;
        }
        .issue, .pr {
            background: #0d1117;
            padding: 15px;
            margin-bottom: 12px;
            border-radius: 6px;
            border: 1px solid #30363d;
            transition: border-color 0.2s;
        }
        .issue:hover, .pr:hover {
            border-color: #58a6ff;
        }
        .issue-title, .pr-title {
            font-size: 1.1em;
            margin-bottom: 8px;
        }
        .issue-title a, .pr-title a {
            color: #58a6ff;
            text-decoration: none;
        }
        .issue-title a:hover, .pr-title a:hover {
            text-decoration: underline;
        }
        .issue-meta, .pr-meta {
            font-size: 0.9em;
            color: #8b949e;
            display: flex;
            flex-wrap: wrap;
            gap: 15px;
            margin-top: 8px;
        }
        .meta-item {
            display: flex;
            align-items: center;
        }
        .status-badge {
            display: inline-block;
            padding: 3px 10px;
            border-radius: 12px;
            font-size: 0.85em;
            font-weight: 600;
            background: #238636;
            color: white;
        }
        .repo-badge {
            background: #30363d;
            padding: 3px 10px;
            border-radius: 12px;
            font-size: 0.85em;
        }
        .time-badge {
            color: #f85149;
            font-weight: 600;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(600px, 1fr));
            gap: 20px;
        }
        .refresh-btn {
            background: #238636;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 1em;
            margin-left: 15px;
        }
        .refresh-btn:hover {
            background: #2ea043;
        }
        .empty {
            color: #8b949e;
            font-style: italic;
            padding: 20px;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>{{.Title}}</h1>
            <div class="last-updated">
                Last updated: {{.LastUpdated}}
                <button class="refresh-btn" onclick="location.reload()">Refresh</button>
            </div>
        </header>

        {{if .Error}}
        <div class="error">⚠️ Error: {{.Error}}</div>
        {{end}}

        {{if .ProjectName}}
        <div class="section">
            <h2>📊 Project Board Issues <span class="badge badge-count">{{len .ProjectIssues}}</span></h2>
            {{if .ProjectIssues}}
                {{range .ProjectIssues}}
                <div class="issue">
                    <div class="issue-title">
                        <a href="{{.URL}}" target="_blank">#{{.Number}} - {{.Title}}</a>
                    </div>
                    <div class="issue-meta">
                        <span class="meta-item repo-badge">{{.RepoName}}</span>
                        {{if .ProjectStatus}}
                        <span class="meta-item status-badge">{{.ProjectStatus}}</span>
                        {{end}}
                        <span class="meta-item">Updated: {{.UpdatedAt.Format "2006-01-02"}}</span>
                        {{if not .AddedToProjectAt.IsZero}}
                        <span class="meta-item time-badge">{{daysOnBoard .AddedToProjectAt}} days on board</span>
                        {{end}}
                    </div>
                </div>
                {{end}}
            {{else}}
                <div class="empty">No project issues assigned to you</div>
            {{end}}
        </div>

        <div class="section">
            <h2>⏰ Stale Issues (>3 weeks) <span class="badge badge-count">{{len .StaleIssues}}</span></h2>
            {{if .StaleIssues}}
                {{range .StaleIssues}}
                <div class="issue">
                    <div class="issue-title">
                        <a href="{{.URL}}" target="_blank">#{{.Number}} - {{.Title}}</a>
                    </div>
                    <div class="issue-meta">
                        <span class="meta-item repo-badge">{{.RepoName}}</span>
                        {{if .ProjectStatus}}
                        <span class="meta-item status-badge">{{.ProjectStatus}}</span>
                        {{end}}
                        <span class="meta-item">Updated: {{.UpdatedAt.Format "2006-01-02"}}</span>
                        {{if not .AddedToProjectAt.IsZero}}
                        <span class="meta-item time-badge">{{daysOnBoard .AddedToProjectAt}} days on board</span>
                        {{end}}
                    </div>
                </div>
                {{end}}
            {{else}}
                <div class="empty">No stale issues found</div>
            {{end}}
        </div>
        {{end}}

        <div class="section">
            <h2>📋 All Assigned Issues <span class="badge badge-count">{{len .AssignedIssues}}</span></h2>
            {{if .AssignedIssues}}
                {{range .AssignedIssues}}
                <div class="issue">
                    <div class="issue-title">
                        <a href="{{.URL}}" target="_blank">#{{.Number}} - {{.Title}}</a>
                    </div>
                    <div class="issue-meta">
                        <span class="meta-item repo-badge">{{.RepoName}}</span>
                        <span class="meta-item">State: {{.State}}</span>
                        <span class="meta-item">Updated: {{.UpdatedAt.Format "2006-01-02"}}</span>
                    </div>
                </div>
                {{end}}
            {{else}}
                <div class="empty">No issues assigned to you</div>
            {{end}}
        </div>

        <div class="section">
            <h2>✅ Recently Merged PRs (12h) <span class="badge badge-count">{{len .RecentPRs}}</span></h2>
            {{if .RecentPRs}}
                {{range .RecentPRs}}
                <div class="pr">
                    <div class="pr-title">
                        <a href="{{.URL}}" target="_blank">#{{.Number}} - {{.Title}}</a>
                    </div>
                    <div class="pr-meta">
                        <span class="meta-item repo-badge">{{.RepoName}}</span>
                        <span class="meta-item">Merged: {{.MergedAt.Format "2006-01-02 15:04"}}</span>
                    </div>
                </div>
                {{end}}
            {{else}}
                <div class="empty">No PRs merged in the last 12 hours</div>
            {{end}}
        </div>
    </div>

    <script>
        // Helper function to calculate days on board
        function daysOnBoard(date) {
            const now = new Date();
            const added = new Date(date);
            const diff = now - added;
            return Math.floor(diff / (1000 * 60 * 60 * 24));
        }
    </script>
</body>
</html>
`
