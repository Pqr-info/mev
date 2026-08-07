package main

import (
	"context"
	"time"
)

// ============================================================
// ============ Received Identity Frame (RIF) ==================
// ============================================================

type HarmonicSovereignReceptionReceivedIdentityFrame struct {
	EpochIndex             int64     `json:"epoch_index"`
	FrameHash              string    `json:"frame_hash"`
	ReconstructionFidelity float64   `json:"reconstruction_fidelity"`
	AttenuationRecovered   float64   `json:"attenuation_recovered"`
	LocalCoherenceScore    float64   `json:"local_coherence_score"`
	ReceivedAt             time.Time `json:"received_at"`
}

// ============================================================
// ===================== Reception State =======================
// ============================================================

type HarmonicSovereignReceptionState struct {
	LastFrame           *HarmonicSovereignReceptionReceivedIdentityFrame         `json:"last_frame"`
	LastPacket          *HarmonicSovereignDistributionTemporalDistributionPacket `json:"last_packet"`
	ReceptionDelta      float64                                                  `json:"reception_delta"`
	ReconstructionScore float64                                                  `json:"reconstruction_score"`
	LastReceptionTime   time.Time                                                `json:"last_reception_time"`
}

// ============================================================
// ===================== Reception Metrics =====================
// ============================================================

type HarmonicSovereignReceptionMetrics struct {
	FramesReconstructed    int64     `json:"frames_reconstructed"`
	FidelityEvaluations    int64     `json:"fidelity_evaluations"`
	AttenuationCorrections int64     `json:"attenuation_corrections"`
	LatencyAverageMs       float64   `json:"latency_average_ms"`
	LocalCoherenceAverage  float64   `json:"local_coherence_average"`
	LastAuditTime          time.Time `json:"last_audit_time"`
}

// ============================================================
// ===================== Reception Kernel ======================
// ============================================================

type HarmonicSovereignReceptionKernel interface {
	DecodePacket(ctx context.Context, packet *HarmonicSovereignDistributionTemporalDistributionPacket) (*HarmonicSovereignReceptionReceivedIdentityFrame, error)
	ComputeReceptionDelta(ctx context.Context, packet *HarmonicSovereignDistributionTemporalDistributionPacket) (float64, error)
	ApplyReconstruction(ctx context.Context, frame *HarmonicSovereignReceptionReceivedIdentityFrame) error
}

type DefaultHarmonicSovereignReceptionKernel struct {
	NodeID  string
	State   *HarmonicSovereignReceptionState
	Metrics *HarmonicSovereignReceptionMetrics
}

func NewDefaultHarmonicSovereignReceptionKernel(nodeID string) *DefaultHarmonicSovereignReceptionKernel {
	return &DefaultHarmonicSovereignReceptionKernel{
		NodeID:  nodeID,
		State:   &HarmonicSovereignReceptionState{},
		Metrics: &HarmonicSovereignReceptionMetrics{},
	}
}

func (k *DefaultHarmonicSovereignReceptionKernel) DecodePacket(
	ctx context.Context,
	packet *HarmonicSovereignDistributionTemporalDistributionPacket,
) (*HarmonicSovereignReceptionReceivedIdentityFrame, error) {
	f := &HarmonicSovereignReceptionReceivedIdentityFrame{
		EpochIndex:             packet.EpochIndex,
		FrameHash:              "DECODED_" + packet.PacketHash,
		ReconstructionFidelity: 1.0,
		AttenuationRecovered:   1.0,
		LocalCoherenceScore:    1.0,
		ReceivedAt:             time.Now(),
	}
	k.State.LastFrame = f
	k.State.LastPacket = packet
	k.State.LastReceptionTime = time.Now()
	k.Metrics.FramesReconstructed++
	return f, nil
}

func (k *DefaultHarmonicSovereignReceptionKernel) ComputeReceptionDelta(
	ctx context.Context,
	packet *HarmonicSovereignDistributionTemporalDistributionPacket,
) (float64, error) {
	return 0.0, nil
}

func (k *DefaultHarmonicSovereignReceptionKernel) ApplyReconstruction(
	ctx context.Context,
	frame *HarmonicSovereignReceptionReceivedIdentityFrame,
) error {
	k.Metrics.AttenuationCorrections++
	return nil
}

// ============================================================
// ===================== Reception Loop ========================
// ============================================================

type HarmonicSovereignReceptionLoop interface {
	RunOnce(ctx context.Context) (*HarmonicSovereignReceptionReceivedIdentityFrame, error)
	RunContinuous(ctx context.Context) error
}

type DefaultHarmonicSovereignReceptionLoop struct {
	NodeID  string
	Kernel  HarmonicSovereignReceptionKernel
	Source  HarmonicSovereignDistributionEngine
	State   *HarmonicSovereignReceptionState
	Metrics *HarmonicSovereignReceptionMetrics
}

func NewDefaultHarmonicSovereignReceptionLoop(
	nodeID string,
	kernel HarmonicSovereignReceptionKernel,
	source HarmonicSovereignDistributionEngine,
) *DefaultHarmonicSovereignReceptionLoop {
	return &DefaultHarmonicSovereignReceptionLoop{
		NodeID:  nodeID,
		Kernel:  kernel,
		Source:  source,
		State:   &HarmonicSovereignReceptionState{},
		Metrics: &HarmonicSovereignReceptionMetrics{},
	}
}

func (l *DefaultHarmonicSovereignReceptionLoop) RunOnce(ctx context.Context) (*HarmonicSovereignReceptionReceivedIdentityFrame, error) {
	lastPacket := l.Source.State().LastPacket
	if lastPacket == nil {
		lastPacket = &HarmonicSovereignDistributionTemporalDistributionPacket{
			EpochIndex:              time.Now().Unix(),
			PacketHash:              "MOCK_PACKET_HASH",
			RoutingVector:           "MOCK_ROUTING",
			BroadcastTopology:       "MOCK_TOPOLOGY",
			AttenuationCompensation: 1.0,
			GeneratedAt:             time.Now(),
		}
	}

	f, err := l.Kernel.DecodePacket(ctx, lastPacket)
	if err != nil {
		return nil, err
	}

	err = l.Kernel.ApplyReconstruction(ctx, f)
	if err != nil {
		return nil, err
	}

	l.Metrics.FidelityEvaluations++
	return f, nil
}

func (l *DefaultHarmonicSovereignReceptionLoop) RunContinuous(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, _ = l.RunOnce(ctx)
		}
	}
}

// ============================================================
// ===================== Reception Engine ======================
// ============================================================

type HarmonicSovereignReceptionEngine interface {
	Kernel() HarmonicSovereignReceptionKernel
	Loop() HarmonicSovereignReceptionLoop
	State() *HarmonicSovereignReceptionState
	Metrics() *HarmonicSovereignReceptionMetrics
}

type DefaultHarmonicSovereignReceptionEngine struct {
	KernelEngine HarmonicSovereignReceptionKernel
	LoopEngine   HarmonicSovereignReceptionLoop
	StateData    *HarmonicSovereignReceptionState
	MetricsData  *HarmonicSovereignReceptionMetrics
}

func NewDefaultHarmonicSovereignReceptionEngine(
	nodeID string,
	distribution HarmonicSovereignDistributionEngine,
) *DefaultHarmonicSovereignReceptionEngine {
	kernel := NewDefaultHarmonicSovereignReceptionKernel(nodeID)
	loop := NewDefaultHarmonicSovereignReceptionLoop(nodeID, kernel, distribution)

	return &DefaultHarmonicSovereignReceptionEngine{
		KernelEngine: kernel,
		LoopEngine:   loop,
		StateData:    kernel.State,
		MetricsData:  kernel.Metrics,
	}
}

func (e *DefaultHarmonicSovereignReceptionEngine) Kernel() HarmonicSovereignReceptionKernel {
	return e.KernelEngine
}

func (e *DefaultHarmonicSovereignReceptionEngine) Loop() HarmonicSovereignReceptionLoop {
	return e.LoopEngine
}

func (e *DefaultHarmonicSovereignReceptionEngine) State() *HarmonicSovereignReceptionState {
	return e.StateData
}

func (e *DefaultHarmonicSovereignReceptionEngine) Metrics() *HarmonicSovereignReceptionMetrics {
	return e.MetricsData
}

// ============================================================
// ===== Identity Harmonic Sovereign Reception Engine ==========
// ============================================================

type IdentityHarmonicSovereignReceptionEngine interface {
	Reception() HarmonicSovereignReceptionEngine
}

type DefaultIdentityHarmonicSovereignReceptionEngine struct {
	Engine HarmonicSovereignReceptionEngine
}

func NewDefaultIdentityHarmonicSovereignReceptionEngine(
	nodeID string,
	distribution HarmonicSovereignDistributionEngine,
) *DefaultIdentityHarmonicSovereignReceptionEngine {
	return &DefaultIdentityHarmonicSovereignReceptionEngine{
		Engine: NewDefaultHarmonicSovereignReceptionEngine(nodeID, distribution),
	}
}

func (e *DefaultIdentityHarmonicSovereignReceptionEngine) Reception() HarmonicSovereignReceptionEngine {
	return e.Engine
}
