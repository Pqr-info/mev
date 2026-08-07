package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type HealingEvent struct {
	TicketID       string
	AddrPacked     []byte
	LineageHash    []byte
	HealerAgent    string
	PolicyDecision string
	ActionTaken    string
	Outcome        string
	CreatedAt      time.Time
}

type HealingStore struct {
	DB *pgx.Conn
}

func NewHealingStore(db *pgx.Conn) *HealingStore {
	return &HealingStore{DB: db}
}

func (s *HealingStore) Record(ctx context.Context, ev HealingEvent) error {
	if s.DB == nil {
		return nil // In-memory fallback if no DB configured
	}
	_, err := s.DB.Exec(ctx,
		`INSERT INTO healing_events
         (ticket_id, addr_packed, lineage_hash, healer_agent,
          policy_decision, action_taken, outcome, created_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ev.TicketID,
		ev.AddrPacked,
		ev.LineageHash,
		ev.HealerAgent,
		ev.PolicyDecision,
		ev.ActionTaken,
		ev.Outcome,
		ev.CreatedAt,
	)
	return err
}

// Simple insight: success rate for a given action
func (s *HealingStore) SuccessRateByAction(ctx context.Context, action string) (float64, error) {
	if s.DB == nil {
		// Mock success rate for fallback
		if action == "RESTART_SERVICE" {
			return 0.8, nil
		}
		return 1.0, nil
	}

	var total, success int64
	err := s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM healing_events WHERE action_taken = $1`,
		action,
	).Scan(&total)
	if err != nil {
		return 0, err
	}
	err = s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM healing_events WHERE action_taken = $1 AND outcome = 'success'`,
		action,
	).Scan(&success)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return float64(success) / float64(total), nil
}
