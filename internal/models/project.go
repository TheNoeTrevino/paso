package models

import (
	"fmt"
	"strconv"
	"time"
)

// Project represents a container for kanban columns and tasks
// Projects are the top-level organizational unit in Paso
type Project struct {
	ID          int
	Name        string
	Description string
	GitBranch   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *Project) PickerLabel() string {
	return fmt.Sprintf("%s (#%d)", p.Name, p.ID)
}

func (p *Project) PickerValue() string {
	return strconv.Itoa(p.ID)
}
