package gtfsparser

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var timeRegex = regexp.MustCompile(`^\d+:[0-5]\d:[0-5]\d$`)

// ValidateAll validates all parsed GTFS data against the GTFS Schedule specification.
// All errors are collected and returned; validation does not stop at the first error.
// ParseAll should be called before ValidateAll.
func (gtfs *GTFS) ValidateAll() []ValidationError {
	var errs []ValidationError

	errs = append(errs, gtfs.validateRequiredFiles()...)
	errs = append(errs, gtfs.validateAgency()...)
	errs = append(errs, gtfs.validateStops()...)
	errs = append(errs, gtfs.validateRoutes()...)
	errs = append(errs, gtfs.validateTrips()...)
	errs = append(errs, gtfs.validateStopTimes()...)
	errs = append(errs, gtfs.validateCalendar()...)

	return errs
}

// validateRequiredFiles checks that required GTFS files were parsed and are non-empty
func (gtfs *GTFS) validateRequiredFiles() []ValidationError {
	var errs []ValidationError

	required := []struct {
		name  string
		count int
	}{
		{"agency.txt", len(gtfs.AgencyData)},
		{"stops.txt", len(gtfs.StopData)},
		{"routes.txt", len(gtfs.RouteData)},
		{"trips.txt", len(gtfs.TripData)},
		{"stop_times.txt", len(gtfs.StopTimeData)},
	}

	for _, f := range required {
		if f.count == 0 {
			errs = append(errs, ValidationError{
				File:    f.name,
				Message: "required file is missing or empty",
			})
		}
	}

	if len(gtfs.CalendarData) == 0 && len(gtfs.CalendarDates) == 0 {
		errs = append(errs, ValidationError{
			File:    "calendar.txt / calendar_dates.txt",
			Message: "at least one of calendar.txt or calendar_dates.txt must be present and non-empty",
		})
	}

	return errs
}

func (gtfs *GTFS) validateAgency() []ValidationError {
	var errs []ValidationError

	for _, a := range gtfs.AgencyData {
		id := a.AgencyID
		if a.AgencyName == "" {
			errs = append(errs, ValidationError{File: "agency.txt", ID: id, Field: "agency_name", Message: "required field is empty"})
		}
		if a.AgencyURL == "" {
			errs = append(errs, ValidationError{File: "agency.txt", ID: id, Field: "agency_url", Message: "required field is empty"})
		}
		if a.AgencyTimezone == "" {
			errs = append(errs, ValidationError{File: "agency.txt", ID: id, Field: "agency_timezone", Message: "required field is empty"})
		}
	}

	return errs
}

func (gtfs *GTFS) validateStops() []ValidationError {
	var errs []ValidationError

	for _, s := range gtfs.StopData {
		id := s.StopID
		if id == "" {
			errs = append(errs, ValidationError{File: "stops.txt", Field: "stop_id", Message: "required field is empty"})
		}

		// stop_name required for location_type 0 and 1
		if s.LocationType == StopPlatform || s.LocationType == Station {
			if s.StopName == "" {
				errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_name", Message: "required for location_type 0 and 1"})
			}
		}

		// if one coordinate is provided, the other must be too
		if s.StopLat != nil && s.StopLon == nil {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lon", Message: "stop_lon is required when stop_lat is provided"})
		}
		if s.StopLon != nil && s.StopLat == nil {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lat", Message: "stop_lat is required when stop_lon is provided"})
		}

		// coordinate range
		if s.StopLat != nil && (*s.StopLat < -90 || *s.StopLat > 90) {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lat", Message: fmt.Sprintf("value %f out of range [-90, 90]", *s.StopLat)})
		}
		if s.StopLon != nil && (*s.StopLon < -180 || *s.StopLon > 180) {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lon", Message: fmt.Sprintf("value %f out of range [-180, 180]", *s.StopLon)})
		}

		// enum range
		if s.LocationType < StopPlatform || s.LocationType > BoardingArea {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "location_type", Message: fmt.Sprintf("invalid value %d, must be 0-4", s.LocationType)})
		}
		if s.WheelchairBoarding < NoAccessibilityInfo || s.WheelchairBoarding > NotAccessible {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "wheelchair_boarding", Message: fmt.Sprintf("invalid value %d, must be 0-2", s.WheelchairBoarding)})
		}
	}

	return errs
}

func (gtfs *GTFS) validateRoutes() []ValidationError {
	var errs []ValidationError

	validRouteTypes := map[RouteType]bool{
		Tram: true, Subway: true, Rail: true, Bus: true,
		Ferry: true, CableTram: true, AerialLift: true, Funicular: true,
		11: true, 12: true, // trolleybus, monorail
	}

	for _, r := range gtfs.RouteData {
		id := r.RouteID
		if id == "" {
			errs = append(errs, ValidationError{File: "routes.txt", Field: "route_id", Message: "required field is empty"})
		}
		if r.RouteShortName == "" && r.RouteLongName == "" {
			errs = append(errs, ValidationError{File: "routes.txt", ID: id, Field: "route_short_name/route_long_name", Message: "at least one of route_short_name or route_long_name is required"})
		}
		if !validRouteTypes[r.RouteType] {
			errs = append(errs, ValidationError{File: "routes.txt", ID: id, Field: "route_type", Message: fmt.Sprintf("invalid value %d", r.RouteType)})
		}
		if r.AgencyID == nil && len(gtfs.AgencyData) > 1 {
			errs = append(errs, ValidationError{File: "routes.txt", ID: id, Field: "agency_id", Message: "required when feed has multiple agencies"})
		}
	}

	return errs
}

func (gtfs *GTFS) validateTrips() []ValidationError {
	var errs []ValidationError

	for _, t := range gtfs.TripData {
		id := t.TripID
		if id == "" {
			errs = append(errs, ValidationError{File: "trips.txt", Field: "trip_id", Message: "required field is empty"})
		}
		if t.RouteID == nil {
			errs = append(errs, ValidationError{File: "trips.txt", ID: id, Field: "route_id", Message: "references a route_id that does not exist"})
		}
		if t.ServiceID == nil {
			errs = append(errs, ValidationError{File: "trips.txt", ID: id, Field: "service_id", Message: "references a service_id that does not exist in calendar.txt or calendar_dates.txt"})
		}
		if t.DirectionID < OutboundTravel || t.DirectionID > InboundTravel {
			errs = append(errs, ValidationError{File: "trips.txt", ID: id, Field: "direction_id", Message: fmt.Sprintf("invalid value %d, must be 0 or 1", t.DirectionID)})
		}
	}

	return errs
}

func (gtfs *GTFS) validateStopTimes() []ValidationError {
	var errs []ValidationError

	// Group stop_times by trip to check sequence ordering
	type seqEntry struct {
		seq int
		idx int
	}
	tripSeqs := make(map[string][]seqEntry)

	for i, st := range gtfs.StopTimeData {
		tripID := ""
		if st.TripID != nil {
			tripID = st.TripID.TripID
		}

		if st.TripID == nil {
			errs = append(errs, ValidationError{File: "stop_times.txt", Field: "trip_id", Message: "references a trip_id that does not exist"})
		}
		if st.StopID == nil {
			errs = append(errs, ValidationError{File: "stop_times.txt", ID: tripID, Field: "stop_id", Message: "references a stop_id that does not exist"})
		}

		// time format
		if st.ArrivalTime != "" && !timeRegex.MatchString(st.ArrivalTime) {
			errs = append(errs, ValidationError{File: "stop_times.txt", ID: tripID, Field: "arrival_time", Message: fmt.Sprintf("invalid format %q, expected HH:MM:SS", st.ArrivalTime)})
		}
		if st.DepartureTime != "" && !timeRegex.MatchString(st.DepartureTime) {
			errs = append(errs, ValidationError{File: "stop_times.txt", ID: tripID, Field: "departure_time", Message: fmt.Sprintf("invalid format %q, expected HH:MM:SS", st.DepartureTime)})
		}

		if tripID != "" {
			tripSeqs[tripID] = append(tripSeqs[tripID], seqEntry{seq: st.StopSequence, idx: i})
		}
	}

	// Check stop_sequence is increasing within each trip
	for tripID, seqs := range tripSeqs {
		sort.Slice(seqs, func(i, j int) bool {
			return seqs[i].seq < seqs[j].seq
		})
		for i := 1; i < len(seqs); i++ {
			if seqs[i].seq <= seqs[i-1].seq {
				errs = append(errs, ValidationError{
					File:    "stop_times.txt",
					ID:      tripID,
					Field:   "stop_sequence",
					Message: fmt.Sprintf("sequence %d is not greater than previous %d", seqs[i].seq, seqs[i-1].seq),
				})
			}
		}
	}

	return errs
}

func (gtfs *GTFS) validateCalendar() []ValidationError {
	var errs []ValidationError

	for _, c := range gtfs.CalendarData {
		id := c.ServiceID
		if id == "" {
			errs = append(errs, ValidationError{File: "calendar.txt", Field: "service_id", Message: "required field is empty"})
		}
		if c.StartDate == "" {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "start_date", Message: "required field is empty"})
		}
		if c.EndDate == "" {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "end_date", Message: "required field is empty"})
		}
		// days must be 0 or 1
		days := map[string]int{
			"monday": c.Monday, "tuesday": c.Tuesday, "wednesday": c.Wednesday,
			"thursday": c.Thursday, "friday": c.Friday, "saturday": c.Saturday, "sunday": c.Sunday,
		}
		for field, val := range days {
			if val != 0 && val != 1 {
				errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: field, Message: fmt.Sprintf("invalid value %d, must be 0 or 1", val)})
			}
		}
		// validate date format YYYYMMDD
		if c.StartDate != "" && !isValidDate(c.StartDate) {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "start_date", Message: fmt.Sprintf("invalid date format %q, expected YYYYMMDD", c.StartDate)})
		}
		if c.EndDate != "" && !isValidDate(c.EndDate) {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "end_date", Message: fmt.Sprintf("invalid date format %q, expected YYYYMMDD", c.EndDate)})
		}
	}

	return errs
}

func isValidDate(s string) bool {
	if len(s) != 8 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	month, _ := strconv.Atoi(s[4:6])
	day, _ := strconv.Atoi(s[6:8])
	return month >= 1 && month <= 12 && day >= 1 && day <= 31
}

// formatValidationErrors returns a human-readable summary of all validation errors
func formatValidationErrors(errs []ValidationError) string {
	if len(errs) == 0 {
		return "no validation errors"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d validation error(s):\n", len(errs)))
	for _, e := range errs {
		sb.WriteString("  " + e.Error() + "\n")
	}
	return sb.String()
}
