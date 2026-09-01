package gateway

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
	"github.com/yz4230/arpmon/pkg/testutil"
	"github.com/yz4230/arpmon/pkg/usecase"
)

type dashboardProbeRepositoryStub struct {
	probes []*domain.Probe
}

func (r *dashboardProbeRepositoryStub) Create(_ context.Context, _ *sql.Tx, probe *domain.Probe) (*domain.Probe, error) {
	return probe, nil
}

func (r *dashboardProbeRepositoryStub) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.Probe, error) {
	for _, probe := range r.probes {
		if probe.ID == id {
			return probe, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *dashboardProbeRepositoryStub) List(_ context.Context, _ *sql.Tx) ([]*domain.Probe, error) {
	return r.probes, nil
}

func (r *dashboardProbeRepositoryStub) DeleteByID(_ context.Context, _ *sql.Tx, _ int) error {
	return nil
}

var _ repository.ProbeRepository = (*dashboardProbeRepositoryStub)(nil)

func newGatewayWithDashboardRepositories(t *testing.T, memberRepository repository.MemberRepository, probeRepository repository.ProbeRepository) (*echo.Echo, *testutil.InMemoryDB) {
	t.Helper()

	tdb := testutil.NewInMemoryDB()
	injector := do.New()
	do.ProvideValue(injector, tdb.Conn)
	do.ProvideValue(injector, memberRepository)
	do.ProvideValue(injector, probeRepository)
	do.ProvideValue(injector, usecase.ProbeUseCase(&probeUseCaseStub{}))
	e := echo.New()
	RegisterHandlers(e, injector)
	return e, tdb
}

func TestDashboardRouteShowsStatistics(t *testing.T) {
	now := time.Now()
	_, target, err := net.ParseCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}

	memberRepository := &memberRepositoryStub{members: []*domain.Member{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}}
	probeRepository := &dashboardProbeRepositoryStub{probes: []*domain.Probe{
		{ID: 1, Target: target, CreatedAt: now.Add(-2 * 24 * time.Hour)},
		{ID: 2, Target: target, CreatedAt: now.Add(-6 * 24 * time.Hour)},
		{ID: 3, Target: target, CreatedAt: now.Add(-8 * 24 * time.Hour)},
	}}
	e, tdb := newGatewayWithDashboardRepositories(t, memberRepository, probeRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Dashboard",
		`data-stat="member-count">2<`,
		`data-stat="recent-probe-count">2<`,
		`data-stat="total-probe-count">3<`,
		`href="/probes"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("dashboard body is missing %q", want)
		}
	}
	if strings.Contains(body, "No probes found.") {
		t.Fatal("dashboard must not render the probe list empty state")
	}
}

func TestProbesRouteShowsProbeHistory(t *testing.T) {
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	memberRepository := &memberRepositoryStub{}
	probeRepository := &dashboardProbeRepositoryStub{probes: []*domain.Probe{{ID: 1, Target: target}}}
	e, tdb := newGatewayWithDashboardRepositories(t, memberRepository, probeRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/probes", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{"Probes", "192.168.1.0/24", `href="/probes/new"`} {
		if !strings.Contains(body, want) {
			t.Errorf("probe history body is missing %q", want)
		}
	}
}

func TestNewProbeRouteShowsStandalonePage(t *testing.T) {
	memberRepository := &memberRepositoryStub{members: []*domain.Member{{ID: 1, Name: "Alice"}}}
	e, tdb := newGatewayWithMemberRepository(t, memberRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/probes/new", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{"Run a new probe", "Alice", `action="/probes" method="post"`} {
		if !strings.Contains(body, want) {
			t.Errorf("new probe body is missing %q", want)
		}
	}
	if strings.Contains(body, "modal") || strings.Contains(body, "data-bs-toggle") {
		t.Fatal("new probe page must not render modal controls")
	}
}
