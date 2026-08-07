package main

import (
	"time"
)

type TimeslipSegment struct {
	BaseEpoch    string          // original epoch
	SlipID       string          // unique ID for this timeslip
	Events       []TemporalEvent // transformed events
	StartTime    time.Time
	EndTime      time.Time
	TransformTag string // e.g. "compressed-2x", "shifted+5m"
}

type TimeslipGenerator struct {
	tme *TemporalMemoryEngine
}

func NewTimeslipGenerator(t *TemporalMemoryEngine) *TimeslipGenerator {
	return &TimeslipGenerator{
		tme: t,
	}
}

func (g *TimeslipGenerator) GenerateTimeslip(epoch string, transform string) TimeslipSegment {
	events := g.tme.GetEventsByEpoch(epoch)
	if len(events) == 0 {
		return TimeslipSegment{
			BaseEpoch:    epoch,
			SlipID:       "slip-" + epoch + "-empty",
			Events:       []TemporalEvent{},
			TransformTag: transform,
		}
	}

	transformed := make([]TemporalEvent, len(events))
	copy(transformed, events)

	if transform == "shifted+5m" {
		for i := range transformed {
			transformed[i].Timestamp = transformed[i].Timestamp.Add(5 * time.Minute)
		}
	} else if transform == "compressed-2x" {
		start := events[0].Timestamp
		for i := range transformed {
			delta := transformed[i].Timestamp.Sub(start)
			transformed[i].Timestamp = start.Add(delta / 2)
		}
	}

	start := transformed[0].Timestamp
	end := transformed[len(transformed)-1].Timestamp

	return TimeslipSegment{
		BaseEpoch:    epoch,
		SlipID:       "slip-" + epoch + "-" + transform,
		Events:       transformed,
		StartTime:    start,
		EndTime:      end,
		TransformTag: transform,
	}
}
