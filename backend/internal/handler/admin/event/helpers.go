package event

import (
	"encoding/json"
	"strings"
)

func parseUserArray(s string) []string {
	if s == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err == nil {
		return arr
	}
	return strings.Split(s, ",")
}
