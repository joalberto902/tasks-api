package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/joalberto902/tasks-api/internal/model"
	"github.com/joalberto902/tasks-api/internal/repository"
)

func TestList(t *testing.T) {
	repo := repository.NewTaskRepository()
	th := NewTaskHandler(repo)

	r := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()

	th.List(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("wrong status code for empty repository, expected=%d, got=%d",
			http.StatusOK, resp.StatusCode)
	}

	contentHeader := resp.Header.Get("Content-Type")
	if contentHeader != "application/json" {
		t.Fatalf("Content-Type header is wrong for empty repository, expected=%s, got=%s",
			"application/json", contentHeader)
	}

	var listEmpty []model.Task
	if err := json.NewDecoder(resp.Body).Decode(&listEmpty); err != nil {
		t.Fatalf("failed to decode payload of empty list request: %v", err)
	}

	if listEmpty == nil {
		t.Fatalf("empty list is wrong, expected=%v, got=%v", []model.Task{}, listEmpty)
	}

	// Adding two cases
	ctx := context.Background()
	task1, err := repo.Create(ctx, repository.CreateTaskParams{Title: "test1"})
	if err != nil {
		t.Fatalf("failed to create task 1 for tests: %v", err)
	}
	task2, err := repo.Create(ctx, repository.CreateTaskParams{Title: "test2"})
	if err != nil {
		t.Fatalf("failed to create task 2 for tests: %v", err)
	}
	expectedList := []model.Task{task1, task2}

	newReq := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	newW := httptest.NewRecorder()

	th.List(newW, newReq)

	newResp := newW.Result()
	defer newResp.Body.Close()

	if newResp.StatusCode != http.StatusOK {
		t.Fatalf("wrong status code for filled repository, expected=%d, got=%d",
			http.StatusOK, newResp.StatusCode)
	}

	contentHeader = newResp.Header.Get("Content-Type")
	if contentHeader != "application/json" {
		t.Fatalf("Content-Type header is wrong for filled repository, expected=%s, got=%s",
			"application/json", contentHeader)
	}

	var resultList []model.Task
	if err := json.NewDecoder(newResp.Body).Decode(&resultList); err != nil {
		t.Fatalf("failed to decode payload for filled list request: %v", err)
	}

	if len(resultList) != 2 {
		t.Fatalf("Size of the filled list is wrong, expected=%d, got=%d",
			2, len(resultList))
	}

	for i, task := range resultList {
		if task.ID != expectedList[i].ID {
			t.Errorf("Task %d received is different from the created, expected=%+v, got=%+v",
				i+1, expectedList[i].ID, task.ID)
		}
	}
}

func TestGet(t *testing.T) {
	repo := repository.NewTaskRepository()
	th := NewTaskHandler(repo)

	task, err := repo.Create(context.Background(), repository.CreateTaskParams{
		Title: "test task",
	})
	if err != nil {
		t.Fatalf("failed to create the test task: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks/{id}", th.Get)

	tests := []struct {
		name            string
		url             string
		expectedCode    int
		expectedTask    uuid.UUID
		expectedMessage string
	}{
		{
			name:         "existent valid id",
			url:          fmt.Sprintf("/tasks/%s", task.ID.String()),
			expectedCode: http.StatusOK,
			expectedTask: task.ID,
		},
		{
			name:            "inexistent valid id",
			url:             fmt.Sprintf("/tasks/%s", uuid.Nil.String()),
			expectedCode:    http.StatusNotFound,
			expectedMessage: "task not found",
		},
		{
			name:            "invalid id",
			url:             "/tasks/abc",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "invalid id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Wrong status code, expected=%d, got=%d", tt.expectedCode, resp.StatusCode)
			}

			contentHeader := resp.Header.Get("Content-Type")
			if contentHeader != "application/json" {
				t.Errorf("Wrong Content-Type Header, expected=%s, got=%s", "application/json", contentHeader)
			}

			if tt.expectedMessage == "" {
				var body model.Task
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode task in '%s': %v", tt.name, err)
				}

				if body.ID != tt.expectedTask {
					t.Fatalf("wrong task, expected=%v, got=%v", tt.expectedTask, body.ID)
				}

				return
			}

			var body map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode error message in '%s': %v", tt.name, err)
			}

			msg, ok := body["error"]
			if !ok {
				t.Fatalf("error key not in body, expected=%v, got=%v", true, ok)
			}

			if msg != tt.expectedMessage {
				t.Errorf("Wrong error message in the body, expected=%s, got=%s", tt.expectedMessage, msg)
			}

		})
	}
}
