package swagger

import (
	"os"
	"strings"
	"testing"
)

func TestNamedAdminResponseDTOsAreDocumented(t *testing.T) {
	source, err := os.ReadFile("swagger.go")
	if err != nil {
		t.Fatalf("read swagger.go: %v", err)
	}
	text := string(source)
	for _, annotation := range []string{
		"response.Resp{data=adminsetuphandler.DebugTokenConfigResponse}",
		"response.Resp{data=admindingtalkservice.SettingsResponse}",
		"response.Resp{data=admindingtalkservice.UserBindingList}",
		"response.Resp{data=admindingtalkservice.PerfReviewList}",
		"response.Resp{data=admindingtalkservice.PerfReviewDetail}",
		"response.Resp{data=admindingtalkservice.PerfHistoryList}",
		"response.Resp{data=formkitschema.FormSchema}",
		"response.Resp{data=adminsurveyhandler.EvalExprResponse}",
	} {
		if !strings.Contains(text, annotation) {
			t.Fatalf("Swagger source missing named response annotation %q", annotation)
		}
	}
}
