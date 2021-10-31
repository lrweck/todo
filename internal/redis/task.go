package redis

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/lrweck/todo/internal"
)

type Task struct {
	client *redis.Client
}

// NewTask instantiates the Task repository.
func NewTask(client *redis.Client) *Task {
	return &Task{
		client: client,
	}
}

// Created publishes a message indicating a task was created.
func (t *Task) Created(ctx context.Context, task internal.Task) error {
	return t.publish(ctx, "Task.Created", "tasks.event.created", task)
}

// Deleted publishes a message indicating a task was deleted.
func (t *Task) Deleted(ctx context.Context, id string) error {
	return t.publish(ctx, "Task.Deleted", "tasks.event.deleted", id)
}

// Updated publishes a message indicating a task was updated.
func (t *Task) Updated(ctx context.Context, task internal.Task) error {
	return t.publish(ctx, "Task.Updated", "tasks.event.updated", task)
}

func (t *Task) publish(ctx context.Context, spanName, channel string, e interface{}) error {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("redis").Start(ctx, spanName)
	defer span.End()

	span.SetAttributes(
		semconv.DBSystemRedis,
		attribute.KeyValue{
			Key:   "db.statement",
			Value: attribute.StringValue("PUBLISH"),
		},
	)

	//-

	var b bytes.Buffer

	if err := json.NewEncoder(&b).Encode(e); err != nil {
		return internal.WrapErrorf(err, internal.ErrCodeUnknown, "json.Encode")
	}

	res := t.client.Publish(ctx, channel, b.Bytes())
	if err := res.Err(); err != nil {
		return internal.WrapErrorf(err, internal.ErrCodeUnknown, "client.Publish")
	}

	return nil
}
