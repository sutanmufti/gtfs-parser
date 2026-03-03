package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

const maxList = 20

// ANSI colour helpers
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

func colorf(code, s string) string  { return code + s + reset }
func id(s string) string            { return colorf(yellow+bold, s) }
func header(s string) string        { return colorf(bold, s) }
func label(s string) string         { return colorf(dim, s) }
func errorf(s string) string        { return colorf(red, s) }
func highlight(s string) string     { return colorf(green, s) }
func commandf(s string) string      { return colorf(cyan, s) }

func main() {
	file := flag.String("f", "", "path to GTFS zip file")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, errorf("usage: gtfs-cli -f <feed.zip>"))
		os.Exit(1)
	}

	gtfs := gtfsparser.GTFS{FileName: *file}
	if err := gtfs.ParseAll(); err != nil {
		fmt.Fprintln(os.Stderr, errorf("parse error: "+err.Error()))
		os.Exit(1)
	}
	gtfs.Compile()

	fmt.Printf("Loaded %s  (%s routes  %s trips  %s stops)\n",
		highlight(*file),
		highlight(fmt.Sprintf("%d", len(gtfs.RouteData))),
		highlight(fmt.Sprintf("%d", len(gtfs.TripData))),
		highlight(fmt.Sprintf("%d", len(gtfs.StopData))),
	)
	fmt.Printf("%s: %s  %s  %s  %s  %s  %s  %s  %s\n",
		label("commands"),
		commandf("routes"), commandf("route <id>"),
		commandf("trips"), commandf("trip <id>"),
		commandf("stops"), commandf("stop <id>"),
		commandf("help"), commandf("quit"),
	)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n" + colorf(cyan+bold, "gtfs> "))
		if !scanner.Scan() {
			break
		}
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}

		switch cmd {

		case "routes":
			printed := 0
			for _, r := range gtfs.RouteData {
				name := r.RouteLongName
				if name == "" {
					name = r.RouteShortName
				}
				fmt.Printf("  %-30s %s\n", id(r.RouteID), name)
				printed++
				if printed == maxList {
					fmt.Printf("  %s\n", label(fmt.Sprintf("... (%d total)", len(gtfs.RouteData))))
					break
				}
			}

		case "route":
			if arg == "" {
				fmt.Println(errorf("usage: route <route_id>"))
				continue
			}
			found := false
			for i := range gtfs.RouteData {
				r := &gtfs.RouteData[i]
				if r.RouteID != arg {
					continue
				}
				found = true
				trips := gtfs.RouteTrips[r]
				fmt.Printf("%s %s %s %s\n",
					header("Route"), id(r.RouteID),
					label("—"), highlight(fmt.Sprintf("%d trip(s)", len(trips))),
				)
				for _, t := range trips {
					fmt.Printf("  %s %s  %s %s\n",
						label("trip:"), id(t.TripID),
						label("headsign:"), t.TripHeadsign,
					)
				}
			}
			if !found {
				fmt.Printf("%s\n", errorf(fmt.Sprintf("route %q not found", arg)))
			}

		case "trips":
			printed := 0
			for _, t := range gtfs.TripData {
				routeID := ""
				if t.RouteID != nil {
					routeID = t.RouteID.RouteID
				}
				fmt.Printf("  %-35s %s %-18s %s\n",
					id(t.TripID),
					label("route:"), id(routeID),
					t.TripHeadsign,
				)
				printed++
				if printed == maxList {
					fmt.Printf("  %s\n", label(fmt.Sprintf("... (%d total)", len(gtfs.TripData))))
					break
				}
			}

		case "trip":
			if arg == "" {
				fmt.Println(errorf("usage: trip <trip_id>"))
				continue
			}
			found := false
			for i := range gtfs.TripData {
				t := &gtfs.TripData[i]
				if t.TripID != arg {
					continue
				}
				found = true
				stopTimes := gtfs.TripStopTimes[t]
				fmt.Printf("%s %s %s %s\n",
					header("Trip"), id(t.TripID),
					label("—"), highlight(fmt.Sprintf("%d stop(s)", len(stopTimes))),
				)
				for _, st := range stopTimes {
					stopID := ""
					if st.StopID != nil {
						stopID = st.StopID.StopID
					}
					fmt.Printf("  %s %-4d %s %-25s %s %s  %s %s\n",
						label("seq"), st.StopSequence,
						label("stop"), id(stopID),
						label("arr"), st.ArrivalTime,
						label("dep"), st.DepartureTime,
					)
				}
			}
			if !found {
				fmt.Printf("%s\n", errorf(fmt.Sprintf("trip %q not found", arg)))
			}

		case "stops":
			printed := 0
			for _, s := range gtfs.StopData {
				fmt.Printf("  %-30s %s\n", id(s.StopID), s.StopName)
				printed++
				if printed == maxList {
					fmt.Printf("  %s\n", label(fmt.Sprintf("... (%d total)", len(gtfs.StopData))))
					break
				}
			}

		case "stop":
			if arg == "" {
				fmt.Println(errorf("usage: stop <stop_id>"))
				continue
			}
			found := false
			for i := range gtfs.StopData {
				s := &gtfs.StopData[i]
				if s.StopID != arg {
					continue
				}
				found = true
				routes := gtfs.StopRoutes[s]
				fmt.Printf("%s %s %s %s %s\n",
					header("Stop"), id(s.StopID),
					label("("+s.StopName+")"),
					label("—"), highlight(fmt.Sprintf("%d route(s)", len(routes))),
				)
				for _, r := range routes {
					fmt.Printf("  %s %s\n", label("route:"), id(r.RouteID))
				}
			}
			if !found {
				fmt.Printf("%s\n", errorf(fmt.Sprintf("stop %q not found", arg)))
			}

		case "help":
			fmt.Printf("  %-30s list all routes (first %d)\n", commandf("routes"), maxList)
			fmt.Printf("  %-30s show trips for a route\n", commandf("route <id>"))
			fmt.Printf("  %-30s list all trips (first %d)\n", commandf("trips"), maxList)
			fmt.Printf("  %-30s show stop times for a trip\n", commandf("trip <id>"))
			fmt.Printf("  %-30s list all stops (first %d)\n", commandf("stops"), maxList)
			fmt.Printf("  %-30s show routes serving a stop\n", commandf("stop <id>"))
			fmt.Printf("  %-30s exit\n", commandf("quit / exit"))

		case "quit", "exit":
			return

		default:
			fmt.Printf("%s — %s\n",
				errorf(fmt.Sprintf("unknown command %q", cmd)),
				label("type 'help' for available commands"),
			)
		}
	}
}
