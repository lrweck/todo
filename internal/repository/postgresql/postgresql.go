package postgresql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lrweck/todo/internal"
	"github.com/lrweck/todo/internal/repository/postgresql/db"
)

//go:generate sqlc generate

func convertPriority(p db.Priority) (internal.Priority, error) {
	switch p {
	case db.PriorityNone:
		return internal.PriorityNone, nil
	case db.PriorityLow:
		return internal.PriorityLow, nil
	case db.PriorityMedium:
		return internal.PriorityMedium, nil
	case db.PriorityHigh:
		return internal.PriorityHigh, nil
	}

	return internal.Priority(-1), fmt.Errorf("unknown value: %s", p)
}

func newNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

func newPriority(p internal.Priority) db.Priority {
	switch p {
	case internal.PriorityNone:
		return db.PriorityNone
	case internal.PriorityLow:
		return db.PriorityLow
	case internal.PriorityMedium:
		return db.PriorityMedium
	case internal.PriorityHigh:
		return db.PriorityHigh
	}

	// XXX: because we are using an enum type, postgres will fail with the following value.

	return "invalid"
}
