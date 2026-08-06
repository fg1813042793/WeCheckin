package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestDiagnoseOldAgentWorkNotificationReportsNumericAgentID(t *testing.T) {
	var sent map[string]interface{}
	client := defaultDingTalkIdentityClient{
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/topapi/message/corpconversation/asyncsend_v2" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			body, _ := json.Marshal(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"task_id": 7,
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		})},
	}
	step := client.diagnoseOldAgentWorkNotificationContext(context.Background(), "token-1", "342080997", "manager001")
	if step.Status != "success" {
		t.Fatalf("diagnosis status = %q, error = %q", step.Status, step.Error)
	}
	if got := step.Request["agent_id"]; got != int64(342080997) {
		t.Fatalf("diagnosis request agent_id = %#v, want int64 342080997", got)
	}
	if got := sent["agent_id"]; got != float64(342080997) {
		t.Fatalf("sent request agent_id = %#v, want numeric 342080997", got)
	}
}
