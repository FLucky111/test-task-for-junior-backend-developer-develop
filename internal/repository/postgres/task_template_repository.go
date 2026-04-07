package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/domain/tasktemplate"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskTemplateRepository struct {
	db *pgxpool.Pool
}

func NewTaskTemplateRepository(db *pgxpool.Pool) *TaskTemplateRepository {
	return &TaskTemplateRepository{db: db}
}

func (r *TaskTemplateRepository) Create(ctx context.Context, tpl tasktemplate.TaskTemplate) (tasktemplate.TaskTemplate, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return tasktemplate.TaskTemplate{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var templateID int64

	err = tx.QueryRow(ctx, `
		INSERT INTO task_templates (title, description, status, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		tpl.Title,
		tpl.Description,
		tpl.Status,
		tpl.Active,
		tpl.CreatedAt,
		tpl.UpdatedAt,
	).Scan(&templateID)
	if err != nil {
		return tasktemplate.TaskTemplate{}, fmt.Errorf("insert task_templates: %w", err)
	}

	var specificDatesJSON []byte
	if len(tpl.Recurrence.SpecificDates) > 0 {
		specificDatesJSON, err = json.Marshal(tpl.Recurrence.SpecificDates)
		if err != nil {
			return tasktemplate.TaskTemplate{}, fmt.Errorf("marshal specific dates: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO task_recurrences (
			template_id,
			type,
			start_date,
			end_date,
			every_n_days,
			day_of_month,
			parity,
			specific_dates
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		templateID,
		tpl.Recurrence.Type,
		tpl.Recurrence.StartDate,
		tpl.Recurrence.EndDate,
		tpl.Recurrence.EveryNDays,
		tpl.Recurrence.DayOfMonth,
		tpl.Recurrence.Parity,
		specificDatesJSON,
	)
	if err != nil {
		return tasktemplate.TaskTemplate{}, fmt.Errorf("insert task_recurrences: %w", err)
	}

	tpl.ID = templateID

	if err := tx.Commit(ctx); err != nil {
		return tasktemplate.TaskTemplate{}, fmt.Errorf("commit tx: %w", err)
	}

	return tpl, nil
}

func (r *TaskTemplateRepository) CreateTask(ctx context.Context, newTask task.Task) (task.Task, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO tasks (title, description, status, scheduled_for, template_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, description, status, scheduled_for, template_id, created_at, updated_at
	`,
		newTask.Title,
		newTask.Description,
		newTask.Status,
		newTask.ScheduledFor,
		newTask.TemplateID,
		newTask.CreatedAt,
		newTask.UpdatedAt,
	).Scan(
		&newTask.ID,
		&newTask.Title,
		&newTask.Description,
		&newTask.Status,
		&newTask.ScheduledFor,
		&newTask.TemplateID,
		&newTask.CreatedAt,
		&newTask.UpdatedAt,
	)
	if err != nil {
		return task.Task{}, fmt.Errorf("insert task: %w", err)
	}

	return newTask, nil
}
