package repository

import (
	"context"
	"errors"
	"fmt"
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

	if !compareTasks(t, recordedTask, task) {
		return
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

	if recordedTask.CreatedAt.Location() != time.UTC {
		t.Fatalf("wrong location at created at, expected=%v, got=%v",
			time.UTC, recordedTask.CreatedAt.Location())
	}

	if recordedTask.UpdatedAt.Location() != time.UTC {
		t.Fatalf("wrong location at created at, expected=%v, got=%v",
			time.UTC, recordedTask.UpdatedAt.Location())
	}

	if !compareTasks(t, recordedTask, task) {
		return
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

func compareTasks(t *testing.T, recordedTask model.Task, task model.Task) bool {
	t.Helper()
	if recordedTask.ID != task.ID {
		t.Errorf("wrong id recorded, expected=%v, got=%v",
			task.ID, recordedTask.ID)

		return false
	}

	if !recordedTask.CreatedAt.Equal(task.CreatedAt) {
		t.Errorf("wrong created at, expected=%v, got=%v",
			task.CreatedAt, recordedTask.CreatedAt)
		return false
	}

	if !recordedTask.UpdatedAt.Equal(task.UpdatedAt) {
		t.Errorf("wrong updated at, expected=%v, got=%v",
			task.UpdatedAt, recordedTask.UpdatedAt)
		return false
	}

	if recordedTask.Title != task.Title {
		t.Errorf("wrong title recorded, expected=%v, got=%v",
			task.Title, recordedTask.Title)
		return false
	}

	if recordedTask.Status != task.Status {
		t.Errorf("wrong status recorded, expected=%v, got=%v",
			task.Status, recordedTask.Status)
		return false
	}

	if recordedTask.Description != task.Description {
		t.Errorf("wrong description recorded, expected=%v, got=%v",
			task.Description, recordedTask.Description)
		return false
	}

	return true
}

func TestPostgresList(t *testing.T) {
	pool := newTestPool(t)
	repo := &postgresRepository{pool: pool}
	ctx := context.Background()

	testStorage := make([]model.Task, 3)
	for i := range 3 {
		task, err := repo.Create(ctx, CreateTaskParams{
			Title:       fmt.Sprintf("title %d", i),
			Status:      model.StatusOngoing,
			Description: "description",
		})
		if err != nil {
			t.Fatalf("creating task %d: %v", i, err)
		}
		t.Cleanup(func() {
			err := repo.Delete(ctx, task.ID)
			if err != nil {
				t.Errorf("failed to cleanup: %v", err)
			}
		})

		testStorage[i] = task
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("getting list: %v", err)
	}

	i := 0
	for j, recordedTask := range list {
		if i < 3 {
			task := testStorage[i]

			if recordedTask.ID == task.ID {
				if !compareTasks(t, recordedTask, task) {
					return
				}
				i++
			}
		}

		if recordedTask.CreatedAt.Location() != time.UTC {
			t.Fatalf("wrong created at time in task %d, expected=%v, got=%v",
				j, time.UTC, recordedTask.CreatedAt.Location())
		}

		if recordedTask.UpdatedAt.Location() != time.UTC {
			t.Fatalf("wrong updated at time in task %d, expected=%v, got=%v",
				j, time.UTC, recordedTask.UpdatedAt.Location())
		}
	}

	if i != 3 {
		// This is to maintain a count AND an order in the list,
		// is List skips any of the values i does not update accordingly
		t.Fatalf("wrong number of tasks tested, expected=%d, got=%d",
			3, i)
	}
}

func TestPostgresUpdate(t *testing.T) {
	tests := []struct {
		name                string
		title               *string
		status              *model.TaskStatus
		description         *string
		expectedTitle       string
		expectedStatus      model.TaskStatus
		expectedDescription string
		expectedErr         error
	}{
		{
			name:                "title changes",
			title:               new("   new  "),
			expectedTitle:       "new",
			expectedStatus:      model.StatusOngoing,
			expectedDescription: "old",
			expectedErr:         nil,
		},
		{
			name:                "empty status",
			status:              new(model.TaskStatus("")),
			expectedTitle:       "old",
			expectedStatus:      model.StatusTodo,
			expectedDescription: "old",
			expectedErr:         nil,
		},
		{
			name:                "status changes",
			status:              new(model.StatusDone),
			expectedTitle:       "old",
			expectedStatus:      model.StatusDone,
			expectedDescription: "old",
			expectedErr:         nil,
		},
		{
			name:                "empty description",
			description:         new(""),
			expectedTitle:       "old",
			expectedStatus:      model.StatusOngoing,
			expectedDescription: "",
			expectedErr:         nil,
		},
		{
			name:                "nothings changes",
			expectedTitle:       "old",
			expectedStatus:      model.StatusOngoing,
			expectedDescription: "old",
			expectedErr:         nil,
		},
		{
			name:        "invalid title",
			title:       new(""),
			expectedErr: ErrInvalidTitle,
		},
		{
			name:        "invalid status",
			status:      new(model.TaskStatus("banana")),
			expectedErr: ErrInvalidStatus,
		},
		{
			name:        "invalid id",
			expectedErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := newTestPool(t)
			repo := &postgresRepository{pool: pool}

			task, err := repo.Create(context.Background(), CreateTaskParams{
				Title:       "old",
				Status:      model.StatusOngoing,
				Description: "old",
			})
			if err != nil {
				t.Fatalf("creating task in '%s': %v", tt.name, err)
			}
			t.Cleanup(func() {
				err := repo.Delete(context.Background(), task.ID)
				if err != nil {
					t.Errorf("failed while cleaning up in test '%s': %v",
						tt.name, err)
				}
			})

			id := task.ID
			if errors.Is(tt.expectedErr, ErrNotFound) {
				id = uuid.Nil
			}

			recordedTask, err := repo.Update(context.Background(), id,
				UpdateTaskParams{
					Title:       tt.title,
					Status:      tt.status,
					Description: tt.description,
				})

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("wrong error while updating, expected=%v, got=%v",
					tt.expectedErr, err)
			}

			if err != nil {
				recordedTask, err := repo.Get(context.Background(), task.ID)
				if err != nil {
					t.Fatalf("reading task from database: %v", err)
				}

				compareTasks(t, recordedTask, task)
				return
			}

			if recordedTask.CreatedAt.Location() != time.UTC {
				t.Fatalf("wrong location in created at, expected=%v, got=%v",
					time.UTC, recordedTask.CreatedAt.Location())
			}

			if recordedTask.UpdatedAt.Location() != time.UTC {
				t.Fatalf("wrong location in updated at, expected=%v, got=%v",
					time.UTC, recordedTask.UpdatedAt.Location())
			}

			if !recordedTask.CreatedAt.Equal(task.CreatedAt) {
				t.Fatalf("wrong created at after update, expected=%v, got=%v",
					task.CreatedAt, recordedTask.CreatedAt)
			}

			getTask, err := repo.Get(context.Background(), recordedTask.ID)
			if err != nil {
				t.Fatalf("reading task from database: %v", err)
			}

			if !compareTasks(t, getTask, recordedTask) {
				return
			}

			if !getTask.UpdatedAt.After(task.UpdatedAt) {
				t.Fatal("updated time is equal or before previous time")
			}

			if getTask.Title != tt.expectedTitle {
				t.Fatalf("wrong title, expected=%s, got=%s",
					tt.expectedTitle, getTask.Title)
			}

			if getTask.Status != tt.expectedStatus {
				t.Fatalf("wrong status, expected=%v, got=%v",
					tt.expectedStatus, getTask.Status)
			}

			if getTask.Description != tt.expectedDescription {
				t.Fatalf("wrong description, expected=%v, got=%v",
					tt.expectedDescription, getTask.Description)
			}

		})
	}
}
