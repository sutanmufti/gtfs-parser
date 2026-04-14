# Release Notes

## V1.2.0

### New lookup maps in `Compile()`

Five new maps are now populated when `Compile()` is called:

| Map | Type | Description |
|---|---|---|
| `TransfersToStop` | `map[*Stop][]Transfer` | Inbound transfers keyed by destination stop — reverse of `TransfersFromStop` |
| `RouteStops` | `map[*Route][]*Stop` | All unique stops served by a route (derived via `RouteTrips` + `TripStopTimes`) |
| `StopTripTimes` | `map[*Stop][]StopTime` | All stop-time records keyed by stop |
| `TripIndex` | `map[string]*Trip` | Direct trip lookup by `trip_id` string |

`TransfersFromStop` and `TransfersToStop` are now built in a single pass over `TransferData`.

---

## V1.1.4

trims white space.

## CLI Explorer
- **Welcome message** — startup now displays a banner with a short description of the tool and feed stats (routes, trips, stops)
- **Transfer support** — `stop <id>` now shows outbound transfers, including the destination stop and transfer type, plus the routes serving that destination stop
- **Richer route detail** — `route <id>` now shows route type, agency, and short name (when both short and long names are present)
- **Richer trip detail** — `trip <id>` now shows the parent route, service ID, headsign, and direction

## Installation

Download the binary for your platform from the assets below, make it executable, and run it:

```sh
chmod +x gtfs-cli
./gtfs-cli -f path/to/feed.zip
```

### Available Commands

| Command | Description |
|---|---|
| `routes` | List all routes |
| `route <id>` | Show trips for a route |
| `trips` | List all trips |
| `trip <id>` | Show stop times for a trip |
| `stops` | List all stops |
| `stop <id>` | Show routes serving a stop and outbound transfers |
| `next` / `prev` | Paginate list results |
| `help` | Print command reference |
| `quit` / `exit` | Exit |
