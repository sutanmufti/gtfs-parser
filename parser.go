package gtfsparser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// ParseAll parses all GTFS files in dependency order.
// Required files return an error if absent; optional files are silently skipped.
// Call ValidateAll after ParseAll to check the parsed data for specification errors.
func (gtfs *GTFS) ParseAll() error {
	parsers := []struct {
		name string
		fn   func() error
	}{
		{"agency", gtfs.ParseAgency},
		{"level", gtfs.ParseLevel},
		{"calendar", gtfs.ParseCalendar},
		{"shape", gtfs.ParseShape},
		{"stop", gtfs.ParseStop},
		{"route", gtfs.ParseRoute},
		{"calendar_dates", gtfs.ParseCalendarDate},
		{"trip", gtfs.ParseTrip},
		{"stop_time", gtfs.ParseStopTime},
		{"frequency", gtfs.ParseFrequency},
		{"transfer", gtfs.ParseTransfer},
		{"pathway", gtfs.ParsePathway},
		{"fare_attribute", gtfs.ParseFareAttribute},
		{"fare_rule", gtfs.ParseFareRule},
		{"feed_info", gtfs.ParseFeedInfo},
		{"attribution", gtfs.ParseAttribution},
		{"translation", gtfs.ParseTranslation},
	}

	for _, p := range parsers {
		if err := p.fn(); err != nil {
			return fmt.Errorf("ParseAll: error parsing %s: %w", p.name, err)
		}
	}

	return nil
}

// readCSVHeaders reads the header row and returns a column index map.
func readCSVHeaders(reader *csv.Reader) (map[string]int, error) {
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	headers = sanitizeHeaders(headers)
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}
	return col, nil
}

// ParseAgency reads agency.txt from the feed zip and populates AgencyData.
// Returns an error if the file is absent or contains duplicate agency_id values.
func (gtfs *GTFS) ParseAgency() error {
	rc, err := openFileFromZip(gtfs.FileName, "agency.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return fmt.Errorf("agency.txt not found in %s", gtfs.FileName)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

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
			AgencyID:       getCol(row, col, "agency_id"),
			AgencyName:     getCol(row, col, "agency_name"),
			AgencyURL:      getCol(row, col, "agency_url"),
			AgencyTimezone: getCol(row, col, "agency_timezone"),
			AgencyLang:     getCol(row, col, "agency_lang"),
			AgencyPhone:    getCol(row, col, "agency_phone"),
			AgencyFareURL:  getCol(row, col, "agency_fare_url"),
			AgencyEmail:    getCol(row, col, "agency_email"),
		}

		if seen[agency.AgencyID] {
			duplicates = append(duplicates, agency.AgencyID)
		} else {
			seen[agency.AgencyID] = true
		}

		agencies = append(agencies, agency)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate agency_id(s) found: %v", duplicates)
	}

	gtfs.AgencyData = agencies
	return nil
}

// ParseRoute reads routes.txt from the feed zip and populates RouteData.
// The agency_id foreign key is resolved against AgencyData.
// ParseAgency must be called before ParseRoute.
func (gtfs *GTFS) ParseRoute() error {
	rc, err := openFileFromZip(gtfs.FileName, "routes.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return fmt.Errorf("routes.txt not found in %s", gtfs.FileName)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

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

		var agencyPtr *Agency
		agencyIDStr := getCol(row, col, "agency_id")
		for i := range gtfs.AgencyData {
			if gtfs.AgencyData[i].AgencyID == agencyIDStr {
				agencyPtr = &gtfs.AgencyData[i]
				break
			}
		}

		route := Route{
			RouteID:           getCol(row, col, "route_id"),
			AgencyID:          agencyPtr,
			RouteShortName:    getCol(row, col, "route_short_name"),
			RouteLongName:     getCol(row, col, "route_long_name"),
			RouteDesc:         getCol(row, col, "route_desc"),
			RouteType:         RouteType(routeType),
			RouteURL:          getCol(row, col, "route_url"),
			RouteColor:        getCol(row, col, "route_color"),
			RouteTextColor:    getCol(row, col, "route_text_color"),
			RouteSortOrder:    sortOrder,
			ContinuousPickup:  PickupDropOffType(contPickup),
			ContinuousDropOff: PickupDropOffType(contDropOff),
			NetworkID:         getCol(row, col, "network_id"),
		}

		if seen[route.RouteID] {
			duplicates = append(duplicates, route.RouteID)
		} else {
			seen[route.RouteID] = true
		}

		routes = append(routes, route)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate route_id(s) found: %v", duplicates)
	}

	gtfs.RouteData = routes
	return nil
}

// ParseStop reads stops.txt from the feed zip and populates StopData.
// ParentStation and LevelID foreign keys are resolved in a second pass after
// all stops are loaded into memory.
// Returns an error if the file is absent or contains duplicate stop_id values.
func (gtfs *GTFS) ParseStop() error {
	rc, err := openFileFromZip(gtfs.FileName, "stops.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return fmt.Errorf("stops.txt not found in %s", gtfs.FileName)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	type stopRaw struct {
		stop            Stop
		parentStationID string
		levelID         string
	}
	var rawStops []stopRaw
	seen := make(map[string]int)
	type duplicate struct {
		stopID string
		line   int
	}
	var duplicates []duplicate
	lineNum := 1

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
			locType = 0
		}
		wheelchair, err := strconv.Atoi(getCol(row, col, "wheelchair_boarding"))
		if err != nil {
			wheelchair = 0
		}

		stop := Stop{
			StopID:             getCol(row, col, "stop_id"),
			StopCode:           getCol(row, col, "stop_code"),
			StopName:           getCol(row, col, "stop_name"),
			StopDesc:           getCol(row, col, "stop_desc"),
			StopLat:            lat,
			StopLon:            lon,
			ZoneID:             getCol(row, col, "zone_id"),
			StopURL:            getCol(row, col, "stop_url"),
			LocationType:       LocationType(locType),
			ParentStation:      nil,
			StopTimezone:       getCol(row, col, "stop_timezone"),
			WheelchairBoarding: WheelchairBoarding(wheelchair),
			LevelID:            nil,
			PlatformCode:       getCol(row, col, "platform_code"),
		}

		if _, exists := seen[stop.StopID]; exists {
			duplicates = append(duplicates, duplicate{stopID: stop.StopID, line: lineNum})
		} else {
			seen[stop.StopID] = lineNum
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
			msg += fmt.Sprintf("\n  stop_id %q at line %d (first seen at line %d)", d.stopID, d.line, seen[d.stopID])
		}
		return fmt.Errorf("%s", msg)
	}

	stops := make([]Stop, len(rawStops))
	for i, rs := range rawStops {
		stops[i] = rs.stop
	}

	stopIndex := make(map[string]*Stop)
	for i := range stops {
		stopIndex[stops[i].StopID] = &stops[i]
	}
	levelIndex := make(map[string]*Level)
	for i := range gtfs.LevelData {
		levelIndex[gtfs.LevelData[i].LevelID] = &gtfs.LevelData[i]
	}
	for i, rs := range rawStops {
		if rs.parentStationID != "" {
			stops[i].ParentStation = stopIndex[rs.parentStationID]
		}
		if rs.levelID != "" {
			stops[i].LevelID = levelIndex[rs.levelID]
		}
	}

	gtfs.StopData = stops
	return nil
}

// ParseTrip reads trips.txt from the feed zip and populates TripData.
// Foreign keys for route_id, service_id, and shape_id are resolved against
// RouteData, CalendarData, and ShapeData respectively.
// ParseRoute, ParseCalendar, and ParseShape must be called before ParseTrip.
func (gtfs *GTFS) ParseTrip() error {
	rc, err := openFileFromZip(gtfs.FileName, "trips.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return fmt.Errorf("trips.txt not found in %s", gtfs.FileName)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	routeIndex := make(map[string]*Route)
	for i := range gtfs.RouteData {
		routeIndex[gtfs.RouteData[i].RouteID] = &gtfs.RouteData[i]
	}
	calendarIndex := make(map[string]*Calendar)
	for i := range gtfs.CalendarData {
		calendarIndex[gtfs.CalendarData[i].ServiceID] = &gtfs.CalendarData[i]
	}
	shapeIndex := make(map[string]*Shape)
	for i := range gtfs.ShapeData {
		shapeIndex[gtfs.ShapeData[i].ShapeID] = &gtfs.ShapeData[i]
	}

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
			RouteID:              routeIndex[getCol(row, col, "route_id")],
			ServiceID:            calendarIndex[getCol(row, col, "service_id")],
			TripID:               getCol(row, col, "trip_id"),
			TripHeadsign:         getCol(row, col, "trip_headsign"),
			TripShortName:        getCol(row, col, "trip_short_name"),
			DirectionID:          DirectionId(directionId),
			BlockID:              getCol(row, col, "block_id"),
			ShapeID:              shapeIndex[getCol(row, col, "shape_id")],
			WheelchairAccessible: WheelchairAccessibleEnum(wheelchairAccessible),
			BikesAllowed:         BikesAllowed(bikesAllowed),
		}

		if seen[trip.TripID] {
			duplicates = append(duplicates, trip.TripID)
		} else {
			seen[trip.TripID] = true
		}

		trips = append(trips, trip)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate trip_id(s) found: %v", duplicates)
	}

	gtfs.TripData = trips
	return nil
}

// ParseCalendar reads calendar.txt from the feed zip and populates CalendarData.
// Returns an error if the file is absent or contains duplicate service_id values.
func (gtfs *GTFS) ParseCalendar() error {
	rc, err := openFileFromZip(gtfs.FileName, "calendar.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return fmt.Errorf("calendar.txt not found in %s", gtfs.FileName)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
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
			ServiceID: getCol(row, col, "service_id"),
			Monday:    monday,
			Tuesday:   tuesday,
			Wednesday: wednesday,
			Thursday:  thursday,
			Friday:    friday,
			Saturday:  saturday,
			Sunday:    sunday,
			StartDate: getCol(row, col, "start_date"),
			EndDate:   getCol(row, col, "end_date"),
		}

		if seen[calendar.ServiceID] {
			duplicates = append(duplicates, calendar.ServiceID)
		} else {
			seen[calendar.ServiceID] = true
		}

		calendars = append(calendars, calendar)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate service_id(s) found: %v", duplicates)
	}

	gtfs.CalendarData = calendars
	return nil
}

// ParseShape reads shapes.txt from the feed zip and populates ShapeData.
// The file is optional; no error is returned if it is absent.
func (gtfs *GTFS) ParseShape() error {
	rc, err := openFileFromZip(gtfs.FileName, "shapes.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	var shapes []Shape
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
			ShapeID:           shapeID,
			ShapePtLat:        lat,
			ShapePtLon:        lon,
			ShapePtSequence:   seq,
			ShapeDistTraveled: distTraveled,
		})
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate (shape_id, shape_pt_sequence) found: %v", duplicates)
	}

	gtfs.ShapeData = shapes
	return nil
}

// ParseStopTime reads stop_times.txt from the feed zip and populates StopTimeData.
// Foreign keys for trip_id and stop_id are resolved against TripData and StopData.
// ParseTrip and ParseStop must be called before ParseStopTime.
func (gtfs *GTFS) ParseStopTime() error {
	rc, err := openFileFromZip(gtfs.FileName, "stop_times.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return fmt.Errorf("stop_times.txt not found in %s", gtfs.FileName)
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	tripIndex := make(map[string]*Trip)
	for i := range gtfs.TripData {
		tripIndex[gtfs.TripData[i].TripID] = &gtfs.TripData[i]
	}
	stopIndex := make(map[string]*Stop)
	for i := range gtfs.StopData {
		stopIndex[gtfs.StopData[i].StopID] = &gtfs.StopData[i]
	}

	var stopTimes []StopTime
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
			TripID:            tripIndex[tripIDStr],
			ArrivalTime:       getCol(row, col, "arrival_time"),
			DepartureTime:     getCol(row, col, "departure_time"),
			StopID:            stopIndex[getCol(row, col, "stop_id")],
			StopSequence:      seq,
			StopHeadsign:      getCol(row, col, "stop_headsign"),
			PickupType:        PickupDropOffType(pickupType),
			DropOffType:       PickupDropOffType(dropOffType),
			ContinuousPickup:  PickupDropOffType(contPickup),
			ContinuousDropOff: PickupDropOffType(contDropOff),
			ShapeDistTraveled: distTraveled,
			Timepoint:         Timepoint(timepoint),
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

// ParseFrequency reads frequencies.txt from the feed zip and populates FrequencyData.
// The file is optional; no error is returned if it is absent.
// ParseTrip must be called before ParseFrequency.
func (gtfs *GTFS) ParseFrequency() error {
	rc, err := openFileFromZip(gtfs.FileName, "frequencies.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	tripIndex := make(map[string]*Trip)
	for i := range gtfs.TripData {
		tripIndex[gtfs.TripData[i].TripID] = &gtfs.TripData[i]
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
			TripID:      tripIndex[getCol(row, col, "trip_id")],
			StartTime:   getCol(row, col, "start_time"),
			EndTime:     getCol(row, col, "end_time"),
			HeadwaySecs: headwaySecs,
			ExactTimes:  ExactTimes(exactTimes),
		})
	}

	gtfs.FrequencyData = frequencies
	return nil
}

// ParseTransfer reads transfers.txt from the feed zip and populates TransferData.
// The file is optional; no error is returned if it is absent.
// ParseStop, ParseRoute, and ParseTrip must be called before ParseTransfer.
func (gtfs *GTFS) ParseTransfer() error {
	rc, err := openFileFromZip(gtfs.FileName, "transfers.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	stopIndex := make(map[string]*Stop)
	for i := range gtfs.StopData {
		stopIndex[gtfs.StopData[i].StopID] = &gtfs.StopData[i]
	}
	routeIndex := make(map[string]*Route)
	for i := range gtfs.RouteData {
		routeIndex[gtfs.RouteData[i].RouteID] = &gtfs.RouteData[i]
	}
	tripIndex := make(map[string]*Trip)
	for i := range gtfs.TripData {
		tripIndex[gtfs.TripData[i].TripID] = &gtfs.TripData[i]
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
			FromStopID:      stopIndex[getCol(row, col, "from_stop_id")],
			ToStopID:        stopIndex[getCol(row, col, "to_stop_id")],
			FromRouteID:     routeIndex[getCol(row, col, "from_route_id")],
			ToRouteID:       routeIndex[getCol(row, col, "to_route_id")],
			FromTripID:      tripIndex[getCol(row, col, "from_trip_id")],
			ToTripID:        tripIndex[getCol(row, col, "to_trip_id")],
			TransferType:    TransferType(transferType),
			MinTransferTime: minTransferTime,
		})
	}

	gtfs.TransferData = transfers
	return nil
}

// ParseCalendarDate reads calendar_dates.txt from the feed zip and populates CalendarDates.
// The file is optional; no error is returned if it is absent.
// ParseCalendar must be called before ParseCalendarDate.
func (gtfs *GTFS) ParseCalendarDate() error {
	rc, err := openFileFromZip(gtfs.FileName, "calendar_dates.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional if calendar.txt exists
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	calendarIndex := make(map[string]*Calendar)
	for i := range gtfs.CalendarData {
		calendarIndex[gtfs.CalendarData[i].ServiceID] = &gtfs.CalendarData[i]
	}

	var calendarDates []CalendarDate
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		exceptionType, _ := strconv.Atoi(getCol(row, col, "exception_type"))

		calendarDates = append(calendarDates, CalendarDate{
			ServiceID:     calendarIndex[getCol(row, col, "service_id")],
			Date:          getCol(row, col, "date"),
			ExceptionType: ExceptionType(exceptionType),
		})
	}

	gtfs.CalendarDates = calendarDates
	return nil
}

// ParseLevel reads levels.txt from the feed zip and populates LevelData.
// The file is optional; no error is returned if it is absent.
func (gtfs *GTFS) ParseLevel() error {
	rc, err := openFileFromZip(gtfs.FileName, "levels.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	var levels []Level
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

		levelIdx, err := strconv.ParseFloat(getCol(row, col, "level_index"), 64)
		if err != nil {
			return fmt.Errorf("invalid level_index for level_id %s: %w", getCol(row, col, "level_id"), err)
		}

		level := Level{
			LevelID:    getCol(row, col, "level_id"),
			LevelIndex: levelIdx,
			LevelName:  getCol(row, col, "level_name"),
		}

		if seen[level.LevelID] {
			duplicates = append(duplicates, level.LevelID)
		} else {
			seen[level.LevelID] = true
		}

		levels = append(levels, level)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate level_id(s) found: %v", duplicates)
	}

	gtfs.LevelData = levels
	return nil
}

// ParsePathway reads pathways.txt from the feed zip and populates PathwayData.
// The file is optional; no error is returned if it is absent.
// ParseStop must be called before ParsePathway.
func (gtfs *GTFS) ParsePathway() error {
	rc, err := openFileFromZip(gtfs.FileName, "pathways.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	stopIndex := make(map[string]*Stop)
	for i := range gtfs.StopData {
		stopIndex[gtfs.StopData[i].StopID] = &gtfs.StopData[i]
	}

	var pathways []Pathway
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

		pathwayMode, _ := strconv.Atoi(getCol(row, col, "pathway_mode"))
		isBidirectional, _ := strconv.Atoi(getCol(row, col, "is_bidirectional"))
		length, _ := strconv.ParseFloat(getCol(row, col, "length"), 64)
		traversalTime, _ := strconv.Atoi(getCol(row, col, "traversal_time"))
		stairCount, _ := strconv.Atoi(getCol(row, col, "stair_count"))
		maxSlope, _ := strconv.ParseFloat(getCol(row, col, "max_slope"), 64)
		minWidth, _ := strconv.ParseFloat(getCol(row, col, "min_width"), 64)

		pathway := Pathway{
			PathwayID:            getCol(row, col, "pathway_id"),
			FromStopID:           stopIndex[getCol(row, col, "from_stop_id")],
			ToStopID:             stopIndex[getCol(row, col, "to_stop_id")],
			PathwayMode:          PathwayMode(pathwayMode),
			IsBidirectional:      isBidirectional,
			Length:               length,
			TraversalTime:        traversalTime,
			StairCount:           stairCount,
			MaxSlope:             maxSlope,
			MinWidth:             minWidth,
			SignpostedAs:         getCol(row, col, "signposted_as"),
			ReversedSignpostedAs: getCol(row, col, "reversed_signposted_as"),
		}

		if seen[pathway.PathwayID] {
			duplicates = append(duplicates, pathway.PathwayID)
		} else {
			seen[pathway.PathwayID] = true
		}

		pathways = append(pathways, pathway)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate pathway_id(s) found: %v", duplicates)
	}

	gtfs.PathwayData = pathways
	return nil
}

// ParseFareAttribute reads fare_attributes.txt from the feed zip and populates FareAttributes.
// The file is optional; no error is returned if it is absent.
// ParseAgency must be called before ParseFareAttribute.
func (gtfs *GTFS) ParseFareAttribute() error {
	rc, err := openFileFromZip(gtfs.FileName, "fare_attributes.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	agencyIndex := make(map[string]*Agency)
	for i := range gtfs.AgencyData {
		agencyIndex[gtfs.AgencyData[i].AgencyID] = &gtfs.AgencyData[i]
	}

	var fareAttributes []FareAttribute
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

		price, err := strconv.ParseFloat(getCol(row, col, "price"), 64)
		if err != nil {
			return fmt.Errorf("invalid price for fare_id %s: %w", getCol(row, col, "fare_id"), err)
		}
		paymentMethod, _ := strconv.Atoi(getCol(row, col, "payment_method"))
		transfers, _ := strconv.Atoi(getCol(row, col, "transfers"))
		transferDuration, _ := strconv.Atoi(getCol(row, col, "transfer_duration"))

		fa := FareAttribute{
			FareID:           getCol(row, col, "fare_id"),
			Price:            price,
			CurrencyType:     getCol(row, col, "currency_type"),
			PaymentMethod:    PaymentMethod(paymentMethod),
			Transfers:        FareTransfers(transfers),
			AgencyID:         agencyIndex[getCol(row, col, "agency_id")],
			TransferDuration: transferDuration,
		}

		if seen[fa.FareID] {
			duplicates = append(duplicates, fa.FareID)
		} else {
			seen[fa.FareID] = true
		}

		fareAttributes = append(fareAttributes, fa)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate fare_id(s) found: %v", duplicates)
	}

	gtfs.FareAttributes = fareAttributes
	return nil
}

// ParseFareRule reads fare_rules.txt from the feed zip and populates FareRules.
// The file is optional; no error is returned if it is absent.
// ParseFareAttribute and ParseRoute must be called before ParseFareRule.
func (gtfs *GTFS) ParseFareRule() error {
	rc, err := openFileFromZip(gtfs.FileName, "fare_rules.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	fareIndex := make(map[string]*FareAttribute)
	for i := range gtfs.FareAttributes {
		fareIndex[gtfs.FareAttributes[i].FareID] = &gtfs.FareAttributes[i]
	}
	routeIndex := make(map[string]*Route)
	for i := range gtfs.RouteData {
		routeIndex[gtfs.RouteData[i].RouteID] = &gtfs.RouteData[i]
	}

	var fareRules []FareRule
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fareRules = append(fareRules, FareRule{
			FareID:        fareIndex[getCol(row, col, "fare_id")],
			RouteID:       routeIndex[getCol(row, col, "route_id")],
			OriginID:      getCol(row, col, "origin_id"),
			DestinationID: getCol(row, col, "destination_id"),
			ContainsID:    getCol(row, col, "contains_id"),
		})
	}

	gtfs.FareRules = fareRules
	return nil
}

// ParseFeedInfo reads feed_info.txt from the feed zip and populates FeedInfo.
// The file is optional; no error is returned if it is absent.
func (gtfs *GTFS) ParseFeedInfo() error {
	rc, err := openFileFromZip(gtfs.FileName, "feed_info.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	var feedInfos []FeedInfo
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		feedInfos = append(feedInfos, FeedInfo{
			FeedPublisherName: getCol(row, col, "feed_publisher_name"),
			FeedPublisherURL:  getCol(row, col, "feed_publisher_url"),
			FeedLang:          getCol(row, col, "feed_lang"),
			DefaultLang:       getCol(row, col, "default_lang"),
			FeedStartDate:     getCol(row, col, "feed_start_date"),
			FeedEndDate:       getCol(row, col, "feed_end_date"),
			FeedVersion:       getCol(row, col, "feed_version"),
			FeedContactEmail:  getCol(row, col, "feed_contact_email"),
			FeedContactURL:    getCol(row, col, "feed_contact_url"),
		})
	}

	gtfs.FeedInfo = feedInfos
	return nil
}

// ParseAttribution reads attributions.txt from the feed zip and populates Attributions.
// The file is optional; no error is returned if it is absent.
// ParseAgency, ParseRoute, and ParseTrip must be called before ParseAttribution.
func (gtfs *GTFS) ParseAttribution() error {
	rc, err := openFileFromZip(gtfs.FileName, "attributions.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	agencyIndex := make(map[string]*Agency)
	for i := range gtfs.AgencyData {
		agencyIndex[gtfs.AgencyData[i].AgencyID] = &gtfs.AgencyData[i]
	}
	routeIndex := make(map[string]*Route)
	for i := range gtfs.RouteData {
		routeIndex[gtfs.RouteData[i].RouteID] = &gtfs.RouteData[i]
	}
	tripIndex := make(map[string]*Trip)
	for i := range gtfs.TripData {
		tripIndex[gtfs.TripData[i].TripID] = &gtfs.TripData[i]
	}

	var attributions []Attribution
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		isProducer, _ := strconv.Atoi(getCol(row, col, "is_producer"))
		isOperator, _ := strconv.Atoi(getCol(row, col, "is_operator"))
		isAuthority, _ := strconv.Atoi(getCol(row, col, "is_authority"))

		attributions = append(attributions, Attribution{
			AttributionID:    getCol(row, col, "attribution_id"),
			AgencyID:         agencyIndex[getCol(row, col, "agency_id")],
			RouteID:          routeIndex[getCol(row, col, "route_id")],
			TripID:           tripIndex[getCol(row, col, "trip_id")],
			OrganizationName: getCol(row, col, "organization_name"),
			IsProducer:       isProducer,
			IsOperator:       isOperator,
			IsAuthority:      isAuthority,
			AttributionURL:   getCol(row, col, "attribution_url"),
			AttributionEmail: getCol(row, col, "attribution_email"),
			AttributionPhone: getCol(row, col, "attribution_phone"),
		})
	}

	gtfs.Attributions = attributions
	return nil
}

// ParseTranslation reads translations.txt from the feed zip and populates Translations.
// The file is optional; no error is returned if it is absent.
func (gtfs *GTFS) ParseTranslation() error {
	rc, err := openFileFromZip(gtfs.FileName, "translations.txt")
	if err != nil {
		return err
	}
	if rc == nil {
		return nil // optional
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	col, err := readCSVHeaders(reader)
	if err != nil {
		return err
	}

	var translations []Translation
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		translations = append(translations, Translation{
			TableName:   getCol(row, col, "table_name"),
			FieldName:   getCol(row, col, "field_name"),
			Language:    getCol(row, col, "language"),
			Translation: getCol(row, col, "translation"),
			RecordID:    getCol(row, col, "record_id"),
			RecordSubID: getCol(row, col, "record_sub_id"),
			FieldValue:  getCol(row, col, "field_value"),
		})
	}

	gtfs.Translations = translations
	return nil
}
