package client

import (
	"slices"
	"strings"
)

type LineSelector struct {
	Line      string
	Direction DirectionMatcher
}

func (l *LineSelector) Matches(s StopTime) bool {
	return s.Trip.Route.Id == l.Line && l.Direction.Matches(s)
}

func (l *LineSelector) String() string {
	return l.Direction.String() + " " + l.Line
}

type DirectionMatcher interface {
	Matches(StopTime) bool
	String() string
}

var _ DirectionMatcher = &HeadsignMatcher{}

type HeadsignMatcher struct {
	Headsign string
}

func (h *HeadsignMatcher) Matches(s StopTime) bool {
	return strings.EqualFold(h.Headsign, s.Headsign)
}

func (h *HeadsignMatcher) String() string {
	return h.Headsign
}

var _ DirectionMatcher = &StationIdMatcher{}

type StationIdMatcher struct {
	stationId string
	childIds  []string
	name      string
}

func NewStationMatcher(stop *Stop) *StationIdMatcher {
	sm := &StationIdMatcher{
		stationId: stop.Id,
		name:      stop.Name,
		childIds:  []string{},
	}
	for _, child := range stop.ChildStops {
		sm.childIds = append(sm.childIds, child.Id)
	}
	return sm
}

func (sm *StationIdMatcher) Matches(s StopTime) bool {
	if sm.stationId == s.Destination.Id {
		return true
	}
	return slices.Contains(sm.childIds, s.Destination.Id)
}

func (sm *StationIdMatcher) String() string {
	return sm.name + " bound"
}
