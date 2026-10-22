package designlint

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// rgb is an sRGB colour with channels in 0..255.
type rgb struct{ r, g, b float64 }

// colour is an sRGB colour with an alpha in 0..1. Alpha 1 is opaque.
type colour struct {
	c rgb
	a float64
}

// key identifies a colour by value, for comparing token declarations.
func (c colour) key() string {
	return fmt.Sprintf("%02x%02x%02x/%.3f", int(math.Round(c.c.r)), int(math.Round(c.c.g)), int(math.Round(c.c.b)), c.a)
}

// hex returns the colour as #rrggbb, for messages. Alpha is shown when it is not 1.
func (c colour) hex() string {
	s := fmt.Sprintf("#%02x%02x%02x", int(math.Round(c.c.r)), int(math.Round(c.c.g)), int(math.Round(c.c.b)))
	if c.a < 1 {
		s += fmt.Sprintf(" at %.0f%% opacity", c.a*100)
	}
	return s
}

var (
	reHexColour = regexp.MustCompile(`#([0-9a-fA-F]{8}|[0-9a-fA-F]{6}|[0-9a-fA-F]{4}|[0-9a-fA-F]{3})\b`)
	reRGBColour = regexp.MustCompile(`(?i)rgba?\(\s*([0-9.]+)\s*[, ]\s*([0-9.]+)\s*[, ]\s*([0-9.]+)\s*(?:[,/]\s*([0-9.]+)(%?)\s*)?\)`)
)

// namedColours are the keywords that cover most real text and background declarations.
// Any other keyword is not a concrete colour, and a rule that uses it is skipped.
var namedColours = map[string]colour{
	"black":       {rgb{0, 0, 0}, 1},
	"white":       {rgb{255, 255, 255}, 1},
	"transparent": {rgb{0, 0, 0}, 0},
}

// parseColor reads the first colour in val: #rgb, #rgba, #rrggbb, #rrggbbaa,
// rgb()/rgba() with 0..255 channels and an optional alpha, or a keyword in namedColours.
func parseColor(val string) (colour, bool) {
	val = strings.TrimSpace(val)
	if m := reHexColour.FindStringSubmatch(val); m != nil {
		h := m[1]
		switch len(h) {
		case 3, 4:
			var b strings.Builder
			for i := 0; i < len(h); i++ {
				b.WriteByte(h[i])
				b.WriteByte(h[i])
			}
			h = b.String()
		}
		n, err := strconv.ParseUint(h, 16, 64)
		if err != nil {
			return colour{}, false
		}
		if len(h) == 8 {
			return colour{rgb{float64(n >> 24 & 0xff), float64(n >> 16 & 0xff), float64(n >> 8 & 0xff)}, float64(n&0xff) / 255}, true
		}
		return colour{rgb{float64(n >> 16 & 0xff), float64(n >> 8 & 0xff), float64(n & 0xff)}, 1}, true
	}
	if m := reRGBColour.FindStringSubmatch(val); m != nil {
		var ch [3]float64
		for i := range ch {
			v, err := strconv.ParseFloat(m[i+1], 64)
			if err != nil || v < 0 || v > 255 {
				return colour{}, false
			}
			ch[i] = v
		}
		alpha := 1.0
		if m[4] != "" {
			a, err := strconv.ParseFloat(m[4], 64)
			if err != nil {
				return colour{}, false
			}
			if m[5] == "%" {
				a /= 100
			}
			alpha = math.Max(0, math.Min(1, a))
		}
		return colour{rgb{ch[0], ch[1], ch[2]}, alpha}, true
	}
	if c, ok := namedColours[strings.ToLower(val)]; ok {
		return c, true
	}
	return colour{}, false
}

// composite returns fg drawn over an opaque bg, as the colour the eye sees.
func composite(fg, bg colour) rgb {
	return rgb{
		r: fg.c.r*fg.a + bg.c.r*(1-fg.a),
		g: fg.c.g*fg.a + bg.c.g*(1-fg.a),
		b: fg.c.b*fg.a + bg.c.b*(1-fg.a),
	}
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

// contrastRatio is the WCAG contrast ratio between two opaque colours, from 1 to 21.
func contrastRatio(a, b rgb) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
