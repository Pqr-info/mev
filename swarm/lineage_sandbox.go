package main

import (
	"errors"
	"sync"
)

// SandboxStatus represents the state of a lineage sandbox.
type SandboxStatus string

const (
	PendingValidation SandboxStatus = "PENDING_VALIDATION"
	Isolated          SandboxStatus = "ISOLATED"
	Approved          SandboxStatus = "APPROVED"
	Rejected          SandboxStatus = "REJECTED"
)

// LineageSandbox represents an isolated timeline state that diverges from the canon.
type LineageSandbox struct {
	SandboxID     string
	BaseHash      string
	DivergentHash string
	Node          string
	Status        SandboxStatus
}

// SandboxManager manages active lineage sandboxes.
type SandboxManager struct {
	mu        sync.RWMutex
	sandboxes map[string]*LineageSandbox
}

// NewSandboxManager creates a new sandbox manager.
func NewSandboxManager() *SandboxManager {
	return &SandboxManager{
		sandboxes: make(map[string]*LineageSandbox),
	}
}

// CreateSandbox creates a new sandbox for validation.
func (sm *SandboxManager) CreateSandbox(sandboxID, baseHash, divergentHash, node string) *LineageSandbox {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sb := &LineageSandbox{
		SandboxID:     sandboxID,
		BaseHash:      baseHash,
		DivergentHash: divergentHash,
		Node:          node,
		Status:        PendingValidation,
	}
	sm.sandboxes[sandboxID] = sb
	return sb
}

// UpdateStatus updates the status of an existing sandbox.
func (sm *SandboxManager) UpdateStatus(sandboxID string, status SandboxStatus) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sb, exists := sm.sandboxes[sandboxID]
	if !exists {
		return errors.New("sandbox not found")
	}

	sb.Status = status
	return nil
}

// GetSandbox returns a sandbox by its ID.
func (sm *SandboxManager) GetSandbox(sandboxID string) (*LineageSandbox, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sb, exists := sm.sandboxes[sandboxID]
	return sb, exists
}
