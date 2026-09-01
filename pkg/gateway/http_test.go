package gateway

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/usecase"
)

type probeUseCaseStub struct {
	target  *net.IPNet
	members []*domain.MemberAttendance
	calls   int
	err     error
}

func (s *probeUseCaseStub) Do(_ context.Context, target *net.IPNet, members []*domain.MemberAttendance) error {
	s.calls++
	s.target = target
	s.members = members
	return s.err
}

var _ usecase.ProbeUseCase = (*probeUseCaseStub)(nil)

func newGatewayWithProbeUseCase(t *testing.T, probeUseCase usecase.ProbeUseCase) *echo.Echo {
	t.Helper()

	injector := do.New()
	do.ProvideValue(injector, probeUseCase)
	e := echo.New()
	RegisterHandlers(e, injector)
	return e
}

func newProbeRequest(target string) *http.Request {
	form := url.Values{"target": {target}}
	return newProbeRequestWithForm(form)
}

func newProbeRequestWithMembers(target string, memberIDs, presentMemberIDs []int) *http.Request {
	form := url.Values{"target": {target}}
	for _, id := range memberIDs {
		form.Add("member_ids", strconv.Itoa(id))
	}
	for _, id := range presentMemberIDs {
		form.Add("present_member_ids", strconv.Itoa(id))
	}
	return newProbeRequestWithForm(form)
}

func newProbeRequestWithForm(form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/probes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestRunProbe(t *testing.T) {
	probeUseCase := &probeUseCaseStub{}
	e := newGatewayWithProbeUseCase(t, probeUseCase)
	req := newProbeRequest(" 192.168.1.1/24 ")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/" {
		t.Fatalf("Location = %q, want %q", got, "/")
	}
	if probeUseCase.calls != 1 {
		t.Fatalf("DoARPScan calls = %d, want 1", probeUseCase.calls)
	}
	if got := probeUseCase.target.String(); got != "192.168.1.0/24" {
		t.Fatalf("target = %q, want %q", got, "192.168.1.0/24")
	}
	if len(probeUseCase.members) != 0 {
		t.Fatalf("members = %v, want no members", probeUseCase.members)
	}
}

func TestRunProbePassesMemberAttendance(t *testing.T) {
	probeUseCase := &probeUseCaseStub{}
	e := newGatewayWithProbeUseCase(t, probeUseCase)
	req := newProbeRequestWithMembers("192.168.1.0/24", []int{1, 2}, []int{2})
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if len(probeUseCase.members) != 2 {
		t.Fatalf("members = %d, want 2", len(probeUseCase.members))
	}
	if got := probeUseCase.members[0]; got.MemberID != 1 || got.IsPresent {
		t.Fatalf("first attendance = %#v, want member 1 absent", got)
	}
	if got := probeUseCase.members[1]; got.MemberID != 2 || !got.IsPresent {
		t.Fatalf("second attendance = %#v, want member 2 present", got)
	}
}

func TestRunProbePassesMemberPresenceChoices(t *testing.T) {
	probeUseCase := &probeUseCaseStub{}
	e := newGatewayWithProbeUseCase(t, probeUseCase)
	form := url.Values{
		"target":            {"192.168.1.0/24"},
		"member_ids":        {"1", "2"},
		"member_presence_1": {"not_present"},
		"member_presence_2": {"present"},
	}
	req := newProbeRequestWithForm(form)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if len(probeUseCase.members) != 2 {
		t.Fatalf("members = %d, want 2", len(probeUseCase.members))
	}
	if got := probeUseCase.members[0]; got.MemberID != 1 || got.IsPresent {
		t.Fatalf("first attendance = %#v, want member 1 absent", got)
	}
	if got := probeUseCase.members[1]; got.MemberID != 2 || !got.IsPresent {
		t.Fatalf("second attendance = %#v, want member 2 present", got)
	}
}

func TestRunProbeRejectsInvalidMemberID(t *testing.T) {
	probeUseCase := &probeUseCaseStub{}
	e := newGatewayWithProbeUseCase(t, probeUseCase)
	form := url.Values{"target": {"192.168.1.0/24"}, "member_ids": {"not-an-id"}}
	req := newProbeRequestWithForm(form)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if probeUseCase.calls != 0 {
		t.Fatalf("Do calls = %d, want 0", probeUseCase.calls)
	}
}

func TestRunProbeHTMXRefreshesAfterSuccess(t *testing.T) {
	probeUseCase := &probeUseCaseStub{}
	e := newGatewayWithProbeUseCase(t, probeUseCase)
	req := newProbeRequest("192.168.1.0/24")
	req.Header.Set("HX-Request", "true")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("HX-Refresh"); got != "true" {
		t.Fatalf("HX-Refresh = %q, want %q", got, "true")
	}
}

func TestRunProbeRejectsInvalidTarget(t *testing.T) {
	probeUseCase := &probeUseCaseStub{}
	e := newGatewayWithProbeUseCase(t, probeUseCase)
	req := newProbeRequest("not-a-cidr")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "valid CIDR") {
		t.Fatalf("body = %q, want a CIDR validation message", recorder.Body.String())
	}
	if probeUseCase.calls != 0 {
		t.Fatalf("DoARPScan calls = %d, want 0", probeUseCase.calls)
	}
}
