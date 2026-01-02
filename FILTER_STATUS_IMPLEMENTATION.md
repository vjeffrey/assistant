# Filter Status Implementation

## Overview

Added `--filter-status` flag to filter GitHub project issues by their status field value. This allows you to narrow down results to only show items with a specific status (e.g., "5-9 january").

## Changes Made

### 1. New Function: `hasMatchingStatus` (github.go:259-276)

```go
func hasMatchingStatus(item ProjectItem, filterStatus string) bool
```

- Returns `true` if the item has a status matching the filter (case-insensitive)
- Returns `true` if no filter is specified (empty string = no filtering)
- Checks `ProjectV2ItemFieldSingleSelectValue` type field values
- Uses the same pattern as `hasExcludedStatus` for consistency

### 2. Updated Functions

**GetProjectIssuesForUser** (github.go:279)
- Added `filterStatus string` parameter
- Calls `hasMatchingStatus()` to filter results

**GetStaleProjectIssues** (github.go:346)
- Added `filterStatus string` parameter
- Calls `hasMatchingStatus()` to filter stale issues

**ShowGitHubAssignmentsWithOptions** (cli.go:211)
- Added `filterStatus string` parameter
- Displays filter status when active
- Passes filter to both stale and assigned issue queries

### 3. CLI Integration (main.go)

- Added `--filter-status` flag (line 22)
- Updated usage help text (line 84)
- Passed flag value to `ShowGitHubAssignmentsWithOptions()`

### 4. Fixed `hasExcludedStatus` (github.go:238-254)

Corrected the status value checking logic:
- Changed from: `fieldName := strings.ToLower(fieldValue.Name)`
- Changed to: `statusValue := strings.ToLower(fieldValue.Name)`
- Now properly checks the status value instead of field name

## Usage Examples

```bash
# Filter assigned issues by status
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --filter-status "5-9 january"

# Combine with stale filter
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --github-stale 3 --filter-status "5-9 january"

# Works with any status value (case-insensitive)
./assistant --github-project PVT_kwDOABpK8s4ApRn8 --filter-status "in progress"
```

## How It Works

1. The filter is applied **after** excluded statuses are checked
2. Filtering is case-insensitive (both "5-9 January" and "5-9 january" work)
3. If no filter is specified, all items pass through (backward compatible)
4. The filter works with both:
   - Regular assigned issues queries
   - Stale issues queries

## Status Field Detection

The code looks for status fields with type `ProjectV2ItemFieldSingleSelectValue`:
- Extracts the `Name` field (which contains the status value)
- Compares it case-insensitively against the filter
- Returns true on first match

This follows the same pattern as the existing `hasExcludedStatus` function, ensuring consistency in how status values are handled throughout the codebase.
