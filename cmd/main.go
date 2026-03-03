package main

import (
	"fmt"

	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

func main() {
	gtfs := gtfsparser.GTFS{FileName: "gtfs_default3.zip"}

	if err := gtfs.ParseAgency(); err != nil {
		fmt.Println("error parsing agency:", err)
		return
	}
	if err := gtfs.ParseRoute(); err != nil {
		fmt.Println("error parsing routes:", err)
		return
	}
	if err := gtfs.ParseCalendar(); err != nil {
		fmt.Println("error parsing calendar:", err)
		return
	}
	if err := gtfs.ParseShape(); err != nil {
		fmt.Println("error parsing shapes:", err)
		return
	}
	if err := gtfs.ParseTrip(); err != nil {
		fmt.Println("error parsing trips:", err)
		return
	}
	if err := gtfs.ParseStop(); err != nil {
		fmt.Println("error parsing stops:", err)
		return
	}
	if err := gtfs.ParseStopTime(); err != nil {
		fmt.Println("error parsing stop times:", err)
		return
	}
	if err := gtfs.ParseFrequency(); err != nil {
		fmt.Println("error parsing frequencies:", err)
		return
	}
	if err := gtfs.ParseTransfer(); err != nil {
		fmt.Println("error parsing transfers:", err)
		return
	}

	fmt.Printf("Parsed %d agency(s)\n", len(gtfs.AgencyData))
	fmt.Printf("Parsed %d route(s)\n", len(gtfs.RouteData))
	fmt.Printf("Parsed %d calendar(s)\n", len(gtfs.CalendarData))
	fmt.Printf("Parsed %d trip(s)\n", len(gtfs.TripData))
	fmt.Printf("Parsed %d stop(s)\n", len(gtfs.StopData))
	fmt.Printf("Parsed %d stop time(s)\n", len(gtfs.StopTimeData))
	fmt.Printf("Parsed %d frequency(s)\n", len(gtfs.FrequencyData))
	fmt.Printf("Parsed %d transfer(s). First 5:\n", len(gtfs.TransferData))
	for i, t := range gtfs.TransferData {
		if i >= 5 {
			break
		}
		fmt.Printf("%+v\n", t)
	}
}
