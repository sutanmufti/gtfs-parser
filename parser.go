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

func getCol(row []string, col map[string]int, name string) string {
	i, ok := col[name]
	if !ok || i >= len(row) {
		return ""
	}
	return row[i]
}

func parseOptionalFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
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
	col := make(map[string]int)
	for i, h := range headers {
		col[h] = i
	}

	// Parse each row into a Stop
	var stops []Stop
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
			parent_station:      getCol(row, col, "parent_station"),
			stop_timezone:       getCol(row, col, "stop_timezone"),
			wheelchair_boarding: WheelchairBoarding(wheelchair),
			level_id:            getCol(row, col, "level_id"),
			platform_code:       getCol(row, col, "platform_code"),
		}

		if seen[stop.stop_id] {
			duplicates = append(duplicates, stop.stop_id)
		} else {
			seen[stop.stop_id] = true
		}

		stops = append(stops, stop)
	}

	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate stop_id(s) found: %v", duplicates)
	}

	gtfs.StopData = stops
	return nil
}
