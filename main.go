package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// Define command-line flags
	daemon := flag.Bool("daemon", false, "Run in daemon mode (background)")
	list := flag.String("list", "", "List entries for: journal, exercise, or reminders")
	questions := flag.Bool("questions", false, "Run the daily questions interactively")
	morning := flag.Bool("morning", false, "Run the morning question about daily focus")
	github := flag.Bool("github", false, "Show GitHub assignments, mentions, and recent merges")
	merged := flag.Bool("merged", false, "Show recently merged PRs (last 12 hours)")
	web := flag.String("web", "", "Start web UI server on specified port (e.g., --web 8080)")
	webDev := flag.String("web-dev", "", "Start web UI server with fake data on specified port (e.g., --web-dev 8080)")
	githubStale := flag.Int("github-stale", 0, "Show issues on project board for more than N weeks (requires GITHUB_PROJECT env var)")
	filterStatus := flag.String("filter-status", "", "Filter issues by status field value (e.g., '5-9 january')")

	flag.Parse()

	// Initialize database
	db, err := NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Web UI mode (unless daemon mode is also enabled)
	if *web != "" && !*daemon {
		server := NewWebServer(db)
		if err := server.Start(*web); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
		return
	}

	// Web Dev mode with fake data
	if *webDev != "" && !*daemon {
		os.Setenv("WEB_DEV_MODE", "true")
		server := NewWebServer(db)
		if err := server.Start(*webDev); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
		return
	}

	// GitHub mode
	if *github || *merged {
		cli := NewCLI(db)
		githubProjectEnv := os.Getenv("GITHUB_PROJECT")
		if err := cli.ShowGitHubAssignmentsWithOptions(githubProjectEnv, *githubStale, *filterStatus, *merged); err != nil {
			log.Fatalf("Error fetching GitHub assignments: %v", err)
		}
		return
	}

	// List mode
	if *list != "" {
		cli := NewCLI(db)
		if err := cli.ListEntries(*list); err != nil {
			log.Fatalf("Error listing entries: %v", err)
		}
		return
	}

	// Questions mode
	if *questions {
		cli := NewCLI(db)
		if err := cli.RunQuestions(); err != nil {
			log.Fatalf("Error running questions: %v", err)
		}
		return
	}

	// Morning question mode
	if *morning {
		cli := NewCLI(db)
		if err := cli.AskMorningQuestion(); err != nil {
			log.Fatalf("Error asking morning question: %v", err)
		}
		return
	}

	// Daemon mode
	if *daemon {
		runDaemon(db, *web)
		return
	}

	// Default: show usage
	fmt.Println("Daily Assistant")
	fmt.Println("\nUsage:")
	fmt.Println("  assistant --daemon                              Run in background with scheduled notifications")
	fmt.Println("  assistant --daemon --web 8080                   Run daemon with web UI on port 8080")
	fmt.Println("  assistant --questions                           Answer daily questions (journal, exercise, symptoms, reminders)")
	fmt.Println("  assistant --morning                             Answer morning question about daily focus")
	fmt.Println("  assistant --github                              Show GitHub assignments, mentions, and recent merges")
	fmt.Println("  assistant --merged                              Show recently merged PRs (last 12 hours)")
	fmt.Println("  assistant --web 8080                            Start web UI server on port 8080")
	fmt.Println("  assistant --web-dev 8080                        Start web UI with fake data (for styling/development)")
	fmt.Println("  GITHUB_PROJECT=<node_id> assistant --github     Show issues assigned to you on a specific project board")
	fmt.Println("  GITHUB_PROJECT=<node_id> assistant --github --github-stale 3  Show issues on board for more than 3 weeks")
	fmt.Println("  GITHUB_PROJECT=<node_id> assistant --github --filter-status '5-9 january'  Filter by status field value")
	fmt.Println("  GITHUB_PROJECT=<node_id> assistant --merged     Show recently merged PRs with project env var set")
	fmt.Println("  assistant --list journal                        List all journal entries")
	fmt.Println("  assistant --list exercise                       List all exercise entries")
	fmt.Println("  assistant --list symptoms                       List all symptom entries")
	fmt.Println("  assistant --list reminders                      List all reminders")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  GITHUB_PROJECT                                  Project GraphQL node ID (e.g., PVT_kwDOABpK8s4ApRn8)")
	fmt.Println("\nScheduled times:")
	fmt.Println("  7:45 AM - Daily summary + special focus question")
	fmt.Println("  1:00 PM - Daily check-in questions")
}

func runDaemon(db *Database, webPort string) {
	// Setup logging
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".assistant")
	logFile := filepath.Join(logDir, "assistant.log")

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("Starting Daily Assistant daemon...")

	// Start web server if port is specified
	if webPort != "" {
		server := NewWebServer(db)
		go func() {
			log.Printf("Starting web server on port %s...\n", webPort)
			if err := server.Start(webPort); err != nil {
				log.Printf("Web server error: %v\n", err)
			}
		}()
	}

	// Start scheduler
	scheduler := NewScheduler(db)
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	fmt.Println("Daily Assistant is now running in the background.")
	fmt.Println("Logs are stored at:", logFile)
	if webPort != "" {
		fmt.Printf("\nWeb UI available at: http://localhost:%s\n", webPort)
	}
	fmt.Println("\nScheduled times:")
	fmt.Println("  7:45 AM - Daily summary (run 'assistant --morning' to set daily focus)")
	fmt.Println("  1:00 PM - Daily check-in (run 'assistant --questions' to respond)")
	fmt.Println("\nPress Ctrl+C to stop the daemon.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Daily Assistant daemon...")
	fmt.Println("\nDaily Assistant stopped.")
}
