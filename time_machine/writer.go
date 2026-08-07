package timemachine

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type MicrostructureEvent map[string]any

func formatTimestampForFilename(ts time.Time) string {
    s := ts.UTC().Format(time.RFC3339Nano)
    s = strings.ReplaceAll(s, ":", "-")
    return s
}

func microstructurePath(root string, ts time.Time) string {
    return filepath.Join(
        root,
        "microstructure",
        fmt.Sprintf("%04d", ts.Year()),
        fmt.Sprintf("%02d", ts.Month()),
        fmt.Sprintf("%02d", ts.Day()),
        fmt.Sprintf("%02d", ts.Hour()),
        fmt.Sprintf("%02d", ts.Minute()),
        formatTimestampForFilename(ts)+".json",
    )
}

func WriteMicrostructureEvent(root string, ts time.Time, event MicrostructureEvent) error {
    path := microstructurePath(root, ts)
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return fmt.Errorf("mkdir: %w", err)
    }
    b, err := json.MarshalIndent(event, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal: %w", err)
    }
    if err := os.WriteFile(path, b, 0o644); err != nil {
        return fmt.Errorf("write: %w", err)
    }
    return nil
}

// WriteMicrostructureEventWithTrades wraps WriteMicrostructureEvent to explicitly inject recommended_trades
func WriteMicrostructureEventWithTrades(root string, ts time.Time, event MicrostructureEvent, recommendedTrades []map[string]any) error {
    if recommendedTrades != nil {
        event["recommended_trades"] = recommendedTrades
    }
    return WriteMicrostructureEvent(root, ts, event)
}

// RequestRollback attempts to switch the active temporal branch backwards in time.
// It mandates the Council of Five quorum constraint check before executing.
func RequestRollback(targetBranchID string, quorum QuorumState) error {
    if err := VerifyQuorum(quorum); err != nil {
        return fmt.Errorf("rollback denied: %w", err)
    }

    // Execute rollback
    fmt.Printf("[Time Machine] Rollback authorized to branch %s by Council of Five.\n", targetBranchID)
    // Actually switch branches on the ledger / file system
    return nil
}
