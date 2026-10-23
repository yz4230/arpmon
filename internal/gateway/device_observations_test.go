package gateway

import (
	"context"
	"database/sql"

	"github.com/yz4230/arpmon/internal/domain"
	"github.com/yz4230/arpmon/internal/repository"
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

type dashboardDeviceObservationRepositoryStub struct {
	scanRuns []*domain.DeviceObservation
}

func (r *dashboardDeviceObservationRepositoryStub) Create(_ context.Context, _ *sql.Tx, scanRun *domain.DeviceObservation) (*domain.DeviceObservation, error) {
	return scanRun, nil
}

func (r *dashboardDeviceObservationRepositoryStub) CreateMany(_ context.Context, _ *sql.Tx, scanRuns []*domain.DeviceObservation) ([]*domain.DeviceObservation, error) {
	return scanRuns, nil
}

func (r *dashboardDeviceObservationRepositoryStub) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.DeviceObservation, error) {
	for _, scanRun := range r.scanRuns {
		if scanRun.ID == id {
			return scanRun, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *dashboardDeviceObservationRepositoryStub) List(_ context.Context, _ *sql.Tx) ([]*domain.DeviceObservation, error) {
	return r.scanRuns, nil
}

func (r *dashboardDeviceObservationRepositoryStub) Update(_ context.Context, _ *sql.Tx, scanRun *domain.DeviceObservation) (*domain.DeviceObservation, error) {
	return scanRun, nil
}

func (r *dashboardDeviceObservationRepositoryStub) DeleteByID(_ context.Context, _ *sql.Tx, _ int) error {
	return nil
}

var _ repository.DeviceObservationRepository = (*dashboardDeviceObservationRepositoryStub)(nil)
