// makeMeasurement will read a JSON file which is a measurement.
//
// Marshal a json text file into the POST variables which will
// be used to request a measurement creation, the variables are:
//   https://atlas.ripe.net/docs/api/v2/reference/#!/measurements/Dns_Type_Measurement_List_POST
// there is a messages/... struct for the JSON to follow as well.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/morrowc/ripe-atlas/messages"
	"github.com/morrowc/ripe-atlas/probes"
)

var (
	getKey = flag.Bool("getKey", false, "Get a new api-key from the atlas system.")
	mDef   = flag.String("measurement", "", "JSON file with prototype measurement request.")
	key    = flag.String("apiKey", "", "RipeAtlas API key, a file with a key as text.")
	count  = flag.Int("count", 5, "How many probes to return upon request.")
	metro  = flag.String("metro", "IAD", "What metro to constrain probe set to.")
	radius = flag.Int("radius", 10, "radius from metro center to constrain probe location.")
	v4     = flag.Bool("v4", true, "Should the probe have ipv4 addressing?")
	v6     = flag.Bool("v6", true, "Should the probe have ipv6 addressing?")
	mType  = flag.String("mType", "dns", "What type of probe request is being requested?")

	// by default the timespan of a measurement is 24h
	measurementDuration = flag.Duration("md", 24*time.Hour, "measurement duration")

	// atlasURLs maps the type of measurement to URL.
	atlasURLs = map[string]string{
		"dns":        "https://atlas.ripe.net:443/api/v2/measurements/dns/",
		"http":       "https://atlas.ripe.net:443/api/v2/measurements/http/",
		"ping":       "https://atlas.ripe.net:443/api/v2/measurements/ping/",
		"sslcert":    "https://atlas.ripe.net:443/api/v2/measurements/sslcert/",
		"traceroute": "https://atlas.ripe.net:443/api/v2/measurements/traceroute/",
		"results":    "https://atlas.ripe.net/api/v2/measurements/%d/",
		"key":        "https://atlas.ripe.net/api/v2/keys/",
	}
)

const (
	contentType = "application/json"
	acceptType  = contentType
	timeFmt     = "2006-01-02T15:04Z"
)

func readKey(key *string) (string, error) {
	k, err := ioutil.ReadFile(*key)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(k), "\n"), nil
}

func readJson(js *string) (*messages.MeasurementRequest, error) {
	j, err := ioutil.ReadFile(*js)
	if err != nil {
		return nil, err
	}

	var m messages.MeasurementRequest
	err = json.Unmarshal(j, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// formatRequest is a debugging utility function, which
// simply pretty-prints the outbound http request.
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string

	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)

	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))

	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

func measurementGroom(m *messages.MeasurementRequest, probeIds []string, metro *string, md *time.Duration) {
	// Create a probes message to add to the measurement.
	m.Probe_Source = append(m.Probe_Source, messages.ProbeSourceMessage{
		Requested: int32(*count),
		Type:      "probes",
		Value:     strings.Join(probeIds, ","),
	})

	// Adjust the description for a measurement definition.
	m.Definitions[0].Description = fmt.Sprintf(
		m.Definitions[0].Description, fmt.Sprintf("%v", *metro))

	// Add the start/stop times to the measurement, default
	startTime := time.Now().Add(10 * time.Minute)
	stopTime := startTime.Add(*md)
	m.StartTime = startTime.UTC().Format(timeFmt)
	m.Definitions[0].StartTime = startTime.UTC().Format(timeFmt)
	m.StopTime = stopTime.UTC().Format(timeFmt)
	m.Definitions[0].StopTime = stopTime.UTC().Format(timeFmt)
}

func restfulRequest(m *messages.MeasurementRequest, k string, au map[string]string, mt *string, rt interface{}) error {
	// marshal the measurement to text for post to the api.
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal the measurement to text: %v\n", err)
	}

	// Make a http POST request with associated headers:
	// Content-Type, Accept. Add the API key as: ?key=<key>
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s?key=%s", au[*mt], k),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return fmt.Errorf("failed to create the http request: %v\n", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", acceptType)

	// fmt.Printf("Request to be sent:\n%v\n", formatRequest(req))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make measurement request:\n%v\n", err)
	}

	// Inbound 'rt' is the json struct to fill with response data, fill it now.
	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	err = json.Unmarshal(c, rt)
	if err != nil {
		return fmt.Errorf("failed to unmarhal json text from the remote api: %v", err)
	}

	return nil
}

func main() {
	flag.Parse()

	// Read api keys and measurement requests.
	apiKey, err := readKey(key)
	if err != nil {
		fmt.Printf("failed to read the api-key file: %v\n", err)
		return
	}

	if *getKey {
		js := messages.ApiKeyRequest{
			Label: fmt.Sprintf("%v", uuid.New().String()),
		}

		jsonBytes, err := json.Marshal(js)
		if err != nil {
			fmt.Printf("Failed to create new key request: %v\n", err)
			return
		}

		// Make a http POST request with associated headers:
		// Content-Type, Accept. Add the API key as: ?key=<key>
		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("%s?key=%s", atlasURLs["key"], apiKey),
			bytes.NewBuffer(jsonBytes),
		)
		if err != nil {
			fmt.Printf("failed to create the http request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Accept", acceptType)

		// fmt.Printf("Request to be sent:\n%v\n", formatRequest(req))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("failed to make measurement request:\n%v\n", err)
			return
		}

		// Inbound 'rt' is the json struct to fill with response data, fill it now.
		c, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("failed to read response body: %v", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("Returned reply: %v\n", string(c))

		rt := &messages.ApiKeyRequest{}
		err = json.Unmarshal(c, rt)
		if err != nil {
			fmt.Printf("failed to unmarhal json text from the remote api: %v", err)
			return
		}

		fmt.Printf("Returned key response: %v\n", *rt)
		return
	}
	// If the measurement type is invalid, stop now.
	if _, ok := atlasURLs[*mType]; !ok {
		fmt.Printf("please select an appropriate measurement type (%v) is invalid.\n", *mType)
		return
	}

	measurement, err := readJson(mDef)
	if err != nil {
		fmt.Printf("failed to unmarshal the JSON measurementRequest: %v\n", err)
		return
	}

	// Locate the probe set to be used in the measurement.
	probes, err := probes.LocateProbes(metro, radius, count, v4, v6)
	if err != nil {
		fmt.Printf("probe gathering failed: %v\n", err)
		return
	}

	var probeIds []string
	for _, p := range probes {
		probeIds = append(probeIds, fmt.Sprintf("%d", p.Id))
	}
	fmt.Printf("For metro: %v found %v probes.\n", *metro, len(probeIds))

	// Clean up the measurement, add duration and probes.
	measurementGroom(measurement, probeIds, metro, measurementDuration)

	var results messages.MeasurementResponse
	err = restfulRequest(measurement, apiKey, atlasURLs, mType, &results)
	if err != nil {
		fmt.Printf("Failed to schedule measurement: %v", err)
		return
	}

	fmt.Println("Measurement scheduled Id:")
	for _, m := range results.Measurements {
		fmt.Printf("\t%s\n", fmt.Sprintf(atlasURLs["results"], m))
	}
}
