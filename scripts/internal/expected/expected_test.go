package expected

import "testing"

func TestJudge(t *testing.T) {
	cases := []struct {
		expected, output, want string
	}{
		{"pass", "", ResultMet},
		{"zero", "0", ResultMet},
		{"zero", "3", ResultNotMet},
		{"all", "5 of 5", ResultMet},
		{"all", "4 of 5", ResultNotMet},
		{"all pages", "pass", ResultMet},
		{"every route", "3 of 4", ResultNotMet},
		{">= 9", "9", ResultMet},
		{">= 9", "8", ResultNotMet},
		{"<= 2 s", "1.5 s", ResultMet},
		{"<= 2 s", "3", ResultNotMeasurable},
		{">= 9", "", ResultNotMeasurable},
	}
	for _, c := range cases {
		got, _, reason := Judge(c.expected, c.output)
		if got != c.want {
			t.Errorf("Judge(%q, %q) = %s (%s), want %s", c.expected, c.output, got, reason, c.want)
		}
	}
}
