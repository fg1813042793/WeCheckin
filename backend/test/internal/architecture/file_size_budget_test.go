package architecture_test

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const maxProductionGoFileLines = 700

func TestWorkflowAndPermissionPackagesKeepProductionFilesReviewable(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位架构测试文件")
	}
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	targets := []string{
		"internal/support/permission",
		"internal/modules/workflow/application",
		"internal/modules/workflow/infrastructure",
	}

	for _, target := range targets {
		target := target
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join(backendRoot, filepath.FromSlash(target))
			err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
					return nil
				}
				lines, err := countFileLines(path)
				if err != nil {
					return err
				}
				if lines > maxProductionGoFileLines {
					t.Errorf("%s 有 %d 行，超过生产 Go 文件预算 %d 行", filepath.ToSlash(path), lines, maxProductionGoFileLines)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("扫描 %s 失败: %v", target, err)
			}
		})
	}
}

func countFileLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	lines := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return lines, nil
}
