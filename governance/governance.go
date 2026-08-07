package governance

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ==========================================
// 1. Core Models and Directives
// ==========================================

type TemporalGovernanceDirective struct {
	Reason             string
	ReplayAllowed      bool
	TimeslipAllowed    bool
	MutationRateCap    float64
	FederationSyncFreq time.Duration
}

// ==========================================
// 2. Constitutional Substrate
// ==========================================

type TemporalConstitution struct {
	ReplayLimits          ReplayLimits
	TimeslipRules         TimeslipRules
	MutationBoundaries    MutationBoundaries
	FederationObligations FederationObligations
	QuorumRules           QuorumRules
	EmergencyPowers       EmergencyPowers
}

type ReplayLimits struct {
	MaxReplayPerEpoch int
	AllowFractured    bool
}

type TimeslipRules struct {
	AllowTimeslips       bool
	MaxTimeslipDeviation float64
}

type MutationBoundaries struct {
	MaxMutationRate  float64
	RequireConsensus bool
}

type FederationObligations struct {
	MinSyncFrequencySeconds int
	DivergenceEscalation    bool
}

type QuorumRules struct {
	RequiredVotes int
	Supermajority bool
}

type EmergencyPowers struct {
	EnableDuringUnstable  bool
	EnableDuringFractured bool
	ClampMutationRate     float64
}

type ConstitutionEngine struct {
	mu           sync.RWMutex
	constitution TemporalConstitution
}

func NewConstitutionEngine(c TemporalConstitution) *ConstitutionEngine {
	return &ConstitutionEngine{
		constitution: c,
	}
}

func (ce *ConstitutionEngine) GetConstitution() TemporalConstitution {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.constitution
}

func (ce *ConstitutionEngine) Interpret(directive TemporalGovernanceDirective) TemporalGovernanceDirective {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	c := ce.constitution
	if !c.ReplayLimits.AllowFractured && directive.Reason == "fractured-epoch" {
		directive.ReplayAllowed = false
	}
	if !c.TimeslipRules.AllowTimeslips {
		directive.TimeslipAllowed = false
	}
	if directive.MutationRateCap > c.MutationBoundaries.MaxMutationRate {
		directive.MutationRateCap = c.MutationBoundaries.MaxMutationRate
	}
	minSync := time.Duration(c.FederationObligations.MinSyncFrequencySeconds) * time.Second
	if directive.FederationSyncFreq < minSync {
		directive.FederationSyncFreq = minSync
	}
	if directive.Reason == "unstable-epoch" && c.EmergencyPowers.EnableDuringUnstable {
		directive.MutationRateCap = c.EmergencyPowers.ClampMutationRate
	}
	if directive.Reason == "fractured-epoch" && c.EmergencyPowers.EnableDuringFractured {
		directive.MutationRateCap = c.EmergencyPowers.ClampMutationRate
	}

	return directive
}

// ==========================================
// 3. Judiciary Substrate
// ==========================================

type TemporalRuling struct {
	RulingID  string
	Epoch     string
	NodeID    string
	Violation string
	Verdict   string
	Directive TemporalGovernanceDirective
	Timestamp time.Time
}

type TemporalJudiciaryEngine struct {
	mu sync.RWMutex
	ce *ConstitutionEngine
}

func NewTemporalJudiciaryEngine(ce *ConstitutionEngine) *TemporalJudiciaryEngine {
	return &TemporalJudiciaryEngine{
		ce: ce,
	}
}

func (je *TemporalJudiciaryEngine) Adjudicate(epoch string, nodeID string, directive TemporalGovernanceDirective) TemporalRuling {
	je.mu.Lock()
	defer je.mu.Unlock()

	c := je.ce.GetConstitution()
	violation := ""
	verdict := "upheld"

	if !c.ReplayLimits.AllowFractured && directive.Reason == "fractured-epoch" && directive.ReplayAllowed {
		violation = "replay-violation"
		directive.ReplayAllowed = false
		verdict = "corrected"
	}
	if !c.TimeslipRules.AllowTimeslips && directive.TimeslipAllowed {
		violation = "timeslip-violation"
		directive.TimeslipAllowed = false
		verdict = "corrected"
	}
	if directive.MutationRateCap > c.MutationBoundaries.MaxMutationRate {
		violation = "mutation-boundary-violation"
		directive.MutationRateCap = c.MutationBoundaries.MaxMutationRate
		verdict = "corrected"
	}
	minSync := time.Duration(c.FederationObligations.MinSyncFrequencySeconds) * time.Second
	if directive.FederationSyncFreq < minSync {
		violation = "federation-obligation-violation"
		directive.FederationSyncFreq = minSync
		verdict = "corrected"
	}

	return TemporalRuling{
		RulingID:  fmt.Sprintf("ruling-%s-%s", epoch, nodeID),
		Epoch:     epoch,
		NodeID:    nodeID,
		Violation: violation,
		Verdict:   verdict,
		Directive: directive,
		Timestamp: time.Now(),
	}
}

// ==========================================
// 4. Legislature and Amendments (Safe Assertions)
// ==========================================

type Amendment struct {
	Field    string
	NewValue interface{}
}

type TemporalLegislature struct {
	mu sync.RWMutex
	ce *ConstitutionEngine
}

func NewTemporalLegislature(ce *ConstitutionEngine) *TemporalLegislature {
	return &TemporalLegislature{
		ce: ce,
	}
}

func (l *TemporalLegislature) ApplyConstitutionAmendment(a Amendment) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.ce.mu.Lock()
	defer l.ce.mu.Unlock()

	switch a.Field {
	case "MaxMutationRate":
		val, ok := a.NewValue.(float64)
		if !ok {
			return errors.New("invalid type for MaxMutationRate: expected float64")
		}
		l.ce.constitution.MutationBoundaries.MaxMutationRate = val
	case "MaxReplayPerEpoch":
		val, ok := a.NewValue.(int)
		if !ok {
			return errors.New("invalid type for MaxReplayPerEpoch: expected int")
		}
		l.ce.constitution.ReplayLimits.MaxReplayPerEpoch = val
	case "AllowFractured":
		val, ok := a.NewValue.(bool)
		if !ok {
			return errors.New("invalid type for AllowFractured: expected bool")
		}
		l.ce.constitution.ReplayLimits.AllowFractured = val
	case "AllowTimeslips":
		val, ok := a.NewValue.(bool)
		if !ok {
			return errors.New("invalid type for AllowTimeslips: expected bool")
		}
		l.ce.constitution.TimeslipRules.AllowTimeslips = val
	default:
		return fmt.Errorf("field %s is non-amendable or unrecognized", a.Field)
	}

	return nil
}

// ==========================================
// 5. Social Contract and Enforcement
// ==========================================

type Contract struct {
	ID        string
	Actor     string
	Epoch     string
	Clauses   []Clause
	SignedAt  time.Time
	Signature string
}

type Clause struct {
	Name     string
	Priority int
	Rule     string
}

type TemporalSocialContractEngine struct {
	mu        sync.RWMutex
	contracts map[string]*Contract
}

func NewTemporalSocialContractEngine() *TemporalSocialContractEngine {
	return &TemporalSocialContractEngine{
		contracts: make(map[string]*Contract),
	}
}

func (tsce *TemporalSocialContractEngine) RegisterContract(c *Contract) {
	tsce.mu.Lock()
	defer tsce.mu.Unlock()
	tsce.contracts[c.ID] = c
}

func (tsce *TemporalSocialContractEngine) GetContract(id string) (*Contract, bool) {
	tsce.mu.RLock()
	defer tsce.mu.RUnlock()
	c, ok := tsce.contracts[id]
	return c, ok
}

// ==========================================
// 6. Deterministic Arbitration
// ==========================================

type ArbitrationEngine struct{}

func NewArbitrationEngine() *ArbitrationEngine {
	return &ArbitrationEngine{}
}

func (ae *ArbitrationEngine) Arbitrate(clauses []Clause) (Clause, []Clause, error) {
	if len(clauses) == 0 {
		return Clause{}, nil, errors.New("no clauses provided for arbitration")
	}

	// Deterministic selection based on priority, epoch time/sorting key stability
	winner := clauses[0]
	var losers []Clause

	for i := 1; i < len(clauses); i++ {
		c := clauses[i]
		if c.Priority > winner.Priority {
			losers = append(losers, winner)
			winner = c
		} else if c.Priority == winner.Priority {
			// Deterministic tie-breaker: sort lexicographically by Clause Name
			if c.Name < winner.Name {
				losers = append(losers, winner)
				winner = c
			} else {
				losers = append(losers, c)
			}
		} else {
			losers = append(losers, c)
		}
	}

	return winner, losers, nil
}

// ==========================================
// 7. Sanctions & Appeals Interlocking (Due Process)
// ==========================================

type Sanction struct {
	SanctionID string
	Actor      string
	Severity   int
	Status     string // "ACTIVE", "PENDING_APPEAL", "RESOLVED"
	CreatedAt  time.Time
}

type Appeal struct {
	AppealID  string
	Actor     string
	Reason    string
	Status    string // "SUBMITTED", "UNDER_REVIEW", "APPROVED", "REJECTED"
	Timestamp time.Time
}

type SanctionsAndAppealsEngine struct {
	mu        sync.RWMutex
	sanctions map[string]*Sanction
	appeals   map[string]*Appeal
}

func NewSanctionsAndAppealsEngine() *SanctionsAndAppealsEngine {
	return &SanctionsAndAppealsEngine{
		sanctions: make(map[string]*Sanction),
		appeals:   make(map[string]*Appeal),
	}
}

func (sae *SanctionsAndAppealsEngine) SubmitAppeal(a *Appeal) {
	sae.mu.Lock()
	defer sae.mu.Unlock()
	sae.appeals[a.Actor] = a

	// Interlock: If there is an active sanction, quarantine it during appeal review
	for _, s := range sae.sanctions {
		if s.Actor == a.Actor && s.Status == "ACTIVE" {
			s.Status = "PENDING_APPEAL"
		}
	}
}

func (sae *SanctionsAndAppealsEngine) GetAppeal(actor string) (*Appeal, bool) {
	sae.mu.RLock()
	defer sae.mu.RUnlock()
	a, ok := sae.appeals[actor]
	return a, ok
}

func (sae *SanctionsAndAppealsEngine) EnforceSanction(s *Sanction) {
	sae.mu.Lock()
	defer sae.mu.Unlock()

	// Check if there is an active appeal already submitted by this actor
	if app, ok := sae.appeals[s.Actor]; ok && (app.Status == "SUBMITTED" || app.Status == "UNDER_REVIEW") {
		s.Status = "PENDING_APPEAL"
	} else {
		s.Status = "ACTIVE"
	}
	sae.sanctions[s.SanctionID] = s
}

func (sae *SanctionsAndAppealsEngine) GetSanction(id string) (*Sanction, bool) {
	sae.mu.RLock()
	defer sae.mu.RUnlock()
	s, ok := sae.sanctions[id]
	return s, ok
}

// ==========================================
// 8. Citizenship and Rights Compliance
// ==========================================

type Citizenship struct {
	NodeID      string
	Authorized  bool
	EpochJoined int
}

type RightsComplianceEngine struct {
	mu          sync.RWMutex
	citizenship map[string]*Citizenship
}

func NewRightsComplianceEngine() *RightsComplianceEngine {
	return &RightsComplianceEngine{
		citizenship: make(map[string]*Citizenship),
	}
}

func (rce *RightsComplianceEngine) RegisterCitizen(c *Citizenship) {
	rce.mu.Lock()
	defer rce.mu.Unlock()
	rce.citizenship[c.NodeID] = c
}

func (rce *RightsComplianceEngine) AssertRights(nodeID string) bool {
	rce.mu.RLock()
	defer rce.mu.RUnlock()
	c, ok := rce.citizenship[nodeID]
	return ok && c.Authorized
}
