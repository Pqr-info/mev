package main

import (
	"sync"
	"time"
)

// QuarantineRecord tracks a node that has been isolated due to safety violations.
type QuarantineRecord struct {
	Node      string
	Reason    string
	Severity  DriftSeverity
	Timestamp time.Time
}

// QuarantineManager handles node isolation and release logic.
type QuarantineManager struct {
	mu          sync.RWMutex
	quarantines map[string]QuarantineRecord
}

// NewQuarantineManager initializes a new quarantine system.
func NewQuarantineManager() *QuarantineManager {
	return &QuarantineManager{
		quarantines: make(map[string]QuarantineRecord),
	}
}

// QuarantineNode flags a node as quarantined, preventing it from participating in consensus.
func (qm *QuarantineManager) QuarantineNode(node, reason string, severity DriftSeverity) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.quarantines[node] = QuarantineRecord{
		Node:      node,
		Reason:    reason,
		Severity:  severity,
		Timestamp: time.Now(),
	}
}

// ReleaseNode removes a node from quarantine.
func (qm *QuarantineManager) ReleaseNode(node string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	delete(qm.quarantines, node)
}

// IsQuarantined checks if a node is currently under quarantine.
func (qm *QuarantineManager) IsQuarantined(node string) bool {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	_, exists := qm.quarantines[node]
	return exists
}

// GetQuarantineRecord retrieves the quarantine details for a node, if any.
func (qm *QuarantineManager) GetQuarantineRecord(node string) (QuarantineRecord, bool) {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	record, exists := qm.quarantines[node]
	return record, exists
}

// GetAllQuarantinedNodes returns a list of all currently quarantined nodes.
func (qm *QuarantineManager) GetAllQuarantinedNodes() []QuarantineRecord {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	nodes := make([]QuarantineRecord, 0, len(qm.quarantines))
	for _, record := range qm.quarantines {
		nodes = append(nodes, record)
	}
	return nodes
}
