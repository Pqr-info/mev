package main

import (
	"errors"
	"sync"
	"time"
)

// TemporalCheckpoint defines a hard constitutional checkpoint in the timeline.
type TemporalCheckpoint struct {
	ID        string
	WALIndex  int64
	Reason    string
	Timestamp time.Time
}

// CheckpointRegistry manages active temporal checkpoints.
type CheckpointRegistry struct {
	mu          sync.RWMutex
	checkpoints map[string]TemporalCheckpoint
	highestWAL  int64
}

// NewCheckpointRegistry creates a new checkpoint registry.
func NewCheckpointRegistry() *CheckpointRegistry {
	return &CheckpointRegistry{
		checkpoints: make(map[string]TemporalCheckpoint),
		highestWAL:  -1,
	}
}

// AddCheckpoint registers a new checkpoint.
func (cr *CheckpointRegistry) AddCheckpoint(cp TemporalCheckpoint) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cp.Timestamp.IsZero() {
		cp.Timestamp = time.Now()
	}

	cr.checkpoints[cp.ID] = cp
	if cp.WALIndex > cr.highestWAL {
		cr.highestWAL = cp.WALIndex
	}
}

// ValidateRollback ensures a rollback does not cross a constitutional checkpoint.
func (cr *CheckpointRegistry) ValidateRollback(targetWALIndex int64) error {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	// If the rollback target is older than our highest registered checkpoint, it's a safety violation.
	if cr.highestWAL != -1 && targetWALIndex < cr.highestWAL {
		return errors.New("SAFETY_VIOLATION: rollback crosses a constitutional checkpoint")
	}

	return nil
}

// GetHighestCheckpoint returns the highest WAL index protected by a checkpoint.
func (cr *CheckpointRegistry) GetHighestCheckpoint() int64 {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.highestWAL
}
