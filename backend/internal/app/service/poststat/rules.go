package poststat

import (
	"encoding/json"

	"wecheckin/backend/pkg/logger"
)

type Rule struct {
	ID              string `json:"id"`
	Action          string `json:"action"`
	ConditionType   string `json:"conditionType"`
	StatField       string `json:"statField"`
	StatScope       string `json:"statScope"`
	NotifyChannel   string `json:"notifyChannel"`
	WebhookType     string `json:"webhookType"`
	WebhookURL      string `json:"webhookUrl"`
	NotifyAdmin     bool   `json:"notifyAdmin"`
	NotifyUserIds   string `json:"notifyUserIds"`
	MessageTemplate string `json:"messageTemplate"`
}

func parseRules(settings string) []Rule {
	var settingsMap map[string]interface{}
	if err := json.Unmarshal([]byte(settings), &settingsMap); err != nil {
		return nil
	}
	raw, ok := settingsMap["logicRules"]
	if !ok {
		return nil
	}
	var rawBytes []byte
	switch v := raw.(type) {
	case string:
		rawBytes = []byte(v)
	default:
		rawBytes, _ = json.Marshal(v)
	}
	var allRules []Rule
	if err := json.Unmarshal(rawBytes, &allRules); err != nil {
		logger.Logger.Printf("[PostStat] parse logicRules error: %v", err)
		return nil
	}
	var out []Rule
	for _, r := range allRules {
		if r.Action == "postStat" {
			out = append(out, r)
		}
	}
	return out
}
