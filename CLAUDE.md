# GTFS Parser — Project Context

## Purpose
A Go library (`github.com/sutanmufti/gtfs-parser`) that:
1. Reads a GTFS `.zip` file
2. Validates the contained `.txt` files against the GTFS Schedule spec
3. Parses the data into idiomatic Go structs for use by downstream projects

This is a **library**, not a standalone application. `cmd/main.go` is for local development/testing only. Exported types and functions must have clean, stable APIs.

## Package
All library files use `package gtfsparser`. Only `cmd/main.go` uses `package main`.

## Build & Test
```
go build ./...
go test ./...
```

## Architecture
- `filereader.go` — zip reading, file discovery
- `structs.go` — all exported Go structs representing GTFS entities
- `parser.go` — CSV parsing logic per GTFS file type; contains `ParseAll()` as the main entry point
- `helpers.go` — shared helpers: `getCol`, `sanitizeHeaders`, `parseOptionalFloat`
- `validate.go` — validation rules (required fields, enum values, referential integrity) — not yet implemented
- `errors.go` — structured error/warning types for validation results — not yet implemented

Validation collects **all errors** before returning — do not fail fast. Return a slice of structured errors so callers can inspect and report them.

## Parser execution order
`ParseAll()` runs parsers in dependency order:
1. `ParseAgency` — no deps
2. `ParseLevel` — no deps
3. `ParseCalendar` — no deps
4. `ParseShape` — no deps
5. `ParseStop` — depends on Level
6. `ParseRoute` — depends on Agency
7. `ParseCalendarDate` — depends on Calendar
8. `ParseTrip` — depends on Route, Calendar, Shape
9. `ParseStopTime` — depends on Trip, Stop
10. `ParseFrequency` — depends on Trip
11. `ParseTransfer` — depends on Stop, Route, Trip
12. `ParsePathway` — depends on Stop
13. `ParseFareAttribute` — depends on Agency
14. `ParseFareRule` — depends on FareAttribute, Route
15. `ParseFeedInfo` — no deps
16. `ParseAttribution` — depends on Agency, Route, Trip
17. `ParseTranslation` — no deps

Optional files return `nil` (not an error) when absent.

## GTFS Spec Reference

### Required files (must exist and be non-empty)
| File | Key entity |
|---|---|
| `agency.txt` | Transit agencies |
| `stops.txt` | Stop locations |
| `routes.txt` | Route definitions |
| `trips.txt` | Trips per route |
| `stop_times.txt` | Arrival/departure times per trip |
| `calendar.txt` OR `calendar_dates.txt` | Service availability (at least one required) |

### Conditionally required / Optional files
| File | Notes |
|---|---|
| `calendar_dates.txt` | Required if `calendar.txt` absent; supplements it if present |
| `shapes.txt` | Optional; referenced by `trips.shape_id` |
| `frequencies.txt` | Optional; headway-based trips |
| `transfers.txt` | Optional; transfer rules between stops |
| `fare_attributes.txt` | Optional |
| `fare_rules.txt` | Optional |
| `feed_info.txt` | Optional; metadata about the feed |
| `pathways.txt` | Optional |
| `levels.txt` | Optional |
| `attributions.txt` | Optional |
| `translations.txt` | Optional |

### Field validation rules
- Required fields must be non-empty
- IDs must be unique within their file (e.g. `stop_id`, `route_id`, `trip_id`)
- Foreign keys must reference existing records (e.g. `trips.route_id` → `routes.route_id`)
- Enum fields must be valid integers within the specified range
- Lat/lon fields must be valid WGS84 coordinates
- Time fields use `HH:MM:SS` format (hours can exceed 24 for post-midnight service)

## Go Conventions
- Use `camelCase` for unexported identifiers, `PascalCase` for exported
- Struct field names use `snake_case` matching GTFS CSV column names exactly (e.g. `stop_id`, `route_long_name`)
- No external dependencies — standard library only
- Errors use the standard `error` interface; define custom types in `errors.go`
- Do not use `panic`; always return errors

## Struct Design Patterns
- **CSV parsing**: read the header row into a `map[string]int` column index map (`col`), then access fields using the `getCol(row, col, "field_name")` helper in `helpers.go`. This safely returns `""` if the column is missing. Do not use struct tags or direct `row[col["field_name"]]` access.
- **Foreign key fields**: use a pointer to the referenced struct (e.g. `agency_id *Agency`, `route_id *Route`) to model relationships. `nil` means not provided or not yet resolved. Resolve inline using an index map built before the parse loop (e.g. `routeIndex := map[string]*Route`).
- **Self-referential FKs**: use a two-pass approach — parse all rows first storing raw string IDs in a helper struct, then resolve pointers in a second pass after the full slice is built. See `ParseStop` for the `stopRaw` pattern.
- **Conditionally required numeric fields**: use a pointer (e.g. `stop_lat *float64`) so `nil` means "not provided" and is distinguishable from a real zero value. Use the `parseOptionalFloat` helper in `helpers.go`.
- **Enum fields**: define a named `int` type and `const` block with `iota` for each enum (e.g. `LocationType`, `RouteType`). Cast the parsed integer to the enum type.
- **BOM handling**: call `sanitizeHeaders(headers)` after reading the header row in every parser to strip UTF-8 BOM from the first column name.
