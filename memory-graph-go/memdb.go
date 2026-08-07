package memorygraph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	memdb "github.com/hashicorp/go-memdb"
)

type MemDBDatabase struct {
	db *memdb.MemDB
}

func NewMemDBDatabase() (*MemDBDatabase, error) {
	// Create the DB schema
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"memory": {
				Name: "memory",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"type": {
						Name:    "type",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Type"},
					},
				},
			},
			"relationship": {
				Name: "relationship",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"from_memory_id": {
						Name:    "from_memory_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "FromMemoryID"},
					},
					"to_memory_id": {
						Name:    "to_memory_id",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "ToMemoryID"},
					},
				},
			},
		},
	}

	// Create database
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}

	return &MemDBDatabase{db: db}, nil
}

func (s *MemDBDatabase) InitializeSchema(ctx context.Context) error {
	// MemDB schema is initialized during creation, nothing to do here.
	return nil
}

func (s *MemDBDatabase) Close(ctx context.Context) error {
	// In-memory DB doesn't strictly need closing, but we can nil it out
	s.db = nil
	return nil
}

func (s *MemDBDatabase) StoreMemory(ctx context.Context, m *Memory) (string, error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}

	// Create write transaction
	txn := s.db.Txn(true)
	defer txn.Abort() // Aborts if not committed

	// Insert
	if err := txn.Insert("memory", m); err != nil {
		return "", err
	}

	txn.Commit()
	return m.ID, nil
}

func (s *MemDBDatabase) GetMemory(ctx context.Context, memoryID string, includeRelationships bool) (*Memory, error) {
	// Create read transaction
	txn := s.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("memory", "id", memoryID)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("memory not found: %s", memoryID)
	}

	m, ok := raw.(*Memory)
	if !ok {
		return nil, fmt.Errorf("internal type error")
	}

	// Make a shallow copy to prevent the caller from mutating the in-memory cache directly
	mCopy := *m
	return &mCopy, nil
}

func (s *MemDBDatabase) CreateRelationship(
	ctx context.Context,
	fromMemoryID string,
	toMemoryID string,
	relationshipType RelationshipType,
	properties *RelationshipProperties,
) (string, error) {
	relID := uuid.NewString()

	var props RelationshipProperties
	if properties != nil {
		props = *properties
	}

	rel := &Relationship{
		ID:           relID,
		FromMemoryID: fromMemoryID,
		ToMemoryID:   toMemoryID,
		Type:         relationshipType,
		Properties:   props,
	}

	txn := s.db.Txn(true)
	defer txn.Abort()

	if err := txn.Insert("relationship", rel); err != nil {
		return "", err
	}

	txn.Commit()
	return relID, nil
}

// Stubs for remaining interface methods
func (s *MemDBDatabase) SearchMemories(ctx context.Context, query SearchQuery) ([]Memory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MemDBDatabase) SearchMemoriesPaginated(ctx context.Context, query SearchQuery) (*PaginatedResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MemDBDatabase) UpdateMemory(ctx context.Context, memory *Memory) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *MemDBDatabase) DeleteMemory(ctx context.Context, memoryID string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *MemDBDatabase) GetRelatedMemories(ctx context.Context, memoryID string, opts *RelatedMemoriesOptions) ([]RelatedMemory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MemDBDatabase) GetMemoryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *MemDBDatabase) GetRecentActivity(ctx context.Context, days int, project *string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
