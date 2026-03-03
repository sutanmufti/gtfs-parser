# gtfs-parser

![Test](https://github.com/sutanmufti/gtfs-parser/actions/workflows/test.yml/badge.svg)

A Go library for parsing and validating [GTFS Schedule](https://gtfs.org/schedule/) zip files.

## Overview

`gtfs-parser` reads a GTFS `.zip` file, parses each `.txt` file into idiomatic Go structs, and validates the feed against the GTFS Schedule specification. It is designed to be used as a library in other Go projects.

## Installation

```sh
go get github.com/sutanmufti/gtfs-parser
```

## Usage

```go
import gtfsparser "github.com/sutanmufti/gtfs-parser"

gtfs := gtfsparser.GTFS{FileName: "path/to/feed.zip"}

// Parse all files in dependency order
if err := gtfs.ParseAll(); err != nil {
    log.Fatal(err)
}

// Validate — returns all errors, does not fail fast
errs := gtfs.ValidateAll()
if len(errs) == 0 {
    fmt.Println("feed is valid")
} else {
    for _, e := range errs {
        fmt.Println(e)
    }
}

// Build lookup indexes for routing algorithms (e.g. RAPTOR)
gtfs.Compile()

// Access parsed data
for _, agency := range gtfs.AgencyData { ... }
for _, stop   := range gtfs.StopData   { ... }
for _, route  := range gtfs.RouteData  { ... }
for _, trip   := range gtfs.TripData   { ... }

// Access compiled indexes
gtfs.TripStopTimes     // map[*Trip][]StopTime   — stop times per trip, sorted by sequence
gtfs.RouteTrips        // map[*Route][]*Trip      — trips per route
gtfs.StopRoutes        // map[*Stop][]*Route      — routes serving each stop
gtfs.TransfersFromStop // map[*Stop][]Transfer    — outbound transfers per stop
gtfs.FrequenciesByTrip // map[*Trip][]Frequency   — frequencies per trip
```

## API

### Types

```go
type GTFS struct {
    FileName       string
    AgencyData     []Agency
    StopData       []Stop
    RouteData      []Route
    TripData       []Trip
    StopTimeData   []StopTime
    CalendarData   []Calendar
    CalendarDates  []CalendarDate
    ShapeData      []Shape
    FrequencyData  []Frequency
    TransferData   []Transfer
    FareAttributes []FareAttribute
    FareRules      []FareRule
    FeedInfo       []FeedInfo
    PathwayData    []Pathway
    LevelData      []Level
    Attributions   []Attribution
    Translations   []Translation
}
```

### Methods

| Method | Description |
|---|---|
| `ParseAll() error` | Parses all GTFS files in dependency order |
| `ValidateAll() []ValidationError` | Validates parsed data; collects all errors |
| `Compile()` | Builds lookup maps (`TripStopTimes`, `RouteTrips`, `StopRoutes`, `TransfersFromStop`, `FrequenciesByTrip`) for use by routing algorithms |

Individual parsers (`ParseAgency`, `ParseStop`, `ParseRoute`, etc.) are also available if you need to parse files selectively. Parsers must be called in dependency order — see [Parser execution order](#parser-execution-order) below.

### ValidationError

```go
type ValidationError struct {
    File    string
    Field   string
    ID      string
    Message string
}
```

Implements `error`. Formatted as `[file] id="..." field="...": message`.

## GTFS Files

### Required

| File | Struct |
|---|---|
| `agency.txt` | `Agency` |
| `stops.txt` | `Stop` |
| `routes.txt` | `Route` |
| `trips.txt` | `Trip` |
| `stop_times.txt` | `StopTime` |
| `calendar.txt` or `calendar_dates.txt` | `Calendar` / `CalendarDate` |

### Optional

`shapes.txt`, `frequencies.txt`, `transfers.txt`, `fare_attributes.txt`, `fare_rules.txt`, `feed_info.txt`, `pathways.txt`, `levels.txt`, `attributions.txt`, `translations.txt`

Optional files return no error when absent.

## Validation Rules

- **Required files** must be present and non-empty
- **agency.txt**: `agency_name`, `agency_url`, `agency_timezone` required
- **stops.txt**: `stop_id` required; `stop_name` required for `location_type` 0/1; if either `stop_lat` or `stop_lon` is provided the other must be too; coordinate ranges [-90,90] / [-180,180]
- **routes.txt**: at least one of `route_short_name` or `route_long_name` required; valid `route_type`; `agency_id` required when feed has multiple agencies
- **trips.txt**: unresolved `route_id`/`service_id` FK references flagged; `direction_id` must be 0 or 1
- **stop_times.txt**: unresolved `trip_id`/`stop_id` FK references flagged; time format `HH:MM:SS` (hours may exceed 23 for post-midnight service); `stop_sequence` must be strictly increasing within each trip
- **calendar.txt**: day values must be 0 or 1; dates must be valid `YYYYMMDD`

## Parser Execution Order

`ParseAll()` runs parsers in the following dependency order:

1. `ParseAgency`, `ParseLevel`, `ParseCalendar`, `ParseShape` — no dependencies
2. `ParseStop` — depends on Level
3. `ParseRoute` — depends on Agency
4. `ParseCalendarDate` — depends on Calendar
5. `ParseTrip` — depends on Route, Calendar, Shape
6. `ParseStopTime` — depends on Trip, Stop
7. `ParseFrequency` — depends on Trip
8. `ParseTransfer` — depends on Stop, Route, Trip
9. `ParsePathway` — depends on Stop
10. `ParseFareAttribute` — depends on Agency
11. `ParseFareRule` — depends on FareAttribute, Route
12. `ParseFeedInfo`, `ParseAttribution`, `ParseTranslation`

## Explorer CLI

An interactive REPL for exploring a GTFS feed is included under `./cli`.

**Build:**

```sh
go build -o gtfs-cli ./cli/
```

**Run:**

```sh
./gtfs-cli -f path/to/feed.zip
```

**Commands:**

| Command | Description |
|---|---|
| `routes` | List all routes (first 20) |
| `route <id>` | Show all trips for a route |
| `trips` | List all trips (first 20) |
| `trip <id>` | Show stop times for a trip |
| `stops` | List all stops (first 20) |
| `stop <id>` | Show routes serving a stop |
| `help` | Print command reference |
| `quit` / `exit` | Exit |

## Development

```sh
go build ./...
go test ./...
```

No external dependencies — standard library only.

## Licence

MIT. See [LICENSE](LICENSE).
