package service

import (
	"context"
	"time"

	"github.com/lrweck/todo/internal"
	"github.com/mercari/go-circuitbreaker"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type TaskRepo interface {
	Create(ctx context.Context, dates internal.CreateParams) (internal.Task, error)
	Delete(ctx context.Context, id string) error
	Find(ctx context.Context, id string) (internal.Task, error)
	Update(ctx context.Context, id, description string, priority internal.Priority, dates internal.Dates, isDone bool) error
}

type TaskSearchRepo interface {
	Search(ctx context.Context, args internal.SearchParams) (internal.SearchResults, error)
}

type TaskMessageBrokerRepo interface {
	Created(ctx context.Context, task internal.Task) error
	Deleted(ctx context.Context, id string) error
	Updated(ctx context.Context, task internal.Task) error
}

type Task struct {
	repo          TaskRepo
	search        TaskSearchRepo
	messageBroker TaskMessageBrokerRepo
	cb            *circuitbreaker.CircuitBreaker
}

func NewTask(logger *zap.Logger,
	repo TaskRepo,
	search TaskSearchRepo,
	messageBroker TaskMessageBrokerRepo,
) *Task {
	return &Task{
		repo:          repo,
		search:        search,
		messageBroker: messageBroker,
		cb: circuitbreaker.New(
			circuitbreaker.WithOpenTimeout(time.Minute),
			circuitbreaker.WithTripFunc(circuitbreaker.NewTripFuncConsecutiveFailures(3)),
			circuitbreaker.WithOnStateChangeHookFn(func(oldState, newState circuitbreaker.State) {
				logger.Info("state changed",
					zap.String("old", string(oldState)),
					zap.String("new", string(newState)),
				)
			}),
		),
	}
}

func (t *Task) By(ctx context.Context, args internal.SearchParams) (_ internal.SearchResults, err error) {

	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("todo.service").Start(ctx, "Task.By")
	defer span.End()

	if !t.cb.Ready() {
		return internal.SearchResults{}, internal.NewErrorf(internal.ErrCodeUnknown, "service not ready")
	}

	defer func() {
		err = t.cb.Done(ctx, err)
	}()

	res, err := t.search.Search(ctx, args)
	if err != nil {
		return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrCodeUnknown, "search")
	}

	return res, nil
}

func (t *Task) Create(ctx context.Context, params internal.CreateParams) (internal.Task, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("todo.service").Start(ctx, "Task.Create")
	defer span.End()

	if err := params.Validate(); err != nil {
		return internal.Task{}, internal.WrapErrorf(err, internal.ErrCodeInvalidArgument, "params.Validate")
	}

	task, err := t.repo.Create(ctx, params)
	if err != nil {
		return internal.Task{}, internal.WrapErrorf(err, internal.ErrCodeUnknown, "repo.Create")
	}

	// XXX: Transactions will be revisited in future episodes.
	_ = t.messageBroker.Created(ctx, task) // XXX: Ignoring errors on purpose

	return task, nil
}

// Delete removes an existing Task from the datastore.
func (t *Task) Delete(ctx context.Context, id string) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("todo.service").Start(ctx, "Task.Delete")
	defer span.End()

	// XXX: We will revisit the number of received arguments in future episodes.
	if err := t.repo.Delete(ctx, id); err != nil {
		return internal.WrapErrorf(err, internal.ErrCodeUnknown, "Delete")
	}

	// XXX: Transactions will be revisited in future episodes.
	_ = t.messageBroker.Deleted(ctx, id) // XXX: Ignoring errors on purpose

	return nil
}

// Task gets an existing Task from the datastore.
func (t *Task) Task(ctx context.Context, id string) (internal.Task, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("todo.service").Start(ctx, "Task.Task")
	defer span.End()

	// XXX: We will revisit the number of received arguments in future episodes.
	task, err := t.repo.Find(ctx, id)
	if err != nil {
		return internal.Task{}, internal.WrapErrorf(err, internal.ErrCodeUnknown, "Find")
	}

	return task, nil
}

// Update updates an existing Task in the datastore.
func (t *Task) Update(ctx context.Context,
	id string,
	description string,
	priority internal.Priority,
	dates internal.Dates,
	isDone bool,
) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("todo.service").Start(ctx, "Task.Update")
	defer span.End()

	// XXX: We will revisit the number of received arguments in future episodes.
	if err := t.repo.Update(ctx, id, description, priority, dates, isDone); err != nil {
		return internal.WrapErrorf(err, internal.ErrCodeUnknown, "repo.Update")
	}

	{
		// XXX: This will be improved when Kafka events are introduced in future episodes
		task, err := t.repo.Find(ctx, id)
		if err == nil {
			// XXX: Transactions will be revisited in future episodes.
			_ = t.messageBroker.Updated(ctx, task) // XXX: Ignoring errors on purpose
		}
	}

	return nil
}
