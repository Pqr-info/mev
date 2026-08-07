package memorygraph

import (
	"context"
	"testing"
)

func TestSQLiteMemoryFlow(t *testing.T) {
	ctx := context.Background()
	
	// Create an in-memory database for testing
	db, err := NewSQLiteDatabase("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close(ctx)

	// Initialize schema
	err = db.InitializeSchema(ctx)
	if err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	// Create a problem memory
	probMem := NewMemory(MemoryTypeProblem, "Discovery Friction", "Could not discover repo name via CLI.")
	probMem.Tags = []string{"discovery", "github"}
	
	probID, err := db.StoreMemory(ctx, probMem)
	if err != nil {
		t.Fatalf("failed to store problem memory: %v", err)
	}
	
	// Create a solution memory
	solMem := NewMemory(MemoryTypeSolution, "Query CLI", "Use gh repo list.")
	solMem.Tags = []string{"workaround"}
	
	solID, err := db.StoreMemory(ctx, solMem)
	if err != nil {
		t.Fatalf("failed to store solution memory: %v", err)
	}

	// Create a relationship (solves)
	props := NewRelationshipProperties()
	props.Context = new(string)
	*props.Context = "Solved via terminal query"
	
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
	if retrieved.Title != "Discovery Friction" {
		t.Errorf("expected 'Discovery Friction', got '%s'", retrieved.Title)
	}
}
