package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
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

func (pr *postgresRepository) Get(ctx context.Context, id uuid.UUID) (model.Task, error) {
	getQuery := `
	SELECT 
		id,
		created_at,
		updated_at,
		title,
		status,
		description
	FROM tasks
	WHERE id = $1;
	`

	row := pr.pool.QueryRow(ctx, getQuery, id)

	var task model.Task

	err := row.Scan(
		&task.ID,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.Title,
		&task.Status,
		&task.Description,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Task{}, ErrNotFound
	}

	if err != nil {
		return model.Task{}, err
	}

	task.CreatedAt = task.CreatedAt.UTC()
	task.UpdatedAt = task.UpdatedAt.UTC()
	return task, nil
}

func (pr *postgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	deleteQuery := `DELETE FROM tasks WHERE id = $1;`

	tag, err := pr.pool.Exec(ctx, deleteQuery, id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (pr *postgresRepository) List(ctx context.Context) ([]model.Task, error) {
	listQuery := `
	SELECT
		id, 
		created_at,
		updated_at,
		title,
		status,
		description
	FROM tasks
	ORDER BY created_at ASC, id;
	`

	rows, err := pr.pool.Query(ctx, listQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		err := rows.Scan(
			&task.ID,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.Title,
			&task.Status,
			&task.Description,
		)

		if err != nil {
			return nil, err
		}

		task.CreatedAt = task.CreatedAt.UTC()
		task.UpdatedAt = task.UpdatedAt.UTC()
		list = append(list, task)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return list, nil
}

func (pr *postgresRepository) Update(
	ctx context.Context,
	id uuid.UUID,
	arg UpdateTaskParams,
) (model.Task, error) {
	_, err := pr.Get(ctx, id)
	if err != nil {
		return model.Task{}, err
	}

	if arg.Title != nil {
		newTitle := strings.TrimSpace(*arg.Title)
		if newTitle == "" {
			return model.Task{}, ErrInvalidTitle
		}
		arg.Title = &newTitle
	}

	if arg.Status != nil {
		newStatus := *arg.Status
		if newStatus == "" {
			newStatus = model.StatusTodo
			arg.Status = &newStatus
		}

		if !newStatus.IsValid() {
			return model.Task{}, ErrInvalidStatus
		}
	}
	updateQuery := `
	UPDATE tasks
	SET updated_at = $1, 
		title = COALESCE($2, title),
		status = COALESCE($3, status),
		description = COALESCE($4, description)
	WHERE id = $5
	RETURNING 
		id,
		created_at,
		updated_at,
		title,
		status,
		description;
	`

	row := pr.pool.QueryRow(ctx, updateQuery,
		time.Now().UTC(),
		arg.Title,
		arg.Status,
		arg.Description,
		id)

	var recordedTask model.Task
	err = row.Scan(
		&recordedTask.ID,
		&recordedTask.CreatedAt,
		&recordedTask.UpdatedAt,
		&recordedTask.Title,
		&recordedTask.Status,
		&recordedTask.Description)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Task{}, ErrNotFound
	}

	if err != nil {
		return model.Task{}, err
	}

	recordedTask.CreatedAt = recordedTask.CreatedAt.UTC()
	recordedTask.UpdatedAt = recordedTask.UpdatedAt.UTC()

	return recordedTask, nil
}
