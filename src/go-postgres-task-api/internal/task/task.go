package task

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Input struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type Store interface {
	Create(ctx context.Context, input Input) (Task, error)
	List(ctx context.Context) ([]Task, error)
	Get(ctx context.Context, id int64) (Task, error)
	Update(ctx context.Context, id int64, input Input) (Task, error)
	Delete(ctx context.Context, id int64) error
}
