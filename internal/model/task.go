package model

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusTodo    TaskStatus = "todo"
	StatusOngoing TaskStatus = "ongoing"
	StatusDone    TaskStatus = "done"
)

// IsValid reports wheter the taskstatus is one of the valid values
func (ts TaskStatus) IsValid() bool {
	switch ts {
	case StatusTodo, StatusOngoing, StatusDone:
		return true
	default:
		return false
	}
}

type Task struct {
	ID          uuid.UUID  `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Title       string     `json:"title"`
	Status      TaskStatus `json:"status"`
	Description string     `json:"description"`
}
