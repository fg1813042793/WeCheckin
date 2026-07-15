package service

import (
	"os"
	"testing"
)

func TestEnrollServiceCodeStaysSplitByResponsibility(t *testing.T) {
	required := []string{"enroll_client.go", "enroll_submission.go", "publish_dept_helpers.go"}
	for _, file := range required {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("enroll service code must keep %s split by responsibility: %v", file, err)
		}
	}
}
