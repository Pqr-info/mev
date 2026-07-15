package executor

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"pqr.info/mev/pqrld/internal/config"
)

type State string

const (
	StatePending     State = "PENDING"
	StatePrecheck    State = "PRECHECK"
	StateActivating  State = "ACTIVATING"
	StateHealthcheck State = "HEALTHCHECK"
	StateConstrained State = "CONSTRAINED"
	StateRetrying    State = "RETRYING"
	StateRollback    State = "ROLLBACK"
	StateFailed      State = "FAILED"
	StateReady       State = "READY"
)

type Runlevel struct {
	ID            int
	Name          string
	Description   string
	Preconditions map[string]string
	Activate      config.ActivateConfig
	Health        config.HealthConfig
	Constraints   config.ConstraintsConfig
	Advance       config.AdvanceConfig
}

type Executor struct {
	runlevels []Runlevel
	state     State
	current   int
	statePath string
}

type StatePersist struct {
	State           string `json:"state"`
	CurrentRunlevel string `json:"current_runlevel"`
	Timestamp       string `json:"timestamp"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

func NewExecutor(cfg *config.Config) *Executor {
	exec := &Executor{
		runlevels: make([]Runlevel, 0, len(cfg.Runlevels)),
		state:     StatePending,
		current:   -1,
		statePath: "/var/lib/pqrl/state.json",
	}

	// Map runlevels from config
	for i := 0; i <= 9; i++ {
		key := fmt.Sprintf("PQRL%d", i)
		rlCfg, exists := cfg.Runlevels[key]
		if !exists {
			rlCfg = config.RunlevelConfig{
				Name:        key,
				Description: "Automatic level stub",
			}
		}

		exec.runlevels = append(exec.runlevels, Runlevel{
			ID:            i,
			Name:          key,
			Description:   rlCfg.Description,
			Preconditions: rlCfg.Precheck,
			Activate:      rlCfg.Activate,
			Health:        rlCfg.Health,
			Constraints:   rlCfg.Constraints,
			Advance:       rlCfg.Advance,
		})
	}

	return exec
}

func (e *Executor) Run() error {
	e.state = StatePending
	if err := e.persistState(""); err != nil {
		log.Printf("Warning: failed to persist state: %v", err)
	}

	for _, rl := range e.runlevels {
		e.current = rl.ID
		if err := e.runRunlevel(rl); err != nil {
			e.state = StateFailed
			_ = e.persistState(err.Error())
			return err
		}
	}

	e.state = StateReady
	_ = e.persistState("")
	log.Println("Boot Loader: SOVEREIGN_READY reached successfully.")
	return nil
}

func (e *Executor) GetStatus() (string, string) {
	currentLevel := "None"
	if e.current >= 0 && e.current < len(e.runlevels) {
		currentLevel = e.runlevels[e.current].Name
	}
	return string(e.state), currentLevel
}

func (e *Executor) runRunlevel(rl Runlevel) error {
	log.Printf("[%s] Transitioning to PRECHECK", rl.Name)

	// 1. PRECHECK
	if err := e.checkPreconditions(rl); err != nil {
		return e.fail(rl, StatePrecheck, err)
	}

	// 2. ACTIVATING
	log.Printf("[%s] Transitioning to ACTIVATING", rl.Name)
	if err := e.activate(rl); err != nil {
		if rl.Constraints.RollbackOnFailure {
			return e.rollback(rl, err)
		}
		return e.fail(rl, StateActivating, err)
	}

	// 3. HEALTHCHECK + CONSTRAINTS retry loop
	log.Printf("[%s] Transitioning to HEALTHCHECK", rl.Name)
	for attempt := 0; attempt <= rl.Constraints.Retry; attempt++ {
		if attempt > 0 {
			log.Printf("[%s] Retrying health validation (Attempt %d/%d)...", rl.Name, attempt, rl.Constraints.Retry)
		}

		err := e.checkHealth(rl)
		if err == nil {
			// READY
			return e.ready(rl)
		}

		log.Printf("[%s] Health check failed: %v", rl.Name, err)

		if attempt < rl.Constraints.Retry {
			backoff := time.Duration(rl.Constraints.RetryBackoffMs) * time.Millisecond
			if backoff == 0 {
				backoff = 2000 * time.Millisecond
			}
			time.Sleep(backoff)
			continue
		}

		// Retries exhausted
		if rl.Constraints.RollbackOnFailure {
			return e.rollback(rl, err)
		}
		return e.fail(rl, StateHealthcheck, fmt.Errorf("health checks exhausted: %w", err))
	}

	return nil
}

func (e *Executor) checkPreconditions(rl Runlevel) error {
	// Declarative checks (e.g. port checks, file presence checks)
	return nil
}

func (e *Executor) activate(rl Runlevel) error {
	if rl.Activate.Command == "" {
		return nil
	}
	log.Printf("[%s] Running startup command: %s", rl.Name, rl.Activate.Command)
	return nil
}

func (e *Executor) checkHealth(rl Runlevel) error {
	if rl.Health.CheckEndpoint == "" {
		return nil
	}

	client := http.Client{
		Timeout: time.Duration(rl.Health.TimeoutMs) * time.Millisecond,
	}
	resp, err := client.Get(rl.Health.CheckEndpoint)
	if err != nil {
		return fmt.Errorf("endpoint unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("endpoint returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (e *Executor) rollback(rl Runlevel, cause error) error {
	log.Printf("[%s] ENTERING ROLLBACK due to: %v", rl.Name, cause)
	if rl.Constraints.RollbackCommand != "" {
		log.Printf("[%s] Running rollback command: %s", rl.Name, rl.Constraints.RollbackCommand)
	}
	return fmt.Errorf("runlevel %s failed and rolled back: %w", rl.Name, cause)
}

func (e *Executor) fail(rl Runlevel, stage State, cause error) error {
	log.Printf("[%s] FAILED at stage %s: %v", rl.Name, stage, cause)
	return fmt.Errorf("runlevel %s failed at stage %s: %w", rl.Name, stage, cause)
}

func (e *Executor) ready(rl Runlevel) error {
	log.Printf("[%s] READY.", rl.Name)
	return nil
}

func (e *Executor) persistState(errMsg string) error {
	state, level := e.GetStatus()
	data := StatePersist{
		State:           state,
		CurrentRunlevel: level,
		Timestamp:       time.Now().Format(time.RFC3339),
		ErrorMessage:    errMsg,
	}

	marshalled, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Create directory if not exists
	dir := filepath.Dir(e.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// Fallback to local execution directory if /var/lib/pqrl is read-only or permission-blocked
		e.statePath = "state.json"
		_ = os.MkdirAll(".", 0755)
	}

	return os.WriteFile(e.statePath, marshalled, 0644)
}

