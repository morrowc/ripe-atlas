// Implements simple icmp probe data gathering.
// Follows the ripe atlas API at:
//   https://atlas.ripe.net/docs/api/v2/manual/
//
// Example measurements to test with:
//  o 3679868 - ping measurement
//  o 7829997 - http measurement
//  o 7837568 - http measurement
//  o 1666834 - traceroute measurement
//
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/morrowc/ripe-atlas/messages"
	"github.com/morrowc/ripe-atlas/probes"
)

var (
	mid     = flag.Int("mid", 0, "Measurement ID to get results for.")
	msmtUrl = "https://atlas.ripe.net/api/v2/measurements/%d/"
)

func prettyPrint(ips []string) string {
	var res []string

	for _, v := range ips {
		res = append(res, v)
	}
	return strings.Join(res, "\n\t")
}

// readStream reads a json stream, report stats as reading continues.
func readStream(r *bufio.Reader, ch chan messages.MeasurementResultMessage, ic chan int32) {
	c := json.NewDecoder(r)
	// Read the first open bracket, don't bother to save the bracket.
	_, err := c.Token()
	if err != nil {
		log.Fatalf("reading the initial token failed: %v", err)
	}

	// Now, read each record and decode that record.
	for c.More() {
		// Create a new record in each iteration.
		var m messages.MeasurementResultMessage
		// decode the record. if EOF, exit and close the channel.
		if err := c.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("*** failed decoding a record: %v", err)
		}
		// Send the measurement data down the channel.
		ic <- m.PrbId
		ch <- m
	}
	close(ch)
}

func main() {
	// Parse flags, to get command-line requested information.
	flag.Parse()

	resurl := fmt.Sprintf(msmtUrl, *mid)
	fmt.Printf("Url: %s\n", resurl)

	// Make a request that gets the measurement details..
	response, err := http.Get(resurl)
	if err != nil {
		fmt.Printf("Failed to Get url: %v Status: %v", resurl, response.Status)
	}
	defer response.Body.Close()

	// Read all of the details (this is minimal data volume, kbytes)
	c, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Failed to read contents.\n")
	}

	var m messages.MeasurementResponseMessage
	// Read a stream of json, decode as the reading continues.
	err = json.Unmarshal(c, &m)
	if err != nil {
		fmt.Printf("failed to unmarshal the response to struct: %v\n", err)
	}

	// print just the resolved_ips
	fmt.Printf("IP addresses polled in this measurement(%d):\n\t%s\n",
		*mid, prettyPrint(m.ResolvedIps))
	fmt.Printf("Results are at:\n%v\n", m.Result)

	// Build infrastructure to gather all result data next, in a pipelined manner.
	//
	// Make channels for other routines to use, message pipelines.
	// ch - measurement results.
	// ic - probe-id channel
	// oc - probe results channel
	ch := make(chan messages.MeasurementResultMessage, 100)
	ic := make(chan int32, 10)
	oc := make(chan messages.ProbeMessage, 10)

	// Request results, send successful results body to the reader.
	response, err = http.Get(m.Result)
	if err != nil {
		log.Fatalf("failed to read the results(%v): %v\n", m.Result, err)
	}
	defer response.Body.Close()

	// Start a routine to query for probe data.
	go probes.GatherProbe(ic, oc)

	// Start the reader routine, read from the channel c records returned.
	go readStream(bufio.NewReader(response.Body), ch, ic)

	// Read the probe stream, add items to the probe map.
	go func() {
		for m := range oc {
			probes.Probes[m.Id] = m
		}
	}()

	// Read the measurementresultsmessage channel, report results.
	var rttNum []float64
	for rec := range ch {
		switch rec.Type {
		case "http":
			fmt.Printf("Http result - Rt: %v\n", rec.Result.Rt)
		case "ping":
			fmt.Printf("Ping result - Rtt: %v\n", rec.Result.Rtt)
		case "traceroute":
			fmt.Printf("Traceroute result - Rtt: %v\n", rec.Result.Rtt)
		case "dns":
			fmt.Printf("Traceroute result - Rtt: %v\n", rec.Result.Rt)
			rttNum = append(rttNum, rec.Result.Rt)
		default:
			fmt.Printf("No idea what type: %v\n", rec.Type)
		}
		fmt.Printf("Max num of probes so far: %d\n", len(probes.Probes))
		fmt.Printf("Num countries for probes: %d\n", len(probes.Countries))

		var total float64 = 0
		for _, val := range rttNum {
			total += val
		}
		fmt.Printf("Avg rtt: %0.5f\n", total/float64(len(rttNum)))
	}
}
