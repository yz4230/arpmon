package usecase

import (
	"context"
	"database/sql"
	"net"

	"github.com/samber/do"
	"github.com/yz4230/arpmon/pkg/arpscan"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
)

type ProbeUseCase interface {
	Do(ctx context.Context, target *net.IPNet) error
}

type probeUseCase struct {
	conn                  *sql.DB
	scanner               arpscan.ARPScanner
	probeRepository       repository.ProbeRepository
	deviceRepository      repository.DeviceRepository
	deviceProbeRepository repository.DeviceProbeRepository
}

func NewProbeUseCase(i *do.Injector) ProbeUseCase {
	return &probeUseCase{
		conn:                  do.MustInvoke[*sql.DB](i),
		scanner:               do.MustInvoke[arpscan.ARPScanner](i),
		probeRepository:       do.MustInvoke[repository.ProbeRepository](i),
		deviceRepository:      do.MustInvoke[repository.DeviceRepository](i),
		deviceProbeRepository: do.MustInvoke[repository.DeviceProbeRepository](i),
	}
}

func (p *probeUseCase) Do(ctx context.Context, target *net.IPNet) error {
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
		deviceMap[device.MACAddress] = device
	}
	deviceProbes := make([]*domain.DeviceProbe, 0, len(entries))
	for _, entry := range entries {
		device, ok := deviceMap[entry.MAC.String()]
		if !ok {
			device, err = p.deviceRepository.Create(ctx, tx, &domain.Device{
				MACAddress: entry.MAC.String(),
			})
			if err != nil {
				return err
			}
			deviceMap[device.MACAddress] = device
		}

		deviceProbes = append(deviceProbes, &domain.DeviceProbe{
			ProbeID:  probe.ID,
			DeviceID: device.ID,
		})
	}

	if _, err := p.deviceProbeRepository.CreateMany(ctx, tx, deviceProbes); err != nil {
		return err
	}

	return tx.Commit()
}
