package poststat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"wecheckin-backend/backend/pkg/logger"
)

func sendWebhook(webhookType, webhookURL, title, msg string) {
	msg = strings.ReplaceAll(msg, "\n", "\n\n")
	client := &http.Client{Timeout: 10 * time.Second}
	var body []byte

	switch webhookType {
	case "dingtalk":
		payload := map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": title,
				"text":  msg,
			},
		}
		body, _ = json.Marshal(payload)
	case "wecom":
		payload := map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": msg,
			},
		}
		body, _ = json.Marshal(payload)
	case "lark":
		payload := map[string]interface{}{
			"msg_type": "interactive",
			"card": map[string]interface{}{
				"header": map[string]interface{}{
					"title":    map[string]string{"tag": "plain_text", "content": title},
					"template": "blue",
				},
				"elements": []map[string]interface{}{
					{"tag": "markdown", "content": msg},
				},
			},
		}
		body, _ = json.Marshal(payload)
	default:
		payload := map[string]interface{}{
			"title":   title,
			"content": msg,
			"type":    "survey_stat",
		}
		body, _ = json.Marshal(payload)
	}

	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		logger.Logger.Printf("[PostStat] webhook send error: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		logger.Logger.Printf("[PostStat] webhook status %d for %s", resp.StatusCode, webhookURL)
	}
}
