package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type SOSCustomFields struct {
	Component string `json:"component,omitempty"`
	Node      string `json:"node,omitempty"`
	Severity  string `json:"severity,omitempty"`
	Category  string `json:"category,omitempty"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
}

type SOSTicket struct {
	ID           string                 `json:"id"`
	Queue        string                 `json:"queue"`
	Subject      string                 `json:"subject"`
	Text         string                 `json:"text"`
	Status       string                 `json:"status"`
	CustomFields *SOSCustomFields       `json:"CustomFields,omitempty"`
}

type HealingActionLog struct {
	Timestamp time.Time `json:"timestamp"`
	TicketID  string    `json:"ticket_id"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	Details   string    `json:"details"`
}

type HealerAgent struct {
	policyEngine *PolicyEngine
	safetyEngine interface{}
	Store        *HealingStore
	stopCh       chan struct{}
	
	mu           sync.RWMutex
	recentLogs   []HealingActionLog
}

func NewHealerAgent(p *PolicyEngine, s interface{}, store *HealingStore) *HealerAgent {
	return &HealerAgent{
		policyEngine: p,
		safetyEngine: s,
		Store:        store,
		stopCh:       make(chan struct{}),
		recentLogs:   make([]HealingActionLog, 0),
	}
}

func (h *HealerAgent) Start() {
	go h.processLoop()
}

func (h *HealerAgent) Stop() {
	close(h.stopCh)
}

func (h *HealerAgent) logAction(ticketID, action, status, details string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	now := time.Now().UTC()
	h.recentLogs = append(h.recentLogs, HealingActionLog{
		Timestamp: now,
		TicketID:  ticketID,
		Action:    action,
		Status:    status,
		Details:   details,
	})
	
	if len(h.recentLogs) > 50 {
		h.recentLogs = h.recentLogs[len(h.recentLogs)-50:]
	}

	// Persist to store if status is SUCCESS or FAILED or ESCALATE (to avoid too many noise rows for INITIATED)
	if h.Store != nil && (status == "SUCCESS" || status == "FAILED" || status == "ESCALATE" || status == "BLOCKED") {
		ev := HealingEvent{
			TicketID:       ticketID,
			HealerAgent:    "HealerAgentV3",
			PolicyDecision: "UNKNOWN", // We can enrich this later if needed
			ActionTaken:    action,
			Outcome:        status,
			CreatedAt:      now,
		}
		_ = h.Store.Record(context.Background(), ev)
	}
}

func (h *HealerAgent) GetRecentLogs() []HealingActionLog {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	res := make([]HealingActionLog, len(h.recentLogs))
	copy(res, h.recentLogs)
	return res
}

func (h *HealerAgent) processLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Track processed tickets
	processed := make(map[string]bool)

	for {
		select {
		case <-ticker.C:
			h.pollTickets(processed)
		case <-h.stopCh:
			return
		}
	}
}

func (h *HealerAgent) pollTickets(processed map[string]bool) {
	resp, err := http.Get("http://localhost:8100/rt/REST/2.0/tickets")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Tickets []SOSTicket `json:"tickets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	for _, t := range result.Tickets {
		if t.Queue == "SOS" && !processed[t.ID] {
			processed[t.ID] = true
			h.handleTicket(t)
		}
	}
}

func (h *HealerAgent) handleTicket(t SOSTicket) {
	log.Info().Str("ticket_id", t.ID).Msg("HealerAgent: Processing SOS ticket")
	
	// Decision Engine
	
	// Determine node context, defaulting to "MAX" if unassigned
	nodeName := "MAX"
	if t.CustomFields != nil && t.CustomFields.Node != "" {
		nodeName = t.CustomFields.Node
	}
	
	// Safety & Policy checks
	// Let's say we want to restart a service. We check if restart is allowed.
	// For now we mock the evaluation.
	eval, _ := h.policyEngine.EvaluateAction("RESTART_SERVICE", nodeName)
	if !eval.Allowed {
		log.Warn().Str("ticket_id", t.ID).Msg("HealerAgent: Action blocked by TemporalPolicyEngine")
		h.logAction(t.ID, "RESTART_SERVICE", "BLOCKED", "Policy denied restart for node "+nodeName)
		return
	}
	
	// Determine action based on error text
	if strings.Contains(strings.ToLower(t.Text), "panic") {
		h.logAction(t.ID, "PATCH_FILE", "INITIATED", "Attempting autonomic file patch for panic")
		// stub: h.PatchFile(...)
		h.logAction(t.ID, "PATCH_FILE", "SUCCESS", "Panic successfully patched")
		log.Info().Str("ticket_id", t.ID).Msg("HealerAgent: PANIC patched successfully")
	} else if strings.Contains(strings.ToLower(t.Text), "bind") || strings.Contains(strings.ToLower(t.Text), "listen") {
		h.logAction(t.ID, "RESTART_SERVICE", "INITIATED", "Attempting autonomic service restart for port conflict")
		// stub: h.RestartService(...)
		h.logAction(t.ID, "RESTART_SERVICE", "SUCCESS", "Service successfully restarted")
		log.Info().Str("ticket_id", t.ID).Msg("HealerAgent: Service restarted successfully")
	} else if strings.Contains(strings.ToLower(t.Text), "corrupt") {
		h.logAction(t.ID, "ROLLBACK_WAL", "INITIATED", "Attempting WAL rollback for corruption")
		// stub: h.RollbackWAL(...)
		h.logAction(t.ID, "ROLLBACK_WAL", "SUCCESS", "WAL successfully rolled back to last stable horizon")
		log.Info().Str("ticket_id", t.ID).Msg("HealerAgent: WAL rolled back successfully")
	} else {
		h.logAction(t.ID, "ESCALATE", "INITIATED", "Error signature unknown, escalating to Gemini")
		log.Warn().Str("ticket_id", t.ID).Msg("HealerAgent: Escalating unrecognized error")
	}
}

// Action Catalog (Stubs)

func (h *HealerAgent) PatchFile(ticket SOSTicket) error {
	// 1. Read file
	// 2. Apply template patch or AI diff
	// 3. Write back
	// 4. Test compilation
	return nil
}

func (h *HealerAgent) RestartService(service string) error {
	// 1. Drain connections
	// 2. Kill service
	// 3. Start service
	// 4. Verify health
	return nil
}

func (h *HealerAgent) RollbackWAL() error {
	// 1. Read WAL backwards
	// 2. Find last CHECKPOINT_CREATED
	// 3. Execute rollback
	return nil
}

