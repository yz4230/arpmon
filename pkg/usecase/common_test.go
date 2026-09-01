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

type inMemoryProbeRepository struct {
	nextID  int
	created []*domain.Probe
}

func newInMemoryProbeRepository() *inMemoryProbeRepository {
	return &inMemoryProbeRepository{nextID: 1}
}

func (r *inMemoryProbeRepository) Create(_ context.Context, _ *sql.Tx, probe *domain.Probe) (*domain.Probe, error) {
	created := *probe
	created.ID = r.nextID
	r.nextID++
	r.created = append(r.created, &created)
	return &created, nil
}

func (r *inMemoryProbeRepository) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.Probe, error) {
	for _, probe := range r.created {
		if probe.ID == id {
			created := *probe
			return &created, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryProbeRepository) List(_ context.Context, _ *sql.Tx) ([]*domain.Probe, error) {
	probes := make([]*domain.Probe, 0, len(r.created))
	for _, probe := range r.created {
		created := *probe
		probes = append(probes, &created)
	}
	return probes, nil
}

func (r *inMemoryProbeRepository) DeleteByID(_ context.Context, _ *sql.Tx, id int) error {
	for i, probe := range r.created {
		if probe.ID == id {
			r.created = append(r.created[:i], r.created[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

var _ repository.ProbeRepository = (*inMemoryProbeRepository)(nil)

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

type inMemoryDeviceProbeRepository struct {
	nextID          int
	createManyErr   error
	createManyCalls int
	received        [][]*domain.DeviceProbe
	created         []*domain.DeviceProbe
}

func (r *inMemoryDeviceProbeRepository) Create(_ context.Context, _ *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error) {
	created := *probe
	created.ID = r.nextID
	r.nextID++
	r.created = append(r.created, &created)
	return &created, nil
}

func (r *inMemoryDeviceProbeRepository) CreateMany(_ context.Context, _ *sql.Tx, probes []*domain.DeviceProbe) ([]*domain.DeviceProbe, error) {
	r.createManyCalls++

	input := make([]*domain.DeviceProbe, 0, len(probes))
	for _, probe := range probes {
		copy := *probe
		input = append(input, &copy)
	}
	r.received = append(r.received, input)

	if r.createManyErr != nil {
		return nil, r.createManyErr
	}

	created := make([]*domain.DeviceProbe, 0, len(probes))
	for _, probe := range probes {
		copy := *probe
		copy.ID = r.nextID
		r.nextID++
		r.created = append(r.created, &copy)
		created = append(created, &copy)
	}
	return created, nil
}

func (r *inMemoryDeviceProbeRepository) GetByID(_ context.Context, _ *sql.Tx, id int) (*domain.DeviceProbe, error) {
	for _, probe := range r.created {
		if probe.ID == id {
			copy := *probe
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryDeviceProbeRepository) List(_ context.Context, _ *sql.Tx) ([]*domain.DeviceProbe, error) {
	probes := make([]*domain.DeviceProbe, 0, len(r.created))
	for _, probe := range r.created {
		copy := *probe
		probes = append(probes, &copy)
	}
	return probes, nil
}

func (r *inMemoryDeviceProbeRepository) Update(_ context.Context, _ *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error) {
	for i, current := range r.created {
		if current.ID == probe.ID {
			copy := *probe
			r.created[i] = &copy
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *inMemoryDeviceProbeRepository) DeleteByID(_ context.Context, _ *sql.Tx, id int) error {
	for i, probe := range r.created {
		if probe.ID == id {
			r.created = append(r.created[:i], r.created[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

var _ repository.DeviceProbeRepository = (*inMemoryDeviceProbeRepository)(nil)

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
