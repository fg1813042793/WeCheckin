package survey

import (
	"testing"
	"time"
)

func TestShouldResetSurveySession(t *testing.T) {
	now := time.Date(2026, 10, 9, 10, 0, 0, 0, time.Local).UnixMilli()
	tests := []struct {
		name    string
		startAt int64
		want    bool
	}{
		{name: "missing start", startAt: 0, want: true},
		{name: "active session", startAt: now - int64((11*time.Hour + 59*time.Minute).Milliseconds()), want: false},
		{name: "stale session", startAt: now - int64((12*time.Hour + time.Second).Milliseconds()), want: true},
		{name: "future session", startAt: now + int64(time.Minute.Milliseconds()), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldResetSurveySession(now, tt.startAt); got != tt.want {
				t.Fatalf("shouldResetSurveySession(%d, %d) = %v, want %v", now, tt.startAt, got, tt.want)
			}
		})
	}
}
