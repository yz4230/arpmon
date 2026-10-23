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
	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/usecase"
)

type scanRunUseCaseStub struct {
	target  *net.IPNet
	members []*domain.MemberAttendance
	scanRun *domain.ScanRun
	calls   int
	err     error
}

func (s *scanRunUseCaseStub) Do(_ context.Context, target *net.IPNet, members []*domain.MemberAttendance) (*domain.ScanRun, error) {
	s.calls++
	s.target = target
	s.members = members
	if s.scanRun == nil {
		s.scanRun = &domain.ScanRun{ID: 1, TargetNetwork: target}
	}
	return s.scanRun, s.err
}

var _ usecase.ScanRunUseCase = (*scanRunUseCaseStub)(nil)

func newGatewayWithScanRunUseCase(t *testing.T, scanRunUseCase usecase.ScanRunUseCase) *echo.Echo {
	t.Helper()

	injector := do.New()
	do.ProvideValue(injector, scanRunUseCase)
	e := echo.New()
	RegisterHandlers(e, injector)
	return e
}

func newScanRunRequest(target string) *http.Request {
	form := url.Values{"target": {target}}
	return newScanRunRequestWithForm(form)
}

func newScanRunRequestWithMembers(target string, memberIDs, presentMemberIDs []int) *http.Request {
	form := url.Values{"target": {target}}
	for _, id := range memberIDs {
		form.Add("member_ids", strconv.Itoa(id))
	}
	for _, id := range presentMemberIDs {
		form.Add("present_member_ids", strconv.Itoa(id))
	}
	return newScanRunRequestWithForm(form)
}

func newScanRunRequestWithForm(form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/scan-runs", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestRunScanRun(t *testing.T) {
	scanRunUseCase := &scanRunUseCaseStub{scanRun: &domain.ScanRun{ID: 42}}
	e := newGatewayWithScanRunUseCase(t, scanRunUseCase)
	req := newScanRunRequest(" 192.168.1.1/24 ")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if got := recorder.Header().Get("Location"); got != "/scan-runs/42" {
		t.Fatalf("Location = %q, want %q", got, "/scan-runs/42")
	}
	if scanRunUseCase.calls != 1 {
		t.Fatalf("DoARPScan calls = %d, want 1", scanRunUseCase.calls)
	}
	if got := scanRunUseCase.target.String(); got != "192.168.1.0/24" {
		t.Fatalf("target = %q, want %q", got, "192.168.1.0/24")
	}
	if len(scanRunUseCase.members) != 0 {
		t.Fatalf("members = %v, want no members", scanRunUseCase.members)
	}
}

func TestRunScanRunPassesMemberAttendance(t *testing.T) {
	scanRunUseCase := &scanRunUseCaseStub{}
	e := newGatewayWithScanRunUseCase(t, scanRunUseCase)
	req := newScanRunRequestWithMembers("192.168.1.0/24", []int{1, 2}, []int{2})
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if len(scanRunUseCase.members) != 2 {
		t.Fatalf("members = %d, want 2", len(scanRunUseCase.members))
	}
	if got := scanRunUseCase.members[0]; got.MemberID != 1 || got.IsPresent {
		t.Fatalf("first attendance = %#v, want member 1 absent", got)
	}
	if got := scanRunUseCase.members[1]; got.MemberID != 2 || !got.IsPresent {
		t.Fatalf("second attendance = %#v, want member 2 present", got)
	}
}

func TestRunScanRunPassesMemberPresenceChoices(t *testing.T) {
	scanRunUseCase := &scanRunUseCaseStub{}
	e := newGatewayWithScanRunUseCase(t, scanRunUseCase)
	form := url.Values{
		"target":            {"192.168.1.0/24"},
		"member_ids":        {"1", "2"},
		"member_presence_1": {"not_present"},
		"member_presence_2": {"present"},
	}
	req := newScanRunRequestWithForm(form)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if len(scanRunUseCase.members) != 2 {
		t.Fatalf("members = %d, want 2", len(scanRunUseCase.members))
	}
	if got := scanRunUseCase.members[0]; got.MemberID != 1 || got.IsPresent {
		t.Fatalf("first attendance = %#v, want member 1 absent", got)
	}
	if got := scanRunUseCase.members[1]; got.MemberID != 2 || !got.IsPresent {
		t.Fatalf("second attendance = %#v, want member 2 present", got)
	}
}

func TestRunScanRunRejectsInvalidMemberID(t *testing.T) {
	scanRunUseCase := &scanRunUseCaseStub{}
	e := newGatewayWithScanRunUseCase(t, scanRunUseCase)
	form := url.Values{"target": {"192.168.1.0/24"}, "member_ids": {"not-an-id"}}
	req := newScanRunRequestWithForm(form)
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if scanRunUseCase.calls != 0 {
		t.Fatalf("Do calls = %d, want 0", scanRunUseCase.calls)
	}
}

func TestRunScanRunHTMXRedirectsAfterSuccess(t *testing.T) {
	scanRunUseCase := &scanRunUseCaseStub{scanRun: &domain.ScanRun{ID: 42}}
	e := newGatewayWithScanRunUseCase(t, scanRunUseCase)
	req := newScanRunRequest("192.168.1.0/24")
	req.Header.Set("HX-Request", "true")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("HX-Redirect"); got != "/scan-runs/42" {
		t.Fatalf("HX-Redirect = %q, want %q", got, "/scan-runs/42")
	}
	if got := recorder.Header().Get("HX-Refresh"); got != "" {
		t.Fatalf("HX-Refresh = %q, want empty", got)
	}
}

func TestRunScanRunRejectsInvalidTarget(t *testing.T) {
	scanRunUseCase := &scanRunUseCaseStub{}
	e := newGatewayWithScanRunUseCase(t, scanRunUseCase)
	req := newScanRunRequest("not-a-cidr")
	recorder := httptest.NewRecorder()

	e.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if !strings.Contains(recorder.Body.String(), "valid CIDR") {
		t.Fatalf("body = %q, want a CIDR validation message", recorder.Body.String())
	}
	if scanRunUseCase.calls != 0 {
		t.Fatalf("DoARPScan calls = %d, want 0", scanRunUseCase.calls)
	}
}
