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
