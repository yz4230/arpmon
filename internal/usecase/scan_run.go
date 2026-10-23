package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/samber/do"
	"github.com/yz4230/arpmon/internal/arpscan"
	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/repository"
)

type ScanRunUseCase interface {
	Do(ctx context.Context, target *net.IPNet, attendances []*domain.MemberAttendance) (*domain.ScanRun, error)
}

type scanRunUseCase struct {
	conn                        *sql.DB
	scanner                     arpscan.ARPScanner
	scanRunRepository           repository.ScanRunRepository
	deviceRepository            repository.DeviceRepository
	deviceObservationRepository repository.DeviceObservationRepository
	memberAttendanceRepository  repository.MemberAttendanceRepository
}

func NewScanRunUseCase(i *do.Injector) ScanRunUseCase {
	return &scanRunUseCase{
		conn:                        do.MustInvoke[*sql.DB](i),
		scanner:                     do.MustInvoke[arpscan.ARPScanner](i),
		scanRunRepository:           do.MustInvoke[repository.ScanRunRepository](i),
		deviceRepository:            do.MustInvoke[repository.DeviceRepository](i),
		deviceObservationRepository: do.MustInvoke[repository.DeviceObservationRepository](i),
		memberAttendanceRepository:  do.MustInvoke[repository.MemberAttendanceRepository](i),
	}
}

func (u *scanRunUseCase) Do(ctx context.Context, target *net.IPNet, attendances []*domain.MemberAttendance) (*domain.ScanRun, error) {
	entries, err := u.scanner.Scan(target)
	if err != nil {
		return nil, err
	}

	tx, err := u.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	scanRun, err := u.scanRunRepository.Create(ctx, tx, &domain.ScanRun{TargetNetwork: target})
	if err != nil {
		return nil, err
	}

	devices, err := u.deviceRepository.List(ctx, tx)
	if err != nil {
		return nil, err
	}
	deviceMap := make(map[string]*domain.Device)
	for _, device := range devices {
		deviceMap[device.MACAddress.String()] = device
	}

	uniqueEntries := make([]*arpscan.Entry, 0, len(entries))
	seenMACAddresses := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		macAddress := entry.MAC.String()
		if _, seen := seenMACAddresses[macAddress]; seen {
			continue
		}
		seenMACAddresses[macAddress] = struct{}{}
		uniqueEntries = append(uniqueEntries, entry)
	}

	devicesToCreate := make([]*domain.Device, 0)
	for _, entry := range uniqueEntries {
		if _, ok := deviceMap[entry.MAC.String()]; !ok {
			devicesToCreate = append(devicesToCreate, &domain.Device{
				MACAddress: entry.MAC,
			})
		}
	}

	if len(devicesToCreate) > 0 {
		newDevices, err := u.deviceRepository.CreateMany(ctx, tx, devicesToCreate)
		if err != nil {
			return nil, err
		}
		for _, device := range newDevices {
			deviceMap[device.MACAddress.String()] = device
		}
	}

	observations := make([]*domain.DeviceObservation, 0, len(uniqueEntries))
	for _, entry := range uniqueEntries {
		device, ok := deviceMap[entry.MAC.String()]
		if !ok {
			return nil, fmt.Errorf("device with MAC %s not found", entry.MAC.String())
		}

		observations = append(observations, &domain.DeviceObservation{
			ScanRunID: scanRun.ID,
			DeviceID:  device.ID,
			IPAddress: entry.IP,
		})
	}

	if _, err := u.deviceObservationRepository.CreateMany(ctx, tx, observations); err != nil {
		return nil, err
	}

	for _, attendance := range attendances {
		attendance.ScanRunID = scanRun.ID
	}
	if _, err := u.memberAttendanceRepository.CreateMany(ctx, tx, attendances); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return scanRun, nil
}
