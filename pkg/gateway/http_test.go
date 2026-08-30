package gateway

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/usecase"
)

type probeUseCaseStub struct {
	target *net.IPNet
	calls  int
	err    error
}

func (s *probeUseCaseStub) DoARPScan(_ context.Context, target *net.IPNet) error {
	s.calls++
	s.target = target
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
