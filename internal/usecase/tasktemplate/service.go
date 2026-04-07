package tasktemplate

import (
	"context"
	"time"

	"example.com/taskservice/internal/domain/task"
	domain "example.com/taskservice/internal/domain/tasktemplate"
)

type Repository interface {
	Create(ctx context.Context, tpl domain.TaskTemplate) (domain.TaskTemplate, error)
	CreateTask(ctx context.Context, newTask task.Task) (task.Task, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, tpl domain.TaskTemplate) (domain.TaskTemplate, error) {
	createdTemplate, err := s.repo.Create(ctx, tpl)
	if err != nil {
		return domain.TaskTemplate{}, err
	}

	dates := generateDates(createdTemplate.Recurrence, 30)

	for _, scheduledDate := range dates {
		templateID := createdTemplate.ID
		now := time.Now()

		_, err := s.repo.CreateTask(ctx, task.Task{
			Title:        createdTemplate.Title,
			Description:  createdTemplate.Description,
			Status:       createdTemplate.Status,
			ScheduledFor: scheduledDate,
			TemplateID:   &templateID,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
		if err != nil {
			return domain.TaskTemplate{}, err
		}
	}

	return createdTemplate, nil
}

func generateDates(rule domain.RecurrenceRule, daysAhead int) []time.Time {
	now := time.Now()
	end := now.AddDate(0, 0, daysAhead)

	var result []time.Time

	switch rule.Type {
	case domain.RecurrenceTypeDaily:
		if rule.EveryNDays == nil {
			return result
		}

		current := rule.StartDate
		for !current.After(end) {
			if !current.Before(now.Truncate(24 * time.Hour)) {
				result = append(result, current)
			}
			current = current.AddDate(0, 0, *rule.EveryNDays)
		}

	case domain.RecurrenceTypeMonthly:
		if rule.DayOfMonth == nil {
			return result
		}

		current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		for !current.After(end) {
			year, month, _ := current.Date()
			day := *rule.DayOfMonth

			daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, now.Location()).Day()
			if day <= daysInMonth {
				date := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
				if !date.Before(rule.StartDate) && !date.Before(now.Truncate(24*time.Hour)) && !date.After(end) {
					result = append(result, date)
				}
			}

			current = current.AddDate(0, 1, 0)
		}

	case domain.RecurrenceTypeSpecificDates:
		for _, date := range rule.SpecificDates {
			if !date.Before(now.Truncate(24*time.Hour)) && !date.After(end) {
				result = append(result, date)
			}
		}

	case domain.RecurrenceTypeDayParity:
		if rule.Parity == nil {
			return result
		}

		current := now.Truncate(24 * time.Hour)
		if current.Before(rule.StartDate) {
			current = rule.StartDate
		}

		for !current.After(end) {
			day := current.Day()

			if *rule.Parity == domain.DayParityEven && day%2 == 0 {
				result = append(result, current)
			}
			if *rule.Parity == domain.DayParityOdd && day%2 != 0 {
				result = append(result, current)
			}

			current = current.AddDate(0, 0, 1)
		}
	}

	if rule.EndDate != nil {
		filtered := make([]time.Time, 0, len(result))
		for _, date := range result {
			if !date.After(*rule.EndDate) {
				filtered = append(filtered, date)
			}
		}
		return filtered
	}

	return result
}
