package timemachine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SnapshotMeta struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

type WALEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"ts"`
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	OldHash     string    `json:"old_hash"`
	NewHash     string    `json:"new_hash"`
	PayloadHash string    `json:"payload_hash"`
}

type WAL struct {
	StartID string     `json:"start_id"`
	EndID   string     `json:"end_id"`
	Events  []WALEvent `json:"events"`
}

type ModelManifest struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Quant    string `json:"quant"`
	PathHash string `json:"path_hash"`
}

type ServiceManifest struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	ConfigHash string `json:"config_hash"`
	Status     string `json:"status"`
}

type BrainManifest struct {
	PathHash       string `json:"path_hash"`
	TranscriptHash string `json:"transcript_hash"`
}

// StateManifest represents the entire state snapshot of a JetWeb node.
type StateManifest struct {
	Snapshot SnapshotMeta      `json:"snapshot"`
	WAL      WAL               `json:"wal"`
	Models   []ModelManifest   `json:"models"`
	Services []ServiceManifest `json:"services"`
	Brain    BrainManifest     `json:"brain"`
}

// WriteStateManifest saves the given StateManifest to the specified root directory.
func WriteStateManifest(root string, manifest *StateManifest) error {
	path := filepath.Join(root, "manifest.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// ReadStateManifest reads the StateManifest from the specified root directory.
func ReadStateManifest(root string) (*StateManifest, error) {
	path := filepath.Join(root, "manifest.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var manifest StateManifest
	if err := json.Unmarshal(b, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &manifest, nil
}

func (m *StateManifest) StateHash() (string, error) {
	// Canonicalize to JSON
	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]), nil
}

func (m *StateManifest) Equivalent(other *StateManifest) bool {
	if other == nil {
		return false
	}
	h1, err1 := m.StateHash()
	h2, err2 := other.StateHash()
	if err1 != nil || err2 != nil {
		return false
	}
	return h1 == h2
}
