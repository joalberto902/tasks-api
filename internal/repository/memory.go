package repository

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joalberto902/tasks-api/internal/model"
)

type taskRepository struct {
	TaskRepository
	data map[uuid.UUID]*model.Task
	rwm  sync.RWMutex
}

func NewTaskRepository() TaskRepository {
	data := make(map[uuid.UUID]*model.Task)

	return &taskRepository{data: data}
}

func (tr *taskRepository) Create(
	ctx context.Context,
	arg CreateTaskParams,
) (model.Task, error) {
	//Validating input
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

	task := model.Task{
		ID:          id,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Title:       arg.Title,
		Status:      arg.Status,
		Description: arg.Description,
	}

	tr.rwm.Lock()
	defer tr.rwm.Unlock()

	tr.data[id] = &task

	return task, nil
}

func (tr *taskRepository) Get(ctx context.Context, id uuid.UUID) (model.Task, error) {
	tr.rwm.RLock()
	defer tr.rwm.RUnlock()
	task, ok := tr.data[id]
	if !ok {
		return model.Task{}, ErrNotFound
	}

	rtask := *task
	return rtask, nil
}

func (tr *taskRepository) List(ctx context.Context) ([]model.Task, error) {
	//Maybe the function should error if the context timesout
	tr.rwm.RLock()
	defer tr.rwm.RUnlock()

	taskSlice := make([]model.Task, 0, len(tr.data))
	for _, task := range tr.data {
		taskSlice = append(taskSlice, *task)
	}

	cmp := func(a, b model.Task) int {
		timeComp := a.CreatedAt.Compare(b.CreatedAt)
		if timeComp != 0 {
			return timeComp
		}

		return bytes.Compare(a.ID[:], b.ID[:])
	}

	slices.SortFunc(taskSlice, cmp)

	return taskSlice, nil
}

func (tr *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tr.rwm.Lock()
	defer tr.rwm.Unlock()

	//Verifying if id is in repo
	_, ok := tr.data[id]
	if !ok {
		return ErrNotFound
	}

	delete(tr.data, id)

	return nil
}

func (tr *taskRepository) Update(
	ctx context.Context,
	id uuid.UUID,
	arg UpdateTaskParams,
) (model.Task, error) {
	tr.rwm.Lock()
	defer tr.rwm.Unlock()

	task, ok := tr.data[id]
	if !ok {
		return model.Task{}, ErrNotFound
	}

	newTitle := task.Title
	if arg.Title != nil {
		newTitle = strings.TrimSpace(*arg.Title)
		if newTitle == "" {
			return model.Task{}, ErrInvalidTitle
		}
	}

	newStatus := task.Status
	if arg.Status != nil {
		newStatus = *arg.Status
		if newStatus == "" {
			newStatus = model.StatusTodo
		}

		if !newStatus.IsValid() {
			return model.Task{}, ErrInvalidStatus
		}
	}

	newDesc := task.Description
	if arg.Description != nil {
		newDesc = *arg.Description
	}

	task.Title = newTitle
	task.Status = newStatus
	task.Description = newDesc
	task.UpdatedAt = time.Now().UTC()

	return *task, nil
}
