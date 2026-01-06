package styles

import (
	"fmt"
	"strings"
)

func DimColor(hex string, intensity float64) string {
	r, g, b, ok := parseHex(hex)
	if !ok {
		return hex
	}

	if intensity < 0 {
		intensity = 0
	} else if intensity > 1 {
		intensity = 1
	}

	factor := 1 - intensity
	r = uint8(float64(r) * factor)
	g = uint8(float64(g) * factor)
	b = uint8(float64(b) * factor)

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func parseHex(hex string) (r, g, b uint8, ok bool) {
	hex = strings.TrimPrefix(hex, "#")

	switch len(hex) {
	case 3:
		var ri, gi, bi int
		_, err := fmt.Sscanf(hex, "%1x%1x%1x", &ri, &gi, &bi)
		if err != nil {
			return 0, 0, 0, false
		}
		return uint8(ri * 17), uint8(gi * 17), uint8(bi * 17), true
	case 6:
		var ri, gi, bi int
		_, err := fmt.Sscanf(hex, "%2x%2x%2x", &ri, &gi, &bi)
		if err != nil {
			return 0, 0, 0, false
		}
		return uint8(ri), uint8(gi), uint8(bi), true
	default:
		return 0, 0, 0, false
	}
}
