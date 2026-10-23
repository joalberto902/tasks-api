package repository

import (
	"context"
	"github.com/google/uuid"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joalberto902/tasks-api/internal/model"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

func (pr *postgresRepository) dbCreate(
	ctx context.Context,
	task model.Task,
) (model.Task, error) {
	query := `
	INSERT INTO tasks(
		id,
		created_at,
		updated_at,
		title,
		status,
		description
	)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6
	)
	RETURNING 
		id,
		created_at,
		updated_at,
		title,
		status,
		description;`

	row := pr.pool.QueryRow(ctx, query,
		task.ID, task.CreatedAt, task.UpdatedAt, task.Title, task.Status, task.Description)

	var newTask model.Task
	if err := row.Scan(
		&newTask.ID,
		&newTask.CreatedAt,
		&newTask.UpdatedAt,
		&newTask.Title,
		&newTask.Status,
		&newTask.Description,
	); err != nil {
		return model.Task{}, err
	}

	newTask.CreatedAt = newTask.CreatedAt.UTC()
	newTask.UpdatedAt = newTask.UpdatedAt.UTC()

	return newTask, nil
}

func (pr *postgresRepository) Create(
	ctx context.Context,
	arg CreateTaskParams,
) (model.Task, error) {
	arg.Title = strings.TrimSpace(arg.Title)
	if arg.Title == "" {
		return model.Task{}, ErrInvalidTitle
	}

	if arg.Status == "" {
		arg.Status = model.StatusTodo
	}

	if !arg.Status.IsValid() {
		return model.Task{}, ErrInvalidStatus
	}

	// Generating new id
	id, err := uuid.NewV7()
	if err != nil {
		return model.Task{}, err
	}

	createdAt := time.Now().UTC()
	updatedAt := createdAt

	task, err := pr.dbCreate(ctx, model.Task{
		ID:          id,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Title:       arg.Title,
		Status:      arg.Status,
		Description: arg.Description,
	})

	if err != nil {
		return model.Task{}, err
	}

	return task, nil
}
