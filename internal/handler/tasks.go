package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/joalberto902/tasks-api/internal/model"
	"github.com/joalberto902/tasks-api/internal/repository"
)

const (
	maxBytes = 1 << 20
)

type TaskHandler struct {
	repo repository.TaskRepository
}

// NewTaskHandler function consumes a repository.TaskRepository type
// and returns a pointer to a new TaskHandler. It returns a pointer
// because it receives a lot simultaneous requests, so they all use
// the same instance
func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func (th *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := th.repo.List(r.Context())
	if err != nil {
		slog.Error("listing tasks",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path)
		writeError(w, r,
			http.StatusInternalServerError,
			"internal server error")
		return
	}

	writeJSON(w, r, http.StatusOK, list)
}

func (th *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	arg := struct {
		Title       string           `json:"title"`
		Status      model.TaskStatus `json:"status"`
		Description string           `json:"description"`
	}{}

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(&arg); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			writeError(w, r,
				http.StatusRequestEntityTooLarge,
				"json payload too large")
			return
		}

		writeError(w, r,
			http.StatusBadRequest,
			"invalid json")
		return
	}

	task, err := th.repo.Create(r.Context(), repository.CreateTaskParams{
		Title:       arg.Title,
		Status:      arg.Status,
		Description: arg.Description,
	})
	if errors.Is(err, repository.ErrInvalidTitle) {
		writeError(w, r,
			http.StatusBadRequest,
			"title field is invalid")
		return
	}

	if errors.Is(err, repository.ErrInvalidStatus) {
		writeError(w, r,
			http.StatusBadRequest,
			"status field is invalid")
		return
	}

	if err != nil {
		slog.Error("creating task",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path)

		writeError(w, r,
			http.StatusInternalServerError,
			"internal server error")
		return
	}

	writeJSON(w, r, http.StatusCreated, task)
}

func (th *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)
	if err != nil {
		writeError(w, r,
			http.StatusBadRequest,
			"invalid id")
		return
	}

	task, err := th.repo.Get(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, r,
			http.StatusNotFound,
			"task not found")
		return
	}

	if err != nil {
		slog.Error("reading task",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path)

		writeError(w, r,
			http.StatusInternalServerError,
			"internal server error")
		return
	}

	writeJSON(w, r, http.StatusOK, task)
}

func (th *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)
	if err != nil {
		writeError(w, r,
			http.StatusBadRequest,
			"invalid id")
		return
	}

	err = th.repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, r,
			http.StatusNotFound,
			"task not found")
		return
	}

	if err != nil {
		slog.Error("deleting task",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path)

		writeError(w, r,
			http.StatusInternalServerError,
			"internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (th *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)
	if err != nil {
		writeError(w, r,
			http.StatusBadRequest,
			"invalid id")
		return
	}

	arg := struct {
		Title       *string           `json:"title,omitempty"`
		Status      *model.TaskStatus `json:"status,omitempty"`
		Description *string           `json:"description,omitempty"`
	}{}

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.Body).Decode(&arg); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			writeError(w, r,
				http.StatusRequestEntityTooLarge,
				"json payload too large")
			return
		}
		writeError(w, r,
			http.StatusBadRequest,
			"invalid json")
		return
	}

	task, err := th.repo.Update(r.Context(), id, repository.UpdateTaskParams{
		Title:       arg.Title,
		Status:      arg.Status,
		Description: arg.Description,
	})
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, r,
			http.StatusNotFound,
			"task not found")
		return
	}

	if errors.Is(err, repository.ErrInvalidTitle) {
		writeError(w, r,
			http.StatusBadRequest,
			"title field is invalid")
		return
	}

	if errors.Is(err, repository.ErrInvalidStatus) {
		writeError(w, r,
			http.StatusBadRequest,
			"status field is invalid")
		return
	}

	if err != nil {
		slog.Error("updating task",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path)

		writeError(w, r,
			http.StatusInternalServerError,
			"internal server error")
		return
	}

	writeJSON(w, r, http.StatusOK, task)
}
