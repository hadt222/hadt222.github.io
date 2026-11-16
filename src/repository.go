package task

import (
	"context"
	"database/sql"
	"errors"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (Task, error) {
	var t Task
	err := s.Scan(
		&t.ID, &t.Title, &t.Description, &t.Completed,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return t, err
}

func (s *PostgresStore) Create(ctx context.Context, in Input) (Task, error) {
	const query = `
        INSERT INTO tasks (title, description, completed)
        VALUES ($1, $2, $3)
        RETURNING id, title, description, completed, created_at, updated_at`
	return scanTask(s.db.QueryRowContext(ctx, query, in.Title, in.Description, in.Completed))
}

func (s *PostgresStore) List(ctx context.Context) ([]Task, error) {
	const query = `
        SELECT id, title, description, completed, created_at, updated_at
        FROM tasks ORDER BY id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PostgresStore) Get(ctx context.Context, id int64) (Task, error) {
	const query = `
        SELECT id, title, description, completed, created_at, updated_at
        FROM tasks WHERE id = $1`
	return scanTask(s.db.QueryRowContext(ctx, query, id))
}

func (s *PostgresStore) Update(ctx context.Context, id int64, in Input) (Task, error) {
	const query = `
        UPDATE tasks
        SET title = $2, description = $3, completed = $4, updated_at = NOW()
        WHERE id = $1
        RETURNING id, title, description, completed, created_at, updated_at`
	return scanTask(s.db.QueryRowContext(ctx, query, id, in.Title, in.Description, in.Completed))
}

func (s *PostgresStore) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}
