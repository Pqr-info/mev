package governance

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"pqr.info/jetweb-core/core/go/pikr"
)

func TestSwarmSupervisorSelfHealing(t *testing.T) {
	sup := NewSwarmSupervisor()

	psi := make([]byte, 32)
	sfi := make([]byte, 16)
	tsi := make([]byte, 16)
	qii := make([]byte, 32)
	qri := make([]byte, 8)

	id, _ := pikr.NewIdentity5(psi, sfi, tsi, qii, qri)

	spec := AgentSpec{
		Name:       "TestAgent-1",
		Role:       "Sentinel",
		Lineage:    "corridor.east",
		MaxRetries: 3,
	}

	err := sup.RegisterAgent(spec, id)
	if err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sup.StartSwarm(ctx)

	status, errorsCount, err := sup.GetAgentStatus("TestAgent-1")
	if err != nil || status != StatusRunning {
		t.Errorf("expected RUNNING status, got %s (errors: %v)", status, err)
	}

	// Trigger 1st crash
	sup.SimulateAgentFailure("TestAgent-1", errors.New("oom"))

	status, _, _ = sup.GetAgentStatus("TestAgent-1")
	if status != StatusError {
		t.Errorf("expected ERROR status immediately after crash, got %s", status)
	}

	// Wait for supervisor to restart it
	time.Sleep(1200 * time.Millisecond)

	status, errorsCount, _ = sup.GetAgentStatus("TestAgent-1")
	if status != StatusRunning {
		t.Errorf("expected self-healing to restart agent to RUNNING, got %s", status)
	}
	if errorsCount != 1 {
		t.Errorf("expected 1 error count registered, got %d", errorsCount)
	}

	// Trigger 2 more crashes to reach max retries
	sup.SimulateAgentFailure("TestAgent-1", errors.New("panic"))
	time.Sleep(1200 * time.Millisecond)
	sup.SimulateAgentFailure("TestAgent-1", errors.New("timeout"))
	time.Sleep(1200 * time.Millisecond)

	// Fourth crash (beyond limit)
	sup.SimulateAgentFailure("TestAgent-1", errors.New("network"))
	time.Sleep(1200 * time.Millisecond)

	status, errorsCount, _ = sup.GetAgentStatus("TestAgent-1")
	if status != StatusError {
		t.Errorf("expected agent to remain in ERROR status after exceeding max retries, got %s", status)
	}
	if errorsCount != 4 {
		t.Errorf("expected 4 error counts, got %d", errorsCount)
	}

	sup.StopSwarm()
}

func TestSwarmSupervisorPolicyEnforcement(t *testing.T) {
	sup := NewSwarmSupervisor()

	psi := make([]byte, 32)
	sfi := make([]byte, 16)
	tsi := make([]byte, 16)
	qii := make([]byte, 32)
	qri := make([]byte, 8)

	id, _ := pikr.NewIdentity5(psi, sfi, tsi, qii, qri)

	// Unauthorized lineage for Sentinel role
	specBad := AgentSpec{
		Name:       "BadAgent-1",
		Role:       "Sentinel",
		Lineage:    "unauthorized.north.lineage",
		MaxRetries: 3,
	}

	err := sup.RegisterAgent(specBad, id)
	if err == nil {
		t.Error("expected supervisor registration to fail for unauthorized lineage")
	}

	// Authorized lineage for Sentinel role
	specGood := AgentSpec{
		Name:       "GoodAgent-1",
		Role:       "Sentinel",
		Lineage:    "corridor.east",
		MaxRetries: 3,
	}

	err = sup.RegisterAgent(specGood, id)
	if err != nil {
		t.Errorf("unexpected error for authorized lineage: %v", err)
	}
}

func TestSwarmSupervisorPolicyHotReload(t *testing.T) {
	sup := NewSwarmSupervisor()

	psi := make([]byte, 32)
	sfi := make([]byte, 16)
	tsi := make([]byte, 16)
	qii := make([]byte, 32)
	qri := make([]byte, 8)

	id, _ := pikr.NewIdentity5(psi, sfi, tsi, qii, qri)

	// Unauthorized lineage Initially
	spec := AgentSpec{
		Name:       "HotReloadAgent-1",
		Role:       "Sentinel",
		Lineage:    "corridor.new_horizon",
		MaxRetries: 3,
	}

	err := sup.RegisterAgent(spec, id)
	if err == nil {
		t.Error("expected registration to fail initially")
	}

	// Create temp policy config
	tmpFile, err := os.CreateTemp("", "policy_*.json")
	if err != nil {
		t.Fatalf("failed to create temp policy config: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write initial policy
	initialPolicy := `{"lineage_rules": {"Sentinel": ["corridor.east"]}}`
	tmpFile.WriteString(initialPolicy)
	tmpFile.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.New(os.Stdout, "[TestWatcher] ", log.LstdFlags)
	go StartPolicyWatcher(ctx, tmpFile.Name(), sup.GetPolicyEngine(), logger)

	// Update policy to allow corridor.new_horizon
	updatedPolicy := `{"lineage_rules": {"Sentinel": ["corridor.east", "corridor.new_horizon"]}}`
	err = os.WriteFile(tmpFile.Name(), []byte(updatedPolicy), 0644)
	if err != nil {
		t.Fatalf("failed to update policy config: %v", err)
	}

	// Give the watcher a moment to pick up the file update
	time.Sleep(1200 * time.Millisecond)

	// Register should now succeed
	err = sup.RegisterAgent(spec, id)
	if err != nil {
		t.Errorf("expected registration to succeed after policy hot-reload, got: %v", err)
	}
}


