package client

import (
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"time"
)

const (
	UPTOWN   string = "Uptown"
	DOWNTOWN string = "Downtown"
)

type Stop struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	StopTimes   []StopTime   `json:"stopTimes"`
	ServiceMaps []ServiceMap `json:"serviceMaps`
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

type StopsResponse struct {
	Stops  []Stop `json:"stops"`
	NextId string `json:"nextId"`
}

type ServiceMap struct {
	Routes []Route `json:"routes"`
}

func (s *Stop) Lines() []string {
	lines := map[string]bool{}
	for _, svcMap := range s.ServiceMaps {
		for _, route := range svcMap.Routes {
			lines[route.Id] = true
		}
	}
	return slices.Sorted(maps.Keys(lines))
}

func (s *SubwayTime) UnmarshalJSON(data []byte) error {
	raw := &struct {
		Time string
	}{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}
	if raw.Time == "" {
		s.Time = time.Time{}
		return nil
	}
	unixTime, err := strconv.ParseInt(raw.Time, 10, 64)
	if err != nil {
		return err
	}
	s.Time = time.Unix(unixTime, 0)
	return nil
}
