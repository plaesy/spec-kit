package mdlint

import "testing"

// MD036Options.Punctuation was declared, parsed from .markdownlint.json, and
// then discarded by checkMD036 (`_ = opt[MD036Options](rs, "MD036")`). Setting
// the option therefore changed nothing, while looking exactly like a setting
// that had been honoured. An allowance that fails to apply is
// indistinguishable from one already applied: it reads as success.
//
// The default is markdownlint's own: an emphasized run ending in sentence
// punctuation is a label or caption ("**Example Requirements:**" introducing a
// code block), not a section title.
func TestMD036PunctuationExemptsCaptions(t *testing.T) {
	cases := []struct {
		name     string
		config   string
		body     string
		wantRule int
	}{
		{
			name:     "default exempts a trailing colon",
			config:   minimalConfig,
			body:     "text\n\n**Example Requirements:**\n\n```\ncode\n```\n",
			wantRule: 0,
		},
		{
			// The rule must still fire on a bare emphasized run. This is the
			// case that proves the exemption is narrow and not a blanket
			// disabling of MD036.
			name:     "default still flags a bare emphasized run",
			config:   minimalConfig,
			body:     "text\n\n**Data Layer**\n\nbody\n",
			wantRule: 1,
		},
		{
			name:     "default exempts a trailing period",
			config:   minimalConfig,
			body:     "text\n\n**Note.**\n\nbody\n",
			wantRule: 0,
		},
		{
			// The ? and the full-width ？ are both in markdownlint's default
			// set. An earlier version of the constant here omitted them, so a
			// question-shaped label was reported by this linter but not by a
			// conformant markdownlint run.
			name:     "default exempts a trailing question mark",
			config:   minimalConfig,
			body:     "text\n\n**How might we {{X}}?**\n\nbody\n",
			wantRule: 0,
		},
		{
			name:     "default exempts a trailing full-width question mark",
			config:   minimalConfig,
			body:     "text\n\n**Recommendation?**\n\nbody\n",
			wantRule: 0,
		},
		{
			// Setting punctuation to "" disables the exemption, which is how
			// a consumer opts back into the stricter behaviour.
			name:     "empty punctuation disables the exemption",
			config:   `{"default": true, "MD036": {"punctuation": ""}}`,
			body:     "text\n\n**Example Requirements:**\n\nbody\n",
			wantRule: 1,
		},
		{
			// A custom set is honoured, proving the option is actually read
			// rather than hardcoded to the default.
			name:     "custom punctuation set is honoured",
			config:   `{"default": true, "MD036": {"punctuation": "!"}}`,
			body:     "text\n\n**Example Requirements:**\n\nbody\n",
			wantRule: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := writeFiles(t, map[string]string{
				".markdownlint.json": tc.config,
				"a.md":               tc.body,
			})
			res, err := Lint(root, nil, nil, nil)
			if err != nil {
				t.Fatalf("Lint: %v", err)
			}
			if got := res.ByRule["MD036"]; got != tc.wantRule {
				t.Errorf("MD036 = %d, want %d", got, tc.wantRule)
			}
		})
	}
}
