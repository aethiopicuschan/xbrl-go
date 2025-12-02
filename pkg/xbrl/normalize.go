package xbrl

import "strings"

// normalizeSpace replaces several space-like runes with ASCII space
// and collapses consecutive whitespace into a single space.
func normalizeSpace(s string) string {
	if s == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"\u00A0", " ",
		"\u3000", " ",
	)
	s = replacer.Replace(s)

	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}
