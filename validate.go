package gtfsparser

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var timeRegex = regexp.MustCompile(`^\d+:[0-5]\d:[0-5]\d$`)

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
		id := a.agency_id
		if a.agency_name == "" {
			errs = append(errs, ValidationError{File: "agency.txt", ID: id, Field: "agency_name", Message: "required field is empty"})
		}
		if a.agency_url == "" {
			errs = append(errs, ValidationError{File: "agency.txt", ID: id, Field: "agency_url", Message: "required field is empty"})
		}
		if a.agency_timezone == "" {
			errs = append(errs, ValidationError{File: "agency.txt", ID: id, Field: "agency_timezone", Message: "required field is empty"})
		}
	}

	return errs
}

func (gtfs *GTFS) validateStops() []ValidationError {
	var errs []ValidationError

	for _, s := range gtfs.StopData {
		id := s.stop_id
		if id == "" {
			errs = append(errs, ValidationError{File: "stops.txt", Field: "stop_id", Message: "required field is empty"})
		}

		// stop_name required for location_type 0 and 1
		if s.location_type == StopPlatform || s.location_type == Station {
			if s.stop_name == "" {
				errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_name", Message: "required for location_type 0 and 1"})
			}
		}

		// if one coordinate is provided, the other must be too
		if s.stop_lat != nil && s.stop_lon == nil {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lon", Message: "stop_lon is required when stop_lat is provided"})
		}
		if s.stop_lon != nil && s.stop_lat == nil {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lat", Message: "stop_lat is required when stop_lon is provided"})
		}

		// coordinate range
		if s.stop_lat != nil && (*s.stop_lat < -90 || *s.stop_lat > 90) {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lat", Message: fmt.Sprintf("value %f out of range [-90, 90]", *s.stop_lat)})
		}
		if s.stop_lon != nil && (*s.stop_lon < -180 || *s.stop_lon > 180) {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "stop_lon", Message: fmt.Sprintf("value %f out of range [-180, 180]", *s.stop_lon)})
		}

		// enum range
		if s.location_type < StopPlatform || s.location_type > BoardingArea {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "location_type", Message: fmt.Sprintf("invalid value %d, must be 0-4", s.location_type)})
		}
		if s.wheelchair_boarding < NoAccessibilityInfo || s.wheelchair_boarding > NotAccessible {
			errs = append(errs, ValidationError{File: "stops.txt", ID: id, Field: "wheelchair_boarding", Message: fmt.Sprintf("invalid value %d, must be 0-2", s.wheelchair_boarding)})
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
		id := r.route_id
		if id == "" {
			errs = append(errs, ValidationError{File: "routes.txt", Field: "route_id", Message: "required field is empty"})
		}
		if r.route_short_name == "" && r.route_long_name == "" {
			errs = append(errs, ValidationError{File: "routes.txt", ID: id, Field: "route_short_name/route_long_name", Message: "at least one of route_short_name or route_long_name is required"})
		}
		if !validRouteTypes[r.route_type] {
			errs = append(errs, ValidationError{File: "routes.txt", ID: id, Field: "route_type", Message: fmt.Sprintf("invalid value %d", r.route_type)})
		}
		if r.agency_id == nil && len(gtfs.AgencyData) > 1 {
			errs = append(errs, ValidationError{File: "routes.txt", ID: id, Field: "agency_id", Message: "required when feed has multiple agencies"})
		}
	}

	return errs
}

func (gtfs *GTFS) validateTrips() []ValidationError {
	var errs []ValidationError

	for _, t := range gtfs.TripData {
		id := t.trip_id
		if id == "" {
			errs = append(errs, ValidationError{File: "trips.txt", Field: "trip_id", Message: "required field is empty"})
		}
		if t.route_id == nil {
			errs = append(errs, ValidationError{File: "trips.txt", ID: id, Field: "route_id", Message: "references a route_id that does not exist"})
		}
		if t.service_id == nil {
			errs = append(errs, ValidationError{File: "trips.txt", ID: id, Field: "service_id", Message: "references a service_id that does not exist in calendar.txt or calendar_dates.txt"})
		}
		if t.direction_id < OutboundTravel || t.direction_id > InboundTravel {
			errs = append(errs, ValidationError{File: "trips.txt", ID: id, Field: "direction_id", Message: fmt.Sprintf("invalid value %d, must be 0 or 1", t.direction_id)})
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
		if st.trip_id != nil {
			tripID = st.trip_id.trip_id
		}

		if st.trip_id == nil {
			errs = append(errs, ValidationError{File: "stop_times.txt", Field: "trip_id", Message: "references a trip_id that does not exist"})
		}
		if st.stop_id == nil {
			errs = append(errs, ValidationError{File: "stop_times.txt", ID: tripID, Field: "stop_id", Message: "references a stop_id that does not exist"})
		}

		// time format
		if st.arrival_time != "" && !timeRegex.MatchString(st.arrival_time) {
			errs = append(errs, ValidationError{File: "stop_times.txt", ID: tripID, Field: "arrival_time", Message: fmt.Sprintf("invalid format %q, expected HH:MM:SS", st.arrival_time)})
		}
		if st.departure_time != "" && !timeRegex.MatchString(st.departure_time) {
			errs = append(errs, ValidationError{File: "stop_times.txt", ID: tripID, Field: "departure_time", Message: fmt.Sprintf("invalid format %q, expected HH:MM:SS", st.departure_time)})
		}

		if tripID != "" {
			tripSeqs[tripID] = append(tripSeqs[tripID], seqEntry{seq: st.stop_sequence, idx: i})
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
		id := c.service_id
		if id == "" {
			errs = append(errs, ValidationError{File: "calendar.txt", Field: "service_id", Message: "required field is empty"})
		}
		if c.start_date == "" {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "start_date", Message: "required field is empty"})
		}
		if c.end_date == "" {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "end_date", Message: "required field is empty"})
		}
		// days must be 0 or 1
		days := map[string]int{
			"monday": c.monday, "tuesday": c.tuesday, "wednesday": c.wednesday,
			"thursday": c.thursday, "friday": c.friday, "saturday": c.saturday, "sunday": c.sunday,
		}
		for field, val := range days {
			if val != 0 && val != 1 {
				errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: field, Message: fmt.Sprintf("invalid value %d, must be 0 or 1", val)})
			}
		}
		// validate date format YYYYMMDD
		if c.start_date != "" && !isValidDate(c.start_date) {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "start_date", Message: fmt.Sprintf("invalid date format %q, expected YYYYMMDD", c.start_date)})
		}
		if c.end_date != "" && !isValidDate(c.end_date) {
			errs = append(errs, ValidationError{File: "calendar.txt", ID: id, Field: "end_date", Message: fmt.Sprintf("invalid date format %q, expected YYYYMMDD", c.end_date)})
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
