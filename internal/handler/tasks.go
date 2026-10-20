package handler

import (
	"log"
	"net/http"

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
