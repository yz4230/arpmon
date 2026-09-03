package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/yz4230/arpmon/pkg/domain"
)

type MemberAttendanceRepository interface {
	Create(ctx context.Context, tx *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error)
	CreateMany(ctx context.Context, tx *sql.Tx, attendances []*domain.MemberAttendance) ([]*domain.MemberAttendance, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.MemberAttendance, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.MemberAttendance, error)
	Update(ctx context.Context, tx *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type memberAttendanceRepository struct{}

func (r *memberAttendanceRepository) Create(ctx context.Context, tx *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO member_attendances (member_id, scan_run_id, is_present)
		VALUES ($1, $2, $3)
		RETURNING id, member_id, scan_run_id, is_present, created_at, updated_at
	`, attendance.MemberID, attendance.ScanRunID, attendance.IsPresent)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.ScanRunID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberAttendanceRepository) CreateMany(ctx context.Context, tx *sql.Tx, attendances []*domain.MemberAttendance) ([]*domain.MemberAttendance, error) {
	if len(attendances) == 0 {
		return nil, nil
	}

	values := make([]string, len(attendances))
	args := make([]any, 0, len(attendances)*3)
	for i, attendance := range attendances {
		values[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args = append(args, attendance.MemberID, attendance.ScanRunID, attendance.IsPresent)
	}

	rows, err := tx.QueryContext(ctx, `
		INSERT INTO member_attendances (member_id, scan_run_id, is_present)
		VALUES `+strings.Join(values, ", ")+`
		RETURNING id, member_id, scan_run_id, is_present, created_at, updated_at
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	created := make([]*domain.MemberAttendance, 0, len(attendances))
	for rows.Next() {
		var attendance domain.MemberAttendance
		if err := rows.Scan(&attendance.ID, &attendance.MemberID, &attendance.ScanRunID, &attendance.IsPresent, &attendance.CreatedAt, &attendance.UpdatedAt); err != nil {
			return nil, err
		}
		created = append(created, &attendance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *memberAttendanceRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.MemberAttendance, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, member_id, scan_run_id, is_present, created_at, updated_at
		FROM member_attendances
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.ScanRunID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberAttendanceRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.MemberAttendance, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, member_id, scan_run_id, is_present, created_at, updated_at
		FROM member_attendances
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var attendances []*domain.MemberAttendance
	for rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.ScanRunID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		attendances = append(attendances, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attendances, nil
}

func (r *memberAttendanceRepository) Update(ctx context.Context, tx *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE member_attendances
		SET member_id = $1, scan_run_id = $2, is_present = $3
		WHERE id = $4
		RETURNING id, member_id, scan_run_id, is_present, created_at, updated_at
	`, attendance.MemberID, attendance.ScanRunID, attendance.IsPresent, attendance.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.ScanRunID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberAttendanceRepository) DeleteByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM member_attendances
		WHERE id = $1
	`, id)
	return err
}

func NewMemberAttendanceRepository() MemberAttendanceRepository {
	return &memberAttendanceRepository{}
}
