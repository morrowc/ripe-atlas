# ripe-atlas

Utilities to enable turning up measurements for ripe-atlas.
For testing this sort of thing:
  o run the program with sensible inputs:
    $ go run makeMeasurement.go -apiKey api-keys \
           -metro FRA  \
           -mType dns -measurement gdns_v6.json -v6

  o collect the output JSON which was sent, and validate that with:
    https://jsonlint.com/

  o copy the validated JSON into /tmp/json

  o curl the request to get the proper/full error message:
    curl -H "Content-type: application/json" \
         -H "Accept: application/json" -X POST \
         -d "$(cat /tmp/json)" \
         "https://atlas.ripe.net:443/api/v2/measurements/dns/?key=<hey>"

  o perhaps also jam the error returned into the json validator for readabiltiy.

NOTE about access:
  o There are api keys needed for:
    - RIPE Atlas requests
    - Geocoding via github.com/kelvins/geocoder
  o The program will download openflights data from github if
    there is not a cached version in /tmp.
