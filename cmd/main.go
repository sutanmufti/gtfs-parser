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

	fmt.Printf("Parsed %d agency record(s)\n", len(gtfs.AgencyData))
	fmt.Printf("Parsed %d route(s). First 5:\n", len(gtfs.RouteData))
	for i, r := range gtfs.RouteData {
		if i >= 5 {
			break
		}
		fmt.Printf("%+v\n", r)
	}
}
