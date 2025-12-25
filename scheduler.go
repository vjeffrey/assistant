package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
	db   *Database
}

func NewScheduler(db *Database) *Scheduler {
	return &Scheduler{
		cron: cron.New(),
		db:   db,
	}
}

func (s *Scheduler) Start() error {
	// 7:45 AM daily summary
	_, err := s.cron.AddFunc("45 7 * * *", s.morningRoutine)
	if err != nil {
		return fmt.Errorf("failed to schedule morning routine: %w", err)
	}

	// 1:00 PM daily questions
	_, err = s.cron.AddFunc("0 13 * * *", s.afternoonRoutine)
	if err != nil {
		return fmt.Errorf("failed to schedule afternoon routine: %w", err)
	}

	s.cron.Start()
	log.Println("Scheduler started. Morning routine at 7:45 AM, afternoon routine at 1:00 PM")
	return nil
}

func (s *Scheduler) morningRoutine() {
	log.Println("Running morning routine...")

	message := "Good morning! Here's your daily summary:\n\n"

	// Get last exercise time
	lastExercise, err := s.db.GetLastExerciseTime()
	if err != nil {
		log.Printf("Error getting last exercise time: %v", err)
	} else if lastExercise != nil {
		hoursSince := time.Since(*lastExercise).Hours()
		message += fmt.Sprintf("⏱ Last exercise: %.1f hours ago\n\n", hoursSince)
	} else {
		message += "⏱ No exercise logged yet\n\n"
	}

	// Get today's reminders
	reminders, err := s.db.GetTodaysReminders()
	if err != nil {
		log.Printf("Error getting reminders: %v", err)
	} else if len(reminders) > 0 {
		message += "📋 Today's reminders:\n"
		for _, r := range reminders {
			message += fmt.Sprintf("  - %s (at %s)\n", r.Item, r.ReminderTime.Format("3:04 PM"))
		}
		message += "\n"
	}

	// Check if there's a daily focus
	focus, err := s.db.GetTodaysFocus()
	if err != nil {
		log.Printf("Error getting daily focus: %v", err)
	} else if focus != nil {
		message += fmt.Sprintf("🎯 Today's focus: %s\n\n", focus.Focus)
	}

	message += "Click this notification to add anything special you want to work on today."

	// Send notification
	err = beeep.Notify("Daily Assistant - Morning Summary", message, "")
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
	}

	// The actual prompt for "anything special" will be triggered via the CLI
	// when the user interacts with the notification or runs the program
}

func (s *Scheduler) afternoonRoutine() {
	log.Println("Running afternoon routine...")

	message := "Time for your daily check-in!\n\n"
	message += "Please run the assistant to answer:\n"
	message += "• Journal entry\n"
	message += "• Exercise update\n"
	message += "• Symptoms\n"
	message += "• Reminders\n\n"

	// Check if there's a daily focus to remind about
	focus, err := s.db.GetTodaysFocus()
	if err != nil {
		log.Printf("Error getting daily focus: %v", err)
	} else if focus != nil {
		message += fmt.Sprintf("🎯 Remember: %s", focus.Focus)
	}

	err = beeep.Notify("Daily Assistant - 1 PM Check-in", message, "")
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
	}
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
