/*
Using https://anti-fish.bitflow.dev/ API and its /check endpoint.
This service functions without restriction, query just needs to be regexp-filtered to avoid unnecessary requests.
*/

package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type AntiFish struct {
	Match   bool `json:"match"`
	Matches []struct {
		Followed    bool    `json:"followed"`
		Domain      string  `json:"domain"`
		URL         string  `json:"url"`
		Source      string  `json:"source"`
		Type        string  `json:"type"`
		TrustRating float64 `json:"trust_rating"`
	} `json:"matches"`
}

var (
	antiFishScheme   = "https"
	antiFishHost     = "anti-fish.bitflow.dev"
	antiFishURL      = fmt.Sprintf("%s://%s/", antiFishScheme, antiFishHost)
	antiFishHostPath = "check"
)

func AntiFishQuery(phisingQuery string) (*AntiFish, error) {
	antiFish := AntiFish{}
	phisingQuery = strings.Replace(phisingQuery, "\n", " ", -1)
	queryString := fmt.Sprintf(`{"message":"%s"}`, phisingQuery)

	u := &url.URL{
		Scheme: antiFishScheme,
		Host:   antiFishHost,
		Path:   antiFishHostPath,
	}

	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(queryString)) // URL-encoded payload
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "*/*")
	r.Header.Add("Content-Length", strconv.Itoa(len(queryString)))
	r.Header.Add("User-Agent", "Mozilla-PAGST1.12")

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		antiFish.Match = false
		return &antiFish, nil

	} else if resp.StatusCode != 200 {
		respError := fmt.Errorf("Unable to fetch data from that site, status-code %d", resp.StatusCode)
		return nil, respError
	}
	r.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stringBody := json.Unmarshal(bytes, &antiFish)
	if stringBody != nil {
		return nil, stringBody
	}

	return &antiFish, nil
}
