// messages implements the structs used to manage json messages to and from
// the RIPE atlas system.
package messages

// MultiMeasurementRespnseMessage is returned when querying for
// measurements instead of requesting a direct measurement.
type MultiMeasurementResponseMessage struct {
	Count    int32                        `json:"count"`
	Next     int32                        `json:"next"`
	Previous int32                        `json:"previous"`
	Results  []MeasurementResponseMessage `json:"results"`
}

// MeasurementResponseMessage is the JSON struct returned by a ripe
// 'request measurement status' request.
type MeasurementResponseMessage struct {
	Af               int            `json:"af"`
	CreationTime     int32          `json:"creation_time"`
	Description      string         `json:"description"`
	Group            string         `json:"group"`
	GroupId          int32          `json:"group_id"`
	Id               int32          `json:"id"`
	InWifiGroup      bool           `json:"in_wifi_group"`
	Interval         int32          `json:"interval"`
	IsAllScheduled   bool           `json:"is_all_scheduled"`
	IsOneoff         bool           `json:"is_one_off"`
	IsPublic         bool           `json:"is_public"`
	PacketInterval   int32          `json:"packet_interval"`
	Packets          int            `json:"packets"`
	ParticipantCount int32          `json:"participant_count"`
	ProbesRequested  int            `json:"probes_requested"`
	ProbesScheduled  int            `json:"probes_scheduled"`
	ResolveOnProbe   bool           `json:"resolve_on_probe"`
	ResolvedIps      []string       `json:"resolved_ips"`
	Result           string         `json:"result"`
	Size             int32          `json:"size"`
	Spread           int32          `json:"spread"`
	StartTime        int32          `json:"start_time"`
	Status           ResponseStatus `json:"status"`
	StopTime         int32          `json:"stop_time"`
	Target           string         `json:"target"`
	TargetAsn        int32          `json:"target_asn"`
	TargetIp         string         `json:"target_ip"`
	Type             string         `json:"type"`
}

// ResponseStatus is the JSON struct used to hold 'status' data for a measurement's status.
// Id is an integer, Name is a string value:
// (0: Specified, 1: Scheduled, 2: Ongoing, 4: Stopped,
//  5: Forced to stop, 6: No suitable probes, 7: Failed, 8: Archived)
type ResponseStatus struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// MeasurementResultMessage is the JSON struct returned by a ripe
// 'results' request.
type MeasurementResultMessage struct {
	Af        int     `json:"af"`
	Avg       float64 `json:"avg"`
	DstAddr   string  `json:"dst_addr"`
	DstName   string  `json:"dst_name"`
	Dup       int     `json:"dup"`
	From      string  `json:"from"`
	Fw        int     `json:"fw"`
	GroupId   int32   `json:"group_id"`
	Lts       int     `json:"lts"`
	Max       float64 `json:"max"`
	Min       float64 `json:"min"`
	MsmId     int32   `json:"msm_id"`
	MsmName   string  `json:"msm_name"`
	PrbId     int32   `json:"prb_id"`
	Proto     string  `json:"proto"`
	Rcvd      int     `json:"rcvd"`
	Result    Results `json:"result"`
	Sent      int     `json:"sent"`
	Size      int     `json:"size"`
	SrcAddr   string  `json:"src_addr"`
	Step      int     `json:"step"`
	Timestamp int32   `json:"timestamp"`
	Ttl       int     `json:"ttl"`
	Type      string  `json:"type"`
	Uri       string  `json:"uri"`
}

// ProbeQueryResults is the JSON struct returned by a ripe
// /?probes request.
type ProbeQueryResults struct {
	Count    int            `json:"count"`
	Next     int            `json:"next"`
	Previous int            `json:"previous"`
	Results  []ProbeMessage `json:"results"`
}

// Results is the individual result content.
type Results struct {
	Af      int     `json:"af"`
	Bsize   int32   `json:"bsize"`
	DstAddr string  `json:"dst_addr"`
	Hsize   int32   `json:"hsize"`
	Methad  string  `json:"methad"`
	Res     int32   `json:"res"`
	Rt      float64 `json:"rt"`
	SrcAddr string  `json:"src_addr"`
	Ver     string  `json:"ver"`
	Rtt     float64 `json:"rtt"`
}

// ProbeMessage is the JSON struct returned by a ripe probe query.
type ProbeMessage struct {
	AddressV4      string      `json:"address_v4"`
	AddressV6      string      `json:"address_v6"`
	ASNv4          int32       `json:"asn_v4"`
	ASNv6          int32       `json:"asn_v6"`
	CountryCode    string      `json:"country_code"`
	Description    string      `json:"description"`
	FirstConnected int32       `json:"first_connected"`
	Geometry       GeometryMsg `json:"geometry"`
	Id             int32       `json:"id"`
	IsAnchor       bool        `json:"is_anchor"`
	IsPublic       bool        `json:"is_public"`
	LastConnected  int32       `json:"last_connected"`
	PrefixV4       string      `json:"prefix_v4"`
	PrefixV6       string      `json:"prefix_v6"`
	Status         ProbeStatus `json:"status"`
	StatusSince    int32       `json:"status_since"`
	Tags           []TagSet    `json:"tags"`
	TotalUptime    int32       `json:"total_uptime"`
	Type           string      `json:"type"`
}

type TagSet struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type GeometryMsg struct {
	Type        string    `json:"type"`
	Coordinates []float32 `json:"coordinates"`
}

type ProbeStatus struct {
	Since string `json:"since"`
	Id    int32  `json:"id"`
	Name  string `json:"name"`
}

// A measurement request contains a probes list.
// Currently 'type' should be "probes"
// value is a comma separated list probe Ids.
type ProbeSourceMessage struct {
	Requested int32  `json:"requested"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

// MeasurementRequest is a JSON struct to hold POST Variables required
// to generate a new measurement request to RipeAtlas, based upon the
// contents at:
// https://atlas.ripe.net/docs/api/v2/reference/#!/measurements/Measurement_List_POST
type MeasurementRequest struct {
	StartTime    string                  `json:"start_time"`
	StopTime     string                  `json:"stop_time"`
	Definitions  []MeasurementDefinition `json:"definitions"`
	Probe_Source []ProbeSourceMessage    `json:"probes"`
}

// MeasurementDefinition is the definition described in the MeasurementRequest
// api content for a v2 POST request.
// NOTE: Only Core && DNS measurement variables implemented initially.
type MeasurementDefinition struct {
	IsPublic            bool     `json:"is_public"`
	Description         string   `json:"description"`
	GroupId             int32    `json:"group_id,omitempty"`
	AF                  int32    `json:"af"` // 4 or 6
	IsOneOff            bool     `json:"is_oneoff"`
	Interval            int32    `json:"interval"`
	Spread              int32    `json:"spread"` // seconds to spread the requests over, 400 max
	ResolveOnProbe      bool     `json:"resolve_on_probe"`
	StartTime           string   `json:"start_time"` // unix timestamp
	StopTime            string   `json:"stop_time"`  // unix timestamp
	Type                string   `json:"type"`       // ping, traceroute, dns, http, ntp, wifi
	Tags                []string `json:"tags"`
	Target              string   `json:"target"`
	UdpPayloadSize      int32    `json:"udp_payload_size"`
	UseProbeResolver    bool     `json:"use_probe_resolver"`
	SetRDBit            bool     `json:"set_rd_bit"`
	PrependProbeId      bool     `json:"prepend_probe_id"`
	Protocol            string   `json:"protocol"` // UDP or TCP
	Retry               int32    `json:"retry"`
	IncludeQBuf         bool     `json:"include_q_buf"`
	SetNSIDBit          bool     `json:"set_nsid_bit"`
	QueryClass          string   `json:"query_class"` // IN or CHAOS
	QueryArgument       string   `json:"query_argument"`
	QueryType           string   `json:"query_type"` // A, AAAA, ANY, CNAME, DNSKEY, DS, TXT ...
	SetCDBit            bool     `json:"set_cd_bit"`
	UseMacros           bool     `json:"use_macros"`
	Timeout             int32    `json:"timeout"`
	TLS                 bool     `json:"tls"`
	Port                int32    `json:"port"`
	DefaultClientSubnet bool     `json:"default_client_subnet"`
}

type MeasurementResponse struct {
	Measurements []int32 `json:"measurements"`
}

// Permissions granted to API keys.
type Grant struct {
	Permission string `json:"permission"`
	Target     string `json:"target"`
}

// An API Key Request struct.
type ApiKeyRequest struct {
	UUID      string  `json:"uuid"`
	ValidFrom string  `json:"valid_from"`
	ValidTo   string  `json:"valid_to"`
	Enabled   bool    `json:"enabled"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	Label     string  `json:"label"`
	Grants    []Grant `json:"grants"`
}
