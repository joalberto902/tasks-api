package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/joalberto902/tasks-api/internal/model"
	"github.com/joalberto902/tasks-api/internal/repository"
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
		log.Println(err)
		writeError(w,
			http.StatusInternalServerError,
			"internal server error")
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (th *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	arg := struct {
		Title       string           `json:"title"`
		Status      model.TaskStatus `json:"status"`
		Description string           `json:"description"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&arg); err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}

	task, err := th.repo.Create(r.Context(), repository.CreateTaskParams{
		Title:       arg.Title,
		Status:      arg.Status,
		Description: arg.Description,
	})
	if errors.Is(err, repository.ErrInvalidTitle) {
		writeError(w, http.StatusBadRequest, "title field is invalid")
		return
	}

	if errors.Is(err, repository.ErrInvalidStatus) {
		writeError(w, http.StatusBadRequest, "status field is invalid")
		return
	}

	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (th *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	task, err := th.repo.Get(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}

	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (th *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")

	id, err := uuid.Parse(idString)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = th.repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}

	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
