package storage

import (
	"fmt"
	"regexp"
)

var tableNamingRegex = regexp.MustCompile(`^[a-z0-9_]+$`)

// ValidateTableName ensures the name is safe for SQL injection and
// fits standard DB naming conventions.
func ValidateTableName(name string) error {
	if !tableNamingRegex.MatchString(name) {
		return fmt.Errorf(
			"invalid table name %q: only lowercase alphanumeric and underscores allowed",
			name,
		)
	}
	return nil
}
