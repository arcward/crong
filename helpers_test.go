package crong

import (
	"strings"
	"testing"
)

// assertEqual is a helper function to compare two values
func assertEqual[V comparable](t testing.TB, val V, expected V) {
	t.Helper()
	if val != expected {
		t.Errorf("expected %v, got %v", expected, val)
	}
}

func slicesEqual(t testing.TB, val []int, expect []int) bool {
	t.Helper()
	if len(val) != len(expect) {
		return false
	}
	for _, v := range expect {
		if !hasValue(t, val, v) {
			return false
		}
	}
	return true
}

func hasValue(t testing.TB, a []int, expect int) bool {
	t.Helper()
	for _, v := range a {
		if v == expect {
			return true
		}
	}
	return false
}

func requireErr(t testing.TB, err error, msg ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error (%s)", strings.Join(msg, "- \n"))
	}
}
