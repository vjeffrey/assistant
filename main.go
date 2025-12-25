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

	flag.Parse()

	// Initialize database
	db, err := NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// GitHub mode
	if *github {
		cli := NewCLI(db)
		if err := cli.ShowGitHubAssignments(); err != nil {
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
		runDaemon(db)
		return
	}

	// Default: show usage
	fmt.Println("Daily Assistant")
	fmt.Println("\nUsage:")
	fmt.Println("  assistant --daemon          Run in background with scheduled notifications")
	fmt.Println("  assistant --questions       Answer daily questions (journal, exercise, symptoms, reminders)")
	fmt.Println("  assistant --morning         Answer morning question about daily focus")
	fmt.Println("  assistant --github          Show GitHub assignments, mentions, and recent merges")
	fmt.Println("  assistant --list journal    List all journal entries")
	fmt.Println("  assistant --list exercise   List all exercise entries")
	fmt.Println("  assistant --list symptoms   List all symptom entries")
	fmt.Println("  assistant --list reminders  List all reminders")
	fmt.Println("\nScheduled times:")
	fmt.Println("  7:45 AM - Daily summary + special focus question")
	fmt.Println("  1:00 PM - Daily check-in questions")
}

func runDaemon(db *Database) {
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

	// Start scheduler
	scheduler := NewScheduler(db)
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	fmt.Println("Daily Assistant is now running in the background.")
	fmt.Println("Logs are stored at:", logFile)
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
