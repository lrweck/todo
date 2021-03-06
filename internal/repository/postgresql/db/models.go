package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityNone   Priority = "none"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

func (e *Priority) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Priority(s)
	case string:
		*e = Priority(s)
	default:
		return fmt.Errorf("unsupported scan type for Priority: %T", src)
	}
	return nil
}

type Tasks struct {
	ID          uuid.UUID
	Description string
	Priority    Priority
	StartDate   sql.NullTime
	DueDate     sql.NullTime
	Done        bool
}
