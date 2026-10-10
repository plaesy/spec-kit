package mdlint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ruleFunc is one rule's check. Options come from the ruleSet so a rule never
// has to reach back into the config.
type ruleFunc func(d *doc, rs *ruleSet) []violation

type ruleSpec struct {
	id      string
	aliases []string
	// options decodes the rule's config object into its option struct.
	options func() any
	check   ruleFunc
	// fixable marks rules that Fix() can repair mechanically.
	fixable bool
}

// ruleSet is a compiled configuration: the enabled rules with decoded options.
type ruleSet struct {
	rules   []ruleSpec
	options map[string]any
}

// opt returns the decoded options for a rule, or the zero value when the config
// did not mention it.
func opt[T any](rs *ruleSet, id string) T {
	if v, ok := rs.options[id]; ok {
		switch typed := v.(type) {
		case T:
			return typed
		case *T:
			return *typed
		}
	}
	var zero T
	return zero
}

// compile turns a config into a ruleSet. Unknown rule names, unknown option
// keys, and non-boolean shorthands are errors: a config that half-loads is a
// check that reports success while checking nothing.
func compile(cfg *Config) (*ruleSet, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	byKey := map[string]*ruleSpec{}
	for i := range allRules {
		spec := &allRules[i]
		byKey[spec.id] = spec
		for _, a := range spec.aliases {
			byKey[a] = spec
		}
	}

	rs := &ruleSet{options: map[string]any{}}
	explicitOff := map[string]bool{}
	mentioned := map[string]bool{}

	keys := make([]string, 0, len(cfg.Rules))
	for k := range cfg.Rules {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		raw := cfg.Rules[key]
		spec, ok := byKey[key]
		if !ok {
			// A rule this linter does not implement, switched OFF, is already
			// off: ignoring it is the same as honouring it, and rejecting it would
			// refuse to load a working .markdownlint.json just because it mentions
			// a rule we skipped. Switching one ON, or configuring it, is a
			// different matter — the user is asking for a check that would not
			// happen, which is exactly the silent no-op this validation exists to
			// prevent.
			var off bool
			if err := json.Unmarshal(raw, &off); err == nil && !off {
				continue
			}
			return nil, fmt.Errorf("config: rule %q is not implemented, so enabling or configuring it would silently do nothing — implemented rules are: %s (run 'plaesy validate markdown --list-rules')", key, strings.Join(implementedRuleIDs(), ", "))
		}
		mentioned[spec.id] = true
		var on bool
		if err := json.Unmarshal(raw, &on); err == nil {
			if !on {
				explicitOff[spec.id] = true
			}
			continue
		}
		if spec.options == nil {
			return nil, fmt.Errorf("config: rule %q takes no options — give it true or false", spec.id)
		}
		target := spec.options()
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(target); err != nil {
			return nil, fmt.Errorf("config: rule %q: %w", spec.id, err)
		}
		rs.options[spec.id] = target
	}

	for i := range allRules {
		spec := &allRules[i]
		if explicitOff[spec.id] {
			continue
		}
		if !cfg.Default && !mentioned[spec.id] {
			// "default": false — only rules the config names are active.
			continue
		}
		rs.rules = append(rs.rules, *spec)
	}
	sort.Slice(rs.rules, func(i, j int) bool { return rs.rules[i].id < rs.rules[j].id })
	return rs, nil
}

func implementedRuleIDs() []string {
	out := make([]string, 0, len(allRules))
	for i := range allRules {
		out = append(out, allRules[i].id)
	}
	sort.Strings(out)
	return out
}
