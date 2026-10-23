package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"

	"github.com/yz4230/arpmon/internal/domain"
)

type DeviceObservationRepository interface {
	Create(ctx context.Context, tx *sql.Tx, observation *domain.DeviceObservation) (*domain.DeviceObservation, error)
	CreateMany(ctx context.Context, tx *sql.Tx, observations []*domain.DeviceObservation) ([]*domain.DeviceObservation, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.DeviceObservation, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.DeviceObservation, error)
	Update(ctx context.Context, tx *sql.Tx, observation *domain.DeviceObservation) (*domain.DeviceObservation, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type deviceObservationRepository struct{}

func (r *deviceObservationRepository) Create(ctx context.Context, tx *sql.Tx, observation *domain.DeviceObservation) (*domain.DeviceObservation, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO device_observations (device_id, scan_run_id, ip_address)
		VALUES ($1, $2, $3)
		RETURNING id, device_id, scan_run_id, ip_address, created_at, updated_at
	`, observation.DeviceID, observation.ScanRunID, observation.IPAddress.String())
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		p, err := scanDeviceObservation(rows)
		if err != nil {
			return nil, err
		}
		return p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceObservationRepository) CreateMany(ctx context.Context, tx *sql.Tx, observations []*domain.DeviceObservation) ([]*domain.DeviceObservation, error) {
	if len(observations) == 0 {
		return nil, nil
	}

	values := make([]string, len(observations))
	args := make([]any, 0, len(observations)*3)
	for i, observation := range observations {
		values[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args = append(args, observation.DeviceID, observation.ScanRunID, observation.IPAddress.String())
	}

	rows, err := tx.QueryContext(ctx, `
		INSERT INTO device_observations (device_id, scan_run_id, ip_address)
		VALUES `+strings.Join(values, ", ")+`
		RETURNING id, device_id, scan_run_id, ip_address, created_at, updated_at
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	created := make([]*domain.DeviceObservation, 0, len(observations))
	for rows.Next() {
		observation, err := scanDeviceObservation(rows)
		if err != nil {
			return nil, err
		}
		created = append(created, observation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *deviceObservationRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.DeviceObservation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, device_id, scan_run_id, ip_address, created_at, updated_at
		FROM device_observations
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		p, err := scanDeviceObservation(rows)
		if err != nil {
			return nil, err
		}
		return p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceObservationRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.DeviceObservation, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, device_id, scan_run_id, ip_address, created_at, updated_at
		FROM device_observations
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var observations []*domain.DeviceObservation
	for rows.Next() {
		p, err := scanDeviceObservation(rows)
		if err != nil {
			return nil, err
		}
		observations = append(observations, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return observations, nil
}

func (r *deviceObservationRepository) Update(ctx context.Context, tx *sql.Tx, observation *domain.DeviceObservation) (*domain.DeviceObservation, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE device_observations
		SET device_id = $1, scan_run_id = $2, ip_address = $3
		WHERE id = $4
		RETURNING id, device_id, scan_run_id, ip_address, created_at, updated_at
	`, observation.DeviceID, observation.ScanRunID, observation.IPAddress.String(), observation.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		p, err := scanDeviceObservation(rows)
		if err != nil {
			return nil, err
		}
		return p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceObservationRepository) DeleteByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM device_observations
		WHERE id = $1
	`, id)
	return err
}

func NewDeviceObservationRepository() DeviceObservationRepository {
	return &deviceObservationRepository{}
}

func scanDeviceObservation(rows rowScanner) (*domain.DeviceObservation, error) {
	var observation domain.DeviceObservation
	var ipAddress string
	if err := rows.Scan(&observation.ID, &observation.DeviceID, &observation.ScanRunID, &ipAddress, &observation.CreatedAt, &observation.UpdatedAt); err != nil {
		return nil, err
	}
	observation.IPAddress = net.ParseIP(ipAddress)
	if observation.IPAddress == nil {
		return nil, fmt.Errorf("parse observation IP address %q", ipAddress)
	}
	return &observation, nil
}
