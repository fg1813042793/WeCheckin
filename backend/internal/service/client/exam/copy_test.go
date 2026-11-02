package exam

import (
	"encoding/json"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestBuildExamCopyResetsIdentityPublicationAndShareLink(t *testing.T) {
	source := model.Exam{
		ID: 9, Title: "月度考试", Description: "说明", Category: "培训", Tags: "安全",
		Visibility: 2, AllowMulti: 1, Anonymous: 1, ShowResult: 1, PaperID: 12,
		Schema:    `{"questions":[{"id":"q1"}]}`,
		Settings:  `{"publicUrl":"https://example.com/ef1/9","logicRules":"[]","answerVisible":true}`,
		StartTime: 100, EndTime: 200, Duration: 60, MaxAttempts: 2, ShowScore: 1,
		MaxResponse: 50, DeptIds: "1,2", Mode: "exam", PublishDepts: "1", QR: "old-qr",
		Status: 1, Order: 3, DeptID: 4, CreateBy: 5, UpdateBy: 6, UpdateDeptID: 7,
		AddTime: 10, EditTime: 20,
	}

	copied := buildExamCopy(source, 21, 22, 1000)
	if copied.ID != 0 || copied.Title != "月度考试 (副本)" || copied.Status != 0 || copied.QR != "" {
		t.Fatalf("copy identity/publication fields = %#v", copied)
	}
	if copied.CreateBy != 21 || copied.UpdateBy != 21 || copied.DeptID != 22 || copied.UpdateDeptID != 22 {
		t.Fatalf("copy ownership fields = %#v", copied)
	}
	if copied.AddTime != 1000 || copied.EditTime != 1000 {
		t.Fatalf("copy timestamps = %d/%d", copied.AddTime, copied.EditTime)
	}
	if copied.Schema != source.Schema || copied.PaperID != source.PaperID || copied.StartTime != source.StartTime || copied.EndTime != source.EndTime {
		t.Fatalf("exam configuration was not preserved: %#v", copied)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(copied.Settings), &settings); err != nil {
		t.Fatalf("decode copied settings: %v", err)
	}
	if settings["publicUrl"] != "" || settings["answerVisible"] != true || settings["logicRules"] != "[]" {
		t.Fatalf("copied settings = %#v", settings)
	}
}
