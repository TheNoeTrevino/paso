package styles

// Tree connector characters for visual hierarchy (used by quiet mode)
const (
	TreeBranch     = "├── " // middle child connector
	TreeLastBranch = "└── " // last child connector
	TreeVertical   = "│   " // continuation line for non-last siblings
	TreeSpace      = "    " // continuation for last sibling (4 spaces)
)

// CompletedDimIntensity controls how much completed tasks are dimmed (0.0-1.0)
const CompletedDimIntensity = 0.6
