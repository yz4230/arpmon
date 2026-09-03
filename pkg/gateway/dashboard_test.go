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

type dashboardScanRunRepositoryStub struct {
	scanRuns []*domain.ScanRun
}

func (r *dashboardScanRunRepositoryStub) Create(_ context.Context, _ *sql.Tx, scanRun *domain.ScanRun) (*domain.ScanRun, error) {
	return scanRun, nil
}

func (r *dashboardScanRunRepositoryStub) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.ScanRun, error) {
	for _, scanRun := range r.scanRuns {
		if scanRun.ID == id {
			return scanRun, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *dashboardScanRunRepositoryStub) List(_ context.Context, _ *sql.Tx) ([]*domain.ScanRun, error) {
	return r.scanRuns, nil
}

func (r *dashboardScanRunRepositoryStub) DeleteByID(_ context.Context, _ *sql.Tx, _ int) error {
	return nil
}

var _ repository.ScanRunRepository = (*dashboardScanRunRepositoryStub)(nil)

type dashboardMemberAttendanceRepositoryStub struct {
	attendances []*domain.MemberAttendance
}

func (r *dashboardMemberAttendanceRepositoryStub) Create(_ context.Context, _ *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error) {
	return attendance, nil
}

func (r *dashboardMemberAttendanceRepositoryStub) CreateMany(_ context.Context, _ *sql.Tx, attendances []*domain.MemberAttendance) ([]*domain.MemberAttendance, error) {
	return attendances, nil
}

func (r *dashboardMemberAttendanceRepositoryStub) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.MemberAttendance, error) {
	for _, attendance := range r.attendances {
		if attendance.ID == id {
			return attendance, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *dashboardMemberAttendanceRepositoryStub) List(_ context.Context, _ *sql.Tx) ([]*domain.MemberAttendance, error) {
	return r.attendances, nil
}

func (r *dashboardMemberAttendanceRepositoryStub) Update(_ context.Context, _ *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error) {
	return attendance, nil
}

func (r *dashboardMemberAttendanceRepositoryStub) DeleteByID(_ context.Context, _ *sql.Tx, _ int) error {
	return nil
}

var _ repository.MemberAttendanceRepository = (*dashboardMemberAttendanceRepositoryStub)(nil)

func newGatewayWithDashboardRepositories(t *testing.T, memberRepository repository.MemberRepository, scanRunRepository repository.ScanRunRepository, attendanceRepositories ...repository.MemberAttendanceRepository) (*echo.Echo, *testutil.InMemoryDB) {
	return newGatewayWithDashboardRepositoriesAndDevices(
		t,
		memberRepository,
		scanRunRepository,
		attendanceRepositories,
		&dashboardDeviceRepositoryStub{},
		&dashboardDeviceObservationRepositoryStub{},
	)
}

func newGatewayWithDashboardRepositoriesAndDevices(t *testing.T, memberRepository repository.MemberRepository, scanRunRepository repository.ScanRunRepository, attendanceRepositories []repository.MemberAttendanceRepository, deviceRepository repository.DeviceRepository, deviceObservationRepository repository.DeviceObservationRepository) (*echo.Echo, *testutil.InMemoryDB) {
	t.Helper()

	tdb := testutil.NewInMemoryDB()
	injector := do.New()
	do.ProvideValue(injector, tdb.Conn)
	do.ProvideValue(injector, memberRepository)
	do.ProvideValue(injector, scanRunRepository)
	do.ProvideValue(injector, deviceRepository)
	do.ProvideValue(injector, deviceObservationRepository)
	if len(attendanceRepositories) > 0 {
		do.ProvideValue(injector, attendanceRepositories[0])
	}
	do.ProvideValue(injector, usecase.ScanRunUseCase(&scanRunUseCaseStub{}))
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
	scanRunRepository := &dashboardScanRunRepositoryStub{scanRuns: []*domain.ScanRun{
		{ID: 1, TargetNetwork: target, CreatedAt: now.Add(-2 * 24 * time.Hour)},
		{ID: 2, TargetNetwork: target, CreatedAt: now.Add(-6 * 24 * time.Hour)},
		{ID: 3, TargetNetwork: target, CreatedAt: now.Add(-8 * 24 * time.Hour)},
	}}
	e, tdb := newGatewayWithDashboardRepositories(t, memberRepository, scanRunRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Dashboard",
		`data-stat="uptime">`,
		`data-stat="total-scan-run-count">3<`,
		`data-stat="member-count">2<`,
		`href="/scan-runs"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("dashboard body is missing %q", want)
		}
	}
	if strings.Contains(body, "No scanRuns found.") {
		t.Fatal("dashboard must not render the scan run list empty state")
	}
	if strings.Contains(body, "Scan runs (last 7 days)") || strings.Contains(body, "recent-scan-run-count") {
		t.Fatal("dashboard must not contain the removed recent scan-run metric")
	}
}

func TestScanRunsRouteShowsHistory(t *testing.T) {
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	memberRepository := &memberRepositoryStub{}
	scanRunRepository := &dashboardScanRunRepositoryStub{scanRuns: []*domain.ScanRun{{ID: 1, TargetNetwork: target}}}
	e, tdb := newGatewayWithDashboardRepositories(t, memberRepository, scanRunRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/scan-runs", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{"Scan runs", "192.168.1.0/24", `href="/scan-runs/1"`, `href="/scan-runs/new"`} {
		if !strings.Contains(body, want) {
			t.Errorf("scan run history body is missing %q", want)
		}
	}
	if strings.Contains(body, `<th scope="col">Updated</th>`) {
		t.Error("scan run history body includes the removed Updated column")
	}
}

func TestScanRunRouteShowsDetails(t *testing.T) {
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	memberRepository := &memberRepositoryStub{members: []*domain.Member{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}}
	scanRunRepository := &dashboardScanRunRepositoryStub{scanRuns: []*domain.ScanRun{{ID: 7, TargetNetwork: target}}}
	attendanceRepository := &dashboardMemberAttendanceRepositoryStub{attendances: []*domain.MemberAttendance{
		{ID: 1, MemberID: 1, ScanRunID: 7, IsPresent: true},
		{ID: 2, MemberID: 2, ScanRunID: 7, IsPresent: false},
		{ID: 3, MemberID: 1, ScanRunID: 8, IsPresent: false},
	}}
	deviceMAC, err := net.ParseMAC("00:11:22:33:44:55")
	if err != nil {
		t.Fatal(err)
	}
	otherDeviceMAC, err := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatal(err)
	}
	probeDeviceRepository := &dashboardDeviceRepositoryStub{devices: []*domain.Device{
		{ID: 10, MACAddress: deviceMAC},
		{ID: 11, MACAddress: otherDeviceMAC},
		{ID: 12, MACAddress: net.HardwareAddr{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc}},
	}}
	deviceObservationRepository := &dashboardDeviceObservationRepositoryStub{scanRuns: []*domain.DeviceObservation{
		{ID: 1, DeviceID: 10, ScanRunID: 7, IPAddress: net.ParseIP("192.168.1.10")},
		{ID: 2, DeviceID: 11, ScanRunID: 7, IPAddress: net.ParseIP("192.168.1.11")},
		{ID: 3, DeviceID: 12, ScanRunID: 8, IPAddress: net.ParseIP("192.168.1.10")},
	}}
	e, tdb := newGatewayWithDashboardRepositoriesAndDevices(t, memberRepository, scanRunRepository, []repository.MemberAttendanceRepository{attendanceRepository}, probeDeviceRepository, deviceObservationRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/scan-runs/7", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"Scan run #7",
		"192.168.1.0/24",
		"Member attendance",
		"Alice",
		"Bob",
		"Present",
		"Not Present",
		`<th scope="col">Member</th>`,
		"Device observations",
		deviceMAC.String(),
		otherDeviceMAC.String(),
		`<th scope="col">MAC address</th>`,
		`<th scope="col">IP address</th>`,
		`<code>192.168.1.10</code>`,
		`<code>192.168.1.11</code>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("scan run detail body is missing %q", want)
		}
	}
	if strings.Contains(body, `<th scope="col">Recorded</th>`) {
		t.Error("scan run detail body includes the removed Recorded column")
	}
	if strings.Contains(body, "12:34:56:78:9a:bc") {
		t.Error("scan run detail body includes a device scanRun from another scanRun")
	}
	if strings.Index(body, "Member attendance") > strings.Index(body, "Device observations") {
		t.Error("device scanRuns appear before member attendance")
	}
}

func TestNewScanRunRouteShowsStandalonePage(t *testing.T) {
	memberRepository := &memberRepositoryStub{members: []*domain.Member{{ID: 1, Name: "Alice"}}}
	e, tdb := newGatewayWithMemberRepository(t, memberRepository)
	t.Cleanup(func() { _ = tdb.Close() })

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/scan-runs/new", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{"Run a new scan", "Alice", `action="/scan-runs" method="post"`} {
		if !strings.Contains(body, want) {
			t.Errorf("new scanRun body is missing %q", want)
		}
	}
	if strings.Contains(body, "modal") || strings.Contains(body, "data-bs-toggle") {
		t.Fatal("new scan run page must not render modal controls")
	}
}
