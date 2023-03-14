// probes implements routines relevant to management and observation of probes.
package probes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/morrowc/ripe-atlas/messages"
)

var (
	Probes    = map[int32]messages.ProbeMessage{}
	Countries = make(map[string]int)
	prbUrl    = "https://atlas.ripe.net/api/v2/probes/%d/"
)

// GatherProbe queries RIPE for probe specific data, only if the probe has
// not been seen prior. Reads probes to query from chan ic, sends results
// on chan oc.
func GatherProbe(ic chan int32, oc chan messages.ProbeMessage) {
	for pi := range ic {
		// Check the current probe map, if the probe isn't here, go get data
		// about the probe.
		_, ok := Probes[pi]
		if !ok {
			u := fmt.Sprintf(prbUrl, pi)
			response, err := http.Get(u)
			if err != nil {
				fmt.Printf("Failed to query probe id(%d): %v\n", pi, err)
				continue
			}
			defer response.Body.Close()

			c, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("Failed to read probe id request body(%d): %v\n", pi, err)
				continue
			}
			m := messages.ProbeMessage{}
			err = json.Unmarshal(c, &m)
			if err != nil {
				fmt.Printf("Failed to parse the probe message json(%d): %v\n", pi, err)
			}
			_, ok := Countries[m.CountryCode]
			if !ok {
				Countries[m.CountryCode] = 0
			} else {
				Countries[m.CountryCode]++
			}
			oc <- m
		}
	}
}
