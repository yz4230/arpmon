package gateway

import (
	"fmt"

	"github.com/yz4230/arpmon/internal/domain"
)

type deviceObservationRow struct {
	MACAddress string
	IPAddress  string
}

func newDeviceObservationRows(observations []*domain.DeviceObservation, devices []*domain.Device, scanRunID int) []deviceObservationRow {
	deviceMACAddresses := make(map[int]string, len(devices))
	for _, device := range devices {
		if device == nil {
			continue
		}
		deviceMACAddresses[device.ID] = device.MACAddress.String()
	}

	rows := make([]deviceObservationRow, 0, len(observations))
	for _, observation := range observations {
		if observation == nil || observation.ScanRunID != scanRunID {
			continue
		}

		macAddress := deviceMACAddresses[observation.DeviceID]
		if macAddress == "" {
			macAddress = fmt.Sprintf("Device #%d", observation.DeviceID)
		}
		rows = append(rows, deviceObservationRow{
			MACAddress: macAddress,
			IPAddress:  observation.IPAddress.String(),
		})
	}

	return rows
}
