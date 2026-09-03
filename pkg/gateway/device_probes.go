package gateway

import (
	"fmt"

	"github.com/yz4230/arpmon/pkg/domain"
)

type deviceProbeRow struct {
	MACAddress string
	IsPresent  bool
}

func newDeviceProbeRows(probes []*domain.DeviceProbe, devices []*domain.Device, probeID int) []deviceProbeRow {
	deviceMACAddresses := make(map[int]string, len(devices))
	for _, device := range devices {
		if device == nil {
			continue
		}
		deviceMACAddresses[device.ID] = device.MACAddress.String()
	}

	rows := make([]deviceProbeRow, 0, len(probes))
	for _, probe := range probes {
		if probe == nil || probe.ProbeID != probeID {
			continue
		}

		macAddress := deviceMACAddresses[probe.DeviceID]
		if macAddress == "" {
			macAddress = fmt.Sprintf("Device #%d", probe.DeviceID)
		}
		rows = append(rows, deviceProbeRow{
			MACAddress: macAddress,
			IsPresent:  probe.IsPresent,
		})
	}

	return rows
}
