package memorygraph

import (
	"time"
)

type MemoryType string

const (
	MemoryTypeTask          MemoryType = "task"
	MemoryTypeCodePattern   MemoryType = "code_pattern"
	MemoryTypeProblem       MemoryType = "problem"
	MemoryTypeSolution      MemoryType = "solution"
	MemoryTypeProject       MemoryType = "project"
	MemoryTypeTechnology    MemoryType = "technology"
	MemoryTypeError         MemoryType = "error"
	MemoryTypeFix           MemoryType = "fix"
	MemoryTypeCommand       MemoryType = "command"
	MemoryTypeFileContext   MemoryType = "file_context"
	MemoryTypeWorkflow      MemoryType = "workflow"
	MemoryTypeGeneral       MemoryType = "general"
	MemoryTypeConversation  MemoryType = "conversation"
)

type RelationshipType string

const (
	RelCauses        RelationshipType = "CAUSES"
	RelTriggers      RelationshipType = "TRIGGERS"
	RelLeadsTo       RelationshipType = "LEADS_TO"
	RelPrevents      RelationshipType = "PREVENTS"
	RelBreaks        RelationshipType = "BREAKS"
	RelSolves        RelationshipType = "SOLVES"
	RelAddresses     RelationshipType = "ADDRESSES"
	RelAlternativeTo RelationshipType = "ALTERNATIVE_TO"
	RelImproves      RelationshipType = "IMPROVES"
	RelReplaces      RelationshipType = "REPLACES"
	RelOccursIn      RelationshipType = "OCCURS_IN"
	RelAppliesTo     RelationshipType = "APPLIES_TO"
	RelWorksWith     RelationshipType = "WORKS_WITH"
	RelRequires      RelationshipType = "REQUIRES"
	RelUsedIn        RelationshipType = "USED_IN"
	RelBuildsOn      RelationshipType = "BUILDS_ON"
	RelContradicts   RelationshipType = "CONTRADICTS"
	RelConfirms      RelationshipType = "CONFIRMS"
	RelGeneralizes   RelationshipType = "GENERALIZES"
	RelSpecializes   RelationshipType = "SPECIALIZES"
	RelSimilarTo     RelationshipType = "SIMILAR_TO"
	RelVariantOf     RelationshipType = "VARIANT_OF"
	RelRelatedTo     RelationshipType = "RELATED_TO"
	RelAnalogyTo     RelationshipType = "ANALOGY_TO"
	RelOppositeOf    RelationshipType = "OPPOSITE_OF"
	RelFollows       RelationshipType = "FOLLOWS"
	RelDependsOn     RelationshipType = "DEPENDS_ON"
	RelEnables       RelationshipType = "ENABLES"
	RelBlocks        RelationshipType = "BLOCKS"
	RelParallelTo    RelationshipType = "PARALLEL_TO"
	RelEffectiveFor  RelationshipType = "EFFECTIVE_FOR"
	RelIneffective   RelationshipType = "INEFFECTIVE_FOR"
	RelPreferredOver RelationshipType = "PREFERRED_OVER"
	RelDeprecatedBy  RelationshipType = "DEPRECATED_BY"
	RelValidatedBy   RelationshipType = "VALIDATED_BY"
)

type MemoryContext struct {
	ProjectPath        *string                `json:"project_path,omitempty"`
	FilesInvolved      []string               `json:"files_involved"`
	Languages          []string               `json:"languages"`
	Frameworks         []string               `json:"frameworks"`
	Technologies       []string               `json:"technologies"`
	GitCommit          *string                `json:"git_commit,omitempty"`
	GitBranch          *string                `json:"git_branch,omitempty"`
	WorkingDirectory   *string                `json:"working_directory,omitempty"`
	Timestamp          string                 `json:"timestamp"`
	SessionID          *string                `json:"session_id,omitempty"`
	UserID             *string                `json:"user_id,omitempty"`
	AdditionalMetadata map[string]interface{} `json:"additional_metadata"`
	TenantID           *string                `json:"tenant_id,omitempty"`
	TeamID             *string                `json:"team_id,omitempty"`
	Visibility         string                 `json:"visibility"`
	CreatedBy          *string                `json:"created_by,omitempty"`
}

type Memory struct {
	ID             string                  `json:"id,omitempty"`
	Type           MemoryType              `json:"type"`
	Title          string                  `json:"title"`
	Content        string                  `json:"content"`
	Summary        *string                 `json:"summary,omitempty"`
	Tags           []string                `json:"tags"`
	Context        *MemoryContext          `json:"context,omitempty"`
	Importance     float64                 `json:"importance"`
	Confidence     float64                 `json:"confidence"`
	Effectiveness  *float64                `json:"effectiveness,omitempty"`
	UsageCount     int                     `json:"usage_count"`
	CreatedAt      string                  `json:"created_at"`
	UpdatedAt      string                  `json:"updated_at"`
	LastAccessed   *string                 `json:"last_accessed,omitempty"`
	Version        int                     `json:"version"`
	UpdatedBy      *string                 `json:"updated_by,omitempty"`
	Relationships  map[string][]string     `json:"relationships,omitempty"`
	MatchInfo      map[string]interface{}  `json:"match_info,omitempty"`
	ContextSummary *string                 `json:"context_summary,omitempty"`
}

type RelationshipProperties struct {
	Strength             float64 `json:"strength"`
	Confidence           float64 `json:"confidence"`
	Context              *string `json:"context,omitempty"`
	EvidenceCount        int     `json:"evidence_count"`
	SuccessRate          *float64`json:"success_rate,omitempty"`
	CreatedAt            string  `json:"created_at"`
	LastValidated        string  `json:"last_validated"`
	ValidationCount      int     `json:"validation_count"`
	CounterEvidenceCount int     `json:"counter_evidence_count"`
	ValidFrom            string  `json:"valid_from"`
	ValidUntil           *string `json:"valid_until,omitempty"`
	RecordedAt           string  `json:"recorded_at"`
	InvalidatedBy        *string `json:"invalidated_by,omitempty"`
}

type Relationship struct {
	ID           string                 `json:"id,omitempty"`
	FromMemoryID string                 `json:"from_memory_id"`
	ToMemoryID   string                 `json:"to_memory_id"`
	Type         RelationshipType       `json:"type"`
	Properties   RelationshipProperties `json:"properties"`
	Description  *string                `json:"description,omitempty"`
	Bidirectional bool                  `json:"bidirectional"`
}

type SearchQuery struct {
	Query               *string  `json:"query,omitempty"`
	Terms               []string `json:"terms"`
	MemoryTypes         []string `json:"memory_types"`
	Tags                []string `json:"tags"`
	ProjectPath         *string  `json:"project_path,omitempty"`
	Languages           []string `json:"languages"`
	Frameworks          []string `json:"frameworks"`
	MinImportance       *float64 `json:"min_importance,omitempty"`
	MinConfidence       *float64 `json:"min_confidence,omitempty"`
	MinEffectiveness    *float64 `json:"min_effectiveness,omitempty"`
	CreatedAfter        *string  `json:"created_after,omitempty"`
	CreatedBefore       *string  `json:"created_before,omitempty"`
	Limit               int      `json:"limit"`
	Offset              int      `json:"offset"`
	IncludeRelationships bool    `json:"include_relationships"`
	SearchTolerance     string   `json:"search_tolerance"`
	MatchMode           string   `json:"match_mode"`
	RelationshipFilter  []string `json:"relationship_filter,omitempty"`
}

type PaginatedResult struct {
	Results    []Memory `json:"results"`
	TotalCount int      `json:"total_count"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	HasMore    bool     `json:"has_more"`
	NextOffset *int     `json:"next_offset,omitempty"`
}

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func NewMemory(memType MemoryType, title, content string) *Memory {
	return &Memory{
		Type:       memType,
		Title:      title,
		Content:    content,
		Tags:       []string{},
		Importance: 0.5,
		Confidence: 0.8,
		UsageCount: 0,
		CreatedAt:  NowISO(),
		UpdatedAt:  NowISO(),
		Version:    1,
	}
}

func NewRelationshipProperties() RelationshipProperties {
	now := NowISO()
	return RelationshipProperties{
		Strength:             0.5,
		Confidence:           0.8,
		EvidenceCount:        1,
		CreatedAt:            now,
		LastValidated:        now,
		ValidationCount:      0,
		CounterEvidenceCount: 0,
		ValidFrom:            now,
		RecordedAt:           now,
	}
}
