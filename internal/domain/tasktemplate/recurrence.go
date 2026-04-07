package tasktemplate

import "time"

type RecurrenceType string

const (
	RecurrenceTypeDaily         RecurrenceType = "daily"
	RecurrenceTypeMonthly       RecurrenceType = "monthly"
	RecurrenceTypeSpecificDates RecurrenceType = "specific_dates"
	RecurrenceTypeDayParity     RecurrenceType = "day_parity"
)

type DayParity string

const (
	DayParityEven DayParity = "even"
	DayParityOdd  DayParity = "odd"
)

type RecurrenceRule struct {
	Type          RecurrenceType `json:"type"`
	StartDate     time.Time      `json:"start_date"`
	EndDate       *time.Time     `json:"end_date,omitempty"`
	EveryNDays    *int           `json:"every_n_days,omitempty"`
	DayOfMonth    *int           `json:"day_of_month,omitempty"`
	Parity        *DayParity     `json:"parity,omitempty"`
	SpecificDates []time.Time    `json:"specific_dates,omitempty"`
}

func (t RecurrenceType) Valid() bool {
	switch t {
	case RecurrenceTypeDaily, RecurrenceTypeMonthly, RecurrenceTypeSpecificDates, RecurrenceTypeDayParity:
		return true
	default:
		return false
	}
}

func (p DayParity) Valid() bool {
	switch p {
	case DayParityEven, DayParityOdd:
		return true
	default:
		return false
	}
}
