package repository

import (
	"context"
	"database/sql"

	"github.com/yz4230/arpmon/pkg/domain"
)

type ProbeRepository interface {
	CreateProbe(ctx context.Context, tx *sql.Tx) (*domain.Probe, error)
	GetProbeByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Probe, error)
	ListProbes(ctx context.Context, tx *sql.Tx) ([]*domain.Probe, error)
	DeleteProbeByID(ctx context.Context, tx *sql.Tx, id int) error
}

type probeRepository struct{}

func (r *probeRepository) CreateProbe(ctx context.Context, tx *sql.Tx) (*domain.Probe, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO probes DEFAULT VALUES
		RETURNING id, created_at, updated_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var p domain.Probe
		if err := rows.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		return &p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *probeRepository) GetProbeByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Probe, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, created_at, updated_at
		FROM probes
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var p domain.Probe
		if err := rows.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		return &p, nil
	}

	return nil, sql.ErrNoRows
}

func (r *probeRepository) ListProbes(ctx context.Context, tx *sql.Tx) ([]*domain.Probe, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, created_at, updated_at
		FROM probes
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var probes []*domain.Probe
	for rows.Next() {
		var p domain.Probe
		if err := rows.Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		probes = append(probes, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return probes, nil
}

func (r *probeRepository) DeleteProbeByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM probes
		WHERE id = $1
	`, id)
	return err
}

func NewProbeRepository() ProbeRepository {
	return &probeRepository{}
}
