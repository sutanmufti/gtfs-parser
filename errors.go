// Package gtfsparser provides a parser and validator for GTFS Schedule feeds.
//
// A GTFS (General Transit Feed Specification) feed is a collection of CSV files
// distributed as a zip archive. This package reads the zip, parses each file into
// typed Go structs, and validates the data against the GTFS Schedule specification.
//
// Typical usage:
//
//	gtfs := gtfsparser.GTFS{FileName: "feed.zip"}
//	if err := gtfs.ParseAll(); err != nil {
//	    log.Fatal(err)
//	}
//	errs := gtfs.ValidateAll()
//	for _, e := range errs {
//	    fmt.Println(e)
//	}
package gtfsparser

import "fmt"

// ValidationError describes a single validation failure within a GTFS feed.
// File identifies the source file, Field identifies the column, ID identifies
// the record, and Message describes the problem.
type ValidationError struct {
	File    string
	Field   string
	ID      string
	Message string
}

// Error implements the error interface, formatting the validation failure
// as [file] id="..." field="...": message.
func (e ValidationError) Error() string {
	return fmt.Sprintf("[%s] id=%q field=%q: %s", e.File, e.ID, e.Field, e.Message)
}
