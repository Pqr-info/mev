package main

import "time"

type TemporalMemoryEngine struct {
	events       []TemporalEvent
	byEpoch      map[string][]TemporalEvent
	byAnchor     map[string][]TemporalEvent
	byTimeBucket map[int64][]TemporalEvent
}

func NewTemporalMemoryEngine() *TemporalMemoryEngine {
	return &TemporalMemoryEngine{
		events:       make([]TemporalEvent, 0),
		byEpoch:      make(map[string][]TemporalEvent),
		byAnchor:     make(map[string][]TemporalEvent),
		byTimeBucket: make(map[int64][]TemporalEvent),
	}
}

func (t *TemporalMemoryEngine) AppendEvent(ev TemporalEvent) {
	t.events = append(t.events, ev)

	// epoch bucket
	t.byEpoch[ev.Epoch] = append(t.byEpoch[ev.Epoch], ev)

	// anchor bucket
	t.byAnchor[ev.Anchor] = append(t.byAnchor[ev.Anchor], ev)

	// time bucket (minute-level)
	bucket := ev.Timestamp.Unix() / 60
	t.byTimeBucket[bucket] = append(t.byTimeBucket[bucket], ev)
}

func (t *TemporalMemoryEngine) GetRecentEvents() []TemporalEvent {
	if len(t.events) == 0 {
		return []TemporalEvent{
			{
				Timestamp:  time.Now(),
				SigmaID:    "mock-node",
				Agent:      "mock-agent",
				Risk:       0.1,
				Confidence: 0.9,
			},
		}
	}
	return t.events
}

func (t *TemporalMemoryEngine) GetEventsByEpoch(epoch string) []TemporalEvent {
	return t.byEpoch[epoch]
}

func (t *TemporalMemoryEngine) GetEventsByAnchor(anchor string) []TemporalEvent {
	return t.byAnchor[anchor]
}

func (t *TemporalMemoryEngine) GetEventsByTimeBucket(bucket int64) []TemporalEvent {
	return t.byTimeBucket[bucket]
}

func (t *TemporalMemoryEngine) DetectRecurrence(epoch string) float64 {
	events := t.byEpoch[epoch]
	if len(events) < 2 {
		return 0.0
	}
	return 0.05
}

func (t *TemporalMemoryEngine) BuildReplaySegment(epoch string) []TemporalEvent {
	return t.byEpoch[epoch]
}
