package repository

import (
	"context"
	"database/sql"

	"github.com/yz4230/arpmon/pkg/domain"
)

type MemberRepository interface {
	CreateMember(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error)
	GetMemberByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Member, error)
	ListMembers(ctx context.Context, tx *sql.Tx) ([]*domain.Member, error)
	UpdateMember(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error)
	DeleteMemberByID(ctx context.Context, tx *sql.Tx, id int) error
}

type memberRepository struct{}

func (r *memberRepository) CreateMember(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error) {
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

func (r *memberRepository) GetMemberByID(ctx context.Context, tx *sql.Tx, id int) (*domain.Member, error) {
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

func (r *memberRepository) ListMembers(ctx context.Context, tx *sql.Tx) ([]*domain.Member, error) {
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

func (r *memberRepository) UpdateMember(ctx context.Context, tx *sql.Tx, member *domain.Member) (*domain.Member, error) {
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

func (r *memberRepository) DeleteMemberByID(ctx context.Context, tx *sql.Tx, id int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM members
		WHERE id = $1
	`, id)
	return err
}

func NewMemberRepository() MemberRepository {
	return &memberRepository{}
}
