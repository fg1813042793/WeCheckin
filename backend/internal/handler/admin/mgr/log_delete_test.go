package mgr

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route/param"
	"wecheckin/backend/internal/model"
)

func TestLogDeleteRejectsMalformedAndUnboundedRequests(t *testing.T) {
	for _, handler := range []app.HandlerFunc{(&AdminMgrHandler{}).DeleteLogs, (&AdminMgrHandler{}).DeleteLoginLogs} {
		for _, tc := range []struct{ id, body string }{
			{"", ""}, {"", "{}"}, {"", `{"ids":[]}`}, {"", `{"ids":[0]}`}, {"", `{"ids":[1,1]}`},
			{"", `{"ids":[-1]}`}, {"", `{"ids":[1.5]}`}, {"", `{"ids":["1"]}`}, {"", `{"ids":[1],"all":true}`},
			{"", `{"ids":[1]} {}`}, {"", `{"ids":[` + strings.Repeat("1,", 100) + "2]}"},
			{"0", ""}, {"bad", ""}, {"1", `{"ids":[2]}`},
		} {
			c := app.NewContext(1)
			c.Set("admin", &model.Admin{ID: 7})
			c.Params = param.Params{{Key: "id", Value: tc.id}}
			c.Request.SetBodyString(tc.body)
			handler(context.Background(), c)
			var body struct {
				Code int
				Msg  string
			}
			if err := json.Unmarshal(c.Response.Body(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code == 0 || strings.Contains(body.Msg, "稍后重试") {
				t.Fatalf("invalid request did not fail validation: id=%q body=%s response=%s", tc.id, tc.body, c.Response.Body())
			}
		}
	}
}
