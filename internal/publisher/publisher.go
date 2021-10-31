package publisher

import (
	"context"

	"github.com/lrweck/todo/internal"
)

type TaskPublisherClient interface {
	Publish(ctx context.Context, spanName, channel string, e interface{}) error
}

type TaskPublisher interface {
	Created(ctx context.Context, task internal.Task) error
	Deleted(ctx context.Context, id string) error
	Updated(ctx context.Context, task internal.Task) error
}

type Task struct {
	client TaskPublisherClient
}

func New(pubClient TaskPublisherClient) *Task {
	return &Task{
		client: pubClient,
	}
}

func (p *Task) Created(ctx context.Context, task internal.Task) error {
	return p.client.Publish(ctx, "Task.Created", "task.event.created", task)
}

func (p *Task) Deleted(ctx context.Context, id string) error {
	return p.client.Publish(ctx, "Task.Deleted", "task.event.deleted", id)
}

func (p *Task) Updated(ctx context.Context, task internal.Task) error {
	return p.client.Publish(ctx, "Task.Updated", "task.event.updated", task)
}
