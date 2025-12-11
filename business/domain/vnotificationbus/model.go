package vnotificationbus

import "github.com/google/uuid"

// ValueWithVision represents a value and its life vision for notifications.
// This is a read-only view model used exclusively for generating notification
// message content. It intentionally uses primitive types (string, int) instead
// of strong business types because:
// 1. Data comes directly from a database view (view_notification_content)
// 2. Values are already validated when stored in the source tables
// 3. This model is never used for writes or user input validation
// 4. Using primitives simplifies the read-only query layer
type ValueWithVision struct {
	UserID            uuid.UUID
	ValueID           uuid.UUID
	ValueContent      string
	ValueFacet        string
	ValueOrder        int
	LifeVisionID      *uuid.UUID
	LifeVisionContent *string
}
