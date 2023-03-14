package probes

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/kelvins/geocoder"
	"github.com/morrowc/ripe-atlas/messages"
)

const (
	// JSON rest requests require proper accept/content-type headers.
	contentType = "application/json"
	acceptType  = contentType

	// geocodingKey is a key for google's places API.
	geocodingKey = "AIzaSyDT_jA3hK1NiGlstMY4L9ng0hpmQGTadzs"
	atlasURL     = "https://atlas.ripe.net/api/v2/probes/?status=1&is_public=true&radius=%f,%f:%d&sort=id"
	// Use the openflights data from github, cache locally for repeat usage.
	openflightsURL = "https://raw.githubusercontent.com/jpatokal/openflights/master/data/airports.dat"
	tmpAirport     = "/tmp/airports.dat"
)

// 2,"Madang Airport","Madang","Papua New Guinea",
//   "MAG","AYMD",-5.20707988739,145.789001465,20,10,
//   "U","Pacific/Port_Moresby","airport","OurAirports"

type airport struct {
	id                               int32
	name, city, country, code, kcode string
	lat, long                        float64
	altitude, tz                     int32
	dst, tzDatabase, recType, source string
}

type airports []airport

func (a *airports) matchMetro(m string) *airport {
	for _, t := range *a {
		if strings.ToUpper(t.code) == strings.ToUpper(m) {
			return &t
		}
	}
	return nil
}

// parseAirports downloads the airport data or uses a local cache, parsing
// the content into a slice of airport structs.
func parseAirports(u string) (*airports, error) {
	// If the cache file does not exist, get it and parse result.
	if _, err := os.Stat(tmpAirport); os.IsNotExist(err) {
		resp, err := http.Get(u)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(tmpAirport, data, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Now, parse the cached file.
	fd, err := os.Open(tmpAirport)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(fd)
	// There are instances of \"foo\" inside a field
	// in this data, lazyquotes avoids erroring.
	csvReader.LazyQuotes = true
	var aps airports
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		var a airport
		tmp, err := strconv.ParseInt(rec[0], 10, 32)
		if err != nil {
			a.id = 0
		}
		a.id = int32(tmp)
		a.name = rec[1]
		a.city = rec[2]
		a.country = rec[3]
		a.code = rec[4]
		a.kcode = rec[5]
		a.lat, err = strconv.ParseFloat(rec[6], 64)
		if err != nil {
			a.lat = 0
		}
		a.long, err = strconv.ParseFloat(rec[7], 64)
		if err != nil {
			a.long = 0
		}
		tmp, err = strconv.ParseInt(rec[8], 10, 32)
		if err != nil {
			a.altitude = 0
		}
		a.altitude = int32(tmp)
		tmp, err = strconv.ParseInt(rec[9], 10, 32)
		if err != nil {
			a.tz = 0
		}
		a.tz = int32(tmp)
		a.dst = rec[10]
		a.tzDatabase = rec[11]
		a.recType = rec[12]
		a.source = rec[13]
		aps = append(aps, a)
	}

	return &aps, nil
}

// metroToCity attempts to convert a metro (FRA) to a city/country: Frankfurt, DE.
// Metro to city is really airport to city/country, with data provided from:
// https://raw.githubusercontent.com/jpatokal/openflights/master/data/airports.dat
// Download and parse the data, download only if the data does not already exist in
// temp location.
func metroToCity(metro *string) (*string, *string, error) {
	if len(*metro) != 3 {
		return nil, nil, fmt.Errorf("invound metro(%v) was not properly formatted.", metro)
	}

	aps, err := parseAirports(openflightsURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find/parse the airports data: %v", err)
	}

	// There was a single metro request, so there will always be a single response entry.
	rec := aps.matchMetro(*metro)
	if rec != nil {
		return &rec.city, &rec.country, nil
	}
	return nil, nil, fmt.Errorf("failed to find a city/country match for: %v", *metro)
}

// geolocate finds the latitude/longitude data based on an airport-code/metro code.
func geolocate(city, country *string) (*geocoder.Location, error) {
	// Setup the geocoder API request basics, address and run the conversion.
	geocoder.ApiKey = geocodingKey
	address := geocoder.Address{
		City:    *city,
		Country: *country,
	}

	location, err := geocoder.Geocoding(address)
	if err != nil {
		return nil, fmt.Errorf("failed to convert address -> location: %v\n", err)
	}

	return &location, nil
}

// LocateProbes queries the RIPE Atlas system for probes which match defined criteria.
func LocateProbes(metro *string, radius, count *int, v4, v6 *bool) ([]messages.ProbeMessage, error) {
	city, country, err := metroToCity(metro)
	if err != nil {
		return nil, fmt.Errorf("Failed to get ciy/country from metro: %v", err)
	}

	location, err := geolocate(city, country)
	if err != nil {
		return nil, fmt.Errorf("failed to geolocate the city/country: %v", err)
	}

	// Set the URL properly, and request from RIPE.
	url := fmt.Sprintf(atlasURL, location.Latitude, location.Longitude, *radius)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get the RIPE Probe Request: %v", err)
	}
	defer resp.Body.Close()

	// slurp the results,
	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response Body: %v", err)
	}

	var probes messages.ProbeQueryResults
	err = json.Unmarshal(c, &probes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the json probe data: %v", err)
	}
	var results []messages.ProbeMessage
	for _, p := range probes.Results {
		// Limit returned probes to those which are "Connected" and "Public"
		if p.Status.Name == "Connected" && p.IsPublic {
			switch {
			case *v4 && !*v6:
				if len(p.AddressV4) > 0 && len(p.AddressV6) == 0 {
					results = append(results, p)
				}
			case !*v4 && *v6:
				if len(p.AddressV4) == 0 && len(p.AddressV6) > 0 {
					results = append(results, p)
				}
			case *v4 && *v6:
				if len(p.AddressV4) > 0 && len(p.AddressV6) > 0 {
					results = append(results, p)
				}
			}
		}
	}
	// If there are enough valid probes trim the result set to the max requested.
	if len(results) > *count {
		return results[0:*count], nil
	}
	return results, nil
}
