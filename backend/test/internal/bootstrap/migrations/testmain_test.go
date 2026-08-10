package migrations_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Chdir("../../../../internal/bootstrap"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
