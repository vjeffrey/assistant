package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type JournalEntry struct {
	Entry string
	Time  time.Time
}

type ExerciseEntry struct {
	Exercises string
	Time      time.Time
}

type Reminder struct {
	ID           int
	Item         string
	Time         time.Time
	ReminderTime time.Time
}

type DailyFocus struct {
	ID           int
	Focus        string
	Time         time.Time
	ReminderTime time.Time
}

type SymptomEntry struct {
	Symptoms string
	Time     time.Time
}

func NewDatabase() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".assistant")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "assistant.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	d := &Database{db: db}
	if err := d.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	return d, nil
}

func (d *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS journal (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entry TEXT NOT NULL,
			time DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS exercise (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			exercises TEXT NOT NULL,
			time DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS reminders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			item TEXT NOT NULL,
			time DATETIME DEFAULT CURRENT_TIMESTAMP,
			reminder_time DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS daily_focus (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			focus TEXT NOT NULL,
			time DATETIME DEFAULT CURRENT_TIMESTAMP,
			reminder_time DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS symptoms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symptoms TEXT NOT NULL,
			time DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (d *Database) AddJournalEntry(entry string) error {
	_, err := d.db.Exec("INSERT INTO journal (entry) VALUES (?)", entry)
	return err
}

func (d *Database) AddExerciseEntry(exercises string) error {
	_, err := d.db.Exec("INSERT INTO exercise (exercises) VALUES (?)", exercises)
	return err
}

func (d *Database) AddReminder(item string, reminderTime time.Time) error {
	_, err := d.db.Exec("INSERT INTO reminders (item, reminder_time) VALUES (?, ?)", item, reminderTime)
	return err
}

func (d *Database) AddDailyFocus(focus string, reminderTime time.Time) error {
	_, err := d.db.Exec("INSERT INTO daily_focus (focus, reminder_time) VALUES (?, ?)", focus, reminderTime)
	return err
}

func (d *Database) AddSymptomEntry(symptoms string) error {
	_, err := d.db.Exec("INSERT INTO symptoms (symptoms) VALUES (?)", symptoms)
	return err
}

func (d *Database) GetLastExerciseTime() (*time.Time, error) {
	var timeStr string
	err := d.db.QueryRow("SELECT time FROM exercise ORDER BY time DESC LIMIT 1").Scan(&timeStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (d *Database) GetTodaysReminders() ([]Reminder, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := d.db.Query(
		"SELECT id, item, time, reminder_time FROM reminders WHERE reminder_time >= ? AND reminder_time < ? ORDER BY reminder_time",
		startOfDay, endOfDay,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		var timeStr, reminderTimeStr string
		if err := rows.Scan(&r.ID, &r.Item, &timeStr, &reminderTimeStr); err != nil {
			return nil, err
		}

		r.Time, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		r.ReminderTime, _ = time.Parse("2006-01-02 15:04:05", reminderTimeStr)
		reminders = append(reminders, r)
	}

	return reminders, nil
}

func (d *Database) GetTodaysFocus() (*DailyFocus, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var df DailyFocus
	var timeStr, reminderTimeStr string
	err := d.db.QueryRow(
		"SELECT id, focus, time, reminder_time FROM daily_focus WHERE reminder_time >= ? AND reminder_time < ? ORDER BY time DESC LIMIT 1",
		startOfDay, endOfDay,
	).Scan(&df.ID, &df.Focus, &timeStr, &reminderTimeStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	df.Time, _ = time.Parse("2006-01-02 15:04:05", timeStr)
	df.ReminderTime, _ = time.Parse("2006-01-02 15:04:05", reminderTimeStr)
	return &df, nil
}

func (d *Database) ListJournalEntries() ([]JournalEntry, error) {
	rows, err := d.db.Query("SELECT entry, time FROM journal ORDER BY time DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []JournalEntry
	for rows.Next() {
		var e JournalEntry
		var timeStr string
		if err := rows.Scan(&e.Entry, &timeStr); err != nil {
			return nil, err
		}
		e.Time, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		entries = append(entries, e)
	}

	return entries, nil
}

func (d *Database) ListExerciseEntries() ([]ExerciseEntry, error) {
	rows, err := d.db.Query("SELECT exercises, time FROM exercise ORDER BY time DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ExerciseEntry
	for rows.Next() {
		var e ExerciseEntry
		var timeStr string
		if err := rows.Scan(&e.Exercises, &timeStr); err != nil {
			return nil, err
		}
		e.Time, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		entries = append(entries, e)
	}

	return entries, nil
}

func (d *Database) ListReminders() ([]Reminder, error) {
	rows, err := d.db.Query("SELECT id, item, time, reminder_time FROM reminders ORDER BY reminder_time DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		var timeStr, reminderTimeStr string
		if err := rows.Scan(&r.ID, &r.Item, &timeStr, &reminderTimeStr); err != nil {
			return nil, err
		}
		r.Time, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		r.ReminderTime, _ = time.Parse("2006-01-02 15:04:05", reminderTimeStr)
		reminders = append(reminders, r)
	}

	return reminders, nil
}

func (d *Database) ListSymptomEntries() ([]SymptomEntry, error) {
	rows, err := d.db.Query("SELECT symptoms, time FROM symptoms ORDER BY time DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []SymptomEntry
	for rows.Next() {
		var e SymptomEntry
		var timeStr string
		if err := rows.Scan(&e.Symptoms, &timeStr); err != nil {
			return nil, err
		}
		e.Time, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		entries = append(entries, e)
	}

	return entries, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
