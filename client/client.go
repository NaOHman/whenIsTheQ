package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Client struct {
	addr   string
	system string
}

func NewClient(addr string, system string) *Client {
	return &Client{
		addr:   addr,
		system: system,
	}
}

func normalizeStationName(stationName string) string {
	stationName = strings.ToLower(stationName)
	words := strings.Fields(stationName)
	for i, word := range words {
		switch word {
		case "street":
			word = "st"
		case "avenue", "ave":
			word = "av"
		case "square":
			word = "sq"
		}
		words[i] = word
	}
	return strings.Join(words, " ")
}

func (c *Client) FindStationCode(stationName string) ([]Stop, error) {
	// The NYC system feed uses multiple stations to represent different platforms at a single physical location
	// There are also physically separate stations with the same name. This is annoying but what can you do
	stops, err := c.getPaginatedStops()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(stops))
	for i, stop := range stops {
		// ToLower helps reduce the edit distance on the fuzzy search
		names[i] = strings.ToLower(stop.Name)
	}
	// replace square and street with their abbreviations before doing the fuzzy search
	stationName = normalizeStationName(stationName)
	matches := fuzzy.Find(stationName, names)
	matchingStops := []Stop{}
	for _, s := range stops {
		if slices.Contains(matches, strings.ToLower(s.Name)) {
			matchingStops = append(matchingStops, s)
		}
	}
	return matchingStops, nil
}

func (c *Client) GetNextDeparture(station string, line *LineSelector) (time.Time, error) {
	nextDeparture := time.Time{}
	stop, err := c.GetStop(station)
	if err != nil {
		return nextDeparture, err
	}
	for _, stopTime := range stop.StopTimes {
		if !stopTime.Future || !line.Matches(stopTime) {
			continue
		}
		if nextDeparture.IsZero() || nextDeparture.After(stopTime.Arrival.Time) {
			nextDeparture = stopTime.Arrival.Time
		}
	}
	if nextDeparture.IsZero() {
		return nextDeparture, fmt.Errorf("could not find any %s trains at station %s", line.String(), stop.Name)
	}
	return nextDeparture, nil
}

func (c *Client) GetStop(stationId string) (*Stop, error) {
	url, err := url.JoinPath(c.addr, "systems", c.system, "stops", stationId)
	if err != nil {
		return nil, err
	}
	stop := &Stop{}
	err = fetchUrl(url, stop)
	return stop, err
}

func fetchUrl(url string, obj any) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error calling %s %d", url, res.StatusCode)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, obj)
}

func (c *Client) getPaginatedStops(firstId ...string) ([]Stop, error) {
	url, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	url = url.JoinPath("systems", c.system, "stops")
	query := url.Query()
	query.Add("filter_by_type", "true")
	query.Add("type", "STATION")
	if len(firstId) == 1 {
		query.Add("first_id", firstId[0])
	}
	url.RawQuery = query.Encode()
	resp := &StopsResponse{}
	if err := fetchUrl(url.String(), resp); err != nil {
		return nil, err
	}
	if resp.NextId != "" {
		next, err := c.getPaginatedStops(resp.NextId)
		if err != nil {
			return nil, err
		}
		return append(resp.Stops, next...), nil
	}
	return resp.Stops, nil
}
