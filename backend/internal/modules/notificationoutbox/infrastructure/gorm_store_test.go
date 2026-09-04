package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestGormStoreKeepsEnqueueIdempotentAndClaimsWithSkipLocked(t *testing.T) {
	source, err := os.ReadFile("gorm_store.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		`clause.OnConflict{Columns: []clause.Column{{Name: "idempotency_key"}}, DoNothing: true}`,
		`clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}`,
		`database.QueryContext(ctx)`,
		`notificationmodel.StatusSending, staleBefore`,
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("gorm store must contain %q", required)
		}
	}
}
