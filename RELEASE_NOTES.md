# Release Notes

## CLI Explorer

- **Welcome message** — startup now displays a banner with a short description of the tool and feed stats (routes, trips, stops)
- **Transfer support** — `stop <id>` now shows outbound transfers, including the destination stop and transfer type, plus the routes serving that destination stop

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
