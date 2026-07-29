package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminListDTOsAvoidHeavyFieldsAndGenericPayloads(t *testing.T) {
	checks := []struct {
		file     string
		required []string
		forbid   []string
	}{
		{
			file: filepath.Join("survey", "dto.go"),
			required: []string{
				"type surveyListItem struct",
				"List  []surveyListItem `json:\"list\"`",
			},
			forbid: []string{
				"model.Survey\n\tResponseCount",
			},
		},
		{
			file: filepath.Join("exam", "dto.go"),
			required: []string{
				"type examListItem struct",
				"List  []examListItem `json:\"list\"`",
			},
			forbid: []string{
				"[]model.Exam `json:\"list\"`",
			},
		},
		{
			file: filepath.Join("event", "dto.go"),
			required: []string{
				"type eventListItem struct",
				"List  []eventListItem `json:\"list\"`",
			},
			forbid: []string{
				"type pagedListResponse struct {\n\tList  interface{}",
				"`json:\"forms\"`",
				"`json:\"scoreFields\"`",
			},
		},
		{
			file: filepath.Join("enroll", "dto.go"),
			required: []string{
				"type enrollListItem struct",
				"List  []enrollListItem `json:\"list\"`",
			},
			forbid: []string{
				"type pagedListResponse struct {\n\tList  interface{}",
				"`json:\"forms\"`",
				"`json:\"joinForms\"`",
				"`json:\"userList\"`",
			},
		},
	}
	for _, check := range checks {
		src, err := os.ReadFile(check.file)
		if err != nil {
			t.Fatalf("read %s: %v", check.file, err)
		}
		text := string(src)
		for _, required := range check.required {
			if !strings.Contains(text, required) {
				t.Fatalf("%s list DTO should include %q", check.file, required)
			}
		}
		for _, forbidden := range check.forbid {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s list DTO should avoid heavy or generic field %q", check.file, forbidden)
			}
		}
	}
}
