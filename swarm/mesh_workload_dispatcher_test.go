package main

import (
	"context"
	"testing"
)

func TestMeshWorkloadDispatcher(t *testing.T) {
	replay := &TimeMachineReplay{}
	router := NewMeshPredictiveRouter5D(replay)

	// Register multiple nodes across different 5D positions
	router.RegisterNode(PredictiveModelNode{
		NodeID:   "nuremburg-01",
		Address:  Address5D{X: 1.0, Y: 2.0, Z: 3.0, T: 1.0, W: 0.1},
		IsActive: true,
		Model:    "gemini-3.5-flash",
	})
	router.RegisterNode(PredictiveModelNode{
		NodeID:   "nuremburg-02",
		Address:  Address5D{X: 4.5, Y: 2.0, Z: 3.0, T: 1.0, W: 0.2},
		IsActive: true,
		Model:    "gemini-3.5-flash",
	})
	router.RegisterNode(PredictiveModelNode{
		NodeID:   "frankfurt-backup",
		Address:  Address5D{X: 10.0, Y: 20.0, Z: 5.0, T: 2.0, W: 0.8},
		IsActive: false, // Inactive, should be skipped
		Model:    "gemini-3.5-flash",
	})

	dispatcher := NewMeshWorkloadDispatcher5D(router)

	tasks := []WorkloadTask{
		{TaskID: "task-A", StartEpoch: 100, EndEpoch: 110, Complexity: 3},
		{TaskID: "task-B", StartEpoch: 110, EndEpoch: 120, Complexity: 5},
		{TaskID: "task-C", StartEpoch: 120, EndEpoch: 130, Complexity: 2},
	}

	ctx := context.Background()
	results, err := dispatcher.DispatchWorkload(ctx, tasks)
	if err != nil {
		t.Fatalf("expected clean workload dispatch, got error: %v", err)
	}

	if len(results) != len(tasks) {
		t.Errorf("expected %d results, got %d", len(tasks), len(results))
	}

	// Verify inactive nodes are skipped and workloads are distributed among active ones
	for _, res := range results {
		if res.NodeID == "frankfurt-backup" {
			t.Errorf("task routed to inactive node frankfurt-backup")
		}
	}
}
