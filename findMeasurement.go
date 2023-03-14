// findMeasurement looks up a measurement or set of measurements by tag.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/morrowc/ripe-atlas/messages"
)

var (
	tags = flag.String("tags", "", "Comma separated list of tags to use in measurement search.")
)

const (
	reqUrl = "https://atlas.ripe.net:443/api/v2/measurements/"
)

// Search will send a request to RIPEAtlas, and parse the reply (if any) to
// return solely the list of measurement-ids.
func search(tags *string) ([]int32, error) {
	u := fmt.Sprintf("%s?tags=%s", reqUrl, *tags)
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("failed to get measurement data from ripe: %v", err)
	}
	defer resp.Body.Close()

	// Discard the first token, which is a [ typically.
	var mids []int32
	var m messages.MultiMeasurementResponseMessage
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bs, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse some json: %v", err.Error())
	}

	for _, mes := range m.Results {
		mids = append(mids, mes.Id)
	}
	return mids, nil

}

func main() {
	flag.Parse()

	if len(*tags) == 0 {
		fmt.Printf("Please provide a list of tags to search.")
		return
	}

	res, err := search(tags)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	for _, id := range res {
		fmt.Printf("%d\n", id)
	}
}
