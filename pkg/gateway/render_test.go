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

func TestFormatUptimeRoundsToNearestSecond(t *testing.T) {
	got := formatUptime(time.Minute + 35*time.Second + 713*time.Millisecond)
	if want := "1m36s"; got != want {
		t.Fatalf("formatUptime() = %q, want %q", got, want)
	}
}

func TestDashboardTemplateIncludesStats(t *testing.T) {
	var rendered bytes.Buffer
	err := NewTemplateRenderer().ExecuteTemplate(&rendered, "index", map[string]any{
		"Title": "ARP Monitor",
		"Stats": DashboardStats{
			Uptime:          "12m34s",
			TotalProbeCount: 12,
			MemberCount:     3,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate() error = %v", err)
	}

	body := rendered.String()
	for _, want := range []string{
		"Dashboard",
		"Uptime",
		`data-stat="uptime">12m34s<`,
		`data-stat="total-probe-count">12<`,
		`data-stat="member-count">3<`,
		`href="/probes/new"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered dashboard is missing %q", want)
		}
	}
	if strings.Contains(body, "Probes (last 7 days)") || strings.Contains(body, "recent-probe-count") {
		t.Fatal("rendered dashboard must not contain the removed recent probe metric")
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
