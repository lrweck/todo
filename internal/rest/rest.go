package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"go.opentelemetry.io/otel/trace"

	"github.com/lrweck/todo/internal"
)

// ErrorResponse represents a response containing an error message.
type ErrorResponse struct {
	Error       string            `json:"error"`
	Validations validation.Errors `json:"validations,omitempty"`
}

func renderErrorResponse(ctx context.Context, w http.ResponseWriter, msg string, err error) {
	resp := ErrorResponse{Error: msg}
	status := http.StatusInternalServerError

	var ierr *internal.Error
	if !errors.As(err, &ierr) {
		resp.Error = "internal error"
	} else {
		switch ierr.Code() {
		case internal.ErrCodeNotFound:
			status = http.StatusNotFound
		case internal.ErrCodeInvalidArgument:
			status = http.StatusBadRequest

			var verrors validation.Errors
			if errors.As(ierr, &verrors) {
				resp.Validations = verrors
			}
		case internal.ErrCodeUnknown:
			fallthrough
		default:
			status = http.StatusInternalServerError
		}
	}

	if err != nil {
		_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("todo.rest").Start(ctx, "rest.renderErrorResponse")
		defer span.End()

		span.RecordError(err)
	}

	renderResponse(ctx, w, resp, status)
}

func renderResponse(ctx context.Context, w http.ResponseWriter, res interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")

	content, err := json.Marshal(res)
	if err != nil {
		// XXX Do something with the error ;)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(status)

	if _, err = w.Write(content); err != nil {
		_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("todo.rest").Start(ctx, "rest.renderErrorResponse")
		defer span.End()

		span.RecordError(err)
	}
}
