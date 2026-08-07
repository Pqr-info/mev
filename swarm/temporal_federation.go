package main

import "time"

type FederationPacket struct {
	NodeID    string
	Epoch     string
	Events    []TemporalEvent
	Stability TemporalStabilityReport
	Timestamp time.Time
}

type TemporalFederationEngine struct {
	tme *TemporalMemoryEngine
	tso *TemporalStabilityOracle
}

func NewTemporalFederationEngine(
	tme *TemporalMemoryEngine,
	tso *TemporalStabilityOracle,
) *TemporalFederationEngine {
	return &TemporalFederationEngine{
		tme: tme,
		tso: tso,
	}
}

func (f *TemporalFederationEngine) Export(epoch, nodeID string) FederationPacket {
	events := f.tme.GetEventsByEpoch(epoch)
	stability := f.tso.Evaluate(epoch)

	return FederationPacket{
		NodeID:    nodeID,
		Epoch:     epoch,
		Events:    events,
		Stability: stability,
		Timestamp: time.Now(),
	}
}

func (f *TemporalFederationEngine) Import(packet FederationPacket) {
	for _, ev := range packet.Events {
		f.tme.AppendEvent(ev)
	}
}

func (f *TemporalFederationEngine) Compare(packet FederationPacket) float64 {
	local := f.tme.GetEventsByEpoch(packet.Epoch)
	remote := packet.Events

	if len(local) == 0 || len(remote) == 0 {
		return 0.0
	}

	n := len(local)
	if len(remote) < n {
		n = len(remote)
	}

	var sum float64
	for i := 0; i < n; i++ {
		delta := local[i].Drift - remote[i].Drift
		if delta < 0 {
			delta = -delta
		}
		sum += delta
	}

	return sum / float64(n)
}
