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

### GitHub Integration

- **View GitHub Assignments**: See all issues assigned to you across multiple organizations
- **Track Mentions**: Find all issues and PRs where you've been mentioned
- **Monitor Recent Merges**: Track PRs merged in the last 12 hours in key repositories
- **Multi-Organization Support**: Works with separate authentication tokens per organization

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
- GitHub CLI (`gh`) - Required for GitHub integration features
  - Install: `brew install gh` (macOS), `apt install gh` (Ubuntu/Debian), or see [GitHub CLI docs](https://cli.github.com/)

### Build

```bash
# Using Make (recommended)
make build

# Or using the build script directly
./build.sh

# Or manually with Go
CGO_ENABLED=1 go build -o assistant
```

After building, you'll have an `assistant` binary in the current directory.

#### Makefile Commands

The project includes a Makefile with helpful commands:

```bash
make build           # Build the assistant binary
make install         # Install to ~/go/bin/assistant
make clean           # Remove build artifacts
make fmt             # Format Go code
make lint            # Run golangci-lint
make test            # Run tests
make check-creds     # Check for accidentally committed credentials
make pre-commit      # Run all checks before committing (creds, fmt, test)
make run-daemon      # Build and run daemon
make run-web         # Build and run web UI on port 8080
make run-daemon-web  # Build and run daemon with web UI
make help            # Show all available commands
```

## Usage

### Command-line Options

```bash
./assistant                       # Show help and usage information
./assistant --daemon              # Run in background with scheduled notifications
./assistant --daemon --web 8080   # Run daemon with web UI on port 8080
./assistant --web 8080            # Start web UI server only (no daemon)
./assistant --questions           # Answer daily questions (journal, exercise, reminders)
./assistant --morning             # Answer morning question about daily focus
./assistant --github              # Show GitHub assignments, mentions, and recent merges
./assistant --merged              # Show recently merged PRs (last 12 hours)
./assistant --list journal        # List all journal entries
./assistant --list exercise       # List all exercise entries
./assistant --list reminders      # List all reminders
./assistant --list symptoms       # List all symptom entries
```

### Start the daemon (background service)

```bash
./assistant --daemon
```

This will run continuously and send notifications at:
- **7:45 AM** - Morning summary with exercise stats, today's reminders, and daily focus prompt
- **1:00 PM** - Daily check-in for journal, exercise, and reminders

Keep this running in a terminal or set it up as a system service (see below).

### Web UI

The assistant includes a lightweight web interface for viewing your GitHub assignments:

```bash
# Start web UI only
./assistant --web 8080

# Or combine with daemon mode to run both
./assistant --daemon --web 8080
```

The web UI displays:
- Project board issues (if `GITHUB_PROJECT` is set)
- Stale issues (>3 weeks on board)
- All assigned issues from configured organizations
- Recently merged PRs (last 12 hours)
- Real-time refresh capability
- Labels and assignees for all issues and PRs

Access it at `http://localhost:8080` in your browser.

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
./assistant --list symptoms    # All symptom entries with timestamps
```

Entries are displayed in reverse chronological order (newest first).

### GitHub Integration

Track your GitHub work across multiple organizations:

```bash
# Show all assigned issues and recent merges
./assistant --github

# Show only issues on a specific project board
./assistant --github-project PVT_kwDOABpK8s4ApRn8

# Show issues on a project board that have been there for more than 3 weeks
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --github-stale 3
```

This command displays:
- **Assigned Issues**: All issues assigned to you in `mondoohq` and `mondoo-community` organizations
- **Project Board Issues**: Filter to issues on a specific GitHub Project v2 board
- **Stale Issues**: Identify issues that have been on a project board for longer than a threshold
- **Mentions**: Issues and PRs where you've been mentioned (when using `--github` without project filtering)
- **Recent Merges**: PRs merged in the last 12 hours in key repositories (when using `--github` without project filtering)

#### Setup

The GitHub integration requires GitHub CLI (`gh`) and organization-specific access tokens:

1. **Install GitHub CLI**:
   ```bash
   brew install gh  # macOS
   # or
   sudo apt install gh  # Ubuntu/Debian
   ```

2. **Set up environment variables**:
   ```bash
   export GITHUB_TOKEN_ASSISTANT_MONDOOHQ="ghp_your_mondoohq_token"
   export GITHUB_TOKEN_ASSISTANT_MONDOO_COMMUNITY="ghp_your_mondoo_community_token"
   ```

   Add these to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to persist across sessions.

3. **Token Requirements**:
   - Create personal access tokens at [github.com/settings/tokens](https://github.com/settings/tokens)
   - Required scopes:
     - `repo`, `read:org` (for private repositories and issue search)
     - `project` (for GitHub Projects v2 queries - required for `--github-project`)
   - Use separate tokens for each organization if they have different access requirements
   - To add the `project` scope to an existing token: `gh auth refresh -s project`

4. **Get Project Node ID**:

   To use the `--github-project` flag, you need the GraphQL node ID of the project. You can get this using the provided helper script:

   ```bash
   # First, ensure your token has the 'project' scope
   gh auth refresh -s project

   # Then get the project ID (project number 14 from URL https://github.com/orgs/mondoohq/projects/14)
   ./get-project-id.sh mondoohq 14
   ```

   This will output something like:
   ```
   Project ID: PVT_kwDOABpK8s4ApRn8
   Title: Engineering Sprint Board
   URL: https://github.com/orgs/mondoohq/projects/14
   ```

   Use the Project ID with the `--github-project` flag.

#### Example Output

```
🔍 Fetching GitHub assignments...

📋 Issues assigned to vjeffrey:
================================================================================

#1234 - Fix authentication bug
  Repository: mondoohq/server
  URL: https://github.com/mondoohq/server/issues/1234
  State: OPEN
  Updated: 2025-12-25 10:30

Total: 1 issue(s)

💬 Issues and PRs mentioning vjeffrey:
================================================================================

Pull Requests:

#5678 - Add new feature
  Repository: mondoohq/console
  URL: https://github.com/mondoohq/console/pull/5678
  State: OPEN
  Updated: 2025-12-25 09:15

Total: 0 issue(s) and 1 PR(s)

✅ Recently merged PRs (last 12 hours):
================================================================================

#9999 - Update dependencies
  Repository: mondoohq/server
  URL: https://github.com/mondoohq/server/pull/9999
  State: MERGED
  Merged: 2025-12-25 08:00

Total: 1 PR(s) merged in the last 12 hours
```

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
        <string>/Users/vj/go/bin/assistant</string>
        <string>--daemon</string>
        <string>--web</string>
        <string>8080</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>GITHUB_TOKEN_ASSISTANT_MONDOOHQ</key>
        <string>your_token_here</string>
        <key>GITHUB_PROJECT</key>
        <string>PVT_your_project_id</string>
    </dict>
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
ExecStart=/path/to/assistant --daemon --web 8080
Environment="GITHUB_TOKEN_ASSISTANT_MONDOOHQ=your_token_here"
Environment="GITHUB_PROJECT=PVT_your_project_id"
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
├── github.go        # GitHub integration using gh CLI
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
- **GitHub integration** requires GitHub CLI (`gh`) and organization-specific tokens set as environment variables

## Security Best Practices

### Protecting Credentials

This project includes tools to help prevent accidentally committing credentials:

1. **Check for credentials before committing**:
   ```bash
   make check-creds
   ```

   This script checks for:
   - GitHub tokens (ghp_, gho_, ghs_, etc.)
   - AWS credentials
   - Private keys
   - API keys and tokens
   - Database connection strings
   - Hardcoded passwords

2. **Run pre-commit checks**:
   ```bash
   make pre-commit
   ```

   This runs credential checks, code formatting, and tests.

3. **Use environment variables** (never hardcode):
   ```bash
   export GITHUB_TOKEN_ASSISTANT_MONDOOHQ="ghp_your_token"
   ```

   Add these to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) but **never commit them**.

4. **Rotate tokens immediately** if accidentally committed:
   - Go to [GitHub Settings > Tokens](https://github.com/settings/tokens)
   - Delete the exposed token
   - Generate a new one
   - Update your environment variables

### Setting Up Git Hooks (Optional)

For automated checking, you can set up a pre-commit hook:

```bash
# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
make check-creds
EOF

chmod +x .git/hooks/pre-commit
```

This will automatically check for credentials before every commit.

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

### GitHub Integration Issues

**GitHub command shows warnings about missing tokens:**
- Ensure environment variables are set: `GITHUB_TOKEN_ASSISTANT_MONDOOHQ` and `GITHUB_TOKEN_ASSISTANT_MONDOO_COMMUNITY`
- Add exports to your shell profile for persistence
- Verify tokens have correct scopes: `repo`, `read:org`

**No results shown for assigned issues or mentions:**
- Verify you have access to the organizations (mondoohq, mondoo-community)
- Check that you actually have issues assigned or mentions in these orgs
- Try running `gh auth status` to verify authentication

**Error: "gh: command not found":**
- Install GitHub CLI: `brew install gh` (macOS) or see [cli.github.com](https://cli.github.com/)

## Future Enhancements

Potential features to add:
- ✅ ~~Web interface for viewing entries~~ (Implemented for GitHub data)
- Extend web interface to show journal, exercise, and reminders
- Export data to JSON/CSV
- Recurring reminders (daily, weekly, monthly)
- Rich text formatting for journal entries
- Tagging and search functionality
- Mobile companion app
- Data sync across devices
- Reminder notifications at the specified time (currently only shown in morning summary)

## License

This project is provided as-is for personal use.
