package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainInstallsStrictClientIPResolver(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "h.SetClientIPFunc(middleware.ClientIPResolver())") {
		t.Fatal("main server must install the trusted-proxy client IP resolver globally")
	}
}
