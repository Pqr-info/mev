package main

import "time"

type EnforcementPacket struct {
	Domain      string         `json:"domain"`
	Actor       string         `json:"actor"`
	Allowed     bool           `json:"allowed"`
	Reason      string         `json:"reason"`
	ContractID  string         `json:"contract_id"`
	ClauseID    string         `json:"clause_id"`
	Priority    int            `json:"priority"`
	EvaluatedAt time.Time      `json:"evaluated_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	Context     map[string]any `json:"context"`
}

type Sink interface {
	Apply(packet EnforcementPacket) error
}

type Relay interface {
	Emit(packet EnforcementPacket) error
	Broadcast(packet EnforcementPacket) error
	AttachSink(name string, sink Sink) error
}

type relay struct {
	sinks map[string]Sink
}

func NewRelay() Relay {
	return &relay{
		sinks: make(map[string]Sink),
	}
}

func (r *relay) AttachSink(name string, sink Sink) error {
	r.sinks[name] = sink
	return nil
}

func (r *relay) Emit(packet EnforcementPacket) error {
	for _, s := range r.sinks {
		if err := s.Apply(packet); err != nil {
			// log error, continue
		}
	}
	return nil
}

func (r *relay) Broadcast(packet EnforcementPacket) error {
	return r.Emit(packet)
}

type TERBridge struct {
	engine *TemporalSocialContractEngine
	relay  Relay
}

func NewTERBridge(engine *TemporalSocialContractEngine, r Relay) *TERBridge {
	return &TERBridge{engine: engine, relay: r}
}

func (b *TERBridge) EvaluateAndEnforce(
	domain string,
	actor string,
	at time.Time,
	ctx map[string]any,
) error {
	// Simple lookup for demonstration
	allowed := false
	reason := "no contract found"
	contractID := ""
	clauseID := ""
	priority := 0

	for _, c := range b.engine.contracts {
		if c.State == StateActive {
			for _, cl := range c.Clauses {
				if string(cl.Domain) == domain {
					allowed = true
					reason = "active contract match"
					contractID = string(c.ID)
					clauseID = string(cl.ID)
					priority = cl.Priority
					break
				}
			}
		}
	}

	packet := EnforcementPacket{
		Domain:      domain,
		Actor:       actor,
		Allowed:     allowed,
		Reason:      reason,
		ContractID:  contractID,
		ClauseID:    clauseID,
		Priority:    priority,
		EvaluatedAt: at,
		ExpiresAt:   at.Add(5 * time.Second),
		Context:     ctx,
	}

	return b.relay.Broadcast(packet)
}
