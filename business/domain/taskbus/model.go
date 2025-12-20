// Package taskbus provides business logic for tasks management.
package taskbus

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/contribution"
	"github.com/francowini/rafiki/business/types/taskdescription"
	"github.com/francowini/rafiki/business/types/taskstatus"
	"github.com/francowini/rafiki/business/types/tasktitle"
)

// Domain errors
var (
	ErrNotFound                     = errors.New("task not found")
	ErrMissingUserID                = errors.New("userID is required for querying tasks")
	ErrNotObjectiveOwner            = errors.New("user does not own the specified objective")
	ErrObjectiveNotFound            = errors.New("objective not found")
	ErrContributionRequiredForLink  = errors.New("contribution required when task is linked to objective")
	ErrOnlyResultAllowsContribution = errors.New("contribution only valid for result tracking type objectives")
	ErrStatusChangeMustUseMethod    = errors.New("status changes must use Complete/Uncomplete/Cancel methods")
	ErrAlreadyCompleted             = errors.New("task is already completed")
	ErrAlreadyCanceled              = errors.New("task is already canceled")
	ErrNotCompleted                 = errors.New("task is not completed")
	ErrCannotUncompleteInboxTask    = errors.New("cannot uncomplete inbox task (no objective progress to reverse)")
	ErrCannotUpdateTerminalTask     = errors.New("cannot update task with terminal status (completed or canceled)")
	ErrCannotSetContributionOnInbox = errors.New("cannot set contribution on inbox task")
	ErrCannotCancelCompletedTask    = errors.New("cannot cancel completed task (use Uncomplete first)")
)

// Task represents a task in the system (inbox or objective-linked).
type Task struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ObjectiveID  *uuid.UUID // NULL for inbox tasks
	Title        tasktitle.TaskTitle
	Description  *taskdescription.TaskDescription // NULL if not provided
	Contribution *contribution.Contribution       // NULL for inbox tasks, required for objective-linked
	Status       taskstatus.TaskStatus
	CompletedAt  *time.Time
	DateCreated  time.Time
	DateUpdated  time.Time
}

// NewTask contains information needed to create a new task.
type NewTask struct {
	UserID       uuid.UUID
	ObjectiveID  *uuid.UUID // NULL for inbox tasks
	Title        tasktitle.TaskTitle
	Description  *taskdescription.TaskDescription
	Contribution *contribution.Contribution // NULL for inbox, required for objective-linked
}

// UpdateTask contains information needed to update a task.
// Nil pointers mean "no update", except ClearDescription.
type UpdateTask struct {
	Title            *tasktitle.TaskTitle
	Description      *taskdescription.TaskDescription
	ClearDescription bool // When true, sets Description to NULL (takes precedence over Description field)
	Contribution     *contribution.Contribution
}

type (
	// TaskStatus is an alias for external usage.
	TaskStatus = taskstatus.TaskStatus
)

// ParseStatus parses a string into a TaskStatus.
func ParseStatus(s string) (taskstatus.TaskStatus, error) {
	return taskstatus.Parse(s)
}
