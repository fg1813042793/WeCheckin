package dashboard

import (
	"testing"
	"time"
)

func TestNormalizeTrendDays(t *testing.T) {
	for _, test := range []struct {
		input int
		want  int
	}{
		{input: 7, want: 7},
		{input: 30, want: 30},
		{input: 0, want: 7},
		{input: 14, want: 7},
	} {
		if got := normalizeTrendDays(test.input); got != test.want {
			t.Fatalf("normalizeTrendDays(%d) = %d, want %d", test.input, got, test.want)
		}
	}
}

func TestBuildTrendPointsFillsDatesAndAggregatesSeries(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, location)
	millis := func(year int, month time.Month, day, hour int) int64 {
		return time.Date(year, month, day, hour, 0, 0, 0, location).UnixMilli()
	}

	points := buildTrendPoints(now, 3, dashboardTrendTimestamps{
		Users:           []int64{millis(2026, 10, 27, 8), millis(2026, 10, 29, 9)},
		EnrollJoins:     []int64{millis(2026, 10, 28, 10), millis(2026, 10, 28, 11)},
		SurveyResponses: []int64{millis(2026, 10, 29, 10)},
		ExamRecords:     []int64{millis(2026, 10, 27, 10)},
	})

	if len(points) != 3 {
		t.Fatalf("len(points) = %d, want 3", len(points))
	}
	wantDates := []string{"2026-10-27", "2026-10-28", "2026-10-29"}
	for index, want := range wantDates {
		if points[index].Date != want {
			t.Fatalf("points[%d].Date = %q, want %q", index, points[index].Date, want)
		}
	}
	if points[0].UserCount != 1 || points[0].ExamRecordCount != 1 {
		t.Fatalf("first point = %#v", points[0])
	}
	if points[1].EnrollJoinCount != 2 {
		t.Fatalf("second point = %#v", points[1])
	}
	if points[2].UserCount != 1 || points[2].SurveyResponseCount != 1 {
		t.Fatalf("last point = %#v", points[2])
	}
}
