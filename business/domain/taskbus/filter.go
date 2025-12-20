package taskbus

import (
	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/types/taskstatus"
)

// QueryFilter defines filter criteria for querying tasks.
type QueryFilter struct {
	ID          *uuid.UUID
	UserID      *uuid.UUID
	ObjectiveID *uuid.UUID
	Status      *taskstatus.TaskStatus
	InboxOnly   bool // When true, only return tasks with objective_id = NULL
}
