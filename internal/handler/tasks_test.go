package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCreate(t *testing.T) {

	tests := []struct {
		name            string
		request         string
		expectedCode    int
		expectedTitle   string
		expectedStatus  model.TaskStatus
		expectedMessage string
	}{
		{
			name:           "valid request",
			request:        `{"title":"x"}`,
			expectedCode:   http.StatusCreated,
			expectedTitle:  "x",
			expectedStatus: model.StatusTodo,
		},
		{
			name:            "invalid json",
			request:         `{"title":"x"`,
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "invalid json",
		},
		{
			name:            "empty title",
			request:         `{"title": ""}`,
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "title field is invalid",
		},
		{
			name: "invalid status",
			request: `{
			"title": "x",
			"status": "banana"
			}`,
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "status field is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewTaskRepository()
			th := NewTaskHandler(repo)
			reader := strings.NewReader(tt.request)

			r := httptest.NewRequest(http.MethodPost, "/tasks", reader)
			w := httptest.NewRecorder()

			th.Create(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Fatalf("wrong status code, expected=%d, got=%d", tt.expectedCode, resp.StatusCode)
			}

			contentHeader := resp.Header.Get("Content-Type")
			if contentHeader != "application/json" {
				t.Fatalf("wrong content-type header, expected=%s, got=%s", "application/json", contentHeader)
			}

			if tt.expectedCode == http.StatusCreated {
				var body model.Task
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode json response body err: %v", err)
				}

				_, err := repo.Get(context.Background(), body.ID)
				if err != nil {
					t.Fatalf("error fetching created task from the database: %v", err)
				}

				if body.Title != tt.expectedTitle {
					t.Fatalf("wrong title in the response, expected=%s, got=%s", tt.expectedTitle,
						body.Title)
				}

				if body.Status != tt.expectedStatus {
					t.Fatalf("wrong status in the response, expected=%v, got=%v", tt.expectedStatus,
						body.Status)
				}

				return
			}

			var body map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode json response body: %v", err)
			}

			lst, err := repo.List(context.Background())
			if err != nil {
				t.Fatalf("failed to fetch all the task stored in repo: %v", err)
			}

			if len(lst) != 0 {
				t.Fatalf("something was wrongly recorded, expected=%d, got=%d", 0, len(lst))
			}

			msg, ok := body["error"]
			if !ok {
				t.Fatal("expected error message to exist")
			}

			if msg != tt.expectedMessage {
				t.Fatalf("wrong error message, expected=%s, got=%s", tt.expectedMessage, msg)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	repo := repository.NewTaskRepository()
	th := NewTaskHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /tasks/{id}", th.Delete)

	expectedTask, err := repo.Create(context.Background(),
		repository.CreateTaskParams{
			Title: "x",
		})
	if err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}

	tests := []struct {
		name            string
		request         string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:         "existent valid id",
			request:      fmt.Sprintf("/tasks/%s", expectedTask.ID.String()),
			expectedCode: http.StatusNoContent,
		},
		{
			name:            "inexistent valid id",
			request:         fmt.Sprintf("/tasks/%s", uuid.Nil.String()),
			expectedCode:    http.StatusNotFound,
			expectedMessage: "task not found",
		},
		{
			name:            "invalid id",
			request:         "/tasks/abc",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "invalid id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodDelete, tt.request, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Fatalf("wrong status code, expected=%d, got=%d",
					tt.expectedCode, resp.StatusCode)
			}

			if tt.expectedCode == http.StatusNoContent {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("failed to read all the response body: %v", err)
				}

				if len(body) != 0 {
					t.Fatalf("expected empty response body, expected=%q, got=%q",
						"", string(body))
				}
				_, err = repo.Get(context.Background(), expectedTask.ID)
				if !errors.Is(err, repository.ErrNotFound) {
					t.Fatalf("task was not deleted from repository, expected=%v, got=%v",
						repository.ErrNotFound, err)
				}

				return
			}

			contentHeader := resp.Header.Get("Content-Type")
			if contentHeader != "application/json" {
				t.Fatalf("wrong Content-Type header, expected=%s, got=%s",
					"application/json", contentHeader)
			}

			var body map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed decoding response body into json: %v", err)
			}

			msg, ok := body["error"]
			if !ok {
				t.Fatal("expected error message to exist")
			}

			if msg != tt.expectedMessage {
				t.Fatalf("wrong error message, expected=%s, got=%s", tt.expectedMessage, msg)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name                string
		body                string
		expectedCode        int
		expectedTitle       string
		expectedStatus      model.TaskStatus
		expectedDescription string
	}{
		{
			name:                "only title changes",
			body:                `{"title": "new"}`,
			expectedCode:        http.StatusOK,
			expectedTitle:       "new",
			expectedStatus:      model.StatusOngoing,
			expectedDescription: "old",
		},
		{
			name:                "empty new description",
			body:                `{"description": ""}`,
			expectedCode:        http.StatusOK,
			expectedTitle:       "old",
			expectedStatus:      model.StatusOngoing,
			expectedDescription: "",
		},
		{
			name:                "nothing changes",
			body:                `{}`,
			expectedCode:        http.StatusOK,
			expectedTitle:       "old",
			expectedStatus:      model.StatusOngoing,
			expectedDescription: "old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewTaskRepository()
			th := NewTaskHandler(repo)

			mux := http.NewServeMux()
			mux.HandleFunc("PATCH /tasks/{id}", th.Update)

			expectedTask, err := repo.Create(context.Background(),
				repository.CreateTaskParams{
					Title:       "old",
					Status:      model.StatusOngoing,
					Description: "old",
				})
			if err != nil {
				t.Fatalf("failed to create expected test: %v", err)
			}

			requestBody := strings.NewReader(tt.body)

			r := httptest.NewRequest(http.MethodPatch,
				fmt.Sprintf("/tasks/%s", expectedTask.ID.String()), requestBody)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Wrong status code, expected=%d, got=%d",
					tt.expectedCode, resp.StatusCode)
			}

			contentHeader := resp.Header.Get("Content-Type")
			if contentHeader != "application/json" {
				t.Errorf("wrong Content-Type header, expected=%s, got=%s",
					"application/json", contentHeader)
			}

			var body model.Task
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response body into json: %v", err)
			}

			repoTask, err := repo.Get(context.Background(), expectedTask.ID)
			if err != nil {
				t.Fatalf("failed to fetch the updated recorded task: %v", err)
			}

			if repoTask.Title != tt.expectedTitle {
				t.Fatalf("wrong repository title, expected=%s, got=%s",
					tt.expectedTitle, repoTask.Title)
			}

			if repoTask.Status != tt.expectedStatus {
				t.Fatalf("wrong repository task status, expected=%s, got=%s",
					tt.expectedStatus, repoTask.Status)
			}

			if repoTask.Description != tt.expectedDescription {
				t.Fatalf("wrong repository description, expected=%s, got=%s",
					tt.expectedDescription, repoTask.Description)
			}

			if body.ID != expectedTask.ID {
				t.Fatalf("wrong response id, expected=%v, got=%v",
					expectedTask.ID, body.ID)
			}

			if body.Title != tt.expectedTitle {
				t.Fatalf("wrong response title, expected=%s, got=%s",
					tt.expectedTitle, body.Title)
			}

			if body.Status != tt.expectedStatus {
				t.Fatalf("wrong response task status, expected=%s, got=%s",
					tt.expectedStatus, body.Status)
			}

			if body.Description != tt.expectedDescription {
				t.Fatalf("wrong response description, expected=%s, got=%s",
					tt.expectedDescription, body.Description)
			}

		})
	}
}

func TestUpdateErrors(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		body            string
		expectedCode    int
		expectedMessage string
	}{
		{
			name:            "invalid json",
			body:            `{"title": "x"`,
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "invalid json",
		},
		{
			name:            "invalid title",
			body:            `{"title": ""}`,
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "title field is invalid",
		},
		{
			name:            "invalid status",
			body:            `{"status": "banana"}`,
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "status field is invalid",
		},
		{
			name:            "invalid id",
			id:              "abc",
			body:            `{}`,
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "invalid id",
		},
		{
			name:            "inexistent id",
			id:              uuid.Nil.String(),
			body:            `{}`,
			expectedCode:    http.StatusNotFound,
			expectedMessage: "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewTaskRepository()
			th := NewTaskHandler(repo)

			expectedTask, err := repo.Create(context.Background(), repository.CreateTaskParams{
				Title:       "old",
				Status:      model.StatusOngoing,
				Description: "old",
			})
			if err != nil {
				t.Fatalf("failed to create test task: %v", err)
			}

			if tt.id == "" {
				tt.id = expectedTask.ID.String()
			}

			mux := http.NewServeMux()
			mux.HandleFunc("PATCH /tasks/{id}", th.Update)

			requestBody := strings.NewReader(tt.body)
			r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/tasks/%s", tt.id), requestBody)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Fatalf("wrong status code, expected=%d, got=%d",
					tt.expectedCode, resp.StatusCode)
			}

			contentHeader := resp.Header.Get("Content-Type")
			if contentHeader != "application/json" {
				t.Fatalf("wrong Content-Type header, expected=%s, got=%s",
					"application/json", contentHeader)
			}

			repoTask, err := repo.Get(context.Background(), expectedTask.ID)
			if err != nil {
				t.Fatalf("failed to fetch recorded task in the repository: %v", err)
			}

			if !repoTask.UpdatedAt.Equal(expectedTask.UpdatedAt) {
				t.Fatalf("wrong update time in the recorded task, expected=%v, got=%v",
					expectedTask.UpdatedAt, repoTask.UpdatedAt)
			}

			if repoTask.Title != expectedTask.Title {
				t.Fatalf("wrong title in the recorded task, expected=%s, got=%s",
					expectedTask.Title, repoTask.Title)
			}

			if repoTask.Status != expectedTask.Status {
				t.Fatalf("wrong status in the recorded task, expected=%v, got=%v",
					expectedTask.Status, repoTask.Status)
			}

			if repoTask.Description != expectedTask.Description {
				t.Fatalf("wrong description in the recorded task, expected=%s, got=%s",
					expectedTask.Description, repoTask.Description)
			}

			var body map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response body from json: %v", err)
			}

			msg, ok := body["error"]
			if !ok {
				t.Fatal("expected error field to exist in the response")
			}

			if msg != tt.expectedMessage {
				t.Fatalf("wrong message in the response error, expected=%s, got=%s",
					tt.expectedMessage, msg)
			}
		})
	}
}
