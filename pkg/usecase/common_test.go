package usecase

import (
	"context"
	"database/sql"
	"net"

	"github.com/yz4230/arpmon/pkg/arpscan"
	"github.com/yz4230/arpmon/pkg/domain"
	"github.com/yz4230/arpmon/pkg/repository"
)

type inMemoryARPScanner struct {
	entries []*arpscan.Entry
	err     error
	calls   int
	target  *net.IPNet
}

func (s *inMemoryARPScanner) Scan(target *net.IPNet) ([]*arpscan.Entry, error) {
	s.calls++
	s.target = target
	return s.entries, s.err
}

var _ arpscan.ARPScanner = (*inMemoryARPScanner)(nil)

type inMemoryScanRunRepository struct {
	nextID  int
	created []*domain.ScanRun
}

func newInMemoryScanRunRepository() *inMemoryScanRunRepository {
	return &inMemoryScanRunRepository{nextID: 1}
}

func (r *inMemoryScanRunRepository) Create(_ context.Context, _ *sql.Tx, scanRun *domain.ScanRun) (*domain.ScanRun, error) {
	created := *scanRun
	created.ID = r.nextID
	r.nextID++
	r.created = append(r.created, &created)
	return &created, nil
}

func (r *inMemoryScanRunRepository) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.ScanRun, error) {
	for _, scanRun := range r.created {
		if scanRun.ID == id {
			created := *scanRun
			return &created, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryScanRunRepository) List(_ context.Context, _ *sql.Tx) ([]*domain.ScanRun, error) {
	scanRuns := make([]*domain.ScanRun, 0, len(r.created))
	for _, scanRun := range r.created {
		created := *scanRun
		scanRuns = append(scanRuns, &created)
	}
	return scanRuns, nil
}

func (r *inMemoryScanRunRepository) DeleteByID(_ context.Context, _ *sql.Tx, id int) error {
	for i, scanRun := range r.created {
		if scanRun.ID == id {
			r.created = append(r.created[:i], r.created[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

var _ repository.ScanRunRepository = (*inMemoryScanRunRepository)(nil)

type inMemoryDeviceRepository struct {
	nextID          int
	devices         []*domain.Device
	created         []*domain.Device
	createManyErr   error
	createManyCalls int
	received        [][]*domain.Device
}

func newInMemoryDeviceRepository(existing []*domain.Device) *inMemoryDeviceRepository {
	r := &inMemoryDeviceRepository{nextID: 1}
	for _, device := range existing {
		copy := *device
		r.devices = append(r.devices, &copy)
		if copy.ID >= r.nextID {
			r.nextID = copy.ID + 1
		}
	}
	return r
}

func (r *inMemoryDeviceRepository) Create(_ context.Context, _ *sql.Tx, device *domain.Device) (*domain.Device, error) {
	created := *device
	created.ID = r.nextID
	r.nextID++
	r.devices = append(r.devices, &created)
	r.created = append(r.created, &created)
	return &created, nil
}

func (r *inMemoryDeviceRepository) CreateMany(ctx context.Context, tx *sql.Tx, devices []*domain.Device) ([]*domain.Device, error) {
	r.createManyCalls++

	input := make([]*domain.Device, 0, len(devices))
	for _, device := range devices {
		copy := *device
		input = append(input, &copy)
	}
	r.received = append(r.received, input)

	if r.createManyErr != nil {
		return nil, r.createManyErr
	}

	created := make([]*domain.Device, 0, len(devices))
	for _, device := range devices {
		createdDevice, err := r.Create(ctx, tx, device)
		if err != nil {
			return nil, err
		}
		created = append(created, createdDevice)
	}
	return created, nil
}

func (r *inMemoryDeviceRepository) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.Device, error) {
	for _, device := range r.devices {
		if device.ID == id {
			copy := *device
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryDeviceRepository) List(_ context.Context, _ *sql.Tx) ([]*domain.Device, error) {
	devices := make([]*domain.Device, 0, len(r.devices))
	for _, device := range r.devices {
		copy := *device
		devices = append(devices, &copy)
	}
	return devices, nil
}

func (r *inMemoryDeviceRepository) Update(_ context.Context, _ *sql.Tx, device *domain.Device) (*domain.Device, error) {
	for i, current := range r.devices {
		if current.ID == device.ID {
			copy := *device
			r.devices[i] = &copy
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryDeviceRepository) DeleteByID(_ context.Context, _ *sql.Tx, id int) error {
	for i, device := range r.devices {
		if device.ID == id {
			r.devices = append(r.devices[:i], r.devices[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

var _ repository.DeviceRepository = (*inMemoryDeviceRepository)(nil)

type inMemoryDeviceObservationRepository struct {
	nextID          int
	createManyErr   error
	createManyCalls int
	received        [][]*domain.DeviceObservation
	created         []*domain.DeviceObservation
}

func (r *inMemoryDeviceObservationRepository) Create(_ context.Context, _ *sql.Tx, observation *domain.DeviceObservation) (*domain.DeviceObservation, error) {
	created := *observation
	created.ID = r.nextID
	r.nextID++
	r.created = append(r.created, &created)
	return &created, nil
}

func (r *inMemoryDeviceObservationRepository) CreateMany(_ context.Context, _ *sql.Tx, observations []*domain.DeviceObservation) ([]*domain.DeviceObservation, error) {
	r.createManyCalls++

	input := make([]*domain.DeviceObservation, 0, len(observations))
	for _, observation := range observations {
		copy := *observation
		input = append(input, &copy)
	}
	r.received = append(r.received, input)

	if r.createManyErr != nil {
		return nil, r.createManyErr
	}

	created := make([]*domain.DeviceObservation, 0, len(observations))
	for _, observation := range observations {
		copy := *observation
		copy.ID = r.nextID
		r.nextID++
		r.created = append(r.created, &copy)
		created = append(created, &copy)
	}
	return created, nil
}

func (r *inMemoryDeviceObservationRepository) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.DeviceObservation, error) {
	for _, observation := range r.created {
		if observation.ID == id {
			copy := *observation
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryDeviceObservationRepository) List(_ context.Context, _ *sql.Tx) ([]*domain.DeviceObservation, error) {
	observations := make([]*domain.DeviceObservation, 0, len(r.created))
	for _, observation := range r.created {
		copy := *observation
		observations = append(observations, &copy)
	}
	return observations, nil
}

func (r *inMemoryDeviceObservationRepository) Update(_ context.Context, _ *sql.Tx, observation *domain.DeviceObservation) (*domain.DeviceObservation, error) {
	for i, current := range r.created {
		if current.ID == observation.ID {
			copy := *observation
			r.created[i] = &copy
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryDeviceObservationRepository) DeleteByID(_ context.Context, _ *sql.Tx, id int) error {
	for i, observation := range r.created {
		if observation.ID == id {
			r.created = append(r.created[:i], r.created[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

var _ repository.DeviceObservationRepository = (*inMemoryDeviceObservationRepository)(nil)

type inMemoryMemberAttendanceRepository struct{}

func (r *inMemoryMemberAttendanceRepository) Create(_ context.Context, _ *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error) {
	created := *attendance
	created.ID = 1
	return &created, nil
}

func (r *inMemoryMemberAttendanceRepository) CreateMany(_ context.Context, _ *sql.Tx, attendances []*domain.MemberAttendance) ([]*domain.MemberAttendance, error) {
	created := make([]*domain.MemberAttendance, 0, len(attendances))
	for i, attendance := range attendances {
		copy := *attendance
		copy.ID = i + 1
		created = append(created, &copy)
	}
	return created, nil
}

func (r *inMemoryMemberAttendanceRepository) GetByID(_ context.Context, _ *sql.Tx, _ int) (*domain.MemberAttendance, error) {
	return nil, sql.ErrNoRows
}

func (r *inMemoryMemberAttendanceRepository) List(_ context.Context, _ *sql.Tx) ([]*domain.MemberAttendance, error) {
	return nil, nil
}

func (r *inMemoryMemberAttendanceRepository) Update(_ context.Context, _ *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error) {
	updated := *attendance
	return &updated, nil
}

func (r *inMemoryMemberAttendanceRepository) DeleteByID(_ context.Context, _ *sql.Tx, _ int) error {
	return nil
}

var _ repository.MemberAttendanceRepository = (*inMemoryMemberAttendanceRepository)(nil)
