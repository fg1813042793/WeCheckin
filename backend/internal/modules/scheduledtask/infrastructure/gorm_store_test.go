package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestGormStoreUsesLocksAndConditionalUpdatesForRuntimeSafety(t *testing.T) {
	source, err := os.ReadFile("gorm_store.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, want := range []string{
		`clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}`,
		`Where("run_status = ?", scheduledtaskmodel.RunStatusQueued)`,
		`Where("version = ?", expectedVersion)`,
		`gorm.ErrDuplicatedKey`,
		`func (store *GormStore) GenerateDueRuns(`,
		`func (store *GormStore) ListUndeliveredRuns(`,
		`func (store *GormStore) WakeRetryRuns(`,
		`func (store *GormStore) WakeWaitingRuns(`,
		`func (store *GormStore) RecoverStaleRuns(`,
		`func (store *GormStore) CompleteRun(`,
		`func (store *GormStore) HeartbeatRun(`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("gorm store missing %q", want)
		}
	}
}
