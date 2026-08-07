package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkloadTask represents a partition of the simulation / backtesting pipeline.
type WorkloadTask struct {
	TaskID      string
	StartEpoch  int
	EndEpoch    int
	Complexity  int // weight indicator for scheduling
}

// WorkloadResult defines the performance output of a distributed task execution.
type WorkloadResult struct {
	TaskID         string
	NodeID         string
	ProcessedCount int
	AnomaliesFound int
	StabilityScore float64
	ExecutionTime  time.Duration
}

// MeshWorkloadDispatcher5D balances tasks across multi-dimensional mesh coordinates.
type MeshWorkloadDispatcher5D struct {
	router *MeshPredictiveRouter5D
	mu     sync.Mutex
}

// NewMeshWorkloadDispatcher5D initializes a distributed workload controller.
func NewMeshWorkloadDispatcher5D(router *MeshPredictiveRouter5D) *MeshWorkloadDispatcher5D {
	return &MeshWorkloadDispatcher5D{
		router: router,
	}
}

// DispatchWorkload partitions and schedules tasks concurrently across all registered active mesh nodes.
func (d *MeshWorkloadDispatcher5D) DispatchWorkload(ctx context.Context, tasks []WorkloadTask) ([]WorkloadResult, error) {
	d.mu.Lock()
	nodes := d.router.Nodes
	d.mu.Unlock()

	activeWorkers := []PredictiveModelNode{}
	for _, n := range nodes {
		if n.IsActive {
			activeWorkers = append(activeWorkers, n)
		}
	}

	if len(activeWorkers) == 0 {
		return nil, fmt.Errorf("cannot dispatch workload: zero active nodes in 5D mesh topology")
	}

	fmt.Printf("[5D-DISPATCHER] Distributing %d tasks across %d active mesh worker nodes...\n", len(tasks), len(activeWorkers))

	var wg sync.WaitGroup
	resultsChan := make(chan WorkloadResult, len(tasks))
	errChan := make(chan error, len(tasks))

	for i, task := range tasks {
		// Assign task to a node using round-robin distribution
		targetNode := activeWorkers[i%len(activeWorkers)]
		wg.Add(1)

		go func(t WorkloadTask, node PredictiveModelNode) {
			defer wg.Done()

			start := time.Now()
			fmt.Printf("[5D-WORKER-%s] Processing Task %s (Epochs: %d-%d, Complexity: %d) at 5D coordinates (%.2f,%.2f,%.2f,%.2f,%.2f)\n",
				node.NodeID, t.TaskID, t.StartEpoch, t.EndEpoch, t.Complexity,
				node.Address.X, node.Address.Y, node.Address.Z, node.Address.T, node.Address.W,
			)

			// Simulate backtest execution workload
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case <-time.After(time.Duration(10*t.Complexity) * time.Millisecond):
				// Successful execution output mapping
				resultsChan <- WorkloadResult{
					TaskID:         t.TaskID,
					NodeID:         node.NodeID,
					ProcessedCount: (t.EndEpoch - t.StartEpoch) * 500,
					AnomaliesFound: t.Complexity / 2,
					StabilityScore: 0.98 - (float64(t.Complexity) * 0.005),
					ExecutionTime:  time.Since(start),
				}
			}
		}(task, targetNode)
	}

	wg.Wait()
	close(resultsChan)
	close(errChan)

	// Check if context canceled
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	results := make([]WorkloadResult, 0, len(tasks))
	for r := range resultsChan {
		results = append(results, r)
	}

	fmt.Printf("[5D-DISPATCHER] Workload completed. Gathered %d execution profiles successfully.\n", len(results))
	return results, nil
}
