package userdb

import (
	"database/sql"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/sdk/sqldb/dbarray"
	"github.com/francowini/rafiki/business/types/name"
	"github.com/francowini/rafiki/business/types/nulltime"
	"github.com/francowini/rafiki/business/types/role"
	"github.com/francowini/rafiki/business/types/telegramchatid"
)

type user struct {
	ID               uuid.UUID      `db:"user_id"`
	Name             string         `db:"name"`
	Email            string         `db:"email"`
	Roles            dbarray.String `db:"roles"`
	PasswordHash     []byte         `db:"password_hash"`
	Department       sql.NullString `db:"department"`
	Enabled          bool           `db:"enabled"`
	TelegramChatID   sql.NullInt64  `db:"telegram_chat_id"`
	TelegramEnabled  bool           `db:"telegram_enabled"`
	TelegramLinkedAt sql.NullTime   `db:"telegram_linked_at"`
	DateCreated      time.Time      `db:"date_created"`
	DateUpdated      time.Time      `db:"date_updated"`
}

func toDBUser(bus userbus.User) user {
	return user{
		ID:           bus.ID,
		Name:         bus.Name.String(),
		Email:        bus.Email.Address,
		Roles:        role.ParseToString(bus.Roles),
		PasswordHash: bus.PasswordHash,
		Department: sql.NullString{
			String: bus.Department.String(),
			Valid:  bus.Department.Valid(),
		},
		Enabled: bus.Enabled,
		TelegramChatID: sql.NullInt64{
			Int64: bus.TelegramChatID.Value(),
			Valid: bus.TelegramChatID.Valid,
		},
		TelegramEnabled:  bus.TelegramEnabled,
		TelegramLinkedAt: bus.TelegramLinkedAt.ToSQLNullTime(),
		DateCreated:      bus.DateCreated.UTC(),
		DateUpdated:      bus.DateUpdated.UTC(),
	}
}

func toBusUser(db user) (userbus.User, error) {
	addr := mail.Address{
		Address: db.Email,
	}

	roles, err := role.ParseMany(db.Roles)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse: %w", err)
	}

	nme, err := name.Parse(db.Name)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse name: %w", err)
	}

	department, err := name.ParseNull(db.Department.String)
	if err != nil {
		return userbus.User{}, fmt.Errorf("parse department: %w", err)
	}

	// Convert sql.NullInt64 to telegramchatid.Null
	var chatID telegramchatid.Null
	if db.TelegramChatID.Valid {
		parsed, err := telegramchatid.Parse(db.TelegramChatID.Int64)
		if err != nil {
			return userbus.User{}, fmt.Errorf("parse telegram_chat_id: %w", err)
		}
		chatID = telegramchatid.NewNull(parsed)
	}

	bus := userbus.User{
		ID:               db.ID,
		Name:             nme,
		Email:            addr,
		Roles:            roles,
		PasswordHash:     db.PasswordHash,
		Enabled:          db.Enabled,
		Department:       department,
		TelegramChatID:   chatID,
		TelegramEnabled:  db.TelegramEnabled,
		TelegramLinkedAt: nulltime.FromSQLNullTime(db.TelegramLinkedAt),
		DateCreated:      db.DateCreated.UTC(),
		DateUpdated:      db.DateUpdated.UTC(),
	}

	return bus, nil
}

func toBusUsers(dbs []user) ([]userbus.User, error) {
	bus := make([]userbus.User, len(dbs))

	for i, db := range dbs {
		var err error
		bus[i], err = toBusUser(db)
		if err != nil {
			return nil, err
		}
	}

	return bus, nil
}
