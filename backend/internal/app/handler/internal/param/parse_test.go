package param

import (
	"reflect"
	"testing"
)

func TestParseUintSlice(t *testing.T) {
	got := ParseUintSlice("1, 2,abc,0,-1,3")
	want := []uint{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseUintSlice() = %#v, want %#v", got, want)
	}
}
