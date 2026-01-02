# macOS XProtect Issue with Assistant Binary

## Problem

macOS XProtect is aggressively removing the `assistant` binary immediately upon execution or even during build. This happens because:

1. The binary uses CGO (required for SQLite)
2. XProtect flags unsigned CGO binaries as potentially malicious
3. The binary is being deleted with exit code 137 (SIGKILL)

## Symptoms

- Binary builds successfully but disappears when you try to run it
- Exit code 137 when attempting to execute
- Binary is deleted from both local directory and `~/go/bin/`
- XProtect processes visible in `ps aux | grep -i protect`

## Solutions

### Option 1: Add Security Exception (Recommended)

1. Build the binary: `CGO_ENABLED=1 go build`
2. When macOS blocks it, open **System Settings** > **Privacy & Security**
3. Look for a message about "assistant was blocked"
4. Click "Allow Anyway"
5. Try running again, and confirm when prompted

### Option 2: Disable XProtect (Not Recommended)

This is not recommended as it reduces your system security, but if you need to:

```bash
# Disable XProtect (requires admin password)
sudo launchctl unload -w /System/Library/LaunchDaemons/com.apple.XProtect*

# To re-enable later:
sudo launchctl load -w /System/Library/LaunchDaemons/com.apple.XProtect*
```

### Option 3: Use Developer Certificate

If you have an Apple Developer account:

```bash
# Sign with your developer certificate
codesign --force --sign "Developer ID Application: Your Name" ./assistant
```

### Option 4: Run from Docker (Alternative)

Create a Dockerfile:

```dockerfile
FROM golang:1.21-alpine
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 go build -o assistant
CMD ["./assistant"]
```

## Current Build Script Status

The `build.sh` script has been updated to build to a temporary file first, but this doesn't solve the XProtect issue. The binary is still removed on execution.

## Verification

To check if XProtect is the issue:

```bash
# Check for XProtect processes
ps aux | grep -i xprotect | grep -v grep

# Check system logs for XProtect blocks
log show --predicate 'process == "XProtect"' --last 5m
```

## Recommended Workflow

Until you add a security exception:

1. Rebuild before each run: `CGO_ENABLED=1 go build && ./assistant <args>`
2. Or use the working binary we found: `./newassistant <args>`
3. Best: Add the security exception as described in Option 1

## Files That Work

These binaries were tested and work because they were created via `cp` not direct build:
- `newassistant` (built with `go build -ldflags="-w -s"`)
- Any binary created by copying a working one

The key is that `cp` doesn't preserve the `com.apple.provenance` extended attribute that triggers XProtect.
