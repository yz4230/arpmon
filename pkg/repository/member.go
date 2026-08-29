package repository

import (
	"context"
	"database/sql"

	"github.com/yz4230/arpmon/pkg/domain"
)

type MemberRepository interface {
	Create(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error)
	GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Member, error)
	List(ctx context.Context, tx *sql.Tx) ([]*domain.Member, error)
	Update(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error)
	DeleteByID(ctx context.Context, tx *sql.Tx, id int) error
}

type memberRepository struct{}

func (r *memberRepository) Create(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error) {
	rows, err := tx.QueryContext(ctx, `
		INSERT INTO members (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`, member.Name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		return &m, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberRepository) GetByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Member, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM members
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		return &m, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberRepository) List(ctx context.Context, tx *sql.Tx) ([]*domain.Member, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, created_at, updated_at
		FROM members
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *memberRepository) Update(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error) {
	rows, err := tx.QueryContext(ctx, `
		UPDATE members
		SET name = $1
		WHERE id = $2
		RETURNING id, name, created_at, updated_at
	`, member.Name, member.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.ID, &m.Name, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		return &m, nil
	}

	return nil, sql.ErrNoRows
}

func (r *memberRepository) DeleteByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM members
		WHERE id = $1
	`, id)
	return err
}

func NewMemberRepository() MemberRepository {
	return &memberRepository{}
}
