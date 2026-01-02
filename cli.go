package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type CLI struct {
	db *Database
}

func NewCLI(db *Database) *CLI {
	return &CLI{db: db}
}

func (c *CLI) RunQuestions() error {
	reader := bufio.NewReader(os.Stdin)

	// Journal
	fmt.Println("\n📝 Do you have anything to add to your journal? (yes/no)")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "yes" || answer == "y" {
		fmt.Println("Enter your journal entry:")
		entry, _ := reader.ReadString('\n')
		entry = strings.TrimSpace(entry)
		if entry != "" {
			if err := c.db.AddJournalEntry(entry); err != nil {
				return fmt.Errorf("failed to add journal entry: %w", err)
			}
			fmt.Println("✓ Journal entry saved")
		}
	}

	// Exercise
	fmt.Println("\n💪 Do you have anything to add to your exercise log? (yes/no)")
	answer, _ = reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "yes" || answer == "y" {
		fmt.Println("Enter your exercise details:")
		exercise, _ := reader.ReadString('\n')
		exercise = strings.TrimSpace(exercise)
		if exercise != "" {
			if err := c.db.AddExerciseEntry(exercise); err != nil {
				return fmt.Errorf("failed to add exercise entry: %w", err)
			}
			fmt.Println("✓ Exercise entry saved")
		}
	}

	// Symptoms
	fmt.Println("\n🩺 Do you have any symptoms to track? (yes/no)")
	answer, _ = reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "yes" || answer == "y" {
		fmt.Println("Enter your symptoms:")
		symptoms, _ := reader.ReadString('\n')
		symptoms = strings.TrimSpace(symptoms)
		if symptoms != "" {
			if err := c.db.AddSymptomEntry(symptoms); err != nil {
				return fmt.Errorf("failed to add symptom entry: %w", err)
			}
			fmt.Println("✓ Symptom entry saved")
		}
	}

	// Reminders
	fmt.Println("\n⏰ Do you have any reminders to add? (yes/no)")
	answer, _ = reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "yes" || answer == "y" {
		for {
			fmt.Println("Enter reminder item:")
			item, _ := reader.ReadString('\n')
			item = strings.TrimSpace(item)
			if item == "" {
				break
			}

			fmt.Println("Enter reminder time (format: YYYY-MM-DD HH:MM, e.g., 2025-12-25 14:30):")
			timeStr, _ := reader.ReadString('\n')
			timeStr = strings.TrimSpace(timeStr)

			reminderTime, err := time.Parse("2006-01-02 15:04", timeStr)
			if err != nil {
				fmt.Printf("Invalid time format. Please use YYYY-MM-DD HH:MM\n")
				continue
			}

			if err := c.db.AddReminder(item, reminderTime); err != nil {
				return fmt.Errorf("failed to add reminder: %w", err)
			}
			fmt.Println("✓ Reminder saved")

			fmt.Println("\nAdd another reminder? (yes/no)")
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "yes" && answer != "y" {
				break
			}
		}
	}

	return nil
}

func (c *CLI) AskMorningQuestion() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🌅 Good morning! Is there anything special you want to work on today? (or type 'no')")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)

	if answer != "" && strings.ToLower(answer) != "no" && strings.ToLower(answer) != "n" {
		// Set reminder for 1 PM today
		now := time.Now()
		reminderTime := time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, now.Location())

		if err := c.db.AddDailyFocus(answer, reminderTime); err != nil {
			return fmt.Errorf("failed to add daily focus: %w", err)
		}
		fmt.Println("✓ I'll remind you about this at 1 PM today")
	}

	return nil
}

func (c *CLI) ListEntries(category string) error {
	switch category {
	case "journal":
		entries, err := c.db.ListJournalEntries()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("No journal entries found.")
			return nil
		}
		fmt.Println("\n📝 Journal Entries:")
		fmt.Println(strings.Repeat("-", 60))
		for _, e := range entries {
			fmt.Printf("[%s]\n%s\n\n", e.Time.Format("2006-01-02 15:04:05"), e.Entry)
		}

	case "exercise":
		entries, err := c.db.ListExerciseEntries()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("No exercise entries found.")
			return nil
		}
		fmt.Println("\n💪 Exercise Entries:")
		fmt.Println(strings.Repeat("-", 60))
		for _, e := range entries {
			fmt.Printf("[%s]\n%s\n\n", e.Time.Format("2006-01-02 15:04:05"), e.Exercises)
		}

	case "reminders":
		reminders, err := c.db.ListReminders()
		if err != nil {
			return err
		}
		if len(reminders) == 0 {
			fmt.Println("No reminders found.")
			return nil
		}
		fmt.Println("\n⏰ Reminders:")
		fmt.Println(strings.Repeat("-", 60))
		for _, r := range reminders {
			fmt.Printf("[Created: %s] [Remind at: %s]\n%s\n\n",
				r.Time.Format("2006-01-02 15:04:05"),
				r.ReminderTime.Format("2006-01-02 15:04:05"),
				r.Item)
		}

	case "symptoms":
		entries, err := c.db.ListSymptomEntries()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Println("No symptom entries found.")
			return nil
		}
		fmt.Println("\n🩺 Symptom Entries:")
		fmt.Println(strings.Repeat("-", 60))
		for _, e := range entries {
			fmt.Printf("[%s]\n%s\n\n", e.Time.Format("2006-01-02 15:04:05"), e.Symptoms)
		}

	default:
		return fmt.Errorf("unknown category: %s. Use: journal, exercise, symptoms, or reminders", category)
	}

	return nil
}

func (c *CLI) ShowGitHubAssignments() error {
	return c.ShowGitHubAssignmentsWithOptions("", 0, "", false)
}

func (c *CLI) ShowGitHubAssignmentsWithOptions(projectNodeID string, staleThresholdWeeks int, filterStatus string, showMergedOnly bool) error {
	gh := NewGitHubManager("vjeffrey")

	fmt.Println("\n🔍 Fetching GitHub assignments...")

	mondoohqToken := os.Getenv("GITHUB_TOKEN_ASSISTANT_MONDOOHQ")
	if mondoohqToken == "" {
		return fmt.Errorf("GITHUB_TOKEN_ASSISTANT_MONDOOHQ not set")
	}

	// If only merged PRs are requested
	if showMergedOnly {
		fmt.Println("\n✅ Recently merged PRs (last 12 hours):")
		fmt.Println(strings.Repeat("=", 80))

		repos := []string{"mondoohq/server", "mondoohq/console", "mondoohq/test-metrics-bigquery"}
		recentPRs, err := gh.GetRecentMergedPRs(repos, 12, mondoohqToken)
		if err != nil {
			return fmt.Errorf("failed to fetch recent merged PRs: %w", err)
		}
		fmt.Println(FormatPRs(recentPRs))
		fmt.Printf("\nTotal: %d PR(s) merged in the last 12 hours\n", len(recentPRs))

		return nil
	}

	// If project filtering is requested
	if projectNodeID != "" {
		fmt.Printf("\n📊 Fetching issues from project board...\n")
		if filterStatus != "" {
			fmt.Printf("   Filtering by status: %s\n", filterStatus)
		}

		// If stale threshold is specified, only show stale issues
		if staleThresholdWeeks > 0 {
			staleDuration := time.Duration(staleThresholdWeeks) * 7 * 24 * time.Hour
			staleIssues, err := gh.GetStaleProjectIssues(projectNodeID, mondoohqToken, staleDuration, filterStatus)
			if err != nil {
				return fmt.Errorf("failed to fetch stale issues: %w", err)
			}

			fmt.Printf("\n⏰ Issues on board for more than %d weeks:\n", staleThresholdWeeks)
			fmt.Println(strings.Repeat("=", 80))
			fmt.Println(FormatIssues(staleIssues))
			fmt.Printf("\nTotal: %d stale issue(s)\n", len(staleIssues))

			return nil
		}

		// Otherwise show assigned issues
		projectIssues, err := gh.GetProjectIssuesForUser(projectNodeID, mondoohqToken, filterStatus)
		if err != nil {
			return fmt.Errorf("failed to fetch project issues: %w", err)
		}

		fmt.Println("\n📋 Issues assigned to vjeffrey on project board:")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println(FormatIssues(projectIssues))
		fmt.Printf("\nTotal: %d issue(s)\n", len(projectIssues))

		return nil
	}

	// Default behavior - fetch from all orgs
	var allIssues []GitHubIssue
	var allMentionedIssues []GitHubIssue
	var allMentionedPRs []GitHubPR

	// Define orgs and their token environment variables
	orgTokens := map[string]string{
		"mondoohq":         "GITHUB_TOKEN_ASSISTANT_MONDOOHQ",
		"mondoo-community": "GITHUB_TOKEN_ASSISTANT_MONDOO_COMMUNITY",
	}

	// Fetch data for each org with its specific token
	for org, tokenEnv := range orgTokens {
		token := os.Getenv(tokenEnv)
		if token == "" {
			fmt.Printf("⚠️  Warning: %s not set, skipping %s\n", tokenEnv, org)
			continue
		}

		// Get assigned issues for this org
		issues, err := gh.GetAssignedIssues(org, token)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to fetch issues from %s: %v\n", org, err)
		} else {
			allIssues = append(allIssues, issues...)
		}

		// Get mentions for this org
		mentionedIssues, mentionedPRs, err := gh.GetMentionedIssuesAndPRs(org, token)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to fetch mentions from %s: %v\n", org, err)
		} else {
			allMentionedIssues = append(allMentionedIssues, mentionedIssues...)
			allMentionedPRs = append(allMentionedPRs, mentionedPRs...)
		}
	}

	// Display assigned issues
	fmt.Println("\n📋 Issues assigned to vjeffrey:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(FormatIssues(allIssues))
	fmt.Printf("\nTotal: %d issue(s)\n", len(allIssues))

	// Display mentioned issues and PRs
	// fmt.Println("\n💬 Issues and PRs mentioning vjeffrey:")
	// fmt.Println(strings.Repeat("=", 80))

	// if len(allMentionedIssues) > 0 {
	// 	fmt.Println("\nIssues:")
	// 	fmt.Println(FormatIssues(allMentionedIssues))
	// }

	// if len(allMentionedPRs) > 0 {
	// 	fmt.Println("\nPull Requests:")
	// 	fmt.Println(FormatPRs(allMentionedPRs))
	// }

	// fmt.Printf("\nTotal: %d issue(s) and %d PR(s)\n", len(allMentionedIssues), len(allMentionedPRs))

	// Get recently merged PRs (use mondoohq token for these repos)
	fmt.Println("\n✅ Recently merged PRs (last 12 hours):")
	fmt.Println(strings.Repeat("=", 80))

	if mondoohqToken != "" {
		repos := []string{"mondoohq/server", "mondoohq/console", "mondoohq/test-metrics-bigquery"}
		recentPRs, err := gh.GetRecentMergedPRs(repos, 12, mondoohqToken)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to fetch recent merged PRs: %v\n", err)
		} else {
			fmt.Println(FormatPRs(recentPRs))
			fmt.Printf("\nTotal: %d PR(s) merged in the last 12 hours\n", len(recentPRs))
		}
	} else {
		fmt.Println("⚠️  GITHUB_TOKEN_ASSISTANT_MONDOOHQ not set, skipping recent PRs")
	}

	return nil
}
