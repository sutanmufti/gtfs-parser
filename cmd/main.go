package main

import (
	"fmt"

	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

func main() {
	gtfs := gtfsparser.GTFS{FileName: "gtfs_default3.zip"}

	if err := gtfs.ParseAll(); err != nil {
		fmt.Println(err)
		return
	}

	errs := gtfs.ValidateAll()
	if len(errs) == 0 {
		fmt.Println("validation passed: no errors")
	} else {
		fmt.Printf("%d validation error(s):\n", len(errs))
		for _, e := range errs {
			fmt.Println(" ", e)
		}
	}

	gtfs.Compile()

	count := 0
	for trip, stopTimes := range gtfs.TripStopTimes {
		fmt.Printf("\nTrip: %s (route: %s)\n", trip.TripID, trip.RouteID.RouteID)
		for _, st := range stopTimes {
			fmt.Printf("  seq %d  stop %-20s  arr %s  dep %s\n",
				st.StopSequence, st.StopID.StopID, st.ArrivalTime, st.DepartureTime)
		}
		count++
		if count == 3 {
			break
		}
	}

	fmt.Println("\n--- RouteTrips (3 routes) ---")
	count = 0
	for route, trips := range gtfs.RouteTrips {
		fmt.Printf("\nRoute: %s\n", route.RouteID)
		for _, t := range trips {
			fmt.Printf("  trip: %s\n", t.TripID)
		}
		count++
		if count == 3 {
			break
		}
	}

	fmt.Println("\n--- StopRoutes (3 stops) ---")
	count = 0
	for stop, routes := range gtfs.StopRoutes {
		fmt.Printf("\nStop: %s\n", stop.StopID)
		for _, r := range routes {
			fmt.Printf("  route: %s\n", r.RouteID)
		}
		count++
		if count == 3 {
			break
		}
	}
}
