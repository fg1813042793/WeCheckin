package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExamGradingRoutesAndSwaggerContract(t *testing.T) {
	for path, required := range map[string][]string{
		"routes.go": {
			`admin.GET("/exams/:id/records/:recordId/grading", aExam.RecordGrading)`,
			`admin.POST("/exams/:id/records/:recordId/grading", aExam.GradeRecord)`,
		},
		filepath.Join("..", "swagger", "exam_grading.go"): {
			"@Router /api/v2/admin/exams/{id}/records/{recordId}/grading [get]",
			"@Router /api/v2/admin/exams/{id}/records/{recordId}/grading [post]",
			"@Security AdminToken", "examservice.RecordGradingDetail", "examservice.GradeRecordRequest",
			"@Param id path int true", "@Param recordId path int true", "提交不可重复",
		},
	} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, snippet := range required {
			if !strings.Contains(string(source), snippet) {
				t.Errorf("%s missing grading contract %q", path, snippet)
			}
		}
	}
}
