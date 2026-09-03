package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/yz4230/arpmon/pkg/domain"
)

type ScanRunRepository interface {
	Create(ctx context.Context, tx *sql.Tx, scanRun *domain.ScanRun) (*domain.ScanRun, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.ScanRun, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.ScanRun, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type scanRunRepository struct{}

func (r *scanRunRepository) Create(ctx context.Context, tx *sql.Tx, scanRun *domain.ScanRun) (*domain.ScanRun, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO scan_runs (target_network)
		VALUES ($1)
		RETURNING id, target_network, created_at, updated_at
	`, scanRun.TargetNetwork.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanScanRun(rows)
	}

	return nil, sql.ErrNoRows
}

func (r *scanRunRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.ScanRun, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, target_network, created_at, updated_at
		FROM scan_runs
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return scanScanRun(rows)
	}

	return nil, sql.ErrNoRows
}

func (r *scanRunRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.ScanRun, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, target_network, created_at, updated_at
		FROM scan_runs
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scanRuns []*domain.ScanRun
	for rows.Next() {
		p, err := scanScanRun(rows)
		if err != nil {
			return nil, err
		}
		scanRuns = append(scanRuns, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return scanRuns, nil
}

func scanScanRun(rows rowScanner) (*domain.ScanRun, error) {
	var p domain.ScanRun
	var target string
	if err := rows.Scan(&p.ID, &target, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}

	_, network, err := net.ParseCIDR(target)
	if err != nil {
		return nil, fmt.Errorf("parse scan run target network %q: %w", target, err)
	}
	p.TargetNetwork = network
	return &p, nil
}

func (r *scanRunRepository) DeleteByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM scan_runs
		WHERE id = $1
	`, id)
	return err
}

func NewScanRunRepository() ScanRunRepository {
	return &scanRunRepository{}
}
