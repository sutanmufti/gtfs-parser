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

func (gtfs *GTFS) ParseTrip() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find trips.txt in the zip
	var tripsFile *zip.File
	for _, f := range r.File {
		if f.Name == "trips.txt" {
			tripsFile = f
			break
		}
	}
	if tripsFile == nil {
		return fmt.Errorf("trips.txt not found in %s", gtfs.FileName)
	}

	// Open and parse as CSV
	rc, err := tripsFile.Open()
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

	// Build indexes for FK resolution
	routeIndex := make(map[string]*Route)
	for i := range gtfs.RouteData {
		routeIndex[gtfs.RouteData[i].route_id] = &gtfs.RouteData[i]
	}
	calendarIndex := make(map[string]*Calendar)
	for i := range gtfs.CalendarData {
		calendarIndex[gtfs.CalendarData[i].service_id] = &gtfs.CalendarData[i]
	}
	shapeIndex := make(map[string]*Shape)
	for i := range gtfs.ShapeData {
		shapeIndex[gtfs.ShapeData[i].shape_id] = &gtfs.ShapeData[i]
	}

	// Parse each row into a Trip
	var trips []Trip
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

		directionId, err := strconv.Atoi(getCol(row, col, "direction_id"))
		if err != nil {
			directionId = 0
		}
		wheelchairAccessible, err := strconv.Atoi(getCol(row, col, "wheelchair_accessible"))
		if err != nil {
			wheelchairAccessible = 0
		}
		bikesAllowed, err := strconv.Atoi(getCol(row, col, "bikes_allowed"))
		if err != nil {
			bikesAllowed = 0
		}

		trip := Trip{
			route_id:              routeIndex[getCol(row, col, "route_id")],
			service_id:            calendarIndex[getCol(row, col, "service_id")],
			trip_id:               getCol(row, col, "trip_id"),
			trip_headsign:         getCol(row, col, "trip_headsign"),
			trip_short_name:       getCol(row, col, "trip_short_name"),
			direction_id:          DirectionId(directionId),
			block_id:              getCol(row, col, "block_id"),
			shape_id:              shapeIndex[getCol(row, col, "shape_id")],
			wheelchair_accessible: WheelchairAccessibleEnum(wheelchairAccessible),
			bikes_allowed:         BikesAllowed(bikesAllowed),
		}

		if seen[trip.trip_id] {
			duplicates = append(duplicates, trip.trip_id)
		} else {
			seen[trip.trip_id] = true
		}

		trips = append(trips, trip)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate trip_id(s) found: %v", duplicates)
	}

	gtfs.TripData = trips
	return nil
}

func (gtfs *GTFS) ParseCalendar() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	var calendarFile *zip.File
	for _, f := range r.File {
		if f.Name == "calendar.txt" {
			calendarFile = f
			break
		}
	}
	if calendarFile == nil {
		return fmt.Errorf("calendar.txt not found in %s", gtfs.FileName)
	}

	rc, err := calendarFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	var calendars []Calendar
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

		monday, _ := strconv.Atoi(getCol(row, col, "monday"))
		tuesday, _ := strconv.Atoi(getCol(row, col, "tuesday"))
		wednesday, _ := strconv.Atoi(getCol(row, col, "wednesday"))
		thursday, _ := strconv.Atoi(getCol(row, col, "thursday"))
		friday, _ := strconv.Atoi(getCol(row, col, "friday"))
		saturday, _ := strconv.Atoi(getCol(row, col, "saturday"))
		sunday, _ := strconv.Atoi(getCol(row, col, "sunday"))

		calendar := Calendar{
			service_id: getCol(row, col, "service_id"),
			monday:     monday,
			tuesday:    tuesday,
			wednesday:  wednesday,
			thursday:   thursday,
			friday:     friday,
			saturday:   saturday,
			sunday:     sunday,
			start_date: getCol(row, col, "start_date"),
			end_date:   getCol(row, col, "end_date"),
		}

		if seen[calendar.service_id] {
			duplicates = append(duplicates, calendar.service_id)
		} else {
			seen[calendar.service_id] = true
		}

		calendars = append(calendars, calendar)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate service_id(s) found: %v", duplicates)
	}

	gtfs.CalendarData = calendars
	return nil
}

func (gtfs *GTFS) ParseShape() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	var shapesFile *zip.File
	for _, f := range r.File {
		if f.Name == "shapes.txt" {
			shapesFile = f
			break
		}
	}
	if shapesFile == nil {
		return nil // shapes.txt is optional
	}

	rc, err := shapesFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	var shapes []Shape
	seen := make(map[string]bool) // tracks (shape_id, shape_pt_sequence) uniqueness
	var duplicates []string

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		shapeID := getCol(row, col, "shape_id")
		seqStr := getCol(row, col, "shape_pt_sequence")

		lat, err := strconv.ParseFloat(getCol(row, col, "shape_pt_lat"), 64)
		if err != nil {
			return fmt.Errorf("invalid shape_pt_lat for shape_id %s: %w", shapeID, err)
		}
		lon, err := strconv.ParseFloat(getCol(row, col, "shape_pt_lon"), 64)
		if err != nil {
			return fmt.Errorf("invalid shape_pt_lon for shape_id %s: %w", shapeID, err)
		}
		seq, err := strconv.Atoi(seqStr)
		if err != nil {
			return fmt.Errorf("invalid shape_pt_sequence for shape_id %s: %w", shapeID, err)
		}
		distTraveled, _ := strconv.ParseFloat(getCol(row, col, "shape_dist_traveled"), 64)

		key := shapeID + "|" + seqStr
		if seen[key] {
			duplicates = append(duplicates, key)
		} else {
			seen[key] = true
		}

		shapes = append(shapes, Shape{
			shape_id:            shapeID,
			shape_pt_lat:        lat,
			shape_pt_lon:        lon,
			shape_pt_sequence:   seq,
			shape_dist_traveled: distTraveled,
		})
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate (shape_id, shape_pt_sequence) found: %v", duplicates)
	}

	gtfs.ShapeData = shapes
	return nil
}

func (gtfs *GTFS) ParseStopTime() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	var stopTimesFile *zip.File
	for _, f := range r.File {
		if f.Name == "stop_times.txt" {
			stopTimesFile = f
			break
		}
	}
	if stopTimesFile == nil {
		return fmt.Errorf("stop_times.txt not found in %s", gtfs.FileName)
	}

	rc, err := stopTimesFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	// Build indexes for FK resolution
	tripIndex := make(map[string]*Trip)
	for i := range gtfs.TripData {
		tripIndex[gtfs.TripData[i].trip_id] = &gtfs.TripData[i]
	}
	stopIndex := make(map[string]*Stop)
	for i := range gtfs.StopData {
		stopIndex[gtfs.StopData[i].stop_id] = &gtfs.StopData[i]
	}

	var stopTimes []StopTime
	seen := make(map[string]bool) // (trip_id, stop_sequence) uniqueness
	var duplicates []string

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		tripIDStr := getCol(row, col, "trip_id")
		seqStr := getCol(row, col, "stop_sequence")

		seq, err := strconv.Atoi(seqStr)
		if err != nil {
			return fmt.Errorf("invalid stop_sequence for trip_id %s: %w", tripIDStr, err)
		}

		pickupType, _ := strconv.Atoi(getCol(row, col, "pickup_type"))
		dropOffType, _ := strconv.Atoi(getCol(row, col, "drop_off_type"))
		contPickup, _ := strconv.Atoi(getCol(row, col, "continuous_pickup"))
		contDropOff, _ := strconv.Atoi(getCol(row, col, "continuous_drop_off"))
		distTraveled, _ := strconv.ParseFloat(getCol(row, col, "shape_dist_traveled"), 64)
		timepoint, _ := strconv.Atoi(getCol(row, col, "timepoint"))

		stopTime := StopTime{
			trip_id:             tripIndex[tripIDStr],
			arrival_time:        getCol(row, col, "arrival_time"),
			departure_time:      getCol(row, col, "departure_time"),
			stop_id:             stopIndex[getCol(row, col, "stop_id")],
			stop_sequence:       seq,
			stop_headsign:       getCol(row, col, "stop_headsign"),
			pickup_type:         PickupDropOffType(pickupType),
			drop_off_type:       PickupDropOffType(dropOffType),
			continuous_pickup:   PickupDropOffType(contPickup),
			continuous_drop_off: PickupDropOffType(contDropOff),
			shape_dist_traveled: distTraveled,
			timepoint:           Timepoint(timepoint),
		}

		key := tripIDStr + "|" + seqStr
		if seen[key] {
			duplicates = append(duplicates, key)
		} else {
			seen[key] = true
		}

		stopTimes = append(stopTimes, stopTime)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate (trip_id, stop_sequence) found: %v", duplicates)
	}

	gtfs.StopTimeData = stopTimes
	return nil
}

func (gtfs *GTFS) ParseFrequency() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	var frequenciesFile *zip.File
	for _, f := range r.File {
		if f.Name == "frequencies.txt" {
			frequenciesFile = f
			break
		}
	}
	if frequenciesFile == nil {
		return nil // frequencies.txt is optional
	}

	rc, err := frequenciesFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	tripIndex := make(map[string]*Trip)
	for i := range gtfs.TripData {
		tripIndex[gtfs.TripData[i].trip_id] = &gtfs.TripData[i]
	}

	var frequencies []Frequency

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		headwaySecs, err := strconv.Atoi(getCol(row, col, "headway_secs"))
		if err != nil {
			return fmt.Errorf("invalid headway_secs for trip_id %s: %w", getCol(row, col, "trip_id"), err)
		}
		exactTimes, _ := strconv.Atoi(getCol(row, col, "exact_times"))

		frequencies = append(frequencies, Frequency{
			trip_id:      tripIndex[getCol(row, col, "trip_id")],
			start_time:   getCol(row, col, "start_time"),
			end_time:     getCol(row, col, "end_time"),
			headway_secs: headwaySecs,
			exact_times:  ExactTimes(exactTimes),
		})
	}

	gtfs.FrequencyData = frequencies
	return nil
}

func (gtfs *GTFS) ParseTransfer() error {
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return err
	}
	defer r.Close()

	var transfersFile *zip.File
	for _, f := range r.File {
		if f.Name == "transfers.txt" {
			transfersFile = f
			break
		}
	}
	if transfersFile == nil {
		return nil // transfers.txt is optional
	}

	rc, err := transfersFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	reader := csv.NewReader(rc)

	headers, err := reader.Read()
	if err != nil {
		return err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	// Build indexes for FK resolution
	stopIndex := make(map[string]*Stop)
	for i := range gtfs.StopData {
		stopIndex[gtfs.StopData[i].stop_id] = &gtfs.StopData[i]
	}
	routeIndex := make(map[string]*Route)
	for i := range gtfs.RouteData {
		routeIndex[gtfs.RouteData[i].route_id] = &gtfs.RouteData[i]
	}
	tripIndex := make(map[string]*Trip)
	for i := range gtfs.TripData {
		tripIndex[gtfs.TripData[i].trip_id] = &gtfs.TripData[i]
	}

	var transfers []Transfer

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		transferType, _ := strconv.Atoi(getCol(row, col, "transfer_type"))
		minTransferTime, _ := strconv.Atoi(getCol(row, col, "min_transfer_time"))

		transfers = append(transfers, Transfer{
			from_stop_id:      stopIndex[getCol(row, col, "from_stop_id")],
			to_stop_id:        stopIndex[getCol(row, col, "to_stop_id")],
			from_route_id:     routeIndex[getCol(row, col, "from_route_id")],
			to_route_id:       routeIndex[getCol(row, col, "to_route_id")],
			from_trip_id:      tripIndex[getCol(row, col, "from_trip_id")],
			to_trip_id:        tripIndex[getCol(row, col, "to_trip_id")],
			transfer_type:     TransferType(transferType),
			min_transfer_time: minTransferTime,
		})
	}

	gtfs.TransferData = transfers
	return nil
}
