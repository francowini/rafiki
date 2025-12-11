package vnotificationdb

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/francowini/rafiki/business/domain/vnotificationbus"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/logger"
)

// Store manages the set of APIs for vnotification database access.
type Store struct {
	log       *logger.Logger
	db        sqlx.ExtContext
	encryptor encrypt.Encryptor
}

// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store {
	return &Store{
		log:       log,
		db:        db,
		encryptor: encryptor,
	}
}

// QueryByUserID retrieves all values with their life visions for a user.
func (s *Store) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]vnotificationbus.ValueWithVision, error) {
	data := map[string]any{
		"user_id": userID,
	}

	const q = `
	SELECT
		user_id, value_id, value_content, value_facet, value_order,
		life_vision_id, life_vision_content
	FROM
		view_notification_content
	WHERE
		user_id = :user_id
	ORDER BY
		value_order ASC`

	var dbValues []valueWithVision
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbValues); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}

	return toBusValuesWithVision(dbValues, s.encryptor)
}
