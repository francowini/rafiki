package telegramapp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/app/sdk/errs"
	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/types/sessiontype"
	"github.com/francowini/rafiki/business/types/telegramchatid"
	"github.com/francowini/rafiki/foundation/jobqueue"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/telegram"
	"github.com/francowini/rafiki/foundation/web"
)

// app contains dependencies for Telegram webhook handlers.
type app struct {
	log                *logger.Logger
	userBus            userbus.ExtBusiness
	telegramSessionBus *telegramsessionbus.Business
	telegramClient     *telegram.Client
	jobQueue           *jobqueue.Client
}

// newApp constructs the Telegram app handler.
func newApp(
	log *logger.Logger,
	userBus userbus.ExtBusiness,
	telegramSessionBus *telegramsessionbus.Business,
	telegramClient *telegram.Client,
	jobQueue *jobqueue.Client,
) *app {
	return &app{
		log:                log,
		userBus:            userBus,
		telegramSessionBus: telegramSessionBus,
		telegramClient:     telegramClient,
		jobQueue:           jobQueue,
	}
}

// webhook processes incoming Telegram updates.
func (a *app) webhook(ctx context.Context, r *http.Request) web.Encoder {
	var update TelegramUpdate
	if err := web.Decode(r, &update); err != nil {
		a.log.Error(ctx, "telegram_webhook.decode", "err", err)
		return errs.New(errs.InvalidArgument, err)
	}

	// Only process messages (ignore other update types)
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	chatID, err := telegramchatid.Parse(msg.Chat.ID)
	if err != nil {
		a.log.Error(ctx, "telegram_webhook.invalid_chat_id", "chat_id", msg.Chat.ID, "err", err)
		return nil
	}

	// Lookup user by telegram_chat_id
	user, err := a.userBus.QueryByTelegramChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, userbus.ErrNotFound) {
			a.sendMessage(ctx, chatID, Msg().Errors.UnlinkedUser)
			return nil
		}
		a.log.Error(ctx, "telegram_webhook.query_user", "err", err)
		return nil // Don't expose internal errors to Telegram
	}

	text := strings.TrimSpace(msg.Text)

	// Route commands (synchronous) vs user responses (async)
	if strings.HasPrefix(text, "/") {
		return a.routeCommand(ctx, chatID, user.ID, text)
	}

	return a.handleUserResponse(ctx, chatID, user.ID, text)
}

// routeCommand handles bot commands.
func (a *app) routeCommand(ctx context.Context, chatID telegramchatid.TelegramChatID, userID uuid.UUID, text string) web.Encoder {
	// Extract command (handle @botname suffix)
	cmd := strings.Split(text, " ")[0]
	cmd = strings.Split(cmd, "@")[0]

	switch cmd {
	case "/momento":
		return a.handleMomentoCommand(ctx, chatID, userID)
	case "/cancel":
		return a.handleCancelCommand(ctx, chatID, userID)
	case "/ayuda", "/help", "/start":
		a.sendMessage(ctx, chatID, Msg().Commands.Ayuda)
		return nil
	case "/ejemplo":
		a.sendMessage(ctx, chatID, Msg().Commands.Ejemplo)
		return nil
	default:
		// Unknown command - send help
		a.sendMessage(ctx, chatID, Msg().Errors.OutsideSession)
		return nil
	}
}

// handleMomentoCommand starts a new moment tracking session.
func (a *app) handleMomentoCommand(ctx context.Context, chatID telegramchatid.TelegramChatID, userID uuid.UUID) web.Encoder {
	sessionType := sessiontype.MustParse(sessiontype.MomentTracking)

	// Check for existing session
	existing, err := a.telegramSessionBus.QueryByUserAndType(ctx, userID, sessionType)
	if err == nil {
		// Session exists
		a.log.Info(ctx, "telegram_momento.session_exists", "session_id", existing.ID)
		a.sendMessage(ctx, chatID, Msg().Session.Exists)
		return nil
	}

	if !errors.Is(err, telegramsessionbus.ErrNotFound) {
		a.log.Error(ctx, "telegram_momento.query_session", "err", err)
		a.sendMessage(ctx, chatID, Msg().Errors.Technical)
		return nil
	}

	// Create new session
	newSession := telegramsessionbus.NewSession{
		UserID:      userID,
		ChatID:      chatID,
		SessionType: sessionType,
	}

	session, err := a.telegramSessionBus.Create(ctx, newSession)
	if err != nil {
		a.log.Error(ctx, "telegram_momento.create_session", "err", err)
		a.sendMessage(ctx, chatID, Msg().Errors.Technical)
		return nil
	}

	a.log.Info(ctx, "telegram_momento.session_created", "session_id", session.ID, "user_id", userID)

	// Send first step prompt
	a.sendMessage(ctx, chatID, Msg().Steps.Step1)

	return nil
}

// handleCancelCommand cancels the active session.
func (a *app) handleCancelCommand(ctx context.Context, chatID telegramchatid.TelegramChatID, userID uuid.UUID) web.Encoder {
	sessionType := sessiontype.MustParse(sessiontype.MomentTracking)

	session, err := a.telegramSessionBus.QueryByUserAndType(ctx, userID, sessionType)
	if err != nil {
		if errors.Is(err, telegramsessionbus.ErrNotFound) {
			a.sendMessage(ctx, chatID, Msg().Session.NoSession)
			return nil
		}
		a.log.Error(ctx, "telegram_cancel.query_session", "err", err)
		a.sendMessage(ctx, chatID, Msg().Errors.Technical)
		return nil
	}

	// Delete session (hard delete, no partial save)
	if err := a.telegramSessionBus.Delete(ctx, session); err != nil {
		a.log.Error(ctx, "telegram_cancel.delete_session", "err", err)
		a.sendMessage(ctx, chatID, Msg().Errors.Technical)
		return nil
	}

	a.log.Info(ctx, "telegram_cancel.session_deleted", "session_id", session.ID)

	// Neutral confirmation (ACT-aligned per Product feedback)
	a.sendMessage(ctx, chatID, Msg().Session.Canceled)

	return nil
}

// handleUserResponse processes text responses during active session.
func (a *app) handleUserResponse(ctx context.Context, chatID telegramchatid.TelegramChatID, userID uuid.UUID, text string) web.Encoder {
	// Empty response check
	if text == "" {
		a.sendMessage(ctx, chatID, Msg().Errors.EmptyResponse)
		return nil
	}

	// Check for active session
	session, err := a.telegramSessionBus.QueryByChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, telegramsessionbus.ErrNotFound) {
			a.sendMessage(ctx, chatID, Msg().Errors.OutsideSession)
			return nil
		}
		a.log.Error(ctx, "telegram_response.query_session", "err", err)
		a.sendMessage(ctx, chatID, Msg().Errors.Technical)
		return nil
	}

	// Enqueue job for async AI processing
	args := TelegramMessageArgs{
		SessionID: session.ID,
		UserID:    userID,
		ChatID:    chatID,
		Text:      text,
	}

	if _, err := a.jobQueue.Insert(ctx, args, nil); err != nil {
		a.log.Error(ctx, "telegram_response.enqueue_job", "err", err)
		a.sendMessage(ctx, chatID, Msg().Errors.Technical)
		return nil
	}

	a.log.Info(ctx, "telegram_response.job_enqueued",
		"session_id", session.ID,
		"user_id", userID,
		"step", session.CurrentStep.Value(),
	)

	return nil
}

// sendMessage sends a Telegram message (fire-and-forget).
func (a *app) sendMessage(ctx context.Context, chatID telegramchatid.TelegramChatID, text string) {
	if _, err := a.telegramClient.SendMessage(ctx, chatID.Value(), text); err != nil {
		a.log.Error(ctx, "telegram.send_message", "chat_id", chatID, "err", err)
	}
}
