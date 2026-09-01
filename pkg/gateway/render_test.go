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

func TestDashboardTemplateIncludesStats(t *testing.T) {
	var rendered bytes.Buffer
	err := NewTemplateRenderer().ExecuteTemplate(&rendered, "index", map[string]any{
		"Title": "ARP Monitor",
		"Stats": DashboardStats{
			MemberCount:      3,
			TotalProbeCount:  12,
			RecentProbeCount: 5,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	body := rendered.String()
	for _, want := range []string{
		"Dashboard",
		"Probes (last 7 days)",
		`data-stat="member-count">3<`,
		`data-stat="recent-probe-count">5<`,
		`data-stat="total-probe-count">12<`,
		`href="/probes/new"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered dashboard is missing %q", want)
		}
	}
}

func TestNewProbeTemplateIncludesMemberAttendanceChoices(t *testing.T) {
	var rendered bytes.Buffer
	err := NewTemplateRenderer().ExecuteTemplate(&rendered, "new-probe", map[string]any{
		"Title": "ARP Monitor",
		"Members": []*domain.Member{
			{ID: 1, Name: "Alice"},
		},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	body := rendered.String()
	for _, want := range []string{
		"Alice",
		`name="member_ids" value="1"`,
		`name="member_presence_1"`,
		`type="radio"`,
		"Present",
		"Not Present",
		`action="/probes" method="post"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered new probe page is missing %q", want)
		}
	}
	if strings.Contains(body, "modal") {
		t.Fatal("rendered new probe page must not contain modal markup")
	}
}
