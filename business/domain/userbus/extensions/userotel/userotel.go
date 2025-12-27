// Package userotel provides an extension for userbus that adds
// otel tracking.
package userotel

import (
	"context"
	"net/mail"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/sdk/order"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/business/types/telegramchatid"
	"github.com/francowini/rafiki/foundation/otel"
)

// Extension provides a wrapper for otel functionality around the userbus.
type Extension struct {
	bus userbus.ExtBusiness
}

// NewExtension constructs a new extension that wraps the userbus with otel.
func NewExtension() userbus.Extension {
	return func(bus userbus.ExtBusiness) userbus.ExtBusiness {
		return &Extension{
			bus: bus,
		}
	}
}

// NewWithTx does not apply otel.
func (ext *Extension) NewWithTx(tx sqldb.CommitRollbacker) (userbus.ExtBusiness, error) {
	return ext.bus.NewWithTx(tx)
}

// Create applies otel to the user creation process.
func (ext *Extension) Create(ctx context.Context, actorID uuid.UUID, nu userbus.NewUser) (userbus.User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.create")
	defer span.End()

	usr, err := ext.bus.Create(ctx, actorID, nu)
	if err != nil {
		return userbus.User{}, err
	}

	return usr, nil
}

// Update applies otel to the user update process.
func (ext *Extension) Update(ctx context.Context, actorID uuid.UUID, usr userbus.User, uu userbus.UpdateUser) (userbus.User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.update")
	defer span.End()

	usr, err := ext.bus.Update(ctx, actorID, usr, uu)
	if err != nil {
		return userbus.User{}, err
	}

	return usr, nil
}

// Delete applies otel to the user delete process.
func (ext *Extension) Delete(ctx context.Context, usr userbus.User) error {
	ctx, span := otel.AddSpan(ctx, "business.userbus.delete")
	defer span.End()

	return ext.bus.Delete(ctx, usr)
}

// Query applies otel to the user query process.
func (ext *Extension) Query(ctx context.Context, filter userbus.QueryFilter, orderBy order.By, page page.Page) ([]userbus.User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.query")
	defer span.End()

	return ext.bus.Query(ctx, filter, orderBy, page)
}

// Count applies otel to the user count process.
func (ext *Extension) Count(ctx context.Context, filter userbus.QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.count")
	defer span.End()

	return ext.bus.Count(ctx, filter)
}

// QueryByID applies otel to the user query by ID process.
func (ext *Extension) QueryByID(ctx context.Context, userID uuid.UUID) (userbus.User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.querybyid")
	defer span.End()

	return ext.bus.QueryByID(ctx, userID)
}

// QueryByEmail applies otel to the user query by email process.
func (ext *Extension) QueryByEmail(ctx context.Context, email mail.Address) (userbus.User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.querybyemail")
	defer span.End()

	return ext.bus.QueryByEmail(ctx, email)
}

// QueryByTelegramChatID applies otel to the user query by Telegram chat ID process.
func (ext *Extension) QueryByTelegramChatID(ctx context.Context, chatID telegramchatid.TelegramChatID) (userbus.User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.querybytelegramchatid")
	defer span.End()

	return ext.bus.QueryByTelegramChatID(ctx, chatID)
}

// Authenticate applies otel to the user authentication process.
func (ext *Extension) Authenticate(ctx context.Context, email mail.Address, password string) (userbus.User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.authenticate")
	defer span.End()

	return ext.bus.Authenticate(ctx, email, password)
}
