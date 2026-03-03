package gtfsparser

import (
	"archive/zip"
	"os"
	"testing"
)

// makeTempZip writes a zip to a temp file containing the given name→content pairs
// and returns its path. The caller must defer os.Remove on the returned path.
func makeTempZip(t *testing.T, files map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp("", "gtfs-test-*.zip")
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			f.Close()
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			f.Close()
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

// ---- Helper unit tests ----

func TestSanitizeHeaders_StripsBOM(t *testing.T) {
	headers := []string{"\uFEFFstop_id", "stop_name"}
	result := sanitizeHeaders(headers)
	if result[0] != "stop_id" {
		t.Errorf("expected 'stop_id', got %q", result[0])
	}
}

func TestSanitizeHeaders_NoChange(t *testing.T) {
	headers := []string{"stop_id", "stop_name"}
	result := sanitizeHeaders(headers)
	if result[0] != "stop_id" {
		t.Errorf("expected 'stop_id', got %q", result[0])
	}
}

func TestSanitizeHeaders_Empty(t *testing.T) {
	result := sanitizeHeaders([]string{})
	if len(result) != 0 {
		t.Error("expected empty slice")
	}
}

func TestGetCol_Found(t *testing.T) {
	row := []string{"A", "B", "C"}
	col := map[string]int{"x": 0, "y": 1, "z": 2}
	if got := getCol(row, col, "y"); got != "B" {
		t.Errorf("expected 'B', got %q", got)
	}
}

func TestGetCol_MissingKey(t *testing.T) {
	row := []string{"A"}
	col := map[string]int{"x": 0}
	if got := getCol(row, col, "missing"); got != "" {
		t.Errorf("expected empty string for missing key, got %q", got)
	}
}

func TestGetCol_OutOfRange(t *testing.T) {
	row := []string{"A"}
	col := map[string]int{"x": 5}
	if got := getCol(row, col, "x"); got != "" {
		t.Errorf("expected empty string for out-of-range index, got %q", got)
	}
}

func TestParseOptionalFloat_Empty(t *testing.T) {
	v, err := parseOptionalFloat("")
	if err != nil || v != nil {
		t.Errorf("expected (nil, nil) for empty string, got (%v, %v)", v, err)
	}
}

func TestParseOptionalFloat_Valid(t *testing.T) {
	v, err := parseOptionalFloat("3.14")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v == nil || *v != 3.14 {
		t.Errorf("expected 3.14, got %v", v)
	}
}

func TestParseOptionalFloat_Invalid(t *testing.T) {
	_, err := parseOptionalFloat("not-a-number")
	if err == nil {
		t.Error("expected error for invalid float")
	}
}

func TestIsValidDate(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"20250101", true},
		{"20251231", true},
		{"2025010", false},   // too short
		{"202501011", false}, // too long
		{"2025ab01", false},  // non-digit
		{"20250001", false},  // month 0
		{"20251301", false},  // month 13
		{"20250100", false},  // day 0
		{"20250132", false},  // day 32
	}
	for _, tc := range tests {
		if got := isValidDate(tc.input); got != tc.valid {
			t.Errorf("isValidDate(%q) = %v, want %v", tc.input, got, tc.valid)
		}
	}
}

// ---- Parser unit tests ----

func TestParseAgency_Success(t *testing.T) {
	content := "agency_id,agency_name,agency_url,agency_timezone\n" +
		"A1,Agency One,http://a1.com,UTC\n" +
		"A2,Agency Two,http://a2.com,America/New_York\n"
	path := makeTempZip(t, map[string]string{"agency.txt": content})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseAgency(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gtfs.AgencyData) != 2 {
		t.Errorf("expected 2 agencies, got %d", len(gtfs.AgencyData))
	}
	if gtfs.AgencyData[0].agency_name != "Agency One" {
		t.Errorf("unexpected agency_name: %q", gtfs.AgencyData[0].agency_name)
	}
}

func TestParseAgency_BOM(t *testing.T) {
	content := "\uFEFFagency_id,agency_name,agency_url,agency_timezone\n" +
		"A1,Agency One,http://a1.com,UTC\n"
	path := makeTempZip(t, map[string]string{"agency.txt": content})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseAgency(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gtfs.AgencyData[0].agency_id != "A1" {
		t.Errorf("BOM not stripped from header: agency_id = %q", gtfs.AgencyData[0].agency_id)
	}
}

func TestParseAgency_Duplicate(t *testing.T) {
	content := "agency_id,agency_name,agency_url,agency_timezone\n" +
		"A1,Agency One,http://a1.com,UTC\n" +
		"A1,Agency One Dup,http://a1dup.com,UTC\n"
	path := makeTempZip(t, map[string]string{"agency.txt": content})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseAgency(); err == nil {
		t.Error("expected error for duplicate agency_id")
	}
}

func TestParseStop_Success(t *testing.T) {
	content := "stop_id,stop_name,stop_lat,stop_lon\n" +
		"S1,Stop One,1.0,2.0\n" +
		"S2,Stop Two,,\n"
	path := makeTempZip(t, map[string]string{"stops.txt": content})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseStop(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gtfs.StopData) != 2 {
		t.Errorf("expected 2 stops, got %d", len(gtfs.StopData))
	}
	if gtfs.StopData[0].stop_lat == nil || *gtfs.StopData[0].stop_lat != 1.0 {
		t.Errorf("expected stop_lat=1.0 for S1")
	}
	if gtfs.StopData[1].stop_lat != nil {
		t.Errorf("expected stop_lat=nil for S2")
	}
}

func TestParseStop_Duplicate(t *testing.T) {
	content := "stop_id,stop_name\n" +
		"S1,Stop One\n" +
		"S1,Stop One Dup\n"
	path := makeTempZip(t, map[string]string{"stops.txt": content})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseStop(); err == nil {
		t.Error("expected error for duplicate stop_id")
	}
}

func TestParseStop_ParentStation(t *testing.T) {
	content := "stop_id,stop_name,parent_station\n" +
		"STATION,Main Station,\n" +
		"PLATFORM,Platform A,STATION\n"
	path := makeTempZip(t, map[string]string{"stops.txt": content})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseStop(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	platform := gtfs.StopData[1]
	if platform.parent_station == nil {
		t.Fatal("expected parent_station to be resolved")
	}
	if platform.parent_station.stop_id != "STATION" {
		t.Errorf("expected parent stop_id 'STATION', got %q", platform.parent_station.stop_id)
	}
}

func TestParseCalendar_Duplicate(t *testing.T) {
	content := "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\n" +
		"SVC1,1,1,1,1,1,0,0,20250101,20251231\n" +
		"SVC1,0,0,0,0,0,1,1,20250101,20251231\n"
	path := makeTempZip(t, map[string]string{"calendar.txt": content})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseCalendar(); err == nil {
		t.Error("expected error for duplicate service_id")
	}
}

func TestParseShape_Optional(t *testing.T) {
	path := makeTempZip(t, map[string]string{"agency.txt": "agency_id\nA1\n"})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseShape(); err != nil {
		t.Errorf("ParseShape should return nil when shapes.txt is absent, got: %v", err)
	}
	if len(gtfs.ShapeData) != 0 {
		t.Errorf("expected empty ShapeData when file is absent")
	}
}

func TestParseRoute_AgencyFK(t *testing.T) {
	agencyCSV := "agency_id,agency_name,agency_url,agency_timezone\nA1,Agency,http://a.com,UTC\n"
	routeCSV := "route_id,agency_id,route_short_name,route_type\nR1,A1,Bus 1,3\n"
	path := makeTempZip(t, map[string]string{
		"agency.txt": agencyCSV,
		"routes.txt": routeCSV,
	})
	defer os.Remove(path)

	gtfs := GTFS{FileName: path}
	if err := gtfs.ParseAgency(); err != nil {
		t.Fatal(err)
	}
	if err := gtfs.ParseRoute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gtfs.RouteData) != 1 {
		t.Fatalf("expected 1 route, got %d", len(gtfs.RouteData))
	}
	if gtfs.RouteData[0].agency_id == nil {
		t.Error("expected agency_id FK to be resolved")
	}
	if gtfs.RouteData[0].agency_id.agency_id != "A1" {
		t.Errorf("expected agency_id 'A1', got %q", gtfs.RouteData[0].agency_id.agency_id)
	}
}

// ---- Integration tests ----

func TestParseAll_DefaultFeed(t *testing.T) {
	gtfs := GTFS{FileName: "gtfs_default3.zip"}
	if err := gtfs.ParseAll(); err != nil {
		t.Fatalf("ParseAll failed: %v", err)
	}
	if len(gtfs.AgencyData) == 0 {
		t.Error("expected at least one agency")
	}
	if len(gtfs.StopData) == 0 {
		t.Error("expected at least one stop")
	}
	if len(gtfs.RouteData) == 0 {
		t.Error("expected at least one route")
	}
	if len(gtfs.TripData) == 0 {
		t.Error("expected at least one trip")
	}
	if len(gtfs.StopTimeData) == 0 {
		t.Error("expected at least one stop time")
	}
}

func TestValidateAll_KnownErrors(t *testing.T) {
	gtfs := GTFS{FileName: "gtfs_default3.zip"}
	if err := gtfs.ParseAll(); err != nil {
		t.Fatalf("ParseAll failed: %v", err)
	}
	errs := gtfs.ValidateAll()
	if len(errs) != 4 {
		t.Errorf("expected 4 validation errors, got %d:", len(errs))
		for _, e := range errs {
			t.Log(" ", e)
		}
	}
}

// ---- Validation unit tests ----

func TestValidateRequiredFiles_AllMissing(t *testing.T) {
	gtfs := GTFS{}
	errs := gtfs.validateRequiredFiles()
	// 5 required files + 1 for calendar/calendar_dates = 6
	if len(errs) != 6 {
		t.Errorf("expected 6 errors for empty GTFS, got %d", len(errs))
		for _, e := range errs {
			t.Log(" ", e)
		}
	}
}

func TestValidateAgency_MissingFields(t *testing.T) {
	gtfs := GTFS{
		AgencyData: []Agency{
			{agency_id: "1", agency_name: "", agency_url: "", agency_timezone: ""},
		},
	}
	errs := gtfs.validateAgency()
	if len(errs) != 3 {
		t.Errorf("expected 3 errors (name, url, timezone), got %d", len(errs))
	}
}

func TestValidateStops_LatLonPairing(t *testing.T) {
	lat := 1.0
	lon := 2.0
	gtfs := GTFS{
		StopData: []Stop{
			{stop_id: "S1", stop_name: "Stop 1", stop_lat: &lat, stop_lon: nil}, // missing lon
			{stop_id: "S2", stop_name: "Stop 2", stop_lat: nil, stop_lon: &lon}, // missing lat
			{stop_id: "S3", stop_name: "Stop 3", stop_lat: &lat, stop_lon: &lon}, // valid pair
			{stop_id: "S4", stop_name: "Stop 4", stop_lat: nil, stop_lon: nil},   // neither - valid
		},
	}
	errs := gtfs.validateStops()
	if len(errs) != 2 {
		t.Errorf("expected 2 pairing errors, got %d", len(errs))
		for _, e := range errs {
			t.Log(" ", e)
		}
	}
}

func TestValidateStops_CoordinateRange(t *testing.T) {
	badLat := 91.0
	badLon := 181.0
	goodLat := 10.0
	goodLon := 10.0
	gtfs := GTFS{
		StopData: []Stop{
			{stop_id: "S1", stop_name: "Stop 1", stop_lat: &badLat, stop_lon: &goodLon},
			{stop_id: "S2", stop_name: "Stop 2", stop_lat: &goodLat, stop_lon: &badLon},
		},
	}
	errs := gtfs.validateStops()
	if len(errs) != 2 {
		t.Errorf("expected 2 range errors, got %d", len(errs))
		for _, e := range errs {
			t.Log(" ", e)
		}
	}
}

func TestValidateRoutes_InvalidType(t *testing.T) {
	gtfs := GTFS{
		RouteData: []Route{
			{route_id: "R1", route_short_name: "1", route_type: RouteType(99)},
		},
	}
	errs := gtfs.validateRoutes()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for invalid route_type, got %d", len(errs))
	}
}

func TestValidateRoutes_MissingName(t *testing.T) {
	gtfs := GTFS{
		RouteData: []Route{
			{route_id: "R1", route_short_name: "", route_long_name: "", route_type: Bus},
		},
	}
	errs := gtfs.validateRoutes()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for missing short and long name, got %d", len(errs))
	}
}

func TestValidateRoutes_AgencyIDRequired(t *testing.T) {
	a1 := Agency{agency_id: "A1"}
	a2 := Agency{agency_id: "A2"}
	gtfs := GTFS{
		AgencyData: []Agency{a1, a2},
		RouteData: []Route{
			{route_id: "R1", route_short_name: "1", route_type: Bus, agency_id: nil},
		},
	}
	errs := gtfs.validateRoutes()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for missing agency_id with multiple agencies, got %d", len(errs))
	}
}

func TestValidateTrips_MissingFK(t *testing.T) {
	gtfs := GTFS{
		TripData: []Trip{
			{trip_id: "T1", route_id: nil, service_id: nil},
		},
	}
	errs := gtfs.validateTrips()
	if len(errs) != 2 {
		t.Errorf("expected 2 errors (route_id, service_id), got %d", len(errs))
	}
}

func TestValidateTrips_InvalidDirectionID(t *testing.T) {
	route := Route{route_id: "R1"}
	cal := Calendar{service_id: "SVC1"}
	gtfs := GTFS{
		TripData: []Trip{
			{trip_id: "T1", route_id: &route, service_id: &cal, direction_id: DirectionId(5)},
		},
	}
	errs := gtfs.validateTrips()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for invalid direction_id, got %d", len(errs))
	}
}

func TestValidateStopTimes_InvalidTimeFormat(t *testing.T) {
	trip := Trip{trip_id: "T1"}
	stop := Stop{stop_id: "S1"}
	gtfs := GTFS{
		StopTimeData: []StopTime{
			{trip_id: &trip, stop_id: &stop, stop_sequence: 1,
				arrival_time: "25:70:00", departure_time: "08:00:00"},
		},
	}
	errs := gtfs.validateStopTimes()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for invalid arrival_time, got %d", len(errs))
		for _, e := range errs {
			t.Log(" ", e)
		}
	}
}

func TestValidateStopTimes_OutOfOrder(t *testing.T) {
	trip := Trip{trip_id: "T1"}
	stop := Stop{stop_id: "S1"}
	gtfs := GTFS{
		StopTimeData: []StopTime{
			{trip_id: &trip, stop_id: &stop, stop_sequence: 3,
				arrival_time: "08:00:00", departure_time: "08:00:00"},
			{trip_id: &trip, stop_id: &stop, stop_sequence: 1,
				arrival_time: "08:05:00", departure_time: "08:05:00"},
			{trip_id: &trip, stop_id: &stop, stop_sequence: 1, // duplicate sequence
				arrival_time: "08:10:00", departure_time: "08:10:00"},
		},
	}
	errs := gtfs.validateStopTimes()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for duplicate stop_sequence, got %d", len(errs))
	}
}

func TestValidateCalendar_InvalidDate(t *testing.T) {
	gtfs := GTFS{
		CalendarData: []Calendar{
			{service_id: "SVC1", start_date: "baddate", end_date: "20251231",
				monday: 1, tuesday: 1, wednesday: 1, thursday: 1, friday: 1},
		},
	}
	errs := gtfs.validateCalendar()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for invalid start_date, got %d", len(errs))
	}
}

func TestValidateCalendar_InvalidDayValue(t *testing.T) {
	gtfs := GTFS{
		CalendarData: []Calendar{
			{service_id: "SVC1", start_date: "20250101", end_date: "20251231",
				monday: 2}, // must be 0 or 1
		},
	}
	errs := gtfs.validateCalendar()
	if len(errs) != 1 {
		t.Errorf("expected 1 error for invalid monday value, got %d", len(errs))
	}
}
