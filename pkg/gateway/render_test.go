package gateway

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/yz4230/arpmon/pkg/domain"
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

func TestIndexTemplateIncludesMemberAttendanceChecklist(t *testing.T) {
	var rendered bytes.Buffer
	err := NewTemplateRenderer().ExecuteTemplate(&rendered, "index", map[string]any{
		"Title":  "ARP Monitor",
		"Probes": []*domain.Probe{},
		"Members": []*domain.Member{
			{ID: 1, Name: "Alice"},
		},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	body := rendered.String()
	for _, want := range []string{"Alice", `name="member_ids" value="1"`, `name="present_member_ids"`, "Who is present?"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered index is missing %q", want)
		}
	}
}
