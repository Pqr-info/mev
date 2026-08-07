package memorygraph

import (
	"context"
	"testing"
)

func TestMemDBMemoryFlow(t *testing.T) {
	ctx := context.Background()

	// Create MemDB database
	db, err := NewMemDBDatabase()
	if err != nil {
		t.Fatalf("failed to open memdb: %v", err)
	}
	defer db.Close(ctx)

	// Initialize schema (no-op for memdb, but good practice)
	err = db.InitializeSchema(ctx)
	if err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	// Create a problem memory
	probMem := NewMemory(MemoryTypeProblem, "Slow Query", "The graph query is too slow.")
	probMem.Tags = []string{"performance"}

	probID, err := db.StoreMemory(ctx, probMem)
	if err != nil {
		t.Fatalf("failed to store problem memory: %v", err)
	}

	// Create a solution memory
	solMem := NewMemory(MemoryTypeSolution, "Use MemDB", "Switched to go-memdb for L1.")
	solMem.Tags = []string{"architecture"}

	solID, err := db.StoreMemory(ctx, solMem)
	if err != nil {
		t.Fatalf("failed to store solution memory: %v", err)
	}

	// Create a relationship (solves)
	props := NewRelationshipProperties()
	props.Context = new(string)
	*props.Context = "Eliminated OS RAMDrive overhead"

	relID, err := db.CreateRelationship(ctx, solID, probID, RelSolves, &props)
	if err != nil {
		t.Fatalf("failed to create relationship: %v", err)
	}

	if relID == "" {
		t.Fatalf("expected non-empty relationship ID")
	}

	// Retrieve memory
	retrieved, err := db.GetMemory(ctx, probID, false)
	if err != nil {
		t.Fatalf("failed to get memory: %v", err)
	}
	if retrieved.Title != "Slow Query" {
		t.Errorf("expected 'Slow Query', got '%s'", retrieved.Title)
	}

	// Retrieve solution to ensure isolation (shallow copy logic)
	retrievedSol, err := db.GetMemory(ctx, solID, false)
	if err != nil {
		t.Fatalf("failed to get solution memory: %v", err)
	}
	
	// Modify retrieved and make sure the DB is untouched
	retrievedSol.Title = "Mutated"
	
	untouchedSol, _ := db.GetMemory(ctx, solID, false)
	if untouchedSol.Title == "Mutated" {
		t.Errorf("expected in-memory DB copy to be isolated from mutations")
	}
}
