package models

// RelationType constants for task relationships
const (
	RelationTypeParentChild = 1
	RelationTypeBlocking    = 2
	RelationTypeRelated     = 3
)

// TaskType constants
const (
	TaskTypeTask    = 1
	TaskTypeFeature = 2
	TaskTypeBug     = 3
)

// Priority constants
const (
	PriorityTrivial  = 1
	PriorityLow      = 2
	PriorityMedium   = 3
	PriorityHigh     = 4
	PriorityCritical = 5
)

// DefaultTaskPosition is the default position for new tasks (appended at the end)
const DefaultTaskPosition = 9999

// DefaultRelationTypeID is the default relation type ID (Parent/Child relationship)
const DefaultRelationTypeID = RelationTypeParentChild

// MaxRelationTypes is the maximum number of relation types
const MaxRelationTypes = 3
