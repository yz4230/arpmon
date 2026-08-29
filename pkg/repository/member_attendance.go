package repository

import (
	"context"
	"database/sql"

	"github.com/yz4230/arpmon/pkg/domain"
)

type MemberAttendanceRepository interface {
	Create(ctx context.Context, tx *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.MemberAttendance, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.MemberAttendance, error)
	Update(ctx context.Context, tx *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type memberAttendanceRepository struct{}

func (r *memberAttendanceRepository) Create(ctx context.Context, tx *sql.Tx, attendance *domain.MemberAttendance) (*domain.MemberAttendance, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO member_attendances (member_id, is_present)
		VALUES ($1, $2)
		RETURNING id, member_id, is_present, created_at, updated_at
	`, attendance.MemberID, attendance.IsPresent)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberAttendanceRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.MemberAttendance, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, member_id, is_present, created_at, updated_at
		FROM member_attendances
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		return &a, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberAttendanceRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.MemberAttendance, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, member_id, is_present, created_at, updated_at
		FROM member_attendances
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var attendances []*domain.MemberAttendance
	for rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
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
		SET member_id = $1, is_present = $2
		WHERE id = $3
		RETURNING id, member_id, is_present, created_at, updated_at
	`, attendance.MemberID, attendance.IsPresent, attendance.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var a domain.MemberAttendance
		if err := rows.Scan(&a.ID, &a.MemberID, &a.IsPresent, &a.CreatedAt, &a.UpdatedAt); err != nil {
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
