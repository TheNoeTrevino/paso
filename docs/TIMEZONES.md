# Timezone Conventions

## Overview

Paso stores all timestamps in UTC and converts to the user's local timezone at display time. This keeps stored data portable across time zones — if you move from New York to Tokyo, your timestamps adjust automatically.

## Storage

Both SQLite and PostgreSQL schemas use timezone-naive columns (`DATETIME` / `timestamp`, not `timestamptz`). Default values come from `CURRENT_TIMESTAMP`, which produces UTC in both databases.

No application code sets timestamps during writes. The database handles it.

## Retrieval

Timestamps come back as UTC through two paths:

- **SQLite**: The modernc.org/sqlite driver returns strings. `NullTimeFromInterface` in `internal/database/types/nullable.go` parses them with `time.Parse`, which defaults to UTC when no timezone is in the format.
- **PostgreSQL**: The `lib/pq` driver normalizes timestamps to UTC over the wire.

In both cases, the `time.Time` values reaching application code carry UTC.

## Display Conventions

**Human-readable output** (CLI): Call `.Local()` to convert UTC to the user's system timezone before formatting. The standup commands do this for date grouping, time display, and range headers.

**Machine-readable output** (JSON): Call `.UTC()` and format as RFC3339. The trailing `Z` makes the timezone explicit.

When adding new display code, follow the same convention: `.Local()` for humans, `.UTC()` for machines.

## Known Inconsistency

The TUI views (detail panel, comments, activity feed) format timestamps without calling `.Local()` or `.UTC()`. Since the underlying `time.Time` values carry UTC from parsing, these render as UTC times without labeling them as such. This can show times offset from what the user expects.

The standup CLI commands handle this correctly and can be used as a reference.

## Due Dates

Due dates are an exception. The TUI date picker operates in `time.Local`, so the selected date reflects the user's local calendar. These values are stored as-is in the timezone-naive column, meaning the timezone offset may be lost on round-trip.

## No Global Configuration

There is no timezone setting in the app config. The application relies on Go's `time.Local`, which reads the system's `TZ` environment variable or OS timezone. Explicit `.UTC()` and `.Local()` calls handle all conversions.
