package exam

import (
	"os"
	"strings"
	"testing"
)

func TestAdminExamOperationsUseAdminScopedQueries(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(src)
	required := []string{
		"func scopedExamQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error)",
		"access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Exam{}, access.ExamAuditFields)",
		"func (s *Service) ListForAdminContext(",
		"func (s *Service) DetailForAdminContext(ctx context.Context, id uint, adminID uint)",
		"func (s *Service) UpdateForAdminContext(ctx context.Context, id uint, updates map[string]interface{}, adminID uint) error",
		"func (s *Service) SetStatusForAdminContext(ctx context.Context, id uint, status int, adminID uint) error",
		"func (s *Service) DeleteForAdminContext(ctx context.Context, id uint, adminID uint) error",
		"func (s *Service) RecordListForAdminContext(ctx context.Context, examID int, keyword string, page, pageSize int, adminID uint)",
		"func (s *Service) RecordDetailForAdminContext(ctx context.Context, id uint, adminID uint)",
		"func (s *Service) RecordDeleteForAdminContext(ctx context.Context, id uint, adminID uint) error",
		"func (s *Service) StatisticsForAdminContext(ctx context.Context, examID int, adminID uint)",
	}
	for _, snippet := range required {
		if !strings.Contains(text, snippet) {
			t.Fatalf("exam service must protect admin operation with %q", snippet)
		}
	}
}
