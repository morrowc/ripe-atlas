// Package implements a single lookup and display for data about a single probe.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/morrowc/ripe-atlas/messages"
)

const (
	probeReqURL = "https://atlas.ripe.net:443/api/v2/probes/?id__in=%d"
)

var (
	id = flag.Int("id", 0, "Probe id to lookup.")
)

func makeRequest(id int32) (string, error) {
	url := fmt.Sprintf(probeReqURL, id)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var probes messages.ProbeQueryResults
	err = json.Unmarshal(c, &probes)
	if err != nil {
		return "", err
	}

	var results []string
	for _, p := range probes.Results {
		results = append(results, fmt.Sprintf(
			"Id: %d  AddrV4: %v AddrV6: %v Country: %v Public: %v",
			p.Id, p.AddressV4, p.AddressV6, p.CountryCode, p.IsPublic))
	}
	return strings.Join(results, "\n"), nil
}

func main() {
	flag.Parse()

	if *id == 0 {
		fmt.Printf("Provide a probe ID to lookup.")
		return
	}

	res, err := makeRequest(int32(*id))
	if err != nil {
		fmt.Printf("Failed to lookup probe id(%v): %v\n", *id, err)
		return
	}
	fmt.Printf("Probe matches:\n%v\n", res)
}
