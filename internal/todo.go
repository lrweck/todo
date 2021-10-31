package internal

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Priority int8

const (
	PriorityNone Priority = iota
	PriorityLow
	PriorityMedium
	PriorityHigh
)

func (p Priority) Validate() error {
	switch p {
	case PriorityNone, PriorityLow, PriorityMedium, PriorityHigh:
		return nil
	}
	return NewErrorf(ErrCodeInvalidArgument, "invalid priority: %d", p)
}

type Category string

type Dates struct {
	Start time.Time
	Due   time.Time
}

func (d Dates) Validate() error {
	if !d.Start.IsZero() && !d.Due.IsZero() && d.Start.After(d.Due) {
		return NewErrorf(ErrCodeInvalidArgument, "start date should not be after due date")
	}
	return nil
}

type Task struct {
	IsDone      bool
	Priority    Priority
	ID          string
	Description string
	Dates       Dates
	SubTasks    []Task
	Categories  []Category
}

func (t Task) Validate() error {

	err := validation.ValidateStruct(&t,
		validation.Field(&t.Description, validation.Required),
		validation.Field(&t.Priority),
		validation.Field(&t.Dates))

	if err != nil {
		return WrapErrorf(err, ErrCodeInvalidArgument, "invalid task")
	}

	return nil
}
