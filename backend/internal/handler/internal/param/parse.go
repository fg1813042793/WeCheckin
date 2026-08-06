package param

import (
	"strconv"
	"strings"
)

func ParseUintSlice(value string) []uint {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	ids := make([]uint, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.Atoi(part); err == nil && id > 0 {
			ids = append(ids, uint(id))
		}
	}
	return ids
}
