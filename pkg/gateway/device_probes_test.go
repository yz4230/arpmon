package gateway

import (
	"context"
	"database/sql"

	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
)

type dashboardDeviceRepositoryStub struct {
	devices []*domain.Device
}

func (r *dashboardDeviceRepositoryStub) Create(_ context.Context, _ *sql.Tx, device *domain.Device) (*domain.Device, error) {
	return device, nil
}

func (r *dashboardDeviceRepositoryStub) CreateMany(_ context.Context, _ *sql.Tx, devices []*domain.Device) ([]*domain.Device, error) {
	return devices, nil
}

func (r *dashboardDeviceRepositoryStub) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.Device, error) {
	for _, device := range r.devices {
		if device.ID == id {
			return device, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *dashboardDeviceRepositoryStub) List(_ context.Context, _ *sql.Tx) ([]*domain.Device, error) {
	return r.devices, nil
}

func (r *dashboardDeviceRepositoryStub) Update(_ context.Context, _ *sql.Tx, device *domain.Device) (*domain.Device, error) {
	return device, nil
}

func (r *dashboardDeviceRepositoryStub) DeleteByID(_ context.Context, _ *sql.Tx, _ int) error {
	return nil
}

var _ repository.DeviceRepository = (*dashboardDeviceRepositoryStub)(nil)

type dashboardDeviceProbeRepositoryStub struct {
	probes []*domain.DeviceProbe
}

func (r *dashboardDeviceProbeRepositoryStub) Create(_ context.Context, _ *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error) {
	return probe, nil
}

func (r *dashboardDeviceProbeRepositoryStub) CreateMany(_ context.Context, _ *sql.Tx, probes []*domain.DeviceProbe) ([]*domain.DeviceProbe, error) {
	return probes, nil
}

func (r *dashboardDeviceProbeRepositoryStub) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.DeviceProbe, error) {
	for _, probe := range r.probes {
		if probe.ID == id {
			return probe, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *dashboardDeviceProbeRepositoryStub) List(_ context.Context, _ *sql.Tx) ([]*domain.DeviceProbe, error) {
	return r.probes, nil
}

func (r *dashboardDeviceProbeRepositoryStub) Update(_ context.Context, _ *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error) {
	return probe, nil
}

func (r *dashboardDeviceProbeRepositoryStub) DeleteByID(_ context.Context, _ *sql.Tx, _ int) error {
	return nil
}

var _ repository.DeviceProbeRepository = (*dashboardDeviceProbeRepositoryStub)(nil)
