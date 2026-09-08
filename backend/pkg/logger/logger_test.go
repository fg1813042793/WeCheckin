package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSQLWriterWritesDedicatedDailyLog(t *testing.T) {
	originalOutput := log.Writer()
	originalFlags := log.Flags()
	originalLogger := Logger
	originalSQLLogger := sqlLogger
	t.Cleanup(func() {
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
		Logger = originalLogger
		sqlLogger = originalSQLLogger
	})

	logDir := t.TempDir()
	if err := Init(logDir, "info", 7, false); err != nil {
		t.Fatalf("initialize logger: %v", err)
	}

	marker := fmt.Sprintf("slow-query-%d", time.Now().UnixNano())
	SQLWriter().Printf("SLOW SQL %s", marker)

	filename := filepath.Join(logDir, "sql-"+time.Now().Format("2006-01-02")+".log")
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read SQL log: %v", err)
	}
	if !strings.Contains(string(content), marker) {
		t.Fatalf("SQL log does not contain marker %q: %s", marker, content)
	}
}
