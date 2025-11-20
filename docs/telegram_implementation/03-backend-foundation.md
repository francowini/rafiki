# Backend Implementation - Foundation Telegram Package

**Task Category**: Backend / Foundation
**Estimated Time**: 12-14 hours
**Prerequisites**:
- [01-backend-database.md](./01-backend-database.md) - Database schema created
- [02-backend-telegramdb.md](./02-backend-telegramdb.md) - Database access layer implemented
**Dependencies**: Telegram Bot API library

---

## Overview

Create the `foundation/telegram` package - a **reusable infrastructure layer** for all Telegram interactions. This package is feature-agnostic and can be used for moments, habits, goals, or any future Telegram-based feature.

**Key Principle**: Foundation code knows NOTHING about business logic (moments, habits, etc.). It only handles Telegram API communication, routing, and conversation state management.

---

## Install Dependencies

```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
```

---

## Directory Structure

```
foundation/telegram/
├── bot.go                 # Bot lifecycle management
├── bot_test.go
├── client.go              # Telegram API client wrapper
├── client_test.go
├── router.go              # Command routing
├── router_test.go
├── conversation.go        # Generic conversation state machine
├── conversation_test.go
├── webhook.go             # Webhook server
├── webhook_test.go
├── plugin.go              # Plugin interface
├── errors.go              # Telegram-specific errors
└── types/
    ├── chatid.go          # Validated ChatID type
    ├── messageid.go       # Validated MessageID type
    └── userid.go          # Validated TelegramUserID type
```

---

## Task 1: Create Validated Types

**File**: `foundation/telegram/types/chatid.go`

```go
// Package types provides validated Telegram-specific types.
package types

import (
    "fmt"
    "strconv"
)

// ChatID represents a validated Telegram chat ID.
type ChatID struct {
    value int64
}

// ParseChatID validates and creates a ChatID.
func ParseChatID(value int64) (ChatID, error) {
    if value == 0 {
        return ChatID{}, fmt.Errorf("chat ID cannot be zero")
    }
    return ChatID{value: value}, nil
}

// MustParseChatID is like ParseChatID but panics on error. Use in tests only.
func MustParseChatID(value int64) ChatID {
    chatID, err := ParseChatID(value)
    if err != nil {
        panic(err)
    }
    return chatID
}

// Value returns the int64 value of the chat ID.
func (c ChatID) Value() int64 {
    return c.value
}

// String returns the string representation.
func (c ChatID) String() string {
    return strconv.FormatInt(c.value, 10)
}

// Equal provides support for the go-cmp package and testing.
func (c ChatID) Equal(c2 ChatID) bool {
    return c.value == c2.value
}

// MarshalText provides support for logging and any marshal needs.
func (c ChatID) MarshalText() ([]byte, error) {
    return []byte(c.String()), nil
}
```

**File**: `foundation/telegram/types/userid.go`

```go
package types

import (
    "fmt"
    "strconv"
)

// TelegramUserID represents a validated Telegram user ID.
type TelegramUserID struct {
    value int64
}

// ParseTelegramUserID validates and creates a TelegramUserID.
func ParseTelegramUserID(value int64) (TelegramUserID, error) {
    if value <= 0 {
        return TelegramUserID{}, fmt.Errorf("telegram user ID must be positive, got %d", value)
    }
    return TelegramUserID{value: value}, nil
}

// MustParseTelegramUserID is like ParseTelegramUserID but panics on error.
func MustParseTelegramUserID(value int64) TelegramUserID {
    userID, err := ParseTelegramUserID(value)
    if err != nil {
        panic(err)
    }
    return userID
}

// Value returns the int64 value.
func (t TelegramUserID) Value() int64 {
    return t.value
}

// String returns the string representation.
func (t TelegramUserID) String() string {
    return strconv.FormatInt(t.value, 10)
}

// Equal provides support for testing.
func (t TelegramUserID) Equal(t2 TelegramUserID) bool {
    return t.value == t2.value
}

// MarshalText provides support for logging.
func (t TelegramUserID) MarshalText() ([]byte, error) {
    return []byte(t.String()), nil
}
```

**File**: `foundation/telegram/types/messageid.go`

```go
package types

import (
    "fmt"
    "strconv"
)

// MessageID represents a validated Telegram message ID.
type MessageID struct {
    value int
}

// ParseMessageID validates and creates a MessageID.
func ParseMessageID(value int) (MessageID, error) {
    if value <= 0 {
        return MessageID{}, fmt.Errorf("message ID must be positive, got %d", value)
    }
    return MessageID{value: value}, nil
}

// MustParseMessageID is like ParseMessageID but panics on error.
func MustParseMessageID(value int) MessageID {
    msgID, err := ParseMessageID(value)
    if err != nil {
        panic(err)
    }
    return msgID
}

// Value returns the int value.
func (m MessageID) Value() int {
    return m.value
}

// String returns the string representation.
func (m MessageID) String() string {
    return strconv.Itoa(m.value)
}

// Equal provides support for testing.
func (m MessageID) Equal(m2 MessageID) bool {
    return m.value == m2.value
}

// MarshalText provides support for logging.
func (m MessageID) MarshalText() ([]byte, error) {
    return []byte(m.String()), nil
}
```

**Tests**: Create `types/*_test.go` for each type.

---

## Task 2: Create Error Types

**File**: `foundation/telegram/errors.go`

```go
package telegram

import (
    "errors"
    "fmt"
)

var (
    // ErrNotLinked indicates the Telegram user is not linked to any Rafiki account.
    ErrNotLinked = errors.New("telegram: user not linked to any account")

    // ErrAlreadyLinked indicates the user is already linked to a Telegram account.
    ErrAlreadyLinked = errors.New("telegram: user already linked to telegram")

    // ErrInvalidLinkCode indicates the link code is invalid or expired.
    ErrInvalidLinkCode = errors.New("telegram: invalid or expired link code")

    // ErrNoActiveConversation indicates no active conversation exists.
    ErrNoActiveConversation = errors.New("telegram: no active conversation")

    // ErrBotNotStarted indicates the bot has not been started.
    ErrBotNotStarted = errors.New("telegram: bot not started")

    // ErrInvalidConfig indicates bot configuration is invalid.
    ErrInvalidConfig = errors.New("telegram: invalid bot configuration")
)

// APIError wraps Telegram API errors with context.
type APIError struct {
    Code    int    // Telegram error code
    Message string // Error message
    Err     error  // Underlying error
}

func (e *APIError) Error() string {
    return fmt.Sprintf("telegram API error %d: %s: %v", e.Code, e.Message, e.Err)
}

func (e *APIError) Unwrap() error {
    return e.Err
}
```

---

## Task 3: Create Client Wrapper

**File**: `foundation/telegram/client.go`

```go
package telegram

import (
    "context"
    "fmt"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/francowini/rafiki/foundation/logger"
    "github.com/francowini/rafiki/foundation/telegram/types"
)

// Client wraps the Telegram Bot API client.
type Client struct {
    api *tgbotapi.BotAPI
    log *logger.Logger
}

// NewClient creates a new Telegram API client.
func NewClient(token string, log *logger.Logger) (*Client, error) {
    if token == "" {
        return nil, fmt.Errorf("bot token cannot be empty: %w", ErrInvalidConfig)
    }

    api, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        return nil, fmt.Errorf("creating bot api: %w", err)
    }

    log.Info(context.Background(), "telegram client created", "username", api.Self.UserName)

    return &Client{
        api: api,
        log: log,
    }, nil
}

// SendMessage sends a text message to a chat.
func (c *Client) SendMessage(ctx context.Context, chatID types.ChatID, text string) error {
    msg := tgbotapi.NewMessage(chatID.Value(), text)
    msg.ParseMode = "Markdown" // Support markdown formatting

    _, err := c.api.Send(msg)
    if err != nil {
        c.log.Error(ctx, "send message failed", "err", err, "chat_id", chatID)
        return &APIError{
            Message: "failed to send message",
            Err:     err,
        }
    }

    c.log.Debug(ctx, "message sent", "chat_id", chatID, "text_length", len(text))
    return nil
}

// GetMe returns information about the bot.
func (c *Client) GetMe() (tgbotapi.User, error) {
    return c.api.Self, nil
}

// SetWebhook sets the webhook URL for the bot.
func (c *Client) SetWebhook(webhookURL string, secretToken string) error {
    webhook, err := tgbotapi.NewWebhook(webhookURL)
    if err != nil {
        return fmt.Errorf("creating webhook config: %w", err)
    }

    webhook.SecretToken = secretToken

    _, err = c.api.Request(webhook)
    if err != nil {
        return fmt.Errorf("setting webhook: %w", err)
    }

    c.log.Info(context.Background(), "webhook set", "url", webhookURL)
    return nil
}

// DeleteWebhook removes the webhook.
func (c *Client) DeleteWebhook() error {
    _, err := c.api.Request(tgbotapi.DeleteWebhookConfig{})
    if err != nil {
        return fmt.Errorf("deleting webhook: %w", err)
    }

    c.log.Info(context.Background(), "webhook deleted")
    return nil
}

// GetUpdates returns a channel for receiving updates (long polling mode).
func (c *Client) GetUpdates(timeout int) tgbotapi.UpdatesChannel {
    u := tgbotapi.NewUpdate(0)
    u.Timeout = timeout
    return c.api.GetUpdatesChan(u)
}
```

**Test**: Create `client_test.go` with mocked Telegram API.

---

## Task 4: Create Plugin Interface

**File**: `foundation/telegram/plugin.go`

```go
package telegram

import (
    "context"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Plugin represents a feature-specific handler for Telegram commands.
type Plugin interface {
    // Name returns the plugin name (e.g., "moments", "habits").
    Name() string

    // Commands returns the commands this plugin handles (e.g., "/moment", "/habit").
    Commands() []string

    // HandleCommand handles a command message.
    HandleCommand(ctx context.Context, msg *tgbotapi.Message) error

    // HandleMessage handles a non-command message (conversation state).
    HandleMessage(ctx context.Context, msg *tgbotapi.Message) error

    // Description returns help text for this plugin.
    Description() string
}
```

---

## Task 5: Create Router

**File**: `foundation/telegram/router.go`

```go
package telegram

import (
    "context"
    "strings"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/francowini/rafiki/foundation/logger"
)

// Router routes incoming Telegram messages to appropriate plugins.
type Router struct {
    plugins map[string]Plugin // command -> plugin mapping
    log     *logger.Logger
}

// NewRouter creates a new message router.
func NewRouter(log *logger.Logger) *Router {
    return &Router{
        plugins: make(map[string]Plugin),
        log:     log,
    }
}

// RegisterPlugin registers a plugin for handling commands.
func (r *Router) RegisterPlugin(plugin Plugin) {
    for _, cmd := range plugin.Commands() {
        r.plugins[cmd] = plugin
        r.log.Info(context.Background(), "plugin registered",
            "plugin", plugin.Name(),
            "command", cmd)
    }
}

// Route routes a message to the appropriate plugin.
func (r *Router) Route(ctx context.Context, msg *tgbotapi.Message) error {
    if msg == nil {
        return nil
    }

    // Check if message is a command
    if msg.IsCommand() {
        return r.routeCommand(ctx, msg)
    }

    // Non-command messages go to all plugins (they check for active conversations)
    return r.routeMessage(ctx, msg)
}

func (r *Router) routeCommand(ctx context.Context, msg *tgbotapi.Message) error {
    cmd := msg.Command()
    cmdWithSlash := "/" + cmd

    plugin, found := r.plugins[cmdWithSlash]
    if !found {
        r.log.Warn(ctx, "unknown command",
            "command", cmd,
            "user_id", msg.From.ID,
            "chat_id", msg.Chat.ID)
        return nil // Silently ignore unknown commands
    }

    r.log.Info(ctx, "routing command",
        "command", cmd,
        "plugin", plugin.Name(),
        "user_id", msg.From.ID)

    return plugin.HandleCommand(ctx, msg)
}

func (r *Router) routeMessage(ctx context.Context, msg *tgbotapi.Message) error {
    // Try all plugins - they check if they have an active conversation for this user
    for _, plugin := range r.plugins {
        err := plugin.HandleMessage(ctx, msg)
        if err != nil {
            // If plugin handled the message, it will return an error
            // Otherwise, it returns nil (message not for this plugin)
            return err
        }
    }

    r.log.Debug(ctx, "no plugin handled message",
        "text", msg.Text,
        "user_id", msg.From.ID)

    return nil
}

// ListPlugins returns all registered plugins.
func (r *Router) ListPlugins() []Plugin {
    seen := make(map[string]bool)
    var plugins []Plugin

    for _, plugin := range r.plugins {
        if !seen[plugin.Name()] {
            plugins = append(plugins, plugin)
            seen[plugin.Name()] = true
        }
    }

    return plugins
}
```

**Test**: `router_test.go`

```go
package telegram_test

import (
    "context"
    "testing"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/francowini/rafiki/foundation/telegram"
)

// Mock plugin for testing
type mockPlugin struct {
    name         string
    commands     []string
    handleCalled bool
}

func (m *mockPlugin) Name() string                                                { return m.name }
func (m *mockPlugin) Commands() []string                                          { return m.commands }
func (m *mockPlugin) Description() string                                         { return "mock plugin" }
func (m *mockPlugin) HandleCommand(ctx context.Context, msg *tgbotapi.Message) error {
    m.handleCalled = true
    return nil
}
func (m *mockPlugin) HandleMessage(ctx context.Context, msg *tgbotapi.Message) error {
    return nil
}

func TestRouter_RegisterPlugin(t *testing.T) {
    log := setupTestLogger()
    router := telegram.NewRouter(log)

    plugin := &mockPlugin{
        name:     "test",
        commands: []string{"/test", "/t"},
    }

    router.RegisterPlugin(plugin)

    plugins := router.ListPlugins()
    if len(plugins) != 1 {
        t.Errorf("expected 1 plugin, got %d", len(plugins))
    }
}

func TestRouter_RouteCommand(t *testing.T) {
    log := setupTestLogger()
    router := telegram.NewRouter(log)

    plugin := &mockPlugin{
        name:     "test",
        commands: []string{"/test"},
    }

    router.RegisterPlugin(plugin)

    msg := &tgbotapi.Message{
        Text: "/test",
        Entities: []tgbotapi.MessageEntity{
            {Type: "bot_command", Offset: 0, Length: 5},
        },
    }

    err := router.Route(context.Background(), msg)
    if err != nil {
        t.Fatalf("routing command: %v", err)
    }

    if !plugin.handleCalled {
        t.Error("expected plugin.HandleCommand to be called")
    }
}
```

---

## Task 6: Create Conversation Manager

**File**: `foundation/telegram/conversation.go`

```go
package telegram

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/francowini/rafiki/business/sdk/sqldb/telegramdb"
    "github.com/francowini/rafiki/foundation/logger"
    "github.com/google/uuid"
)

// ConversationManager manages multi-step conversations.
type ConversationManager struct {
    store telegramdb.Store
    log   *logger.Logger
}

// NewConversationManager creates a new conversation manager.
func NewConversationManager(store telegramdb.Store, log *logger.Logger) *ConversationManager {
    return &ConversationManager{
        store: store,
        log:   log,
    }
}

// Conversation represents an active conversation.
type Conversation struct {
    ID               uuid.UUID
    TelegramUserID   int64
    UserID           uuid.UUID
    ConversationType string
    CurrentStep      string
    Data             map[string]interface{}
    LastActivity     time.Time
}

// StartConversation begins a new conversation.
func (cm *ConversationManager) StartConversation(ctx context.Context, telegramUserID int64, userID uuid.UUID, conversationType string) (Conversation, error) {
    data, err := json.Marshal(map[string]interface{}{})
    if err != nil {
        return Conversation{}, fmt.Errorf("marshaling initial data: %w", err)
    }

    cs := telegramdb.ConversationState{
        ConversationID:   uuid.New(),
        TelegramUserID:   telegramUserID,
        UserID:           userID,
        ConversationType: conversationType,
        CurrentStep:      conversationType + ":start",
        Data:             data,
        LastActivity:     time.Now(),
        DateCreated:      time.Now(),
        DateUpdated:      time.Now(),
    }

    if err := cm.store.CreateConversationState(ctx, cs); err != nil {
        return Conversation{}, fmt.Errorf("creating conversation state: %w", err)
    }

    cm.log.Info(ctx, "conversation started",
        "conversation_id", cs.ConversationID,
        "telegram_user_id", telegramUserID,
        "type", conversationType)

    return cm.toConversation(cs)
}

// GetConversation retrieves the active conversation for a Telegram user.
func (cm *ConversationManager) GetConversation(ctx context.Context, telegramUserID int64) (Conversation, error) {
    cs, err := cm.store.QueryConversationStateByTelegramUserID(ctx, telegramUserID)
    if err != nil {
        return Conversation{}, ErrNoActiveConversation
    }

    return cm.toConversation(cs)
}

// UpdateConversation updates the conversation state.
func (cm *ConversationManager) UpdateConversation(ctx context.Context, conv Conversation) error {
    data, err := json.Marshal(conv.Data)
    if err != nil {
        return fmt.Errorf("marshaling conversation data: %w", err)
    }

    cs := telegramdb.ConversationState{
        ConversationID:   conv.ID,
        TelegramUserID:   conv.TelegramUserID,
        UserID:           conv.UserID,
        ConversationType: conv.ConversationType,
        CurrentStep:      conv.CurrentStep,
        Data:             data,
        LastActivity:     time.Now(),
        DateUpdated:      time.Now(),
    }

    if err := cm.store.UpdateConversationState(ctx, cs); err != nil {
        return fmt.Errorf("updating conversation state: %w", err)
    }

    cm.log.Info(ctx, "conversation updated",
        "conversation_id", conv.ID,
        "step", conv.CurrentStep)

    return nil
}

// CompleteConversation marks a conversation as complete and deletes it.
func (cm *ConversationManager) CompleteConversation(ctx context.Context, conversationID uuid.UUID) error {
    if err := cm.store.DeleteConversationState(ctx, conversationID); err != nil {
        return fmt.Errorf("deleting conversation state: %w", err)
    }

    cm.log.Info(ctx, "conversation completed", "conversation_id", conversationID)
    return nil
}

// CleanupAbandoned removes conversations abandoned for >5 minutes.
func (cm *ConversationManager) CleanupAbandoned(ctx context.Context) (int, error) {
    count, err := cm.store.DeleteAbandonedConversations(ctx)
    if err != nil {
        return 0, fmt.Errorf("cleanup abandoned conversations: %w", err)
    }

    if count > 0 {
        cm.log.Info(ctx, "abandoned conversations cleaned up", "count", count)
    }

    return count, nil
}

func (cm *ConversationManager) toConversation(cs telegramdb.ConversationState) (Conversation, error) {
    var data map[string]interface{}
    if err := json.Unmarshal(cs.Data, &data); err != nil {
        return Conversation{}, fmt.Errorf("unmarshaling conversation data: %w", err)
    }

    return Conversation{
        ID:               cs.ConversationID,
        TelegramUserID:   cs.TelegramUserID,
        UserID:           cs.UserID,
        ConversationType: cs.ConversationType,
        CurrentStep:      cs.CurrentStep,
        Data:             data,
        LastActivity:     cs.LastActivity,
    }, nil
}
```

---

## Task 7: Create Bot (Lifecycle Management)

**File**: `foundation/telegram/bot.go`

```go
package telegram

import (
    "context"
    "fmt"
    "sync"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/francowini/rafiki/foundation/logger"
)

// Bot manages the Telegram bot lifecycle.
type Bot struct {
    client  *Client
    router  *Router
    log     *logger.Logger
    webhook *WebhookServer

    mu       sync.RWMutex
    running  bool
    stopChan chan struct{}
}

// Config holds bot configuration.
type Config struct {
    Token          string
    WebhookURL     string // If empty, use polling mode
    WebhookPort    int    // Default: 8443
    SecretToken    string // For webhook validation
}

// NewBot creates a new Telegram bot.
func NewBot(cfg Config, router *Router, log *logger.Logger) (*Bot, error) {
    if cfg.Token == "" {
        return nil, ErrInvalidConfig
    }

    client, err := NewClient(cfg.Token, log)
    if err != nil {
        return nil, fmt.Errorf("creating client: %w", err)
    }

    bot := &Bot{
        client:   client,
        router:   router,
        log:      log,
        stopChan: make(chan struct{}),
    }

    // Set up webhook if URL provided
    if cfg.WebhookURL != "" {
        if err := client.SetWebhook(cfg.WebhookURL, cfg.SecretToken); err != nil {
            return nil, fmt.Errorf("setting webhook: %w", err)
        }

        bot.webhook = NewWebhookServer(cfg.WebhookPort, cfg.SecretToken, log)
    }

    return bot, nil
}

// Start starts the bot (webhook or polling mode).
func (b *Bot) Start(ctx context.Context) error {
    b.mu.Lock()
    if b.running {
        b.mu.Unlock()
        return fmt.Errorf("bot already running")
    }
    b.running = true
    b.mu.Unlock()

    b.log.Info(ctx, "bot starting")

    if b.webhook != nil {
        return b.startWebhook(ctx)
    }

    return b.startPolling(ctx)
}

func (b *Bot) startWebhook(ctx context.Context) error {
    b.log.Info(ctx, "bot started in webhook mode")

    updates := b.webhook.Start()

    for {
        select {
        case <-ctx.Done():
            b.log.Info(ctx, "bot stopping (context cancelled)")
            return ctx.Err()
        case <-b.stopChan:
            b.log.Info(ctx, "bot stopping (stop requested)")
            return nil
        case update := <-updates:
            go b.handleUpdate(ctx, update)
        }
    }
}

func (b *Bot) startPolling(ctx context.Context) error {
    b.log.Info(ctx, "bot started in polling mode")

    updates := b.client.GetUpdates(30)

    for {
        select {
        case <-ctx.Done():
            b.log.Info(ctx, "bot stopping (context cancelled)")
            return ctx.Err()
        case <-b.stopChan:
            b.log.Info(ctx, "bot stopping (stop requested)")
            return nil
        case update := <-updates:
            go b.handleUpdate(ctx, update)
        }
    }
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
    if update.Message == nil {
        return
    }

    b.log.Debug(ctx, "update received",
        "update_id", update.UpdateID,
        "message_id", update.Message.MessageID,
        "from_user_id", update.Message.From.ID,
        "text", update.Message.Text)

    if err := b.router.Route(ctx, update.Message); err != nil {
        b.log.Error(ctx, "routing message", "err", err)
    }
}

// Stop stops the bot gracefully.
func (b *Bot) Stop(ctx context.Context) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    if !b.running {
        return nil
    }

    b.log.Info(ctx, "bot stopping")

    close(b.stopChan)

    if b.webhook != nil {
        b.webhook.Stop()
    }

    b.running = false
    b.log.Info(ctx, "bot stopped")

    return nil
}
```

---

## Task 8: Create Webhook Server

**File**: `foundation/telegram/webhook.go`

```go
package telegram

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/francowini/rafiki/foundation/logger"
)

// WebhookServer handles incoming webhook requests from Telegram.
type WebhookServer struct {
    port        int
    secretToken string
    log         *logger.Logger
    server      *http.Server
    updates     chan tgbotapi.Update
}

// NewWebhookServer creates a new webhook server.
func NewWebhookServer(port int, secretToken string, log *logger.Logger) *WebhookServer {
    if port == 0 {
        port = 8443 // Default Telegram webhook port
    }

    return &WebhookServer{
        port:        port,
        secretToken: secretToken,
        log:         log,
        updates:     make(chan tgbotapi.Update, 100),
    }
}

// Start starts the webhook server and returns a channel of updates.
func (ws *WebhookServer) Start() <-chan tgbotapi.Update {
    mux := http.NewServeMux()
    mux.HandleFunc("/webhook", ws.handleWebhook)

    ws.server = &http.Server{
        Addr:    ":" + strconv.Itoa(ws.port),
        Handler: mux,
    }

    go func() {
        ws.log.Info(nil, "webhook server starting", "port", ws.port)
        if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            ws.log.Error(nil, "webhook server error", "err", err)
        }
    }()

    return ws.updates
}

// Stop stops the webhook server.
func (ws *WebhookServer) Stop() error {
    if ws.server != nil {
        return ws.server.Close()
    }
    return nil
}

func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
    // Validate secret token
    if ws.secretToken != "" {
        token := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
        if token != ws.secretToken {
            ws.log.Warn(nil, "invalid secret token", "ip", r.RemoteAddr)
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
    }

    // Parse update
    var update tgbotapi.Update
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        ws.log.Error(nil, "decoding webhook update", "err", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    // Send to updates channel (non-blocking)
    select {
    case ws.updates <- update:
        w.WriteHeader(http.StatusOK)
    default:
        ws.log.Warn(nil, "updates channel full, dropping update")
        w.WriteHeader(http.StatusServiceUnavailable)
    }
}
```

---

## Checklist

### Types
- [ ] Create `types/chatid.go` with validation
- [ ] Create `types/userid.go` with validation
- [ ] Create `types/messageid.go` with validation
- [ ] Write tests for all types

### Errors
- [ ] Create `errors.go` with error constants
- [ ] Create `APIError` type

### Client
- [ ] Create `client.go` with Telegram API wrapper
- [ ] Implement `SendMessage`
- [ ] Implement `SetWebhook` / `DeleteWebhook`
- [ ] Implement `GetUpdates` (polling)
- [ ] Write tests (mock Telegram API)

### Plugin System
- [ ] Create `plugin.go` interface
- [ ] Document plugin contract

### Router
- [ ] Create `router.go` with command routing
- [ ] Implement `RegisterPlugin`
- [ ] Implement `Route` (command vs. message)
- [ ] Write tests with mock plugins

### Conversation Manager
- [ ] Create `conversation.go` with state management
- [ ] Implement `StartConversation`
- [ ] Implement `GetConversation`
- [ ] Implement `UpdateConversation`
- [ ] Implement `CompleteConversation`
- [ ] Implement `CleanupAbandoned`
- [ ] Write tests

### Bot
- [ ] Create `bot.go` with lifecycle management
- [ ] Implement `NewBot` with config
- [ ] Implement `Start` (webhook vs. polling)
- [ ] Implement `Stop` (graceful shutdown)
- [ ] Write tests

### Webhook Server
- [ ] Create `webhook.go` with HTTP server
- [ ] Implement secret token validation
- [ ] Implement update channel
- [ ] Write tests

### Integration Test
- [ ] Create `bot_integration_test.go` with full flow test
- [ ] Test polling mode
- [ ] Test webhook mode
- [ ] Test plugin routing

---

**Status**: ⏭️ Ready for Implementation
**Next Task**: [04-backend-userbus.md](./04-backend-userbus.md) - UserBus Telegram methods
