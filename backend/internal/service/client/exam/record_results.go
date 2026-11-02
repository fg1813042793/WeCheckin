package exam

import (
	"encoding/json"
	"strings"
	examPkg "wecheckin/backend/internal/formkit/exam"
)

func decodeGradingPaperResult(raw string) (examPkg.PaperResult, error) {
	var result examPkg.PaperResult
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &result.Results); err != nil {
			return result, ErrRecordGradingUnavailable
		}
	} else {
		if err := json.Unmarshal([]byte(raw), &result); err != nil || result.Results == nil {
			return result, ErrRecordGradingUnavailable
		}
	}
	return result, nil
}

// DecodeRecordResults supports both schema result arrays and paper result objects.
func DecodeRecordResults(raw string) []examPkg.Result {
	result, err := decodeGradingPaperResult(raw)
	if err != nil {
		return nil
	}
	return result.Results
}
