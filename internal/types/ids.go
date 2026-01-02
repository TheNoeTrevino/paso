package types

// ID type aliases provide semantic meaning and reduce repetitive int conversions.
// These aliases document what each integer represents in the domain model,
// making code more readable and enabling future optimizations without refactoring.

// ProjectID identifies a unique project in the system
type ProjectID int

// ColumnID identifies a unique column within a project
type ColumnID int

// TaskID identifies a unique task within a project
type TaskID int

// LabelID identifies a unique label within a project
type LabelID int

// TypeID identifies a task type (e.g., task, feature, bug)
type TypeID int

// PriorityID identifies a task priority level (e.g., trivial, low, high)
type PriorityID int

// RelationTypeID identifies the type of relationship between tasks (e.g., parent-child, blocking)
type RelationTypeID int

// CommentID identifies a unique comment on a task
type CommentID int

// Constants for common ID values

const (
	// Task type constants
	TaskTypeTask    TypeID = 1
	TaskTypeFeature TypeID = 2
	TaskTypeBug     TypeID = 3

	// Priority constants
	PriorityTrivial  PriorityID = 1
	PriorityLow      PriorityID = 2
	PriorityMedium   PriorityID = 3
	PriorityHigh     PriorityID = 4
	PriorityCritical PriorityID = 5

	// Relation type constants
	RelationTypeParentChild RelationTypeID = 1
	RelationTypeBlocking    RelationTypeID = 2
	RelationTypeRelated     RelationTypeID = 3
)

// ToInt converts type alias back to int for compatibility with legacy code
func (id ProjectID) ToInt() int {
	return int(id)
}

func (id ColumnID) ToInt() int {
	return int(id)
}

func (id TaskID) ToInt() int {
	return int(id)
}

func (id LabelID) ToInt() int {
	return int(id)
}

func (id TypeID) ToInt() int {
	return int(id)
}

func (id PriorityID) ToInt() int {
	return int(id)
}

func (id RelationTypeID) ToInt() int {
	return int(id)
}

func (id CommentID) ToInt() int {
	return int(id)
}

// FromInt creates type aliases from int values
func ProjectIDFromInt(i int) ProjectID {
	return ProjectID(i)
}

func ColumnIDFromInt(i int) ColumnID {
	return ColumnID(i)
}

func TaskIDFromInt(i int) TaskID {
	return TaskID(i)
}

func LabelIDFromInt(i int) LabelID {
	return LabelID(i)
}

func TypeIDFromInt(i int) TypeID {
	return TypeID(i)
}

func PriorityIDFromInt(i int) PriorityID {
	return PriorityID(i)
}

func RelationTypeIDFromInt(i int) RelationTypeID {
	return RelationTypeID(i)
}

func CommentIDFromInt(i int) CommentID {
	return CommentID(i)
}
