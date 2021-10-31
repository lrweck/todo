package internal

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateParams struct {
	Description string
	Priority    Priority
	Dates       Dates
}

func (c CreateParams) Validate() error {

	if c.Priority == PriorityNone {
		return validation.Errors{
			"priority": NewErrorf(ErrCodeInvalidArgument, "priority is required"),
		}
	}

	t := Task{
		Description: c.Description,
		Priority:    c.Priority,
		Dates:       c.Dates,
	}

	if err := validation.ValidateStruct(&t); err != nil {
		return WrapErrorf(err, ErrCodeInvalidArgument, "validation.Validate")
	}

	return nil
}

type SearchParams struct {
	Description *string
	Priority    *Priority
	IsDone      *bool
	From        int64
	Size        int64
}

func (s SearchParams) IsZero() bool {
	return s.Description == nil &&
		s.Priority == nil &&
		s.IsDone == nil
}

type SearchResults struct {
	Tasks []Task
	Total int64
}
