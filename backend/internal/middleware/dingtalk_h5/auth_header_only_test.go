package dingtalkh5

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestDingTalkH5AuthenticationOnlyAcceptsAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		query  string
		want   string
	}{
		{name: "raw token", header: "header-token", want: "header-token"},
		{name: "bearer token", header: "Bearer header-token", want: "header-token"},
		{name: "case insensitive bearer", header: "bearer header-token", want: "header-token"},
		{name: "query token is ignored", query: "query-token", want: ""},
		{name: "header wins over query", header: "header-token", query: "query-token", want: "header-token"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := app.NewContext(0)
			c.Request.SetRequestURI("/api/v2/dingtalk/h5/workflows/definitions?token=" + test.query)
			if test.header != "" {
				c.Request.Header.Set("Authorization", test.header)
			}

			if got := authorizationToken(c); got != test.want {
				t.Fatalf("authorizationToken() = %q, want %q", got, test.want)
			}
		})
	}
}
