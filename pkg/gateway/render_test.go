package gateway

import (
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
)

func TestFormatDateTimeIncludesAbsoluteAndRelativeTime(t *testing.T) {
	value := time.Now().Add(-2 * time.Hour)

	got := formatDateTime(value)
	wantAbsolute := value.Format(time.DateTime)
	wantRelative := humanize.Time(value)

	if !strings.HasPrefix(got, wantAbsolute+" (") {
		t.Fatalf("formatDateTime() = %q, want absolute timestamp prefix %q", got, wantAbsolute)
	}
	if !strings.HasSuffix(got, ")") {
		t.Fatalf("formatDateTime() = %q, want a closing parenthesis", got)
	}
	if !strings.Contains(got, wantRelative) {
		t.Fatalf("formatDateTime() = %q, want relative timestamp %q", got, wantRelative)
	}
}
