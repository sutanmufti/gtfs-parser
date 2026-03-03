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
}
