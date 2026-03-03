package gtfsparser

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

func (gtfs *GTFS) ParseAgency() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find agency.txt in the zip
	var agencyFile *zip.File
	for _, f := range r.File {
		if f.Name == "agency.txt" {
			agencyFile = f
			break
		}
	}
	if agencyFile == nil {
		return fmt.Errorf("agency.txt not found in %s", gtfs.FileName)
	}

	// Open and parse as CSV
	rc, err := agencyFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	// Read header row and build column index map
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	// Parse each row into an Agency
	var agencies []Agency
	seen := make(map[string]bool)
	var duplicates []string

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		agency := Agency{
			agency_id:       getCol(row, col, "agency_id"),
			agency_name:     getCol(row, col, "agency_name"),
			agency_url:      getCol(row, col, "agency_url"),
			agency_timezone: getCol(row, col, "agency_timezone"),
			agency_lang:     getCol(row, col, "agency_lang"),
			agency_phone:    getCol(row, col, "agency_phone"),
			agency_fare_url: getCol(row, col, "agency_fare_url"),
			agency_email:    getCol(row, col, "agency_email"),
		}

		if seen[agency.agency_id] {
			duplicates = append(duplicates, agency.agency_id)
		} else {
			seen[agency.agency_id] = true
		}

		agencies = append(agencies, agency)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate agency_id(s) found: %v", duplicates)
	}

	gtfs.AgencyData = agencies
	return nil
}

func (gtfs *GTFS) ParseRoute() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find routes.txt in the zip
	var routesFile *zip.File
	for _, f := range r.File {
		if f.Name == "routes.txt" {
			routesFile = f
			break
		}
	}
	if routesFile == nil {
		return fmt.Errorf("routes.txt not found in %s", gtfs.FileName)
	}

	// Open and parse as CSV
	rc, err := routesFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	// Read header row and build column index map
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	// Parse each row into a Route
	var routes []Route
	seen := make(map[string]bool)
	var duplicates []string

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		routeType, err := strconv.Atoi(getCol(row, col, "route_type"))
		if err != nil {
			routeType = 0
		}
		sortOrder, err := strconv.Atoi(getCol(row, col, "route_sort_order"))
		if err != nil {
			sortOrder = 0
		}
		contPickup, err := strconv.Atoi(getCol(row, col, "continuous_pickup"))
		if err != nil {
			contPickup = 0
		}
		contDropOff, err := strconv.Atoi(getCol(row, col, "continuous_drop_off"))
		if err != nil {
			contDropOff = 0
		}

		// Resolve agency_id FK to matching Agency in gtfs.AgencyData
		var agencyPtr *Agency
		agencyIDStr := getCol(row, col, "agency_id")
		for i := range gtfs.AgencyData {
			if gtfs.AgencyData[i].agency_id == agencyIDStr {
				agencyPtr = &gtfs.AgencyData[i]
				break
			}
		}

		route := Route{
			route_id:            getCol(row, col, "route_id"),
			agency_id:           agencyPtr,
			route_short_name:    getCol(row, col, "route_short_name"),
			route_long_name:     getCol(row, col, "route_long_name"),
			route_desc:          getCol(row, col, "route_desc"),
			route_type:          RouteType(routeType),
			route_url:           getCol(row, col, "route_url"),
			route_color:         getCol(row, col, "route_color"),
			route_text_color:    getCol(row, col, "route_text_color"),
			route_sort_order:    sortOrder,
			continuous_pickup:   PickupDropOffType(contPickup),
			continuous_drop_off: PickupDropOffType(contDropOff),
			network_id:          getCol(row, col, "network_id"),
		}

		if seen[route.route_id] {
			duplicates = append(duplicates, route.route_id)
		} else {
			seen[route.route_id] = true
		}

		routes = append(routes, route)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate route_id(s) found: %v", duplicates)
	}

	gtfs.RouteData = routes
	return nil
}

func (gtfs *GTFS) ParseStop() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find stops.txt in the zip
	var stopsFile *zip.File
	for _, f := range r.File {
		if f.Name == "stops.txt" {
			stopsFile = f
			break
		}
	}
	if stopsFile == nil {
		return fmt.Errorf("stops.txt not found in %s", gtfs.FileName)
	}

	// Open and parse as CSV
	rc, err := stopsFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	// Read header row and build column index map
	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	// Parse each row into a Stop
	type stopRaw struct {
		stop            Stop
		parentStationID string
		levelID         string
	}
	var rawStops []stopRaw
	seen := make(map[string]int) // stop_id -> first seen line number
	type duplicate struct {
		stop_id string
		line    int
	}
	var duplicates []duplicate
	lineNum := 1 // start at 1 since header is line 1

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		lineNum++

		lat, err := parseOptionalFloat(getCol(row, col, "stop_lat"))
		if err != nil {
			return fmt.Errorf("invalid stop_lat for stop_id %s: %w", getCol(row, col, "stop_id"), err)
		}
		lon, err := parseOptionalFloat(getCol(row, col, "stop_lon"))
		if err != nil {
			return fmt.Errorf("invalid stop_lon for stop_id %s: %w", getCol(row, col, "stop_id"), err)
		}

		locType, err := strconv.Atoi(getCol(row, col, "location_type"))
		if err != nil {
			locType = 0 // default: stop/platform
		}
		wheelchair, err := strconv.Atoi(getCol(row, col, "wheelchair_boarding"))
		if err != nil {
			wheelchair = 0 // default: no info
		}

		stop := Stop{
			stop_id:             getCol(row, col, "stop_id"),
			stop_code:           getCol(row, col, "stop_code"),
			stop_name:           getCol(row, col, "stop_name"),
			stop_desc:           getCol(row, col, "stop_desc"),
			stop_lat:            lat,
			stop_lon:            lon,
			zone_id:             getCol(row, col, "zone_id"),
			stop_url:            getCol(row, col, "stop_url"),
			location_type:       LocationType(locType),
			parent_station:      nil, // resolved in second pass
			stop_timezone:       getCol(row, col, "stop_timezone"),
			wheelchair_boarding: WheelchairBoarding(wheelchair),
			level_id:            nil, // resolved in second pass
			platform_code:       getCol(row, col, "platform_code"),
		}

		if _, exists := seen[stop.stop_id]; exists {
			duplicates = append(duplicates, duplicate{stop_id: stop.stop_id, line: lineNum})
		} else {
			seen[stop.stop_id] = lineNum
		}

		rawStops = append(rawStops, stopRaw{
			stop:            stop,
			parentStationID: getCol(row, col, "parent_station"),
			levelID:         getCol(row, col, "level_id"),
		})
	}

	if len(duplicates) > 0 {
		msg := "duplicate stop_id(s) found:"
		for _, d := range duplicates {
			msg += fmt.Sprintf("\n  stop_id %q at line %d (first seen at line %d)", d.stop_id, d.line, seen[d.stop_id])
		}
		return fmt.Errorf("%s", msg)
	}

	// Build stop index for parent_station resolution
	stops := make([]Stop, len(rawStops))
	for i, rs := range rawStops {
		stops[i] = rs.stop
	}

	// Second pass: resolve parent_station and level_id FKs
	stopIndex := make(map[string]*Stop)
	for i := range stops {
		stopIndex[stops[i].stop_id] = &stops[i]
	}
	levelIndex := make(map[string]*Level)
	for i := range gtfs.LevelData {
		levelIndex[gtfs.LevelData[i].level_id] = &gtfs.LevelData[i]
	}
	for i, rs := range rawStops {
		if rs.parentStationID != "" {
			stops[i].parent_station = stopIndex[rs.parentStationID]
		}
		if rs.levelID != "" {
			stops[i].level_id = levelIndex[rs.levelID]
		}
	}

	gtfs.StopData = stops
	return nil
}
