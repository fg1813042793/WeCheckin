package event

import (
	"os"
	"strings"
	"testing"
)

func TestAdminEventOperationsUseAdminScopedQueries(t *testing.T) {
	cases := []struct {
		file     string
		required []string
	}{
		{
			file: "admin.go",
			required: []string{
				"func scopedEventQueryContext(ctx context.Context, db *gorm.DB, adminID uint) (*gorm.DB, error)",
				"access.ScopedResourceQueryByFieldsContext(ctx, db, adminID, &model.Event{}, access.EventAuditFields)",
				"func ensureEventVisibleContext(ctx context.Context, db *gorm.DB, eventID string, adminID uint) error",
				"func GetAdminEventDetailForAdminContext(ctx context.Context, id string, adminID uint)",
				"func EditEventForAdminContext(",
				"func DelEventForAdminContext(ctx context.Context, id string, adminID uint) error",
				"func StatusEventForAdminContext(ctx context.Context, id string, status int, adminID uint) error",
				"func GetEventParticipantListForAdminContext(ctx context.Context, eventID string, adminID uint)",
				"func DelEventParticipantForAdminContext(ctx context.Context, id string, adminID uint) error",
				"func EditEventParticipantForAdminContext(ctx context.Context, id, forms string, adminID uint) error",
			},
		},
		{
			file: "dynamic.go",
			required: []string{
				"func PostEventDynamicForAdminContext(ctx context.Context, eventID, userID, title, content, images, videos, addIP string, adminID uint) error",
				"func GetEventDynamicsForAdminContext(ctx context.Context, eventID string, page, pageSize int, adminID uint)",
				"func EditEventDynamicForAdminContext(ctx context.Context, id, title, content, images, videos, editIP string, adminID uint) error",
				"func DelEventDynamicForAdminContext(ctx context.Context, id string, adminID uint) error",
			},
		},
		{
			file: "score.go",
			required: []string{
				"func SaveEventScoreForAdminContext(ctx context.Context, eventID, participantID, score, judgeID string, adminID uint) error",
				"func GetEventScoresForAdminContext(ctx context.Context, eventID string, page, pageSize int, adminID uint)",
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
				t.Fatalf("%s must protect admin event operation with %q", tc.file, snippet)
			}
		}
	}
}
