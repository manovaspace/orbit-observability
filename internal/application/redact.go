package application

import "strings"

// RedactString masks "key=value" and "key: value" fragments for the given keys.
// The keys are lower-cased for matching.
func RedactString(s string, keys []string) string {
	out := s
	for _, k := range keys {
		if k == "" {
			continue
		}
		for _, sep := range []string{"=", ": ", ":"} {
			needle := k + sep
			search := 0
			for {
				i := IndexFold(out[search:], needle)
				if i < 0 {
					break
				}
				i += search
				start := i + len(needle)
				if start >= len(out) || isValueBoundary(out[start]) {
					// Nothing to mask (for example the ": " pass already
					// replaced this key). Skip the needle and keep scanning.
					search = start
					continue
				}
				end := start
				for end < len(out) && out[end] != ' ' && out[end] != '"' && out[end] != '\'' && out[end] != ',' && out[end] != '}' {
					end++
				}
				out = out[:start] + "[REDACTED]" + out[end:]
				// Skip the marker so the same occurrence cannot match again.
				search = start + len("[REDACTED]")
			}
		}
	}
	return out
}

// IndexFold finds substr in s, ignoring case.
func IndexFold(s, substr string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(substr))
}

func isValueBoundary(b byte) bool {
	switch b {
	case ' ', '"', '\'', ',', '}':
		return true
	default:
		return false
	}
}
