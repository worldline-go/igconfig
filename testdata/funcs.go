package testdata

import (
	"testing"
	"time"
)

//nolint:golint
func MustParseTime(t *testing.T, s string) time.Time {
	tm, err := time.ParseInLocation(time.RFC3339, s, time.Local)
	if err != nil {
		t.Fatalf("time: %q, err: %s", s, err.Error())
	}

	return tm
}
