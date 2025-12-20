// Package taskdb provides database access for tasks.
package taskdb

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/taskbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/types/contribution"
	"github.com/francowini/rafiki/business/types/taskdescription"
	"github.com/francowini/rafiki/business/types/taskstatus"
	"github.com/francowini/rafiki/business/types/tasktitle"
)

// task represents the database model.
type task struct {
	ID           uuid.UUID  `db:"task_id"`
	UserID       uuid.UUID  `db:"user_id"`
	ObjectiveID  *uuid.UUID `db:"objective_id"`
	Title        string     `db:"title"`       // encrypted
	Description  *string    `db:"description"` // encrypted
	Contribution *int       `db:"contribution"`
	Status       string     `db:"status"`
	CompletedAt  *time.Time `db:"completed_at"`
	DateCreated  time.Time  `db:"date_created"`
	DateUpdated  time.Time  `db:"date_updated"`
}

// toDBTaskEncrypted converts business model to DB model with encryption.
func toDBTaskEncrypted(bus taskbus.Task, enc encrypt.Encryptor) (task, error) {
	// Encrypt title
	title, err := enc.Encrypt(bus.Title.String())
	if err != nil {
		return task{}, fmt.Errorf("encrypt title: %w", err)
	}

	// Encrypt description if provided
	var description *string
	if bus.Description != nil && bus.Description.String() != "" {
		desc, err := enc.Encrypt(bus.Description.String())
		if err != nil {
			return task{}, fmt.Errorf("encrypt description: %w", err)
		}
		description = &desc
	}

	var contributionVal *int
	if bus.Contribution != nil {
		v := bus.Contribution.Value()
		contributionVal = &v
	}

	var completedAt *time.Time
	if bus.CompletedAt != nil {
		t := bus.CompletedAt.UTC()
		completedAt = &t
	}

	return task{
		ID:           bus.ID,
		UserID:       bus.UserID,
		ObjectiveID:  bus.ObjectiveID,
		Title:        title,
		Description:  description,
		Contribution: contributionVal,
		Status:       bus.Status.String(),
		CompletedAt:  completedAt,
		DateCreated:  bus.DateCreated.UTC(),
		DateUpdated:  bus.DateUpdated.UTC(),
	}, nil
}

// toBusTaskDecrypted converts DB model to business model with decryption.
func toBusTaskDecrypted(db task, enc encrypt.Encryptor) (taskbus.Task, error) {
	// Decrypt and parse title
	titleStr, err := enc.Decrypt(db.Title)
	if err != nil {
		return taskbus.Task{}, fmt.Errorf("decrypt title: %w", err)
	}

	title, err := tasktitle.Parse(titleStr)
	if err != nil {
		return taskbus.Task{}, fmt.Errorf("parse title: %w", err)
	}

	// Decrypt and parse description if provided
	var descriptionPtr *taskdescription.TaskDescription
	if db.Description != nil && *db.Description != "" {
		descStr, err := enc.Decrypt(*db.Description)
		if err != nil {
			return taskbus.Task{}, fmt.Errorf("decrypt description: %w", err)
		}

		desc, err := taskdescription.Parse(descStr)
		if err != nil {
			return taskbus.Task{}, fmt.Errorf("parse description: %w", err)
		}
		descriptionPtr = &desc
	}

	// Parse status
	status, err := taskstatus.Parse(db.Status)
	if err != nil {
		return taskbus.Task{}, fmt.Errorf("parse status: %w", err)
	}

	var contributionPtr *contribution.Contribution
	if db.Contribution != nil {
		c, err := contribution.Parse(*db.Contribution)
		if err != nil {
			return taskbus.Task{}, fmt.Errorf("parse contribution: %w", err)
		}
		contributionPtr = &c
	}

	var completedAt *time.Time
	if db.CompletedAt != nil {
		t := db.CompletedAt.UTC()
		completedAt = &t
	}

	return taskbus.Task{
		ID:           db.ID,
		UserID:       db.UserID,
		ObjectiveID:  db.ObjectiveID,
		Title:        title,
		Description:  descriptionPtr,
		Contribution: contributionPtr,
		Status:       status,
		CompletedAt:  completedAt,
		DateCreated:  db.DateCreated.UTC(),
		DateUpdated:  db.DateUpdated.UTC(),
	}, nil
}

// toBusTasksDecrypted converts a slice of DB models to business models.
func toBusTasksDecrypted(dbs []task, enc encrypt.Encryptor) ([]taskbus.Task, error) {
	tasks := make([]taskbus.Task, len(dbs))

	for i, db := range dbs {
		var err error
		tasks[i], err = toBusTaskDecrypted(db, enc)
		if err != nil {
			return nil, fmt.Errorf("record %d (id=%s): %w", i, db.ID, err)
		}
	}

	return tasks, nil
}
