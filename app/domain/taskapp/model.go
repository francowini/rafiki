// Package taskapp provides HTTP handlers for tasks.
package taskapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/business/domain/taskbus"
	"github.com/francowini/rafiki/business/types/contribution"
	"github.com/francowini/rafiki/business/types/taskdescription"
	"github.com/francowini/rafiki/business/types/tasktitle"
)

// Task represents a task for API responses.
type Task struct {
	ID           string  `json:"id"`
	ObjectiveID  *string `json:"objectiveId,omitempty"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	Contribution *int    `json:"contribution,omitempty"`
	Status       string  `json:"status"`
	CompletedAt  *string `json:"completedAt,omitempty"`
	DateCreated  string  `json:"dateCreated"`
	DateUpdated  string  `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (t Task) Encode() ([]byte, string, error) {
	data, err := json.Marshal(t)
	return data, "application/json", err
}

// NewTask represents data for creating a new task.
type NewTask struct {
	ObjectiveID  *string `json:"objectiveId"`
	Title        string  `json:"title" validate:"required"`
	Description  *string `json:"description"`
	Contribution *int    `json:"contribution" validate:"omitempty,min=1,max=10"`
}

// Decode implements the decoder interface.
func (nt *NewTask) Decode(data []byte) error {
	return json.Unmarshal(data, nt)
}

// Validate checks the data for correctness.
func (nt NewTask) Validate() error {
	if err := errs.Check(nt); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// UpdateTask represents data for updating a task.
type UpdateTask struct {
	Title            *string `json:"title"`
	Description      *string `json:"description"`
	ClearDescription bool    `json:"clearDescription"`
	Contribution     *int    `json:"contribution" validate:"omitempty,min=1,max=10"`
}

// Decode implements the decoder interface.
func (ut *UpdateTask) Decode(data []byte) error {
	return json.Unmarshal(data, ut)
}

// Validate checks the data for correctness.
func (ut UpdateTask) Validate() error {
	if err := errs.Check(ut); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// CompleteTaskResponse represents the response after completing a task.
type CompleteTaskResponse struct {
	Task              Task `json:"task"`
	ObjectiveProgress *int `json:"objectiveProgress,omitempty"` // New currentMetric for objective-linked tasks
}

// Encode implements the encoder interface.
func (ctr CompleteTaskResponse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(ctr)
	return data, "application/json", err
}

// ===== Business → App conversions =====

func toAppTask(t taskbus.Task) Task {
	var objectiveID *string
	if t.ObjectiveID != nil {
		id := t.ObjectiveID.String()
		objectiveID = &id
	}

	var description *string
	if t.Description != nil && t.Description.String() != "" {
		d := t.Description.String()
		description = &d
	}

	var contributionVal *int
	if t.Contribution != nil {
		c := t.Contribution.Value()
		contributionVal = &c
	}

	var completedAt *string
	if t.CompletedAt != nil {
		c := t.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		completedAt = &c
	}

	return Task{
		ID:           t.ID.String(),
		ObjectiveID:  objectiveID,
		Title:        t.Title.String(),
		Description:  description,
		Contribution: contributionVal,
		Status:       t.Status.String(),
		CompletedAt:  completedAt,
		DateCreated:  t.DateCreated.Format("2006-01-02T15:04:05Z07:00"),
		DateUpdated:  t.DateUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppTasks(tasks []taskbus.Task) []Task {
	app := make([]Task, len(tasks))
	for i, t := range tasks {
		app[i] = toAppTask(t)
	}
	return app
}

// ===== App → Business conversions =====

func toBusNewTask(nt NewTask, userID uuid.UUID) (taskbus.NewTask, error) {
	var errors errs.FieldErrors

	// Parse title
	title, err := tasktitle.Parse(nt.Title)
	if err != nil {
		errors.Add("title", err)
	}

	// Parse objective ID if provided
	var objectiveIDPtr *uuid.UUID
	if nt.ObjectiveID != nil && *nt.ObjectiveID != "" {
		objectiveID, err := uuid.Parse(*nt.ObjectiveID)
		if err != nil {
			errors.Add("objectiveId", err)
		} else {
			objectiveIDPtr = &objectiveID
		}
	}

	// Parse description if provided
	var descriptionPtr *taskdescription.TaskDescription
	if nt.Description != nil && *nt.Description != "" {
		desc, err := taskdescription.Parse(*nt.Description)
		if err != nil {
			errors.Add("description", err)
		} else {
			descriptionPtr = &desc
		}
	}

	// Parse contribution if provided
	var contributionPtr *contribution.Contribution
	if nt.Contribution != nil {
		c, err := contribution.Parse(*nt.Contribution)
		if err != nil {
			errors.Add("contribution", err)
		} else {
			contributionPtr = &c
		}
	}

	if len(errors) > 0 {
		return taskbus.NewTask{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return taskbus.NewTask{
		UserID:       userID,
		ObjectiveID:  objectiveIDPtr,
		Title:        title,
		Description:  descriptionPtr,
		Contribution: contributionPtr,
	}, nil
}

func toBusUpdateTask(ut UpdateTask) (taskbus.UpdateTask, error) {
	var errors errs.FieldErrors
	var bus taskbus.UpdateTask

	if ut.Title != nil {
		title, err := tasktitle.Parse(*ut.Title)
		if err != nil {
			errors.Add("title", err)
		} else {
			bus.Title = &title
		}
	}

	// Handle ClearDescription flag
	bus.ClearDescription = ut.ClearDescription

	if ut.Description != nil && !ut.ClearDescription {
		desc, err := taskdescription.Parse(*ut.Description)
		if err != nil {
			errors.Add("description", err)
		} else {
			bus.Description = &desc
		}
	}

	if ut.Contribution != nil {
		c, err := contribution.Parse(*ut.Contribution)
		if err != nil {
			errors.Add("contribution", err)
		} else {
			bus.Contribution = &c
		}
	}

	if len(errors) > 0 {
		return taskbus.UpdateTask{}, fmt.Errorf("validate: %w", errors.ToError())
	}

	return bus, nil
}
