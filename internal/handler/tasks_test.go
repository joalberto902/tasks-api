package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
