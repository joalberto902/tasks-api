package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joalberto902/tasks-api/internal/model"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("no DATABASE_URL, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("pinging database: %v", err)
	}

	return pool
}

func TestPostgresCreate(t *testing.T) {
	pool := newTestPool(t)
	repo := &postgresRepository{pool: pool}

	task, err := repo.Create(context.Background(), CreateTaskParams{
		Title: "x",
	})
	if err != nil {
		t.Fatalf("creating task: %v", err)
	}
	t.Cleanup(func() {
		deleteQuery := `DELETE FROM tasks WHERE id = $1;`
		_, err := pool.Exec(context.Background(), deleteQuery, task.ID)
		if err != nil {
			t.Errorf("failed to delete from database: %v", err)
		}
	})

	if task.ID.Version() != uuid.Version(7) {
		t.Fatalf("wrong uuid version, expected=%v, got=%v",
			uuid.Version(7), task.ID.Version())
	}

	if task.CreatedAt.Location() != time.UTC {
		t.Fatalf("wrong location at created time, expected=%v, got=%v",
			time.UTC, task.CreatedAt.Location())
	}

	if task.UpdatedAt.Location() != time.UTC {
		t.Fatalf("wrong location at updated time, expected=%v, got=%v",
			time.UTC, task.UpdatedAt.Location())
	}

	if task.Title != "x" {
		t.Fatalf("wrong title, expected=%s, got=%s",
			"x", task.Title)
	}

	if task.Status != model.StatusTodo {
		t.Fatalf("wrong status, expected=%v, got=%v",
			model.StatusTodo, task.Status)
	}

	if task.Description != "" {
		t.Fatalf("wrong description, expected=%s, got=%s",
			"", task.Description)
	}

	var recordedTask model.Task

	selectQuery := `
	SELECT id,
		created_at,
		updated_at,
		title,
		status, 
		description
	FROM tasks 
	WHERE id = $1;`
	row := pool.QueryRow(context.Background(), selectQuery, task.ID)

	if err := row.Scan(
		&recordedTask.ID,
		&recordedTask.CreatedAt,
		&recordedTask.UpdatedAt,
		&recordedTask.Title,
		&recordedTask.Status,
		&recordedTask.Description,
	); err != nil {
		t.Fatalf("scanning row: %v", err)
	}

	if recordedTask.ID != task.ID {
		t.Fatalf("wrong id recorded, expected=%v, got=%v",
			task.ID, recordedTask.ID)
	}

	if !recordedTask.CreatedAt.Equal(task.CreatedAt) {
		t.Fatalf("wrong created at, expected=%v, got=%v",
			task.CreatedAt, recordedTask.CreatedAt)
	}

	if !recordedTask.UpdatedAt.Equal(task.UpdatedAt) {
		t.Fatalf("wrong updated at, expected=%v, got=%v",
			task.UpdatedAt, recordedTask.UpdatedAt)
	}

	if recordedTask.Title != task.Title {
		t.Fatalf("wrong title recorded, expected=%v, got=%v",
			task.Title, recordedTask.Title)
	}

	if recordedTask.Status != task.Status {
		t.Fatalf("wrong status recorded, expected=%v, got=%v",
			task.Status, recordedTask.Status)
	}

	if recordedTask.Description != task.Description {
		t.Fatalf("wrong description recorded, expected=%v, got=%v",
			task.Description, recordedTask.Description)
	}
}

func TestPostgresGet(t *testing.T) {
	pool := newTestPool(t)
	repo := &postgresRepository{pool: pool}

	ctx := context.Background()
	task, err := repo.Create(ctx, CreateTaskParams{
		Title: "x",
	})
	if err != nil {
		t.Fatalf("creating task: %v", err)
	}
	t.Cleanup(func() {
		deleteQuery := `DELETE FROM tasks WHERE id=$1;`
		_, err := pool.Exec(ctx, deleteQuery, task.ID)
		if err != nil {
			t.Errorf("failed to delete from database: %v", err)
		}
	})

	if _, err = repo.Get(ctx, uuid.Nil); !errors.Is(err, ErrNotFound) {
		t.Errorf("wrong error while not existent valid id, expected=%v, got=%v",
			ErrNotFound, err)
	}

	recordedTask, err := repo.Get(ctx, task.ID)
	if err != nil {
		t.Fatalf("reading task: %v", err)
	}

	if recordedTask.ID != task.ID {
		t.Fatalf("wrong id recorded, expected=%v, got=%v",
			task.ID, recordedTask.ID)
	}

	if !recordedTask.CreatedAt.Equal(task.CreatedAt) {
		t.Fatalf("wrong created at, expected=%v, got=%v",
			task.CreatedAt, recordedTask.CreatedAt)
	}

	if recordedTask.CreatedAt.Location() != time.UTC {
		t.Fatalf("wrong location on created at, expected=%v, got=%v",
			time.UTC, task.CreatedAt.Location())
	}

	if !recordedTask.UpdatedAt.Equal(task.UpdatedAt) {
		t.Fatalf("wrong updated at, expected=%v, got=%v",
			task.UpdatedAt, recordedTask.UpdatedAt)
	}

	if recordedTask.UpdatedAt.Location() != time.UTC {
		t.Fatalf("wrong location on updated at, expected=%v, got=%v",
			time.UTC, task.UpdatedAt.Location())
	}

	if recordedTask.Title != task.Title {
		t.Fatalf("wrong title recorded, expected=%v, got=%v",
			task.Title, recordedTask.Title)
	}

	if recordedTask.Status != task.Status {
		t.Fatalf("wrong status recorded, expected=%v, got=%v",
			task.Status, recordedTask.Status)
	}

	if recordedTask.Description != task.Description {
		t.Fatalf("wrong description recorded, expected=%v, got=%v",
			task.Description, recordedTask.Description)
	}
}

func TestPostgresDelete(t *testing.T) {
	pool := newTestPool(t)
	repo := &postgresRepository{pool: pool}

	task, err := repo.Create(context.Background(), CreateTaskParams{
		Title: "x",
	})
	if err != nil {
		t.Fatalf("creating task: %v", err)
	}
	t.Cleanup(func() {
		deleteQuery := `DELETE FROM tasks WHERE id=$1;`
		_, err := pool.Exec(context.Background(), deleteQuery, task.ID)
		if err != nil {
			t.Errorf("failed to delete from database: %v", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = repo.Delete(ctx, task.ID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("wrong error with context canceled, expected=%v, got=%v",
			context.Canceled, err)
	}

	err = repo.Delete(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("failed to delete existing task, expected%v, got=%v",
			nil, err)
	}

	err = repo.Delete(context.Background(), task.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("wrong error deleting not existent task, expected=%v, got=%v",
			ErrNotFound, err)
	}

}
