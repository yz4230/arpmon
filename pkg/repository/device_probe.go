package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/yz4230/arpmon/pkg/domain"
)

type DeviceProbeRepository interface {
	Create(ctx context.Context, tx *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error)
	CreateMany(ctx context.Context, tx *sql.Tx, probes []*domain.DeviceProbe) ([]*domain.DeviceProbe, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.DeviceProbe, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.DeviceProbe, error)
	Update(ctx context.Context, tx *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type deviceProbeRepository struct{}

func (r *deviceProbeRepository) Create(ctx context.Context, tx *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO device_probes (device_id, probe_id, is_present)
		VALUES ($1, $2, $3)
		RETURNING id, device_id, probe_id, is_present, created_at, updated_at
	`, probe.DeviceID, probe.ProbeID, probe.IsPresent)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var p domain.DeviceProbe
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.ProbeID, &p.IsPresent, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		return &p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceProbeRepository) CreateMany(ctx context.Context, tx *sql.Tx, probes []*domain.DeviceProbe) ([]*domain.DeviceProbe, error) {
	if len(probes) == 0 {
		return nil, nil
	}

	values := make([]string, len(probes))
	args := make([]any, 0, len(probes)*3)
	for i, probe := range probes {
		values[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args = append(args, probe.DeviceID, probe.ProbeID, probe.IsPresent)
	}

	rows, err := tx.QueryContext(ctx, `
		INSERT INTO device_probes (device_id, probe_id, is_present)
		VALUES `+strings.Join(values, ", ")+`
		RETURNING id, device_id, probe_id, is_present, created_at, updated_at
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	created := make([]*domain.DeviceProbe, 0, len(probes))
	for rows.Next() {
		var probe domain.DeviceProbe
		if err := rows.Scan(&probe.ID, &probe.DeviceID, &probe.ProbeID, &probe.IsPresent, &probe.CreatedAt, &probe.UpdatedAt); err != nil {
			return nil, err
		}
		created = append(created, &probe)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *deviceProbeRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.DeviceProbe, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, device_id, probe_id, is_present, created_at, updated_at
		FROM device_probes
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var p domain.DeviceProbe
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.ProbeID, &p.IsPresent, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		return &p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceProbeRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.DeviceProbe, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, device_id, probe_id, is_present, created_at, updated_at
		FROM device_probes
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var probes []*domain.DeviceProbe
	for rows.Next() {
		var p domain.DeviceProbe
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.ProbeID, &p.IsPresent, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		probes = append(probes, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return probes, nil
}

func (r *deviceProbeRepository) Update(ctx context.Context, tx *sql.Tx, probe *domain.DeviceProbe) (*domain.DeviceProbe, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE device_probes
		SET device_id = $1, probe_id = $2, is_present = $3
		WHERE id = $4
		RETURNING id, device_id, probe_id, is_present, created_at, updated_at
	`, probe.DeviceID, probe.ProbeID, probe.IsPresent, probe.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var p domain.DeviceProbe
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.ProbeID, &p.IsPresent, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		return &p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *deviceProbeRepository) DeleteByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM device_probes
		WHERE id = $1
	`, id)
	return err
}

func NewDeviceProbeRepository() DeviceProbeRepository {
	return &deviceProbeRepository{}
}
