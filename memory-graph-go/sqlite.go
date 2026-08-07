package memorygraph

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type SQLiteDatabase struct {
	db *sql.DB
}

func NewSQLiteDatabase(dsn string) (*SQLiteDatabase, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &SQLiteDatabase{db: db}, nil
}

func (s *SQLiteDatabase) InitializeSchema(ctx context.Context) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		summary TEXT,
		tags TEXT,
		context TEXT,
		importance REAL,
		confidence REAL,
		effectiveness REAL,
		usage_count INTEGER,
		created_at TEXT,
		updated_at TEXT,
		last_accessed TEXT,
		version INTEGER,
		updated_by TEXT
	);

	CREATE TABLE IF NOT EXISTS relationships (
		id TEXT PRIMARY KEY,
		from_memory_id TEXT NOT NULL,
		to_memory_id TEXT NOT NULL,
		type TEXT NOT NULL,
		properties TEXT,
		description TEXT,
		bidirectional BOOLEAN,
		FOREIGN KEY(from_memory_id) REFERENCES memories(id),
		FOREIGN KEY(to_memory_id) REFERENCES memories(id)
	);

	CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
	CREATE INDEX IF NOT EXISTS idx_relationships_from ON relationships(from_memory_id);
	CREATE INDEX IF NOT EXISTS idx_relationships_to ON relationships(to_memory_id);
	`
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *SQLiteDatabase) Close(ctx context.Context) error {
	return s.db.Close()
}

func (s *SQLiteDatabase) StoreMemory(ctx context.Context, m *Memory) (string, error) {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}

	tagsJSON, _ := json.Marshal(m.Tags)
	contextJSON, _ := json.Marshal(m.Context)

	query := `
		INSERT INTO memories (
			id, type, title, content, summary, tags, context, importance, confidence,
			effectiveness, usage_count, created_at, updated_at, last_accessed, version, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		m.ID, m.Type, m.Title, m.Content, m.Summary, string(tagsJSON), string(contextJSON),
		m.Importance, m.Confidence, m.Effectiveness, m.UsageCount, m.CreatedAt, m.UpdatedAt,
		m.LastAccessed, m.Version, m.UpdatedBy,
	)
	if err != nil {
		return "", err
	}

	return m.ID, nil
}

func (s *SQLiteDatabase) GetMemory(ctx context.Context, memoryID string, includeRelationships bool) (*Memory, error) {
	query := `SELECT id, type, title, content, summary, tags, context, importance, confidence, usage_count, created_at, updated_at, version FROM memories WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, memoryID)

	var m Memory
	var tagsStr, contextStr, summaryStr sql.NullString
	
	err := row.Scan(
		&m.ID, &m.Type, &m.Title, &m.Content, &summaryStr, &tagsStr, &contextStr,
		&m.Importance, &m.Confidence, &m.UsageCount, &m.CreatedAt, &m.UpdatedAt, &m.Version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("memory not found: %s", memoryID)
		}
		return nil, err
	}

	if summaryStr.Valid {
		m.Summary = &summaryStr.String
	}
	if tagsStr.Valid {
		json.Unmarshal([]byte(tagsStr.String), &m.Tags)
	}
	if contextStr.Valid {
		var c MemoryContext
		if err := json.Unmarshal([]byte(contextStr.String), &c); err == nil {
			m.Context = &c
		}
	}

	return &m, nil
}

func (s *SQLiteDatabase) CreateRelationship(
	ctx context.Context,
	fromMemoryID string,
	toMemoryID string,
	relationshipType RelationshipType,
	properties *RelationshipProperties,
) (string, error) {
	relID := uuid.NewString()

	propsJSON, _ := json.Marshal(properties)

	query := `
		INSERT INTO relationships (
			id, from_memory_id, to_memory_id, type, properties, bidirectional
		) VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, relID, fromMemoryID, toMemoryID, relationshipType, string(propsJSON), false)
	if err != nil {
		return "", err
	}
	return relID, nil
}

// Stubs for remaining interface methods
func (s *SQLiteDatabase) SearchMemories(ctx context.Context, query SearchQuery) ([]Memory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SQLiteDatabase) SearchMemoriesPaginated(ctx context.Context, query SearchQuery) (*PaginatedResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SQLiteDatabase) UpdateMemory(ctx context.Context, memory *Memory) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *SQLiteDatabase) DeleteMemory(ctx context.Context, memoryID string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *SQLiteDatabase) GetRelatedMemories(ctx context.Context, memoryID string, opts *RelatedMemoriesOptions) ([]RelatedMemory, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SQLiteDatabase) GetMemoryStatistics(ctx context.Context) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SQLiteDatabase) GetRecentActivity(ctx context.Context, days int, project *string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
