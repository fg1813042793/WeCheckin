package admincontent

import (
	"os"
	"strings"
	"testing"
)

func TestAdminContentProjectOperationsUseAdminScopedQueries(t *testing.T) {
	cases := []struct {
		file     string
		required []string
	}{
		{
			file: "news.go",
			required: []string{
				"func scopedNewsQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error)",
				"access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.News{}, access.NewsAuditFields)",
				"func GetNewsDetailForAdminContext(ctx context.Context, id string, adminID uint)",
				"func EditNewsForAdminContext(",
				"func DelNewsForAdminContext(ctx context.Context, id string, adminID uint) error",
				"func StatusNewsForAdminContext(ctx context.Context, id string, status int, adminID uint) error",
			},
		},
		{
			file: "enroll.go",
			required: []string{
				"func scopedEnrollQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error)",
				"access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Enroll{}, access.EnrollAuditFields)",
				"func GetEnrollDetailForAdminContext(ctx context.Context, id string, adminID uint)",
				"func EditEnrollForAdminContext(",
				"func DelEnrollForAdminContext(ctx context.Context, id string, adminID uint) error",
				"func ClearEnrollAllForAdminContext(ctx context.Context, id string, adminID uint) error",
				"func StatusEnrollForAdminContext(ctx context.Context, id string, status int, adminID uint) error",
			},
		},
		{
			file: "enroll_records.go",
			required: []string{
				"func ensureEnrollVisibleContext(ctx context.Context, db *gorm.DB, enrollID string, adminID uint) error",
				"func GetEnrollUserListForAdminContext(ctx context.Context, enrollID, keyword string, adminID uint)",
				"func GetEnrollJoinListForAdminContext(ctx context.Context, enrollID, keyword string, page, pageSize int, adminID uint)",
				"func DelEnrollJoinForAdminContext(ctx context.Context, id string, adminID uint) error",
				"func RemoveEnrollUserForAdminContext(ctx context.Context, enrollID, userID string, adminID uint) error",
				"func EditEnrollUserFormsForAdminContext(ctx context.Context, enrollID, userID, forms string, adminID uint) error",
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
				t.Fatalf("%s must protect admin single-record operations with %q", tc.file, snippet)
			}
		}
	}
}
