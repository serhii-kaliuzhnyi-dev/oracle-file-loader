package util

import (
	"regexp"
	"strings"
)

func ToLowerSnakeCase(s string) string {
	// Convert spaces to underscores
	s = strings.ReplaceAll(s, " ", "_")
	// Convert to lower case
	s = strings.ToLower(s)
	// Use regex to replace any sequence of non-word characters with a single underscore
	re := regexp.MustCompile(`[^\w]+`)
	s = re.ReplaceAllString(s, "_")
	// Remove any leading or trailing underscores
	s = strings.Trim(s, "_")
	return s
}
