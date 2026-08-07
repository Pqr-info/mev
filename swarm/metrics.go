package main

import "sync"

type Metrics struct {
	mu        sync.RWMutex
	Inference InferenceMetrics
	Temporal  TemporalMetrics
	Health    HealthMetrics
	Consensus ConsensusMetrics
	Atlas     AtlasMetrics
}

type InferenceMetrics struct {
	TokensTotal       uint64
	RequestsTotal     uint64
	AvgLatencyMs      float64
	GPUUtilization    float64
	KVCacheSaturation float64
}

type TemporalMetrics struct {
	WALMutationsTotal uint64
	SnapshotTotal     uint64
	DriftEventsTotal  uint64
	RollbacksTotal    uint64
	RollforwardsTotal uint64
	PromotionsTotal   uint64
}

type HealthMetrics struct {
	NodeReachabilityFailures uint64
	ManifestFreshnessMs      float64
	StateHashChurnTotal      uint64
}

type ConsensusMetrics struct {
	ProposalsTotal       uint64
	DecisionsTotal       uint64
	AvgConfidence        float64
	ArbitrationTotal     uint64
	AvgDecisionLatencyMs float64
}

type AtlasMetrics struct {
	EventStreamRate    float64
	NodeCardUpdateRate float64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}
