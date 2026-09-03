package usecase

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"

	"github.com/yz4230/arpmon/pkg/arpscan"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/testutil"
)

func TestScanRunUseCaseDo(t *testing.T) {
	existingMAC := mustParseMAC(t, "00:11:22:33:44:55")
	newMAC := mustParseMAC(t, "aa:bb:cc:dd:ee:ff")
	anotherNewMAC := mustParseMAC(t, "12:34:56:78:9a:bc")
	_, target, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}

	scanErr := errors.New("scan failed")
	createDevicesErr := errors.New("create devices failed")
	createDeviceObservationsErr := errors.New("create device observations failed")

	tests := []struct {
		name                           string
		entries                        []*arpscan.Entry
		scanErr                        error
		existingDevices                []*domain.Device
		createDevicesErr               error
		createDeviceObservationsErr    error
		wantErr                        error
		wantCreatedDevices             []wantDevice
		wantNewDeviceMACs              []net.HardwareAddr
		wantDeviceObservations         []wantDeviceObservation
		wantDeviceCreateManyCalls      int
		wantObservationCreateManyCalls int
		wantBeginCount                 int
		wantCommitCount                int
		wantRollbackCount              int
	}{
		{
			name: "device exists",
			entries: []*arpscan.Entry{{
				MAC: existingMAC,
				IP:  net.ParseIP("192.168.1.10"),
			}},
			existingDevices: []*domain.Device{{
				ID:         42,
				MACAddress: existingMAC,
			}},
			wantDeviceObservations: []wantDeviceObservation{{
				deviceID:  42,
				scanRunID: 1,
				ipAddress: net.ParseIP("192.168.1.10"),
			}},
			wantObservationCreateManyCalls: 1,
			wantBeginCount:                 1,
			wantCommitCount:                1,
		},
		{
			name: "device does not exist",
			entries: []*arpscan.Entry{{
				MAC: newMAC,
				IP:  net.ParseIP("192.168.1.11"),
			}},
			wantCreatedDevices: []wantDevice{{
				id:  1,
				mac: newMAC,
			}},
			wantDeviceObservations: []wantDeviceObservation{{
				deviceID:  1,
				scanRunID: 1,
				ipAddress: net.ParseIP("192.168.1.11"),
			}},
			wantNewDeviceMACs:              []net.HardwareAddr{newMAC},
			wantDeviceCreateManyCalls:      1,
			wantObservationCreateManyCalls: 1,
			wantBeginCount:                 1,
			wantCommitCount:                1,
		},
		{
			name: "one device exists and two devices do not exist",
			entries: []*arpscan.Entry{
				{MAC: existingMAC, IP: net.ParseIP("192.168.1.10")},
				{MAC: newMAC, IP: net.ParseIP("192.168.1.11")},
				{MAC: anotherNewMAC, IP: net.ParseIP("192.168.1.12")},
			},
			existingDevices: []*domain.Device{{
				ID:         42,
				MACAddress: existingMAC,
			}},
			wantCreatedDevices: []wantDevice{
				{id: 43, mac: newMAC},
				{id: 44, mac: anotherNewMAC},
			},
			wantDeviceObservations: []wantDeviceObservation{
				{deviceID: 42, scanRunID: 1, ipAddress: net.ParseIP("192.168.1.10")},
				{deviceID: 43, scanRunID: 1, ipAddress: net.ParseIP("192.168.1.11")},
				{deviceID: 44, scanRunID: 1, ipAddress: net.ParseIP("192.168.1.12")},
			},
			wantNewDeviceMACs:              []net.HardwareAddr{newMAC, anotherNewMAC},
			wantDeviceCreateManyCalls:      1,
			wantObservationCreateManyCalls: 1,
			wantBeginCount:                 1,
			wantCommitCount:                1,
		},
		{
			name: "duplicate scanner entries are stored once",
			entries: []*arpscan.Entry{
				{MAC: newMAC, IP: net.ParseIP("192.168.1.11")},
				{MAC: newMAC, IP: net.ParseIP("192.168.1.11")},
			},
			wantCreatedDevices: []wantDevice{{id: 1, mac: newMAC}},
			wantDeviceObservations: []wantDeviceObservation{{
				deviceID:  1,
				scanRunID: 1,
				ipAddress: net.ParseIP("192.168.1.11"),
			}},
			wantNewDeviceMACs:              []net.HardwareAddr{newMAC},
			wantDeviceCreateManyCalls:      1,
			wantObservationCreateManyCalls: 1,
			wantBeginCount:                 1,
			wantCommitCount:                1,
		},
		{
			name:                           "no device is returned by scanner",
			entries:                        nil,
			wantObservationCreateManyCalls: 1,
			wantBeginCount:                 1,
			wantCommitCount:                1,
		},
		{
			name:           "scanner returns an error",
			scanErr:        scanErr,
			wantErr:        scanErr,
			wantBeginCount: 0,
		},
		{
			name: "device repository returns an error",
			entries: []*arpscan.Entry{{
				MAC: newMAC,
				IP:  net.ParseIP("192.168.1.11"),
			}},
			createDevicesErr:          createDevicesErr,
			wantErr:                   createDevicesErr,
			wantNewDeviceMACs:         []net.HardwareAddr{newMAC},
			wantDeviceCreateManyCalls: 1,
			wantBeginCount:            1,
			wantRollbackCount:         1,
		},
		{
			name: "device observation repository returns an error",
			entries: []*arpscan.Entry{{
				MAC: existingMAC,
				IP:  net.ParseIP("192.168.1.10"),
			}},
			existingDevices: []*domain.Device{{
				ID:         42,
				MACAddress: existingMAC,
			}},
			createDeviceObservationsErr:    createDeviceObservationsErr,
			wantErr:                        createDeviceObservationsErr,
			wantDeviceObservations:         []wantDeviceObservation{{deviceID: 42, scanRunID: 1, ipAddress: net.ParseIP("192.168.1.10")}},
			wantObservationCreateManyCalls: 1,
			wantBeginCount:                 1,
			wantRollbackCount:              1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := testutil.NewInMemoryDB()
			t.Cleanup(func() { _ = testDB.Close() })

			scanner := &inMemoryARPScanner{entries: tt.entries, err: tt.scanErr}
			scanRunRepository := newInMemoryScanRunRepository()
			deviceRepository := newInMemoryDeviceRepository(tt.existingDevices)
			deviceRepository.createManyErr = tt.createDevicesErr
			observationRepository := &inMemoryDeviceObservationRepository{createManyErr: tt.createDeviceObservationsErr}

			uc := &scanRunUseCase{
				conn:                        testDB.Conn,
				scanner:                     scanner,
				scanRunRepository:           scanRunRepository,
				deviceRepository:            deviceRepository,
				deviceObservationRepository: observationRepository,
				memberAttendanceRepository:  &inMemoryMemberAttendanceRepository{},
			}

			gotScanRun, gotErr := uc.Do(context.Background(), target, nil)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Fatalf("Do() error = %v, want %v", gotErr, tt.wantErr)
			}

			if scanner.calls != 1 {
				t.Fatalf("Scan() called %d times, want 1", scanner.calls)
			}
			if scanner.target != target {
				t.Fatal("Scan() received a different target")
			}

			if tt.wantErr == nil {
				if gotScanRun == nil || gotScanRun.ID != 1 || gotScanRun.TargetNetwork.String() != target.String() {
					t.Fatalf("returned scan run = %#v, want ID 1 and target network %s", gotScanRun, target)
				}
				if len(scanRunRepository.created) != 1 {
					t.Fatalf("ScanRunRepository.Create() created %d scan runs, want 1", len(scanRunRepository.created))
				}
				if got := scanRunRepository.created[0]; got.ID != 1 || got.TargetNetwork.String() != target.String() {
					t.Fatalf("created scan run = %#v, want ID 1 and target network %s", got, target)
				}
			}

			assertDevices(t, deviceRepository.created, tt.wantCreatedDevices)
			if deviceRepository.createManyCalls != tt.wantDeviceCreateManyCalls {
				t.Fatalf("DeviceRepository.CreateMany() called %d times, want %d", deviceRepository.createManyCalls, tt.wantDeviceCreateManyCalls)
			}
			if deviceRepository.createManyCalls > 0 {
				assertDeviceMACs(t, deviceRepository.received[0], tt.wantNewDeviceMACs)
			}

			if observationRepository.createManyCalls != tt.wantObservationCreateManyCalls {
				t.Fatalf("DeviceObservationRepository.CreateMany() called %d times, want %d", observationRepository.createManyCalls, tt.wantObservationCreateManyCalls)
			}
			if observationRepository.createManyCalls > 0 {
				assertDeviceObservations(t, observationRepository.received[0], tt.wantDeviceObservations)
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
	mac net.HardwareAddr
}

type wantDeviceObservation struct {
	deviceID  int
	scanRunID int
	ipAddress net.IP
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
		if got[i].ID != wantDevice.id || !bytes.Equal(got[i].MACAddress, wantDevice.mac) {
			t.Errorf("created device %d = %#v, want ID %d and MAC %s", i, got[i], wantDevice.id, wantDevice.mac.String())
		}
	}
}

func assertDeviceMACs(t *testing.T, got []*domain.Device, want []net.HardwareAddr) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("devices passed to CreateMany = %d, want %d", len(got), len(want))
	}
	for i, wantMAC := range want {
		if !bytes.Equal(got[i].MACAddress, wantMAC) {
			t.Errorf("device passed to CreateMany at index %d has MAC %q, want %q", i, got[i].MACAddress.String(), wantMAC.String())
		}
	}
}

func assertDeviceObservations(t *testing.T, got []*domain.DeviceObservation, want []wantDeviceObservation) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("created device observations = %d, want %d", len(got), len(want))
	}
	for i, wantObservation := range want {
		if got[i].DeviceID != wantObservation.deviceID || got[i].ScanRunID != wantObservation.scanRunID || !got[i].IPAddress.Equal(wantObservation.ipAddress) {
			t.Errorf("created device observation %d = %#v, want device ID %d, scan run ID %d, and IP address %s", i, got[i], wantObservation.deviceID, wantObservation.scanRunID, wantObservation.ipAddress)
		}
	}
}
