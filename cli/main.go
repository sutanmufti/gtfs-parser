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

func colorf(code, s string) string { return code + s + reset }
func id(s string) string           { return colorf(yellow+bold, s) }
func header(s string) string       { return colorf(bold, s) }
func label(s string) string        { return colorf(dim, s) }
func errorf(s string) string       { return colorf(red, s) }
func highlight(s string) string    { return colorf(green, s) }
func commandf(s string) string     { return colorf(cyan, s) }

func totalFor(gtfs *gtfsparser.GTFS, cmd string) int {
	switch cmd {
	case "routes":
		return len(gtfs.RouteData)
	case "trips":
		return len(gtfs.TripData)
	case "stops":
		return len(gtfs.StopData)
	}
	return 0
}

// printPage prints up to maxList items from the given list command starting at
// offset. Returns the number of items printed and whether more items remain.
func printPage(gtfs *gtfsparser.GTFS, cmd string, offset int) (printed int, hasMore bool) {
	switch cmd {
	case "routes":
		total := len(gtfs.RouteData)
		for i := offset; i < total; i++ {
			r := gtfs.RouteData[i]
			name := r.RouteLongName
			if name == "" {
				name = r.RouteShortName
			}
			fmt.Printf("  %-30s %s\n", id(r.RouteID), name)
			printed++
			if printed == maxList {
				hasMore = i+1 < total
				break
			}
		}
	case "trips":
		total := len(gtfs.TripData)
		for i := offset; i < total; i++ {
			t := gtfs.TripData[i]
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
				hasMore = i+1 < total
				break
			}
		}
	case "stops":
		total := len(gtfs.StopData)
		for i := offset; i < total; i++ {
			s := gtfs.StopData[i]
			fmt.Printf("  %-30s %s\n", id(s.StopID), s.StopName)
			printed++
			if printed == maxList {
				hasMore = i+1 < total
				break
			}
		}
	}
	return printed, hasMore
}

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

	fmt.Println(colorf(cyan+bold, "gtfs-cli") + colorf(dim, " — GTFS feed explorer"))
	fmt.Println(colorf(dim, "Browse routes, trips, and stops from any GTFS zip feed."))
	fmt.Println()
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

	// Pagination state
	pageCmd := ""
	pageStart := 0  // start offset of the currently displayed page
	pageOffset := 0 // start offset of the next page

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

		case "routes", "trips", "stops":
			pageCmd = cmd
			pageStart = 0
			pageOffset = 0
			total := totalFor(&gtfs, cmd)
			n, hasMore := printPage(&gtfs, cmd, pageOffset)
			pageOffset += n
			if hasMore {
				fmt.Printf("  %s\n", label(fmt.Sprintf("showing %d–%d of %d — type 'next' for more", pageStart+1, pageOffset, total)))
			} else {
				fmt.Printf("  %s\n", label(fmt.Sprintf("%d total", total)))
				pageCmd = ""
			}

		case "next":
			if pageCmd == "" {
				fmt.Println(errorf("nothing to page — run routes, trips, or stops first"))
				continue
			}
			total := totalFor(&gtfs, pageCmd)
			pageStart = pageOffset
			n, hasMore := printPage(&gtfs, pageCmd, pageOffset)
			pageOffset += n
			if hasMore {
				fmt.Printf("  %s\n", label(fmt.Sprintf("showing %d–%d of %d — type 'next' / 'prev' for more", pageStart+1, pageOffset, total)))
			} else {
				fmt.Printf("  %s\n", label(fmt.Sprintf("%d total — end of list", total)))
				pageCmd = ""
			}

		case "prev":
			if pageCmd == "" {
				fmt.Println(errorf("nothing to page — run routes, trips, or stops first"))
				continue
			}
			if pageStart == 0 {
				fmt.Println(errorf("already at the beginning"))
				continue
			}
			total := totalFor(&gtfs, pageCmd)
			pageOffset = pageStart
			pageStart = max(0, pageStart-maxList)
			n, hasMore := printPage(&gtfs, pageCmd, pageStart)
			_ = n
			hint := "type 'next' / 'prev' to navigate"
			if !hasMore {
				hint = "type 'next' for more"
			}
			fmt.Printf("  %s\n", label(fmt.Sprintf("showing %d–%d of %d — %s", pageStart+1, pageOffset, total, hint)))

		case "route":
			pageCmd = ""
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
				name := r.RouteLongName
				if name == "" {
					name = r.RouteShortName
				}
				fmt.Printf("%s %s %s %s %s\n",
					header("Route"), id(r.RouteID),
					label("("+name+")"),
					label("—"), highlight(fmt.Sprintf("%d trip(s)", len(trips))),
				)
				if r.RouteShortName != "" && r.RouteLongName != "" {
					fmt.Printf("  %s %s\n", label("short name:"), r.RouteShortName)
				}
				agencyName := ""
				if r.AgencyID != nil {
					agencyName = r.AgencyID.AgencyName
				}
				fmt.Printf("  %s %s  %s %s\n",
					label("type:"), highlight(fmt.Sprintf("%d", r.RouteType)),
					label("agency:"), agencyName,
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

		case "trip":
			pageCmd = ""
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
				routeID := ""
				if t.RouteID != nil {
					routeID = t.RouteID.RouteID
				}
				serviceID := ""
				if t.ServiceID != nil {
					serviceID = t.ServiceID.ServiceID
				}
				fmt.Printf("%s %s %s %s\n",
					header("Trip"), id(t.TripID),
					label("—"), highlight(fmt.Sprintf("%d stop(s)", len(stopTimes))),
				)
				fmt.Printf("  %s %s  %s %s\n",
					label("route:"), id(routeID),
					label("service:"), serviceID,
				)
				fmt.Printf("  %s %s  %s %s\n",
					label("headsign:"), t.TripHeadsign,
					label("direction:"), highlight(fmt.Sprintf("%d", t.DirectionID)),
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

		case "stop":
			pageCmd = ""
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
				transfers := gtfs.TransfersFromStop[s]
				fmt.Printf("%s %s %s %s %s\n",
					header("Stop"), id(s.StopID),
					label("("+s.StopName+")"),
					label("—"), highlight(fmt.Sprintf("%d route(s)", len(routes))),
				)
				for _, r := range routes {
					fmt.Printf("  %s %s\n", label("route:"), id(r.RouteID))
				}
				if len(transfers) > 0 {
					fmt.Printf("  %s\n", label(fmt.Sprintf("%d transfer(s):", len(transfers))))
					for _, t := range transfers {
						toStop := ""
						if t.ToStopID != nil {
							toStop = t.ToStopID.StopID
						}
						fmt.Printf("    %s %s %s %s\n",
							label("-> stop:"), id(toStop),
							label("type:"), highlight(fmt.Sprintf("%d", t.TransferType)),
						)
						for _, r := range gtfs.StopRoutes[t.ToStopID] {
							fmt.Printf("            %s %s\n", label("route:"), highlight(r.RouteID))
						}
					}
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
			fmt.Printf("  %-30s show routes serving a stop and outbound transfers\n", commandf("stop <id>"))
			fmt.Printf("  %-30s show next page of last list\n", commandf("next"))
			fmt.Printf("  %-30s show previous page of last list\n", commandf("prev"))
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
