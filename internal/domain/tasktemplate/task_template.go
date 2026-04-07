package tasktemplate

import (
	"time"

	"example.com/taskservice/internal/domain/task"
)

type TaskTemplate struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      task.Status    `json:"status"`
	Active      bool           `json:"active"`
	Recurrence  RecurrenceRule `json:"recurrence"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
