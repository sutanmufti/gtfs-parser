package main

import (
	"fmt"

	gtfsparser "github.com/sutanmufti/gtfs-parser"
)

func main() {
	gtfs := gtfsparser.GTFS{FileName: "gtfs_default3.zip"}

	err := gtfs.ParseStop()
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Printf("Parsed %d stop(s). First 5:\n", len(gtfs.StopData))
	for i, s := range gtfs.StopData {
		if i >= 5 {
			break
		}
		fmt.Printf("%+v\n", s)
	}
}
