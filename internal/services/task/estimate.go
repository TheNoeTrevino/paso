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
		return nil, fmt.Errorf("estimate string cannot be empty")
	}

	// Regex to match digit sequences followed by unit letters
	re := regexp.MustCompile(`(\d+)([wdhm])`)
	matches := re.FindAllStringSubmatch(estimate, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid estimate format: no valid components found")
	}

	// Verify that the entire string was matched (no extra characters)
	fullMatch := re.FindAllString(estimate, -1)
	var reconstructed string
	for _, match := range fullMatch {
		reconstructed += match
	}
	if reconstructed != estimate {
		return nil, fmt.Errorf("invalid estimate format: contains invalid characters")
	}

	components := make([]EstimateComponent, 0, len(matches))
	seenUnits := make(map[string]bool)

	for _, match := range matches {
		quantity, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("invalid quantity '%s': %w", match[1], err)
		}

		if quantity <= 0 {
			return nil, fmt.Errorf("quantity must be positive, got %d", quantity)
		}

		unit := match[2]

		// Check for duplicate units
		if seenUnits[unit] {
			return nil, fmt.Errorf("duplicate unit '%s' found", unit)
		}
		seenUnits[unit] = true

		components = append(components, EstimateComponent{
			Quantity: quantity,
			Unit:     unit,
		})
	}

	return components, nil
}
