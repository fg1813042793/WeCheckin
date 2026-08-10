package dingtalkh5_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../../../../internal/service/dingtalkh5"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestDingTalkH5ServiceMirrorsHandlerPackages(t *testing.T) {
	for _, file := range []string{
		filepath.Join("auth", "service.go"),
		filepath.Join("account", "service.go"),
		filepath.Join("bootstrap", "service.go"),
		filepath.Join("config", "app_config.go"),
		filepath.Join("config", "corp_config.go"),
		filepath.Join("config", "dingtalk_oapi.go"),
		filepath.Join("performance", "review", "service.go"),
		filepath.Join("performance", "review", "reviews.go"),
		filepath.Join("performance", "template", "service.go"),
		filepath.Join("performance", "user", "service.go"),
	} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("service/dingtalkh5 must mirror handler/dingtalkh5 package %s: %v", file, err)
		}
	}

	for _, dir := range []string{
		"core",
		filepath.Join("performance", "auth"),
		filepath.Join("performance", "config"),
		filepath.Join("performance", "domain"),
		filepath.Join("performance", "notification"),
		filepath.Join("performance", "workbench"),
	} {
		if _, err := os.Stat(dir); err == nil {
			t.Fatalf("service/dingtalkh5/performance should only keep handler-aligned performance subpackages; move %s to the matching service layer", dir)
		}
	}
}
