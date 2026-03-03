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

	fmt.Printf("Parsed %d calendar(s)\n", len(gtfs.CalendarData))
	fmt.Printf("Parsed %d shape(s)\n", len(gtfs.ShapeData))
	fmt.Printf("Parsed %d trip(s). First 5:\n", len(gtfs.TripData))
	for i, t := range gtfs.TripData {
		if i >= 5 {
			break
		}
		fmt.Printf("%+v\n", t)
	}
}
