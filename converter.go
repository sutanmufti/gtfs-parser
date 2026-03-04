package gtfsparser

func (gtfs *GTFS) Compile() {
	// 1. TripStopTimes: trip → sorted stop times
	gtfs.TripStopTimes = make(map[*Trip][]StopTime)
	for _, st := range gtfs.StopTimeData {
		t := st.TripID // already a *Trip from ParseStopTime
		gtfs.TripStopTimes[t] = append(gtfs.TripStopTimes[t], st)
	}
	for t := range gtfs.TripStopTimes {
		list := gtfs.TripStopTimes[t]
		for i := 1; i < len(list); i++ {
			for j := i; j > 0 && list[j].StopSequence < list[j-1].StopSequence; j-- {
				list[j], list[j-1] = list[j-1], list[j]
			}
		}
		gtfs.TripStopTimes[t] = list
	}

	// 2. RouteTrips: route → []*Trip
	gtfs.RouteTrips = make(map[*Route][]*Trip)
	for i := range gtfs.TripData {
		t := &gtfs.TripData[i]
		r := t.RouteID // already a *Route from ParseTrip
		gtfs.RouteTrips[r] = append(gtfs.RouteTrips[r], t)
	}

	// 3. StopRoutes: stop → []*Route (derived via RouteTrips + TripStopTimes)
	gtfs.StopRoutes = make(map[*Stop][]*Route)
	for r, trips := range gtfs.RouteTrips {
		seen := make(map[*Stop]bool)
		for _, t := range trips {
			for _, st := range gtfs.TripStopTimes[t] {
				s := st.StopID // already a *Stop from ParseStopTime
				if !seen[s] {
					seen[s] = true
					gtfs.StopRoutes[s] = append(gtfs.StopRoutes[s], r)
				}
			}
		}
	}

	// 4. TransfersFromStop: stop → []Transfer
	gtfs.TransfersFromStop = make(map[*Stop][]Transfer)
	for _, tr := range gtfs.TransferData {
		s := tr.FromStopID // already a *Stop from ParseTransfer
		gtfs.TransfersFromStop[s] = append(gtfs.TransfersFromStop[s], tr)
	}

	// 5. FrequenciesByTrip: trip → []Frequency
	gtfs.FrequenciesByTrip = make(map[*Trip][]Frequency)
	for _, f := range gtfs.FrequencyData {
		t := f.TripID // already a *Trip from ParseFrequency
		gtfs.FrequenciesByTrip[t] = append(gtfs.FrequenciesByTrip[t], f)
	}
}
