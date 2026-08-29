package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/yz4230/arpmon/pkg/domain"
)

type ProbeRepository interface {
	Create(ctx context.Context, tx *sql.Tx, probe *domain.Probe) (*domain.Probe, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Probe, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.Probe, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type probeRepository struct{}

func (r *probeRepository) Create(ctx context.Context, tx *sql.Tx, probe *domain.Probe) (*domain.Probe, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO probes (target)
		VALUES ($1)
		RETURNING id, target, created_at, updated_at
	`, probe.Target.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanProbe(rows)
	}

	return nil, sql.ErrNoRows
}

func (r *probeRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Probe, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, target, created_at, updated_at
		FROM probes
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanProbe(rows)
	}

	return nil, sql.ErrNoRows
}

func (r *probeRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.Probe, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, target, created_at, updated_at
		FROM probes
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var probes []*domain.Probe
	for rows.Next() {
		p, err := scanProbe(rows)
		if err != nil {
			return nil, err
		}
		probes = append(probes, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return probes, nil
}

func scanProbe(rows *sql.Rows) (*domain.Probe, error) {
	var p domain.Probe
	var target string
	if err := rows.Scan(&p.ID, &target, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}

	_, network, err := net.ParseCIDR(target)
	if err != nil {
		return nil, fmt.Errorf("parse probe target %q: %w", target, err)
	}
	p.Target = network
	return &p, nil
}

func (r *probeRepository) DeleteByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM probes
		WHERE id = $1
	`, id)
	return err
}

func NewProbeRepository() ProbeRepository {
	return &probeRepository{}
}
