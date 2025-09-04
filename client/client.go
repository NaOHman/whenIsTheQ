package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Stop struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	StopTimes []StopTime `json:"stopTimes"`
}

type StopTime struct {
	Arrival   SubwayTime `json:"arrival"`
	Departure SubwayTime `json:"departure"`
	Future    bool       `json:"future"`
	Headsign  string     `json:"headsign"`
	Trip      Trip       `json:"trip"`
}

type Trip struct {
	Id    string `json:"id"`
	Route Route  `json:"route"`
}

type SubwayTime struct {
	Time time.Time
}

type Route struct {
	Id    string `json:"id"`
	Color string `json:"color"`
}

func (s *SubwayTime) UnmarshalJSON(data []byte) error {
	raw := &struct {
		Time string
	}{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}
	unixTime, err := strconv.ParseInt(raw.Time, 10, 64)
	if err != nil {
		return err
	}
	s.Time = time.Unix(unixTime, 0)
	return nil
}

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

func (c *Client) GetNextDeparture(station, line, direction string) (time.Time, error) {
	nextDeparture := time.Time{}
	url, err := url.JoinPath(c.addr, "systems", c.system, "stops", station)
	if err != nil {
		return nextDeparture, err
	}
	res, err := http.Get(url)
	if err != nil {
		return nextDeparture, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nextDeparture, fmt.Errorf("error calling %s %d", url, res.StatusCode)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nextDeparture, err
	}
	stop := &Stop{}
	if err = json.Unmarshal(bytes, stop); err != nil {
		return nextDeparture, err
	}
	for _, stopTime := range stop.StopTimes {
		if !stopTime.Future {
			continue
		}
		if stopTime.Trip.Route.Id != line {
			continue
		}
		if stopTime.Headsign != direction {
			continue
		}
		if nextDeparture.IsZero() || nextDeparture.After(stopTime.Arrival.Time) {
			nextDeparture = stopTime.Arrival.Time
		}
	}
	if nextDeparture.IsZero() {
		return nextDeparture, fmt.Errorf("could not find any %s %s trains at station %s", direction, line, station)
	}
	return nextDeparture, nil
}
