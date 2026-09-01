package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/arpscan"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
)

type ProbeUseCase interface {
	Do(ctx context.Context, target *net.IPNet, members []*domain.MemberAttendance) error
}

type probeUseCase struct {
	conn                       *sql.DB
	scanner                    arpscan.ARPScanner
	probeRepository            repository.ProbeRepository
	deviceRepository           repository.DeviceRepository
	deviceProbeRepository      repository.DeviceProbeRepository
	memberAttendanceRepository repository.MemberAttendanceRepository
}

func NewProbeUseCase(i *do.Injector) ProbeUseCase {
	return &probeUseCase{
		conn:                       do.MustInvoke[*sql.DB](i),
		scanner:                    do.MustInvoke[arpscan.ARPScanner](i),
		probeRepository:            do.MustInvoke[repository.ProbeRepository](i),
		deviceRepository:           do.MustInvoke[repository.DeviceRepository](i),
		deviceProbeRepository:      do.MustInvoke[repository.DeviceProbeRepository](i),
		memberAttendanceRepository: do.MustInvoke[repository.MemberAttendanceRepository](i),
	}
}

func (p *probeUseCase) Do(ctx context.Context, target *net.IPNet, members []*domain.MemberAttendance) error {
	entries, err := p.scanner.Scan(target)
	if err != nil {
		return err
	}

	tx, err := p.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	probe, err := p.probeRepository.Create(ctx, tx, &domain.Probe{Target: target})
	if err != nil {
		return err
	}

	devices, err := p.deviceRepository.List(ctx, tx)
	if err != nil {
		return err
	}
	deviceMap := make(map[string]*domain.Device)
	for _, device := range devices {
		deviceMap[device.MACAddress.String()] = device
	}

	deviceToCreate := make([]*domain.Device, 0)
	for _, entry := range entries {
		if _, ok := deviceMap[entry.MAC.String()]; !ok {
			deviceToCreate = append(deviceToCreate, &domain.Device{
				MACAddress: entry.MAC,
			})
		}
	}

	if len(deviceToCreate) > 0 {
		newDevices, err := p.deviceRepository.CreateMany(ctx, tx, deviceToCreate)
		if err != nil {
			return err
		}
		for _, device := range newDevices {
			deviceMap[device.MACAddress.String()] = device
		}
	}

	deviceProbes := make([]*domain.DeviceProbe, 0, len(entries))
	for _, entry := range entries {
		device, ok := deviceMap[entry.MAC.String()]
		if !ok {
			return fmt.Errorf("device with MAC %s not found", entry.MAC.String())
		}

		deviceProbes = append(deviceProbes, &domain.DeviceProbe{
			ProbeID:   probe.ID,
			DeviceID:  device.ID,
			IsPresent: true,
		})
	}

	if _, err := p.deviceProbeRepository.CreateMany(ctx, tx, deviceProbes); err != nil {
		return err
	}

	for _, member := range members {
		member.ProbeID = probe.ID
	}
	if _, err := p.memberAttendanceRepository.CreateMany(ctx, tx, members); err != nil {
		return err
	}

	return tx.Commit()
}
