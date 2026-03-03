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

	fmt.Printf("agency:         %d\n", len(gtfs.AgencyData))
	fmt.Printf("routes:         %d\n", len(gtfs.RouteData))
	fmt.Printf("calendar:       %d\n", len(gtfs.CalendarData))
	fmt.Printf("calendar_dates: %d\n", len(gtfs.CalendarDates))
	fmt.Printf("shapes:         %d\n", len(gtfs.ShapeData))
	fmt.Printf("levels:         %d\n", len(gtfs.LevelData))
	fmt.Printf("stops:          %d\n", len(gtfs.StopData))
	fmt.Printf("trips:          %d\n", len(gtfs.TripData))
	fmt.Printf("stop_times:     %d\n", len(gtfs.StopTimeData))
	fmt.Printf("frequencies:    %d\n", len(gtfs.FrequencyData))
	fmt.Printf("transfers:      %d\n", len(gtfs.TransferData))
	fmt.Printf("pathways:       %d\n", len(gtfs.PathwayData))
	fmt.Printf("fare_attributes:%d\n", len(gtfs.FareAttributes))
	fmt.Printf("fare_rules:     %d\n", len(gtfs.FareRules))
	fmt.Printf("feed_info:      %d\n", len(gtfs.FeedInfo))
	fmt.Printf("attributions:   %d\n", len(gtfs.Attributions))
	fmt.Printf("translations:   %d\n", len(gtfs.Translations))
}
