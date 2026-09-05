package application

import (
	"strings"
	"unicode/utf8"
)

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

const (
	maxDefinitionNameSearchLength = 50
	maxStarterNameSearchLength    = 50
)

func normalizeDefinitionNameSearch(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > maxDefinitionNameSearchLength {
		return "", ErrDefinitionNameSearchTooLong
	}
	return value, nil
}

func normalizeStarterNameSearch(value string) (string, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > maxStarterNameSearchLength {
		return "", ErrStarterNameSearchTooLong
	}
	return value, nil
}
