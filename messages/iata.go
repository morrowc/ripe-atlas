package messages

// IATA Cities request struct.
type IATACitiesRequest struct {
	Code string `json:"code"`
}

//
// IATA response message from API request.
type IATAResponse struct {
	Request  IATARequest          `json:"request"`
	Response []IATACitiesResponse `json:"response"`
}

// IATA Request, a full fledged request object.
// (no idea why this is returned)
type IATARequest struct {
	Lang     string `json:"lang"`
	Currency string `json:"currency"`
	Time     int    `json:"time"`
	Id       int64  `json:"id"`
	Server   string `json:"server"`
	Host     string `json:"host"`
	PID      int32  `json:"pid"`
	Version  int    `json:"version"`
	Method   string `json:"method"`
}

// IATA Cities response.
type IATACitiesResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
}
