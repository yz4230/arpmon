package usecase

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/yz4230/arpmon/pkg/arpscan"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/testutil"
)

func TestProbeUseCaseDoARPScan(t *testing.T) {
	existingMAC := mustParseMAC(t, "00:11:22:33:44:55")
	newMAC := mustParseMAC(t, "aa:bb:cc:dd:ee:ff")
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	scanErr := errors.New("scan failed")
	createManyErr := errors.New("create device probes failed")

	tests := []struct {
		name                string
		entries             []*arpscan.Entry
		scanErr             error
		existingDevices     []*domain.Device
		createManyErr       error
		wantErr             error
		wantCreatedDevices  []wantDevice
		wantDeviceProbes    []wantDeviceProbe
		wantCreateManyCalls int
		wantBeginCount      int
		wantCommitCount     int
		wantRollbackCount   int
	}{
		{
			name: "device exists",
			entries: []*arpscan.Entry{{
				MAC: existingMAC,
				IP:  net.ParseIP("192.168.1.10"),
			}},
			existingDevices: []*domain.Device{{
				ID:         42,
				MACAddress: existingMAC.String(),
			}},
			wantDeviceProbes: []wantDeviceProbe{{
				deviceID: 42,
				probeID:  1,
			}},
			wantCreateManyCalls: 1,
			wantBeginCount:      1,
			wantCommitCount:     1,
		},
		{
			name: "device does not exist",
			entries: []*arpscan.Entry{{
				MAC: newMAC,
				IP:  net.ParseIP("192.168.1.11"),
			}},
			wantCreatedDevices: []wantDevice{{
				id:  1,
				mac: newMAC.String(),
			}},
			wantDeviceProbes: []wantDeviceProbe{{
				deviceID: 1,
				probeID:  1,
			}},
			wantCreateManyCalls: 1,
			wantBeginCount:      1,
			wantCommitCount:     1,
		},
		{
			name: "one device exists and one device does not exist",
			entries: []*arpscan.Entry{
				{MAC: existingMAC, IP: net.ParseIP("192.168.1.10")},
				{MAC: newMAC, IP: net.ParseIP("192.168.1.11")},
			},
			existingDevices: []*domain.Device{{
				ID:         42,
				MACAddress: existingMAC.String(),
			}},
			wantCreatedDevices: []wantDevice{{
				id:  43,
				mac: newMAC.String(),
			}},
			wantDeviceProbes: []wantDeviceProbe{
				{deviceID: 42, probeID: 1},
				{deviceID: 43, probeID: 1},
			},
			wantCreateManyCalls: 1,
			wantBeginCount:      1,
			wantCommitCount:     1,
		},
		{
			name:                "no device is returned by scanner",
			entries:             nil,
			wantCreateManyCalls: 1,
			wantBeginCount:      1,
			wantCommitCount:     1,
		},
		{
			name:              "scanner returns an error",
			scanErr:           scanErr,
			wantErr:           scanErr,
			wantBeginCount:    0,
			wantCommitCount:   0,
			wantRollbackCount: 0,
		},
		{
			name: "device probe repository returns an error",
			entries: []*arpscan.Entry{{
				MAC: existingMAC,
				IP:  net.ParseIP("192.168.1.10"),
			}},
			existingDevices: []*domain.Device{{
				ID:         42,
				MACAddress: existingMAC.String(),
			}},
			createManyErr: createManyErr,
			wantErr:       createManyErr,
			wantDeviceProbes: []wantDeviceProbe{{
				deviceID: 42,
				probeID:  1,
			}},
			wantCreateManyCalls: 1,
			wantBeginCount:      1,
			wantCommitCount:     0,
			wantRollbackCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := testutil.NewInMemoryDB()
			t.Cleanup(func() { _ = testDB.Close() })

			scanner := &inMemoryARPScanner{entries: tt.entries, err: tt.scanErr}
			probeRepository := newInMemoryProbeRepository()
			deviceRepository := newInMemoryDeviceRepository(tt.existingDevices)
			deviceProbeRepository := &inMemoryDeviceProbeRepository{createManyErr: tt.createManyErr}

			uc := &probeUseCase{
				conn:                  testDB.Conn,
				scanner:               scanner,
				probeRepository:       probeRepository,
				deviceRepository:      deviceRepository,
				deviceProbeRepository: deviceProbeRepository,
			}

			gotErr := uc.Do(context.Background(), target)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Fatalf("DoARPScan() error = %v, want %v", gotErr, tt.wantErr)
			}

			if scanner.calls != 1 {
				t.Fatalf("Scan() called %d times, want 1", scanner.calls)
			}
			if scanner.target != target {
				t.Fatal("Scan() received a different target")
			}

			if tt.wantErr == nil {
				if len(probeRepository.created) != 1 {
					t.Fatalf("ProbeRepository.Create() created %d probes, want 1", len(probeRepository.created))
				}
				if got := probeRepository.created[0]; got.ID != 1 || got.Target.String() != target.String() {
					t.Fatalf("created probe = %#v, want ID 1 and target %s", got, target)
				}
			}

			assertDevices(t, deviceRepository.created, tt.wantCreatedDevices)
			if deviceProbeRepository.createManyCalls != tt.wantCreateManyCalls {
				t.Fatalf("CreateMany() called %d times, want %d", deviceProbeRepository.createManyCalls, tt.wantCreateManyCalls)
			}
			if len(deviceProbeRepository.received) > 0 {
				assertDeviceProbes(t, deviceProbeRepository.received[0], tt.wantDeviceProbes)
			}

			beginCount, commitCount, rollbackCount := testDB.Snapshot()
			if beginCount != tt.wantBeginCount || commitCount != tt.wantCommitCount || rollbackCount != tt.wantRollbackCount {
				t.Fatalf("transaction calls = begin:%d commit:%d rollback:%d, want begin:%d commit:%d rollback:%d", beginCount, commitCount, rollbackCount, tt.wantBeginCount, tt.wantCommitCount, tt.wantRollbackCount)
			}
		})
	}
}

type wantDevice struct {
	id  int
	mac string
}

type wantDeviceProbe struct {
	deviceID int
	probeID  int
}

func mustParseMAC(t *testing.T, value string) net.HardwareAddr {
	t.Helper()

	mac, err := net.ParseMAC(value)
	if err != nil {
		t.Fatalf("net.ParseMAC(%q): %v", value, err)
	}
	return mac
}

func assertDevices(t *testing.T, got []*domain.Device, want []wantDevice) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("created devices = %d, want %d", len(got), len(want))
	}
	for i, wantDevice := range want {
		if got[i].ID != wantDevice.id || got[i].MACAddress != wantDevice.mac {
			t.Errorf("created device %d = %#v, want ID %d and MAC %s", i, got[i], wantDevice.id, wantDevice.mac)
		}
	}
}

func assertDeviceProbes(t *testing.T, got []*domain.DeviceProbe, want []wantDeviceProbe) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("created device probes = %d, want %d", len(got), len(want))
	}
	for i, wantProbe := range want {
		if got[i].DeviceID != wantProbe.deviceID || got[i].ProbeID != wantProbe.probeID {
			t.Errorf("created device probe %d = %#v, want device ID %d and probe ID %d", i, got[i], wantProbe.deviceID, wantProbe.probeID)
		}
	}
}
