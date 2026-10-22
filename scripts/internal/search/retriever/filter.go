package retriever

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Filter is a set of metadata predicates. A chunk is kept only when every
// predicate holds. The zero value keeps everything.
//
// Filters run on the vector and text candidates before they are fused, so a
// chunk that fails a filter never reaches the ranking. Exact-match filters
// (Filters in QueryOptions) are applied by the stores themselves; these
// predicates cover what exact match cannot express: a date range, a path
// prefix, and a file type.
type Filter []predicate

type predicate func(meta map[string]string) bool

// ModifiedKey is the chunk metadata key a date filter reads. The value is the
// document's modification date, YYYY-MM-DD. A chunk without it fails every date
// predicate, so a missing date never matches a date range by accident.
const ModifiedKey = "modified"

// ParseFilter reads a comma-separated filter spec, for example
// "source:markdown,path:internal/search,file_type:md,date:>2024-01-01".
//
//   - source:<value>      exact match on the source metadata ("markdown", "html", "json")
//   - path:<prefix>       the document path starts with the prefix, matched on
//     whole path segments, so "search" matches "a/search/b.go" and not "a/searching.go"
//   - file_type:<ext>     the document's extension, with or without the leading dot
//   - date:<op><YYYY-MM-DD> compares the modification date. op is one of >, >=, <, <=,
//     or = (the default when no op is given)
//
// Every term must hold (AND). An empty spec returns a Filter that keeps everything.
func ParseFilter(spec string) (Filter, error) {
	var f Filter
	for _, term := range strings.Split(spec, ",") {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		key, value, ok := strings.Cut(term, ":")
		if !ok {
			return nil, fmt.Errorf("filter term %q: want key:value", term)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("filter %q: empty value", key)
		}

		switch key {
		case "source":
			f = append(f, func(meta map[string]string) bool {
				return meta["source"] == value
			})
		case "path":
			prefix := "/" + strings.Trim(filepath.ToSlash(value), "/")
			f = append(f, func(meta map[string]string) bool {
				p := "/" + strings.Trim(filepath.ToSlash(meta["path"]), "/")
				// The stored path can be absolute, so the prefix is matched on segments
				// anywhere in it, not only at the start.
				return strings.Contains(p+"/", prefix+"/")
			})
		case "file_type":
			ext := strings.ToLower(strings.TrimPrefix(value, "."))
			f = append(f, func(meta map[string]string) bool {
				return strings.ToLower(strings.TrimPrefix(filepath.Ext(meta["path"]), ".")) == ext
			})
		case "date":
			pred, err := datepredicate(value)
			if err != nil {
				return nil, err
			}
			f = append(f, pred)
		default:
			return nil, fmt.Errorf("unknown filter key %q (want source, path, file_type, or date)", key)
		}
	}
	return f, nil
}

// dateLayout is the only date format the filter accepts.
const dateLayout = "2006-01-02"

// dateOps are checked longest first, so ">=" is not read as ">" followed by "=".
var dateOps = []string{">=", "<=", ">", "<", "="}

func datepredicate(value string) (predicate, error) {
	op := "="
	for _, candidate := range dateOps {
		if strings.HasPrefix(value, candidate) {
			op = candidate
			value = strings.TrimPrefix(value, candidate)
			break
		}
	}
	want, err := time.Parse(dateLayout, strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("date filter %q: want YYYY-MM-DD after the operator: %w", value, err)
	}
	return func(meta map[string]string) bool {
		raw, ok := meta[ModifiedKey]
		if !ok || len(raw) < len(dateLayout) {
			return false
		}
		got, err := time.Parse(dateLayout, raw[:len(dateLayout)])
		if err != nil {
			return false
		}
		switch op {
		case ">":
			return got.After(want)
		case ">=":
			return !got.Before(want)
		case "<":
			return got.Before(want)
		case "<=":
			return !got.After(want)
		default:
			return got.Equal(want)
		}
	}, nil
}

// Match reports whether a chunk's metadata passes every predicate in f.
func (f Filter) Match(meta map[string]string) bool {
	for _, p := range f {
		if !p(meta) {
			return false
		}
	}
	return true
}

// keepBy returns the items whose metadata passes f. It keeps the original order,
// and returns the input unchanged when f is empty.
func keepBy[T any](items []T, meta func(T) map[string]string, f Filter) []T {
	if len(f) == 0 {
		return items
	}
	out := make([]T, 0, len(items))
	for _, it := range items {
		if f.Match(meta(it)) {
			out = append(out, it)
		}
	}
	return out
}
