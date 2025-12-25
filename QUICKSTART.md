# Quick Start Guide

## Setup (One-time)

```bash
# 1. Build the program
./build.sh

# 2. Start the daemon
./assistant --daemon
```

Keep the daemon running. Set up as a system service for automatic startup (see README.md).

## Daily Usage

### Morning (7:45 AM)
1. Get notification with summary
2. Run: `./assistant --morning`
3. Answer: What do you want to work on today?

### Afternoon (1:00 PM)
1. Get notification for check-in
2. Run: `./assistant --questions`
3. Answer three questions:
   - Journal? (yes/no, then type entry)
   - Exercise? (yes/no, then type details)
   - Reminders? (yes/no, then item + time in format `YYYY-MM-DD HH:MM`)

## View History

```bash
./assistant --list journal      # All journal entries
./assistant --list exercise     # All exercise logs
./assistant --list reminders    # All reminders
```

## Tips

- Answer "no" or "n" to skip any question
- You can add multiple reminders in one session
- Daily focus only lasts for today
- Check logs at `~/.assistant/assistant.log` if something goes wrong

## Example Session

```
$ ./assistant --questions

📝 Do you have anything to add to your journal? (yes/no)
yes
Enter your journal entry:
Had a great morning working on the new feature. Made good progress.
✓ Journal entry saved

💪 Do you have anything to add to your exercise log? (yes/no)
yes
Enter your exercise details:
Morning run - 5K in 28 minutes
✓ Exercise entry saved

⏰ Do you have any reminders to add? (yes/no)
yes
Enter reminder item:
Review pull requests
Enter reminder time (format: YYYY-MM-DD HH:MM, e.g., 2025-12-25 14:30):
2025-12-26 10:00
✓ Reminder saved

Add another reminder? (yes/no)
no
```
