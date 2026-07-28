package survey

import (
	"os"
	"strings"
	"testing"
)

func TestAdminSurveyOperationsUseAdminScopedQueries(t *testing.T) {
	cases := []struct {
		file     string
		required []string
	}{
		{
			file: "service.go",
			required: []string{
				"func scopedSurveyQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error)",
				"access.ScopedResourceQueryContext(ctx, db, adminID, &model.Survey{}, \"`survey_dept_id`\", \"`survey_create_by`\")",
				"func (s *SurveyService) ListForAdminContext(",
				"func (s *SurveyService) DetailForAdminContext(ctx context.Context, id uint, adminID uint)",
				"func (s *SurveyService) UpdateForAdminContext(ctx context.Context, sv *model.Survey, adminID uint) error",
				"func (s *SurveyService) SetStatusForAdminContext(ctx context.Context, id uint, status int, adminID uint) error",
				"func (s *SurveyService) DeleteForAdminContext(ctx context.Context, id uint, adminID uint) error",
			},
		},
		{
			file: "response_query.go",
			required: []string{
				"func (r *ResponseService) ListForAdminContext(ctx context.Context, surveyID uint, page, pageSize int, keyword string, adminID uint)",
				"func (r *ResponseService) GetForAdminContext(ctx context.Context, id uint, adminID uint)",
				"func (r *ResponseService) DeleteForAdminContext(ctx context.Context, id uint, adminID uint) error",
				"func (r *ResponseService) BatchDeleteForAdminContext(ctx context.Context, ids []int, adminID uint) error",
			},
		},
	}

	for _, tc := range cases {
		src, err := os.ReadFile(tc.file)
		if err != nil {
			t.Fatalf("read %s: %v", tc.file, err)
		}
		text := string(src)
		for _, snippet := range tc.required {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must protect admin survey operation with %q", tc.file, snippet)
			}
		}
	}
}
