package gateway

import (
	"net"
	"testing"

	"github.com/yz4230/arpmon/internal/domain"
)

func TestNewDeviceObservationRowsFiltersScanRunAndIncludesIPAddress(t *testing.T) {
	rows := newDeviceObservationRows([]*domain.DeviceObservation{
		{DeviceID: 1, ScanRunID: 7, IPAddress: net.ParseIP("192.168.1.10")},
		{DeviceID: 2, ScanRunID: 8, IPAddress: net.ParseIP("192.168.1.11")},
	}, []*domain.Device{{ID: 1, MACAddress: net.HardwareAddr{1, 2, 3, 4, 5, 6}}}, 7)
	if len(rows) != 1 || rows[0].IPAddress != "192.168.1.10" {
		t.Fatalf("rows = %#v, want one observation with its IP address", rows)
	}
}
