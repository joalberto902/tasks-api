package repository

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joalberto902/tasks-api/internal/model"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		status         model.TaskStatus
		description    string
		expectedErr    error
		expectedTitle  string
		expectedStatus model.TaskStatus
	}{
		{name: "empty title", title: "", status: "", description: "", expectedErr: ErrInvalidTitle},
		{name: "only space title", title: " ", status: "", description: "", expectedErr: ErrInvalidTitle},
		{name: "only breakline char title", title: "\n", status: "", description: "", expectedErr: ErrInvalidTitle},
		{name: "only tab char title", title: "\t", status: "", description: "", expectedErr: ErrInvalidTitle},
		{name: "multiple spaces title", title: "   ", status: "", description: "", expectedErr: ErrInvalidTitle},
		{name: "random invalid status", title: "title", status: "banana", description: "", expectedErr: ErrInvalidStatus},
		{name: "Uppercase invalid status", title: "title", status: "TODO", description: "", expectedErr: ErrInvalidStatus},
		{name: "valid empty status", title: "title", status: "", description: "", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusTodo},
		{name: "testing status todo", title: "title", status: model.StatusTodo, description: "", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusTodo},
		{name: "testing status ongoing", title: "title", status: model.StatusOngoing, description: "", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusOngoing},
		{name: "testing status done", title: "title", status: model.StatusDone, description: "", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusDone},
		{name: "status todo with description", title: "title", status: model.StatusTodo, description: "desc", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusTodo},
		{name: "status ongoing with description", title: "title", status: model.StatusOngoing, description: "desc", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusOngoing},
		{name: "status done with description", title: "title", status: model.StatusDone, description: "desc", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusDone},
		{name: "title with spaces and letters", title: "  title  ", status: model.StatusDone, description: "desc", expectedErr: nil, expectedTitle: "title", expectedStatus: model.StatusDone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewTaskRepository()
			startTime := time.Now().UTC()
			task, err := repo.Create(
				context.Background(),
				CreateTaskParams{
					Title:       tt.title,
					Status:      tt.status,
					Description: tt.description,
				})

			endTime := time.Now().UTC()
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("error is different than %v. got=%v", tt.expectedErr, err)
			}

			if err != nil {
				empty := model.Task{}
				if task != empty {
					t.Fatalf("task is not empty if Create is expected to error")
				}
				return
			}

			if task.ID == uuid.Nil {
				t.Fatalf("uuid is nil")
			}

			if task.ID.Version() != uuid.Version(7) {
				t.Fatalf("expected uuid v%d. got = %d", 7, task.ID.Version())
			}

			if task.CreatedAt.Before(startTime) || task.CreatedAt.After(endTime) {
				t.Fatalf("CreatedAt in the incorrect time window")
			}

			if !task.UpdatedAt.Equal(task.CreatedAt) || task.UpdatedAt.Before(startTime) || task.UpdatedAt.After(endTime) {
				t.Fatalf("UpdatedAt in the incorrect time window")
			}

			if task.CreatedAt.Location() != time.UTC {
				t.Fatalf("CreatedAt is not in utc")
			}

			if task.UpdatedAt.Location() != time.UTC {
				t.Fatalf("UpdatedAt is not in utc")
			}

			if task.Title != tt.expectedTitle {
				t.Errorf("Wrong title, expected=%s. got=%s", tt.expectedTitle, task.Title)
			}

			if task.Status != tt.expectedStatus {
				t.Errorf("Wrong status, expected=%v. got=%v", tt.expectedStatus, task.Status)
			}

			if task.Description != tt.description {
				t.Errorf("Wrong description, expected=%s. got=%s", tt.description, task.Description)
			}
		})
	}
}

func TestGet(t *testing.T) {
	//Testing Create with Get
	ctx := context.Background()
	repo := NewTaskRepository()

	expectedTask, err := repo.Create(ctx, CreateTaskParams{
		Title:       "test title",
		Status:      model.StatusTodo,
		Description: "description for get test",
	})

	if err != nil {
		t.Fatalf("Error creating new value in the repository: %v", err)
	}

	inputTask, err := repo.Get(ctx, expectedTask.ID)
	if err != nil {
		t.Fatalf("error does not equal %v. got=%v", nil, err)
	}

	if inputTask != expectedTask {
		t.Errorf("Task value is not %+v. got= %+v", expectedTask, inputTask)
	}

	_, err = repo.Get(ctx, uuid.Nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error does not equal %v. got=%v", ErrNotFound, err)
	}
}

func TestList(t *testing.T) {
	ctx := context.Background()
	repo := NewTaskRepository()

	s, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Error while generating list, %v", err)
	}

	if s == nil {
		t.Fatal("List on a empty repository should return a non-nil slice")
	}

	if len(s) != 0 {
		t.Fatal("List on a empty repository should a zero size slice")
	}

	tests := []CreateTaskParams{
		{Title: "test1", Status: model.StatusTodo},
		{Title: "test2", Status: model.StatusTodo},
		{Title: "test3", Status: model.StatusTodo},
		{Title: "test4", Status: model.StatusTodo},
		{Title: "test5", Status: model.StatusTodo},
	}

	expectedSlice := make([]model.Task, 0, 5)
	for _, tt := range tests {
		task, err := repo.Create(ctx, tt)
		if err != nil {
			t.Fatalf("Failed while creating %s with the error: %v", tt.Title, err)
		}

		expectedSlice = append(expectedSlice, task)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Error while listing: %v", err)
	}

	if !slices.Equal(expectedSlice, list) {
		t.Errorf("Wrong list slice, expected=%v. got=%v", expectedSlice, list)
	}
}
