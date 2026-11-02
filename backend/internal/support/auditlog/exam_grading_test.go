package auditlog

import "testing"

func TestExamGradingAuditClassification(t *testing.T) {
	for _, path := range []string{
		"/api/v2/admin/exams/8/records/99/grading",
		"/api/v2/admin/exams/:id/records/:recordId/grading",
	} {
		op, ok := Classify("POST", path)
		if !ok || op.Source != "admin" || op.Category != "operation" || op.Module != "exam" || op.Action != "grade" || op.TargetType != "exam_record" || op.RelatedType != "exam" || op.Async {
			t.Fatalf("grading audit = %+v, tracked=%v", op, ok)
		}
		if _, ok := Classify("GET", path); ok {
			t.Fatal("reading grading detail must not generate mutation audit")
		}
	}
	for _, path := range []string{"/api/v2/admin/exams", "/api/v2/admin/exams/8/copy", "/api/v2/exams/8/records/99/grading"} {
		if op, _ := Classify("POST", path); op.Action == "grade" {
			t.Fatalf("unrelated route classified as grading: %s", path)
		}
	}
}
