package schema

import (
	"encoding/json"
	"fmt"
)

func jsonUnmarshal(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

func fmtSprintfForReport(i int) string { return fmt.Sprintf("q%d", i) }
