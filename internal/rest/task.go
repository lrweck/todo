package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	router "github.com/gorilla/mux"
	"github.com/lrweck/todo/internal"
)

const uuidRegEx string = `[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`

//go:generate counterfeiter -generate

//counterfeiter:generate -o resttesting/task_service.gen.go . TaskService

// TaskService ...
type TaskService interface {
	By(ctx context.Context, args internal.SearchParams) (internal.SearchResults, error)
	Create(ctx context.Context, params internal.CreateParams) (internal.Task, error)
	Delete(ctx context.Context, id string) error
	Task(ctx context.Context, id string) (internal.Task, error)
	Update(ctx context.Context, id, description string, priority internal.Priority, dates internal.Dates, isDone bool) error
}

// TaskHandler ...
type TaskHandler struct {
	svc TaskService
}

// NewTaskHandler ...
func NewTaskHandler(svc TaskService) *TaskHandler {
	return &TaskHandler{
		svc: svc,
	}
}

// Register connects the handlers to the router.
func (t *TaskHandler) Register(r *router.Router) {

	r.HandleFunc("/tasks", t.create).Methods(http.MethodPost)
	r.HandleFunc(fmt.Sprintf("/task/{id:%s}", uuidRegEx), t.task).Methods(http.MethodGet)
	r.HandleFunc(fmt.Sprintf("/task/{id:%s}", uuidRegEx), t.update).Methods(http.MethodPut)
	r.HandleFunc(fmt.Sprintf("/task/{id:%s}", uuidRegEx), t.delete).Methods(http.MethodDelete)
	r.HandleFunc("/search/tasks", t.search).Methods(http.MethodPost)
}

// Task is an activity that needs to be completed within a period of time.
type Task struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
	IsDone      bool     `json:"is_done"`
}

// CreateTasksRequest defines the request used for creating tasks.
type CreateTasksRequest struct {
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
}

// CreateTasksResponse defines the response returned back after creating tasks.
type CreateTasksResponse struct {
	Task Task `json:"task"`
}

func (t *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(r.Context(), w, "invalid request",
			internal.WrapErrorf(err, internal.ErrCodeInvalidArgument, "json decoder"))

		return
	}

	defer r.Body.Close()

	task, err := t.svc.Create(r.Context(), internal.CreateParams{
		Description: req.Description,
		Priority:    req.Priority.Convert(),
		Dates:       req.Dates.Convert(),
	})
	if err != nil {
		renderErrorResponse(r.Context(), w, "create failed", err)

		return
	}

	renderResponse(r.Context(),
		w,
		&CreateTasksResponse{
			Task: Task{
				ID:          task.ID,
				Description: task.Description,
				Priority:    NewPriority(task.Priority),
				Dates:       NewDates(task.Dates),
			},
		},
		http.StatusCreated)
}

func (t *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	// NOTE: Safe to ignore error, because it's always defined.
	id := router.Vars(r)["id"]

	if err := t.svc.Delete(r.Context(), id); err != nil {
		renderErrorResponse(r.Context(), w, "delete failed", err)

		return
	}

	renderResponse(r.Context(), w, struct{}{}, http.StatusOK)
}

// ReadTasksResponse defines the response returned back after searching one task.
type ReadTasksResponse struct {
	Task Task `json:"task"`
}

func (t *TaskHandler) task(w http.ResponseWriter, r *http.Request) {
	// NOTE: Safe to ignore error, because it's always defined.
	id := router.Vars(r)["id"]

	task, err := t.svc.Task(r.Context(), id)
	if err != nil {
		renderErrorResponse(r.Context(), w, "find failed", err)

		return
	}

	renderResponse(r.Context(),
		w,
		&ReadTasksResponse{
			Task: Task{
				ID:          task.ID,
				Description: task.Description,
				Priority:    NewPriority(task.Priority),
				Dates:       NewDates(task.Dates),
				IsDone:      task.IsDone,
			},
		},
		http.StatusOK)
}

// UpdateTasksRequest defines the request used for updating a task.
type UpdateTasksRequest struct {
	Description string   `json:"description"`
	IsDone      bool     `json:"is_done"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
}

func (t *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	var req UpdateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(r.Context(), w, "invalid request",
			internal.WrapErrorf(err, internal.ErrCodeInvalidArgument, "json decoder"))

		return
	}

	defer r.Body.Close()

	// NOTE: Safe to ignore error, because it's always defined.
	id := router.Vars(r)["id"]

	err := t.svc.Update(r.Context(), id, req.Description, req.Priority.Convert(), req.Dates.Convert(), req.IsDone)
	if err != nil {
		renderErrorResponse(r.Context(), w, "update failed", err)

		return
	}

	renderResponse(r.Context(), w, &struct{}{}, http.StatusOK)
}

// SearchTasksRequest defines the request used for searching tasks.
type SearchTasksRequest struct {
	Description *string   `json:"description"`
	Priority    *Priority `json:"priority"`
	IsDone      *bool     `json:"is_done"`
	From        int64     `json:"from"`
	Size        int64     `json:"size"`
}

// SearchTasksResponse defines the response returned back after searching for any task.
type SearchTasksResponse struct {
	Tasks []Task `json:"tasks"`
	Total int64  `json:"total"`
}

func (t *TaskHandler) search(w http.ResponseWriter, r *http.Request) {
	var req SearchTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(r.Context(), w, "invalid request",
			internal.WrapErrorf(err, internal.ErrCodeInvalidArgument, "json decoder"))

		return
	}

	defer r.Body.Close()

	var priority *internal.Priority

	if req.Priority != nil {
		res := req.Priority.Convert()
		priority = &res
	}

	res, err := t.svc.By(r.Context(), internal.SearchParams{
		Description: req.Description,
		Priority:    priority,
		IsDone:      req.IsDone,
		From:        req.From,
		Size:        req.Size,
	})
	if err != nil {
		renderErrorResponse(r.Context(), w, "search failed", err)

		return
	}

	tasks := make([]Task, len(res.Tasks))

	for i, task := range res.Tasks {
		tasks[i].ID = task.ID
		tasks[i].Description = task.Description
		tasks[i].Priority = NewPriority(task.Priority)
		tasks[i].Dates = NewDates(task.Dates)
	}

	renderResponse(r.Context(),
		w,
		&SearchTasksResponse{
			Tasks: tasks,
			Total: res.Total,
		}, http.StatusOK)
}
