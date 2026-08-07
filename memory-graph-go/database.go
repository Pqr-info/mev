package memorygraph

import (
	"context"
)

type IMemoryDatabase interface {
	InitializeSchema(ctx context.Context) error
	Close(ctx context.Context) error

	StoreMemory(ctx context.Context, memory *Memory) (string, error)
	GetMemory(ctx context.Context, memoryID string, includeRelationships bool) (*Memory, error)
	SearchMemories(ctx context.Context, query SearchQuery) ([]Memory, error)
	SearchMemoriesPaginated(ctx context.Context, query SearchQuery) (*PaginatedResult, error)
	UpdateMemory(ctx context.Context, memory *Memory) (bool, error)
	DeleteMemory(ctx context.Context, memoryID string) (bool, error)

	CreateRelationship(
		ctx context.Context,
		fromMemoryID string,
		toMemoryID string,
		relationshipType RelationshipType,
		properties *RelationshipProperties,
	) (string, error)

	GetRelatedMemories(
		ctx context.Context,
		memoryID string,
		opts *RelatedMemoriesOptions,
	) ([]RelatedMemory, error)

	GetMemoryStatistics(ctx context.Context) (map[string]interface{}, error)
	GetRecentActivity(ctx context.Context, days int, project *string) (map[string]interface{}, error)
}

type RelatedMemoriesOptions struct {
	RelationshipTypes []string
	MaxDepth          int
}

type RelatedMemory struct {
	Memory       Memory
	Relationship Relationship
}
