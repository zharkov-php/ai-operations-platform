package analytics

import (
	"errors"
	"testing"
	"time"
)

func TestParseRangeUsesTimezoneBoundaries(t *testing.T) {
	result, err := ParseRange("2026-08-06", "2026-08-07", "Atlantic/Canary", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if result.From.Format(time.RFC3339) != "2026-08-05T23:00:00Z" || result.To.Format(time.RFC3339) != "2026-08-06T23:00:00Z" {
		t.Fatalf("range=%s..%s", result.From, result.To)
	}
}
func TestParseRangeHandlesDST(t *testing.T) {
	result, err := ParseRange("2026-03-29", "2026-03-30", "Europe/London", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if result.To.Sub(result.From) != 23*time.Hour {
		t.Fatalf("duration=%s", result.To.Sub(result.From))
	}
}
func TestParseRangeRejectsInvalidOrUnbounded(t *testing.T) {
	for _, test := range [][3]string{{"bad", "2026-01-02", "UTC"}, {"2026-01-02", "2026-01-01", "UTC"}, {"2020-01-01", "2026-01-01", "UTC"}, {"2026-01-01", "2026-01-02", "Mars/Base"}} {
		if _, err := ParseRange(test[0], test[1], test[2], time.Now()); !errors.Is(err, ErrInvalidRange) {
			t.Fatalf("input=%v err=%v", test, err)
		}
	}
}
