package gtfsparser

// 1. Create a custom type based on an integer or string
type CemvSupport int

// 2. Use const and iota to define values (starting from 0)
const (
	Empty CemvSupport = iota
	CEMV
	NotSupported
)

type Agency struct {
	agency_id       string
	agency_name     string
	agency_url      string
	agency_timezone string
	agency_lang     string
	agency_phone    string
	agency_fare_url string
	agency_email    string
	cemv_support    CemvSupport
}

type LocationType int

const (
	StopPlatform LocationType = iota // 0
	Station                          // 1
	EntranceExit                     // 2
	GenericNode                      // 3
	BoardingArea                     // 4
)

type WheelchairBoarding int

const (
	NoAccessibilityInfo WheelchairBoarding = iota // 0
	Accessible                                    // 1
	NotAccessible                                 // 2
)

type Stop struct {
	stop_id             string
	stop_code           string
	stop_name           string
	stop_desc           string
	stop_lat            *float64
	stop_lon            *float64
	zone_id             string
	stop_url            string
	location_type       LocationType
	parent_station      *Stop
	stop_timezone       string
	wheelchair_boarding WheelchairBoarding
	level_id            *Level
	platform_code       string
}

// routes.txt
type RouteType int

const (
	Tram       RouteType = iota // 0
	Subway                      // 1
	Rail                        // 2
	Bus                         // 3
	Ferry                       // 4
	CableTram                   // 5
	AerialLift                  // 6
	Funicular                   // 7
)

type PickupDropOffType int

const (
	RegularlyScheduled   PickupDropOffType = iota // 0
	NoPickupDropOff                               // 1
	PhoneAgency                                   // 2
	CoordinateWithDriver                          // 3
)

type Route struct {
	route_id            string
	agency_id           *Agency
	route_short_name    string
	route_long_name     string
	route_desc          string
	route_type          RouteType
	route_url           string
	route_color         string
	route_text_color    string
	route_sort_order    int
	continuous_pickup   PickupDropOffType
	continuous_drop_off PickupDropOffType
	network_id          string
}

// trips.txt
type DirectionId int

const (
	OutboundTravel DirectionId = iota // 0
	InboundTravel                     // 1
)

type WheelchairAccessibleEnum int

const (
	NoWheelchairInfo     WheelchairAccessibleEnum = iota // 0
	WheelchairAccessible                                 // 1
	NoWheelchair                                         // 2
)

type BikesAllowed int

const (
	NoBikeInfo  BikesAllowed = iota // 0
	BikeAllowed                     // 1
	NoBike                          // 2
)

type Trip struct {
	route_id              *Route
	service_id            *Calendar
	trip_id               string
	trip_headsign         string
	trip_short_name       string
	direction_id          DirectionId
	block_id              string
	shape_id              *Shape
	wheelchair_accessible WheelchairAccessibleEnum
	bikes_allowed         BikesAllowed
}

// stop_times.txt
type Timepoint int

const (
	ApproximateTime Timepoint = iota // 0
	ExactTime                        // 1
)

type StopTime struct {
	trip_id             *Trip
	arrival_time        string
	departure_time      string
	stop_id             *Stop
	stop_sequence       int
	stop_headsign       string
	pickup_type         PickupDropOffType
	drop_off_type       PickupDropOffType
	continuous_pickup   PickupDropOffType
	continuous_drop_off PickupDropOffType
	shape_dist_traveled float64
	timepoint           Timepoint
}

// calendar.txt
type Calendar struct {
	service_id string
	monday     int
	tuesday    int
	wednesday  int
	thursday   int
	friday     int
	saturday   int
	sunday     int
	start_date string
	end_date   string
}

// calendar_dates.txt
type ExceptionType int

const (
	ServiceAdded   ExceptionType = iota + 1 // 1
	ServiceRemoved                          // 2
)

type CalendarDate struct {
	service_id     *Calendar
	date           string
	exception_type ExceptionType
}

// shapes.txt
type Shape struct {
	shape_id            string
	shape_pt_lat        float64
	shape_pt_lon        float64
	shape_pt_sequence   int
	shape_dist_traveled float64
}

// frequencies.txt
type ExactTimes int

const (
	FrequencyBased ExactTimes = iota // 0
	ScheduleBased                    // 1
)

type Frequency struct {
	trip_id      *Trip
	start_time   string
	end_time     string
	headway_secs int
	exact_times  ExactTimes
}

// transfers.txt
type TransferType int

const (
	RecommendedTransfer TransferType = iota // 0
	TimedTransfer                           // 1
	MinTimeTransfer                         // 2
	ImpossibleTransfer                      // 3
)

type Transfer struct {
	from_stop_id      *Stop
	to_stop_id        *Stop
	from_route_id     *Route
	to_route_id       *Route
	from_trip_id      *Trip
	to_trip_id        *Trip
	transfer_type     TransferType
	min_transfer_time int
}

// fare_attributes.txt
type PaymentMethod int

const (
	PaidOnBoard        PaymentMethod = iota // 0
	PaidBeforeBoarding                      // 1
)

type FareTransfers int

const (
	NoTransfersPermitted  FareTransfers = iota // 0
	OneTransferPermitted                       // 1
	TwoTransfersPermitted                      // 2
)

type FareAttribute struct {
	fare_id           string
	price             float64
	currency_type     string
	payment_method    PaymentMethod
	transfers         FareTransfers
	agency_id         *Agency
	transfer_duration int
}

// fare_rules.txt
type FareRule struct {
	fare_id        *FareAttribute
	route_id       *Route
	origin_id      string
	destination_id string
	contains_id    string
}

// feed_info.txt
type FeedInfo struct {
	feed_publisher_name string
	feed_publisher_url  string
	feed_lang           string
	default_lang        string
	feed_start_date     string
	feed_end_date       string
	feed_version        string
	feed_contact_email  string
	feed_contact_url    string
}

// pathways.txt
type PathwayMode int

const (
	Walkway       PathwayMode = iota + 1 // 1
	Stairs                               // 2
	MovingWalkway                        // 3
	Escalator                            // 4
	Elevator                             // 5
	FareGate                             // 6
	ExitGate                             // 7
)

type Pathway struct {
	pathway_id             string
	from_stop_id           *Stop
	to_stop_id             *Stop
	pathway_mode           PathwayMode
	is_bidirectional       int
	length                 float64
	traversal_time         int
	stair_count            int
	max_slope              float64
	min_width              float64
	signposted_as          string
	reversed_signposted_as string
}

// levels.txt
type Level struct {
	level_id    string
	level_index float64
	level_name  string
}

// attributions.txt
type Attribution struct {
	attribution_id    string
	agency_id         *Agency
	route_id          *Route
	trip_id           *Trip
	organization_name string
	is_producer       int
	is_operator       int
	is_authority      int
	attribution_url   string
	attribution_email string
	attribution_phone string
}

// translations.txt
type Translation struct {
	table_name    string
	field_name    string
	language      string
	translation   string
	record_id     string
	record_sub_id string
	field_value   string
}

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
}
