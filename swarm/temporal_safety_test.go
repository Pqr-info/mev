package main

import (
	"testing"
)

func TestCheckpointRegistry(t *testing.T) {
	registry := NewCheckpointRegistry()

	// Initially, rollback should be valid anywhere
	err := registry.ValidateRollback(100)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	// Add a checkpoint at index 200
	cp := TemporalCheckpoint{
		ID:       "cp-1",
		WALIndex: 200,
		Reason:   "Safety Check",
	}
	registry.AddCheckpoint(cp)

	// Rollback to 250 should be valid
	if err := registry.ValidateRollback(250); err != nil {
		t.Fatalf("Expected rollback to 250 to be valid")
	}

	// Rollback to 150 should be invalid (crosses checkpoint at 200)
	if err := registry.ValidateRollback(150); err == nil {
		t.Fatalf("Expected error for rollback to 150, got nil")
	}
}

func TestSandboxManager(t *testing.T) {
	manager := NewSandboxManager()

	sb := manager.CreateSandbox("sb-1", "hashA", "hashB", "nodeX")
	if sb.Status != PendingValidation {
		t.Fatalf("Expected initial status to be PENDING_VALIDATION, got %v", sb.Status)
	}

	err := manager.UpdateStatus("sb-1", Approved)
	if err != nil {
		t.Fatalf("Unexpected error updating status: %v", err)
	}

	retrieved, exists := manager.GetSandbox("sb-1")
	if !exists {
		t.Fatalf("Expected sandbox to exist")
	}
	if retrieved.Status != Approved {
		t.Fatalf("Expected status to be APPROVED, got %v", retrieved.Status)
	}
}

func TestQuarantineManager(t *testing.T) {
	manager := NewQuarantineManager()

	if manager.IsQuarantined("nodeY") {
		t.Fatalf("Expected nodeY to not be quarantined initially")
	}

	manager.QuarantineNode("nodeY", "Excessive Drift", DriftCritical)

	if !manager.IsQuarantined("nodeY") {
		t.Fatalf("Expected nodeY to be quarantined")
	}

	record, exists := manager.GetQuarantineRecord("nodeY")
	if !exists || record.Severity != DriftCritical {
		t.Fatalf("Failed to retrieve correct quarantine record")
	}

	manager.ReleaseNode("nodeY")

	if manager.IsQuarantined("nodeY") {
		t.Fatalf("Expected nodeY to be released")
	}
}
