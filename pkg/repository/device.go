package repository

import (
	"context"
	"database/sql"

	"github.com/yz4230/arpmon/pkg/domain"
)

type DeviceRepository interface {
	Create(ctx context.Context, tx *sql.Tx, device *domain.Device) (*domain.Device, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Device, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.Device, error)
	Update(ctx context.Context, tx *sql.Tx, device *domain.Device) (*domain.Device, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type deviceRepository struct{}

func (r *deviceRepository) Create(ctx context.Context, tx *sql.Tx, device *domain.Device) (*domain.Device, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO devices (mac_address)
		VALUES ($1)
		RETURNING id, mac_address, created_at, updated_at
	`, device.MACAddress)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var d domain.Device
		if err := rows.Scan(&d.ID, &d.MACAddress, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		return &d, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Device, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, mac_address, created_at, updated_at
		FROM devices
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var d domain.Device
		if err := rows.Scan(&d.ID, &d.MACAddress, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		return &d, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.Device, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, mac_address, created_at, updated_at
		FROM devices
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		var d domain.Device
		if err := rows.Scan(&d.ID, &d.MACAddress, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}

func (r *deviceRepository) Update(ctx context.Context, tx *sql.Tx, device *domain.Device) (*domain.Device, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE devices
		SET mac_address = $1
		WHERE id = $2
		RETURNING id, mac_address, created_at, updated_at
	`, device.MACAddress, device.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var d domain.Device
		if err := rows.Scan(&d.ID, &d.MACAddress, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		return &d, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceRepository) DeleteByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM devices
		WHERE id = $1
	`, id)
	return err
}

func NewDeviceRepository() DeviceRepository {
	return &deviceRepository{}
}
