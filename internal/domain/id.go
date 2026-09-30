package domain

import (
	"fmt"
	"regexp"
)

// ID — идентификатор сущности. В базе это UUID.
type ID string

var idPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func ParseID(raw string) (ID, error) {
	if !idPattern.MatchString(raw) {
		return "", fmt.Errorf("id: нужен uuid")
	}
	return ID(raw), nil
}

func (id ID) String() string { return string(id) }

func (id ID) IsZero() bool { return id == "" }
