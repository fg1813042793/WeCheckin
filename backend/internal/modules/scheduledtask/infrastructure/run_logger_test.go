package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestGormRunLoggerCapsSegmentsAndSerializesSequence(t *testing.T) {
	source, err := os.ReadFile("run_logger.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, want := range []string{
		`clause.Locking{Strength: "UPDATE"}`,
		`MaxLogSegmentBytes`,
		`MaxLogRunBytes`,
		`log_sequence`,
		`utf8.ValidString`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("run logger missing %q", want)
		}
	}
}
