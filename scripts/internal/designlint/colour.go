package designlint

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

// rgb is an sRGB colour with channels in 0..255. Alpha is ignored: a translucent
// colour is checked as if opaque, which can over-report a failure. That is the
// safe direction for a contrast gate.
type rgb struct{ r, g, b float64 }

var (
	reHexColour = regexp.MustCompile(`#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\b`)
	reRGBColour = regexp.MustCompile(`(?i)rgba?\(\s*([0-9.]+)\s*,\s*([0-9.]+)\s*,\s*([0-9.]+)`)
)

// parseColor reads the first colour in val. It accepts #rgb, #rrggbb and
// rgb()/rgba() with 0..255 channels. Anything else is not a concrete colour.
func parseColor(val string) (rgb, bool) {
	if m := reHexColour.FindStringSubmatch(val); m != nil {
		h := m[1]
		if len(h) == 3 {
			h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		}
		n, err := strconv.ParseUint(h, 16, 32)
		if err != nil {
			return rgb{}, false
		}
		return rgb{float64(n >> 16 & 0xff), float64(n >> 8 & 0xff), float64(n & 0xff)}, true
	}
	if m := reRGBColour.FindStringSubmatch(val); m != nil {
		var c [3]float64
		for i := range c {
			v, err := strconv.ParseFloat(m[i+1], 64)
			if err != nil || v < 0 || v > 255 {
				return rgb{}, false
			}
			c[i] = v
		}
		return rgb{c[0], c[1], c[2]}, true
	}
	return rgb{}, false
}

// hex returns the colour as #rrggbb, for messages.
func (c rgb) hex() string {
	return fmt.Sprintf("#%02x%02x%02x", int(math.Round(c.r)), int(math.Round(c.g)), int(math.Round(c.b)))
}

// luminance is the WCAG 2.x relative luminance of an sRGB colour.
func luminance(c rgb) float64 {
	lin := func(v float64) float64 {
		s := v / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.r) + 0.7152*lin(c.g) + 0.0722*lin(c.b)
}

// contrastRatio is the WCAG contrast ratio between two colours, from 1 to 21.
func contrastRatio(a, b rgb) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
