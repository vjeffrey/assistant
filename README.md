# Daily Assistant

A Go program that runs in the background and helps you track your journal, exercise, and reminders with scheduled notifications.

The assistant sends desktop notifications at specific times each day and provides an interactive CLI to log your activities, set reminders, and review your history.

## Features

### Scheduled Notifications

- **Morning Summary (7:45 AM)**:
  - Shows hours since your last exercise entry
  - Lists all reminders scheduled for today
  - Displays your daily focus if you set one
  - Prompts you to set something special to work on today

- **Afternoon Check-in (1:00 PM)**:
  - Reminds you to update your journal
  - Prompts for exercise log
  - Allows you to add new reminders with custom times
  - Reminds you of your daily focus if set

### Data Management

- **Desktop Notifications**: Native system notifications on macOS, Linux, and Windows
- **SQLite Storage**: All data stored locally in `~/.assistant/assistant.db`
- **CLI Interface**: Easy commands to interact with your data
- **List Commands**: View your complete history of journal entries, exercises, and reminders

## Installation

### Prerequisites

- Go 1.21 or later
- CGO enabled (required for SQLite)
- GCC or compatible C compiler

### Build

```bash
# SQLite requires CGO to be enabled
CGO_ENABLED=1 go build -o assistant

# Or use the provided build script
./build.sh
```

After building, you'll have an `assistant` binary in the current directory.

## Usage

### Command-line Options

```bash
./assistant                    # Show help and usage information
./assistant --daemon           # Run in background with scheduled notifications
./assistant --questions        # Answer daily questions (journal, exercise, reminders)
./assistant --morning          # Answer morning question about daily focus
./assistant --list journal     # List all journal entries
./assistant --list exercise    # List all exercise entries
./assistant --list reminders   # List all reminders
```

### Start the daemon (background service)

```bash
./assistant --daemon
```

This will run continuously and send notifications at:
- **7:45 AM** - Morning summary with exercise stats, today's reminders, and daily focus prompt
- **1:00 PM** - Daily check-in for journal, exercise, and reminders

Keep this running in a terminal or set it up as a system service (see below).

### Respond to daily questions

When you receive the 1:00 PM notification, run:

```bash
./assistant --questions
```

This will prompt you for:
- **Journal entry**: Type "yes" if you want to add an entry, then write your thoughts
- **Exercise log**: Type "yes" to log your workout details
- **Reminders**: Type "yes" to add reminders with custom dates/times
  - Format: `YYYY-MM-DD HH:MM` (e.g., `2025-12-26 09:00`)
  - You can add multiple reminders in one session

### Set your daily focus

When you receive the 7:45 AM notification, run:

```bash
./assistant --morning
```

This asks what you want to work on today. If you provide an answer (anything other than "no"), the assistant will remind you about it at 1:00 PM along with your other tasks.

### List your entries

View your complete history anytime:

```bash
./assistant --list journal     # All journal entries with timestamps
./assistant --list exercise    # All exercise logs with timestamps
./assistant --list reminders   # All reminders with creation and reminder times
```

Entries are displayed in reverse chronological order (newest first).

## Database Schema

All data is stored in SQLite at `~/.assistant/assistant.db`. The database contains four tables:

### journal
- `id` (INTEGER) - Primary key
- `entry` (TEXT) - Your journal entry
- `time` (DATETIME) - Auto-recorded timestamp when entry was created

### exercise
- `id` (INTEGER) - Primary key
- `exercises` (TEXT) - Your exercise details
- `time` (DATETIME) - Auto-recorded timestamp when logged

### reminders
- `id` (INTEGER) - Primary key
- `item` (TEXT) - Reminder description
- `time` (DATETIME) - When reminder was created
- `reminder_time` (DATETIME) - When to be reminded (you specify this)

### daily_focus
- `id` (INTEGER) - Primary key
- `focus` (TEXT) - What you want to work on today
- `time` (DATETIME) - When focus was set (morning)
- `reminder_time` (DATETIME) - When to be reminded (automatically set to 1 PM)

## Data Location

- Database: `~/.assistant/assistant.db`
- Logs: `~/.assistant/assistant.log`

## Platform Support

The assistant uses the `beeep` library for cross-platform desktop notifications:

- **macOS**: Native NSUserNotification API
- **Linux**: D-Bus desktop notifications
- **Windows**: Windows Toast notifications

All three platforms support rich notifications with title, message body, and system integration.

## Running as a Background Service

For production use, you can set up the assistant to run automatically at system startup instead of manually running `./assistant --daemon`.

### macOS (launchd)

Create `~/Library/LaunchAgents/io.mondoo.assistant.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>io.mondoo.assistant</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/vj/go/src/go.mondoo.io/assistant/assistant</string>
        <string>--daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/Users/vj/.assistant/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/vj/.assistant/stderr.log</string>
</dict>
</plist>
```

Then load it:

```bash
launchctl load ~/Library/LaunchAgents/io.mondoo.assistant.plist
launchctl start io.mondoo.assistant
```

To stop or unload:

```bash
launchctl stop io.mondoo.assistant
launchctl unload ~/Library/LaunchAgents/io.mondoo.assistant.plist
```

### Linux (systemd)

Create `~/.config/systemd/user/assistant.service`:

```ini
[Unit]
Description=Daily Assistant
After=default.target

[Service]
Type=simple
ExecStart=/path/to/assistant --daemon
Restart=always

[Install]
WantedBy=default.target
```

Enable and start:

```bash
systemctl --user enable assistant
systemctl --user start assistant
```

Check status and logs:

```bash
systemctl --user status assistant
journalctl --user -u assistant -f
```

To stop:

```bash
systemctl --user stop assistant
systemctl --user disable assistant
```

## Example Workflow

### First-time Setup

1. **Build the program:**
   ```bash
   ./build.sh
   ```

2. **Start the daemon:**
   ```bash
   ./assistant --daemon
   ```
   Leave this running in a terminal, or set up as a system service (see above).

### Daily Routine

1. **Morning (7:45 AM)**:
   - You receive a notification with:
     - Hours since your last exercise
     - Today's reminders
     - Your daily focus (if previously set)
   - Run `./assistant --morning` to set today's special focus
   - Example: "finish the project proposal" or "no"

2. **Afternoon (1:00 PM)**:
   - You receive a notification to check in
   - Run `./assistant --questions` to:
     - Add journal entry: "Had a productive morning working on the proposal"
     - Log exercise: "30 min run, 3 miles"
     - Add reminders: "Call dentist" for tomorrow at 10 AM
   - You'll also see your daily focus reminder if you set one

3. **Anytime**:
   - Review your history:
     ```bash
     ./assistant --list journal
     ./assistant --list exercise
     ./assistant --list reminders
     ```

## Project Structure

```
.
├── main.go          # Entry point, CLI argument parsing, daemon orchestration
├── database.go      # SQLite database management and CRUD operations
├── scheduler.go     # Cron-based scheduler for timed notifications
├── cli.go           # Interactive CLI for questions and listing data
├── build.sh         # Build script with CGO enabled
├── go.mod           # Go module dependencies
└── README.md        # This file
```

## Dependencies

- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver for Go
- [github.com/robfig/cron/v3](https://github.com/robfig/cron) - Cron-style scheduler
- [github.com/gen2brain/beeep](https://github.com/gen2brain/beeep) - Cross-platform desktop notifications

## Notes

- **Time format for reminders**: `YYYY-MM-DD HH:MM` (e.g., `2025-12-25 14:30`)
- **The daemon must be running** to receive notifications at scheduled times
- **All data is stored locally** on your machine in `~/.assistant/`
- **Logs are written** to `~/.assistant/assistant.log` when running in daemon mode
- **CGO is required** because SQLite needs C bindings
- **Daily focus is temporary** - it only reminds you on the day you set it

## Troubleshooting

### Build Issues

**Error: "Binary was compiled with 'CGO_ENABLED=0'"**
- Solution: Use `CGO_ENABLED=1 go build` or run `./build.sh`

**Error: "gcc: command not found"**
- macOS: Install Xcode Command Line Tools: `xcode-select --install`
- Linux: Install build essentials: `sudo apt-get install build-essential` (Ubuntu/Debian)
- Windows: Install MinGW-w64 or TDM-GCC

### Runtime Issues

**Notifications not appearing:**
- Ensure the daemon is running with `./assistant --daemon`
- Check notification permissions in system settings
- Review logs at `~/.assistant/assistant.log`

**Database locked errors:**
- Only run one instance of the daemon at a time
- Make sure no other processes are accessing the database

**Daemon not running at scheduled times:**
- Verify your system time is correct
- Check that the daemon process hasn't crashed (check logs)
- If using a system service, check service status

## Future Enhancements

Potential features to add:
- Web interface for viewing entries
- Export data to JSON/CSV
- Recurring reminders (daily, weekly, monthly)
- Rich text formatting for journal entries
- Tagging and search functionality
- Mobile companion app
- Data sync across devices
- Reminder notifications at the specified time (currently only shown in morning summary)

## License

This project is provided as-is for personal use.
