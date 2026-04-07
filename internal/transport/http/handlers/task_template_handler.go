package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/domain/tasktemplate"
)

type TaskTemplateUsecase interface {
	Create(ctx context.Context, tpl tasktemplate.TaskTemplate) (tasktemplate.TaskTemplate, error)
}

type TaskTemplateHandler struct {
	usecase TaskTemplateUsecase
}

func NewTaskTemplateHandler(usecase TaskTemplateUsecase) *TaskTemplateHandler {
	return &TaskTemplateHandler{usecase: usecase}
}

type createTaskTemplateRequest struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Status      task.Status             `json:"status"`
	Active      *bool                   `json:"active"`
	Recurrence  createRecurrenceRequest `json:"recurrence"`
}

type createRecurrenceRequest struct {
	Type          tasktemplate.RecurrenceType `json:"type"`
	StartDate     string                      `json:"start_date"`
	EndDate       *string                     `json:"end_date,omitempty"`
	EveryNDays    *int                        `json:"every_n_days,omitempty"`
	DayOfMonth    *int                        `json:"day_of_month,omitempty"`
	Parity        *tasktemplate.DayParity     `json:"parity,omitempty"`
	SpecificDates []string                    `json:"specific_dates,omitempty"`
}

func (h *TaskTemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTaskTemplateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTaskTemplateJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	template, err := mapCreateTaskTemplateRequest(req)
	if err != nil {
		writeTaskTemplateJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdTemplate, err := h.usecase.Create(r.Context(), template)
	if err != nil {
		writeTaskTemplateJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, createdTemplate)
}

func mapCreateTaskTemplateRequest(req createTaskTemplateRequest) (tasktemplate.TaskTemplate, error) {
	if req.Title == "" {
		return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("title is required")
	}

	if !req.Status.Valid() {
		return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("status is invalid")
	}

	if !req.Recurrence.Type.Valid() {
		return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("recurrence.type is invalid")
	}

	startDate, err := time.Parse("2006-01-02", req.Recurrence.StartDate)
	if err != nil {
		return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("recurrence.start_date must be in format YYYY-MM-DD")
	}

	var endDate *time.Time
	if req.Recurrence.EndDate != nil {
		parsedEndDate, err := time.Parse("2006-01-02", *req.Recurrence.EndDate)
		if err != nil {
			return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("recurrence.end_date must be in format YYYY-MM-DD")
		}
		endDate = &parsedEndDate
	}

	recurrence := tasktemplate.RecurrenceRule{
		Type:      req.Recurrence.Type,
		StartDate: startDate,
		EndDate:   endDate,
	}

	switch req.Recurrence.Type {
	case tasktemplate.RecurrenceTypeDaily:
		if req.Recurrence.EveryNDays == nil || *req.Recurrence.EveryNDays <= 0 {
			return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("recurrence.every_n_days must be greater than 0 for daily recurrence")
		}
		recurrence.EveryNDays = req.Recurrence.EveryNDays

	case tasktemplate.RecurrenceTypeMonthly:
		if req.Recurrence.DayOfMonth == nil || *req.Recurrence.DayOfMonth < 1 || *req.Recurrence.DayOfMonth > 30 {
			return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("recurrence.day_of_month must be between 1 and 30 for monthly recurrence")
		}
		recurrence.DayOfMonth = req.Recurrence.DayOfMonth

	case tasktemplate.RecurrenceTypeSpecificDates:
		if len(req.Recurrence.SpecificDates) == 0 {
			return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("recurrence.specific_dates must not be empty for specific_dates recurrence")
		}

		specificDates := make([]time.Time, 0, len(req.Recurrence.SpecificDates))
		for _, rawDate := range req.Recurrence.SpecificDates {
			parsedDate, err := time.Parse("2006-01-02", rawDate)
			if err != nil {
				return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("each recurrence.specific_dates value must be in format YYYY-MM-DD")
			}
			specificDates = append(specificDates, parsedDate)
		}
		recurrence.SpecificDates = specificDates

	case tasktemplate.RecurrenceTypeDayParity:
		if req.Recurrence.Parity == nil || !req.Recurrence.Parity.Valid() {
			return tasktemplate.TaskTemplate{}, newTaskTemplateValidationError("recurrence.parity must be even or odd for day_parity recurrence")
		}
		recurrence.Parity = req.Recurrence.Parity
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	now := time.Now()

	return tasktemplate.TaskTemplate{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Active:      active,
		Recurrence:  recurrence,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

type taskTemplateValidationError struct {
	message string
}

func newTaskTemplateValidationError(message string) error {
	return taskTemplateValidationError{message: message}
}

func (e taskTemplateValidationError) Error() string {
	return e.message
}

func writeTaskTemplateJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
