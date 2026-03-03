package main

import (
	"fmt"

	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

func main() {
	// gtfs := gtfsparser.GTFS{FileName: "gtfs_default3.zip"}
	gtfs := gtfsparser.GTFS{FileName: "gtfs_dup.zip"}

	err := gtfs.ParseAgency()
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("Parsed %d agency record(s):\n", len(gtfs.AgencyData))
	for _, a := range gtfs.AgencyData {
		fmt.Printf("%+v\n", a)
	}
}
