package gtfsparser

// CemvSupport indicates whether an agency supports contactless EMV payment.
type CemvSupport int

const (
	Empty        CemvSupport = iota // No information provided.
	CEMV                            // Contactless EMV payment is supported.
	NotSupported                    // Contactless EMV payment is not supported.
)

// Agency represents a transit agency from agency.txt.
type Agency struct {
	AgencyID       string
	AgencyName     string
	AgencyURL      string
	AgencyTimezone string
	AgencyLang     string
	AgencyPhone    string
	AgencyFareURL  string
	AgencyEmail    string
	CemvSupport    CemvSupport
}

// LocationType classifies the type of a stop location as defined in stops.txt.
type LocationType int

const (
	StopPlatform LocationType = iota // A stop or platform where passengers board or alight.
	Station                          // A physical station containing one or more platforms.
	EntranceExit                     // An entrance or exit to a station.
	GenericNode                      // A generic node within a station, not used for boarding.
	BoardingArea                     // A specific boarding area on a platform.
)

// WheelchairBoarding indicates the wheelchair accessibility of a stop.
type WheelchairBoarding int

const (
	NoAccessibilityInfo WheelchairBoarding = iota // No accessibility information available.
	Accessible                                    // At least some vehicles can accommodate a wheelchair.
	NotAccessible                                 // Wheelchair boarding is not possible.
)

// Stop represents a stop, station, or other location from stops.txt.
// StopLat and StopLon are pointers so that nil represents an absent value,
// distinguishable from a valid coordinate of zero.
type Stop struct {
	StopID             string
	StopCode           string
	StopName           string
	StopDesc           string
	StopLat            *float64
	StopLon            *float64
	ZoneID             string
	StopURL            string
	LocationType       LocationType
	ParentStation      *Stop
	StopTimezone       string
	WheelchairBoarding WheelchairBoarding
	LevelID            *Level
	PlatformCode       string
}

// RouteType classifies the mode of transport used on a route.
type RouteType int

const (
	Tram       RouteType = iota // Tram, streetcar, or light rail.
	Subway                      // Underground or metro rail.
	Rail                        // Intercity or long-distance rail.
	Bus                         // Short- or long-distance bus.
	Ferry                       // Boat service.
	CableTram                   // Street-level cable car.
	AerialLift                  // Aerial tramway or gondola.
	Funicular                   // Funicular railway.
)

// PickupDropOffType specifies how passengers may board or alight at a stop.
type PickupDropOffType int

const (
	RegularlyScheduled   PickupDropOffType = iota // Regularly scheduled pickup or drop-off.
	NoPickupDropOff                               // No pickup or drop-off available.
	PhoneAgency                                   // Must phone the agency to arrange.
	CoordinateWithDriver                          // Must coordinate with the driver.
)

// Route represents a transit route from routes.txt.
// A route is a group of trips displayed to riders as a single service.
type Route struct {
	RouteID           string
	AgencyID          *Agency
	RouteShortName    string
	RouteLongName     string
	RouteDesc         string
	RouteType         RouteType
	RouteURL          string
	RouteColor        string
	RouteTextColor    string
	RouteSortOrder    int
	ContinuousPickup  PickupDropOffType
	ContinuousDropOff PickupDropOffType
	NetworkID         string
}

// DirectionId indicates the direction of travel for a trip.
type DirectionId int

const (
	OutboundTravel DirectionId = iota // Travel in the outbound direction.
	InboundTravel                     // Travel in the inbound direction.
)

// WheelchairAccessibleEnum indicates the wheelchair accessibility of a trip.
type WheelchairAccessibleEnum int

const (
	NoWheelchairInfo     WheelchairAccessibleEnum = iota // No accessibility information available.
	WheelchairAccessible                                 // At least one wheelchair can be accommodated.
	NoWheelchair                                         // No wheelchairs can be accommodated.
)

// BikesAllowed indicates whether bicycles are permitted on a trip.
type BikesAllowed int

const (
	NoBikeInfo  BikesAllowed = iota // No bicycle information available.
	BikeAllowed                     // At least one bicycle may be brought aboard.
	NoBike                          // No bicycles are permitted.
)

// Trip represents a journey made by a vehicle from trips.txt.
// A trip is a sequence of stops that occurs at a specific time.
type Trip struct {
	RouteID              *Route
	ServiceID            *Calendar
	TripID               string
	TripHeadsign         string
	TripShortName        string
	DirectionID          DirectionId
	BlockID              string
	ShapeID              *Shape
	WheelchairAccessible WheelchairAccessibleEnum
	BikesAllowed         BikesAllowed
}

// Timepoint indicates the precision of arrival and departure times for a stop.
type Timepoint int

const (
	ApproximateTime Timepoint = iota // Times are approximate.
	ExactTime                        // Times are exact.
)

// StopTime represents a vehicle's arrival and departure at a stop from stop_times.txt.
type StopTime struct {
	TripID            *Trip
	ArrivalTime       string
	DepartureTime     string
	StopID            *Stop
	StopSequence      int
	StopHeadsign      string
	PickupType        PickupDropOffType
	DropOffType       PickupDropOffType
	ContinuousPickup  PickupDropOffType
	ContinuousDropOff PickupDropOffType
	ShapeDistTraveled float64
	Timepoint         Timepoint
}

// Calendar represents a regular weekly service schedule from calendar.txt.
type Calendar struct {
	ServiceID string
	Monday    int
	Tuesday   int
	Wednesday int
	Thursday  int
	Friday    int
	Saturday  int
	Sunday    int
	StartDate string
	EndDate   string
}

// ExceptionType indicates whether service has been added or removed on a given date.
type ExceptionType int

const (
	ServiceAdded   ExceptionType = iota + 1 // Service has been added for this date.
	ServiceRemoved                          // Service has been removed for this date.
)

// CalendarDate represents an exception to a regular service schedule from calendar_dates.txt.
type CalendarDate struct {
	ServiceID     *Calendar
	Date          string
	ExceptionType ExceptionType
}

// Shape represents a single point in a route's geographic path from shapes.txt.
// All points sharing a ShapeID together define the drawn path of a trip.
type Shape struct {
	ShapeID           string
	ShapePtLat        float64
	ShapePtLon        float64
	ShapePtSequence   int
	ShapeDistTraveled float64
}

// ExactTimes indicates whether a frequency-based trip follows an exact schedule.
type ExactTimes int

const (
	FrequencyBased ExactTimes = iota // Trip does not adhere to a fixed schedule.
	ScheduleBased                    // Trip follows a fixed schedule based on the start time.
)

// Frequency represents a headway-based trip interval from frequencies.txt.
type Frequency struct {
	TripID      *Trip
	StartTime   string
	EndTime     string
	HeadwaySecs int
	ExactTimes  ExactTimes
}

// TransferType specifies the type of connection between two stops or routes.
type TransferType int

const (
	RecommendedTransfer TransferType = iota // Transfer is recommended at this point.
	TimedTransfer                           // Transfer is timed; the departing vehicle waits.
	MinTimeTransfer                         // A minimum transfer time is required.
	ImpossibleTransfer                      // Transfer is not possible at this location.
)

// Transfer represents a connection rule between two stops or routes from transfers.txt.
type Transfer struct {
	FromStopID      *Stop
	ToStopID        *Stop
	FromRouteID     *Route
	ToRouteID       *Route
	FromTripID      *Trip
	ToTripID        *Trip
	TransferType    TransferType
	MinTransferTime int
}

// PaymentMethod indicates when a fare must be paid.
type PaymentMethod int

const (
	PaidOnBoard        PaymentMethod = iota // Fare is paid on board.
	PaidBeforeBoarding                      // Fare must be paid before boarding.
)

// FareTransfers specifies the number of transfers permitted on a fare.
type FareTransfers int

const (
	NoTransfersPermitted  FareTransfers = iota // No transfers are permitted.
	OneTransferPermitted                       // One transfer is permitted.
	TwoTransfersPermitted                      // Two transfers are permitted.
)

// FareAttribute represents a fare class from fare_attributes.txt.
type FareAttribute struct {
	FareID           string
	Price            float64
	CurrencyType     string
	PaymentMethod    PaymentMethod
	Transfers        FareTransfers
	AgencyID         *Agency
	TransferDuration int
}

// FareRule represents a rule for applying a fare to an itinerary from fare_rules.txt.
type FareRule struct {
	FareID        *FareAttribute
	RouteID       *Route
	OriginID      string
	DestinationID string
	ContainsID    string
}

// FeedInfo contains metadata about the GTFS feed from feed_info.txt.
type FeedInfo struct {
	FeedPublisherName string
	FeedPublisherURL  string
	FeedLang          string
	DefaultLang       string
	FeedStartDate     string
	FeedEndDate       string
	FeedVersion       string
	FeedContactEmail  string
	FeedContactURL    string
}

// PathwayMode describes the type of pathway connecting two locations within a station.
type PathwayMode int

const (
	Walkway       PathwayMode = iota + 1 // A walkway.
	Stairs                               // A staircase.
	MovingWalkway                        // A moving walkway or travelator.
	Escalator                            // An escalator.
	Elevator                             // A lift.
	FareGate                             // A fare gate or turnstile.
	ExitGate                             // A pathway that leads out of the fare area.
)

// Pathway represents a connection between two locations within a station from pathways.txt.
type Pathway struct {
	PathwayID            string
	FromStopID           *Stop
	ToStopID             *Stop
	PathwayMode          PathwayMode
	IsBidirectional      int
	Length               float64
	TraversalTime        int
	StairCount           int
	MaxSlope             float64
	MinWidth             float64
	SignpostedAs         string
	ReversedSignpostedAs string
}

// Level represents a floor level within a station from levels.txt.
type Level struct {
	LevelID    string
	LevelIndex float64
	LevelName  string
}

// Attribution represents an organisation's role in producing a GTFS dataset from attributions.txt.
type Attribution struct {
	AttributionID    string
	AgencyID         *Agency
	RouteID          *Route
	TripID           *Trip
	OrganizationName string
	IsProducer       int
	IsOperator       int
	IsAuthority      int
	AttributionURL   string
	AttributionEmail string
	AttributionPhone string
}

// Translation represents a translated value for a field in the feed from translations.txt.
type Translation struct {
	TableName   string
	FieldName   string
	Language    string
	Translation string
	RecordID    string
	RecordSubID string
	FieldValue  string
}

// GTFS holds all parsed data from a GTFS feed zip file.
// Set FileName to the path of the zip archive, then call ParseAll to load all
// files, followed by ValidateAll to check the data against the specification.
type GTFS struct {
	FileName       string
	AgencyData     []Agency
	StopData       []Stop
	RouteData      []Route
	TripData       []Trip
	StopTimeData   []StopTime
	CalendarData   []Calendar
	CalendarDates  []CalendarDate
	ShapeData      []Shape
	FrequencyData  []Frequency
	TransferData   []Transfer
	FareAttributes []FareAttribute
	FareRules      []FareRule
	FeedInfo       []FeedInfo
	PathwayData    []Pathway
	LevelData      []Level
	Attributions   []Attribution
	Translations   []Translation

	TripStopTimes     map[*Trip][]StopTime
	TripRoutes        map[*Trip][]StopTime
	StopRoutes        map[*Stop][]Route
	TransfersFromStop map[*Stop][]Transfer
	FrequenciesByTrip map[*Trip][]Frequency
}
