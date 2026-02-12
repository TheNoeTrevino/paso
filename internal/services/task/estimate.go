package task

import (
	"fmt"
	"regexp"
	"strconv"
)

// EstimateComponent represents a single component of a time estimate
type EstimateComponent struct {
	Quantity int
	Unit     string
}

// ParseEstimate parses a time estimate string like "1w2d3h4m" into its components.
// Supported units:
//   - h: hours
//   - d: days
//   - w: weeks
//   - m: months
//
// Examples:
//   - "1w2d3h" -> [{1, "w"}, {2, "d"}, {3, "h"}]
//   - "2d" -> [{2, "d"}]
//   - "4h" -> [{4, "h"}]
//   - "1w" -> [{1, "w"}]
//   - "3m" -> [{3, "m"}]
//
// Returns an error if:
//   - The format is invalid
//   - The same unit appears more than once
//   - The quantity is not a positive integer
func ParseEstimate(estimate string) ([]EstimateComponent, error) {
	if estimate == "" {
		return nil, ErrInvalidEstimateFormat
	}

	// Regex to match digit sequences followed by unit letters
	re := regexp.MustCompile(`(\d+)([wdhm])`)
	matches := re.FindAllStringSubmatch(estimate, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: no valid components found", ErrInvalidEstimateFormat)
	}

	// Verify that the entire string was matched (no extra characters)
	// Use FindAllStringIndex to check coverage without allocating strings
	indexes := re.FindAllStringIndex(estimate, -1)
	expectedPos := 0
	for _, idx := range indexes {
		if idx[0] != expectedPos {
			return nil, fmt.Errorf("%w: contains invalid characters", ErrInvalidEstimateFormat)
		}
		expectedPos = idx[1]
	}
	if expectedPos != len(estimate) {
		return nil, fmt.Errorf("%w: contains invalid characters", ErrInvalidEstimateFormat)
	}

	components := make([]EstimateComponent, 0, len(matches))
	seenUnits := make(map[string]bool)

	for _, match := range matches {
		quantity, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("%w: invalid quantity '%s': %w", ErrInvalidEstimateFormat, match[1], err)
		}

		if quantity <= 0 {
			return nil, fmt.Errorf("%w: quantity must be positive, got %d", ErrInvalidEstimateFormat, quantity)
		}

		unit := match[2]

		// Check for duplicate units
		if seenUnits[unit] {
			return nil, fmt.Errorf("%w: '%s' appears multiple times", ErrDuplicateEstimateUnit, unit)
		}
		seenUnits[unit] = true

		components = append(components, EstimateComponent{
			Quantity: quantity,
			Unit:     unit,
		})
	}

	return components, nil
}
