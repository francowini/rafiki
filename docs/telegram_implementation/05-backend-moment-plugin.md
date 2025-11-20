# Backend Implementation - Moment Conversation Plugin

**Task Category**: Backend / Business Logic
**Estimated Time**: 8-10 hours
**Prerequisites**:
- [03-backend-foundation.md](./03-backend-foundation.md) - Foundation telegram package
- [04-backend-userbus.md](./04-backend-userbus.md) - UserBus Telegram methods
**Dependencies**: momentbus, foundation/telegram

---

## Overview

Implement the moment creation plugin for Telegram. This plugin handles the `/moment` command and guides users through a 7-step conversation to create a moment.

**Conversation Flow**:
1. Situation
2. Thoughts
3. Physical symptoms
4. Behavior
5. Consequences
6. Values reflection
7. Intensity (0-10)

**Timeout**: 5 minutes (auto-abandon if no activity)
**Cancellation**: Not allowed (simplified MVP)

---

## File to Create

```
api/services/partners/plugins/
└── moment.go    # Moment conversation plugin
```

---

## Implementation

**File**: `api/services/partners/plugins/moment.go`

```go
package plugins

import (
    "context"
    "fmt"
    "strconv"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/francowini/rafiki/business/domain/momentbus"
    "github.com/francowini/rafiki/business/domain/userbus"
    "github.com/francowini/rafiki/business/types/content"
    "github.com/francowini/rafiki/business/types/intensity"
    "github.com/francowini/rafiki/foundation/logger"
    "github.com/francowini/rafiki/foundation/telegram"
)

// MomentPlugin handles the /moment command and conversation flow.
type MomentPlugin struct {
    client      *telegram.Client
    convManager *telegram.ConversationManager
    momentBus   *momentbus.Business
    userBus     *userbus.Business
    log         *logger.Logger
}

// NewMomentPlugin creates a new moment plugin.
func NewMomentPlugin(
    client *telegram.Client,
    convManager *telegram.ConversationManager,
    momentBus *momentbus.Business,
    userBus *userbus.Business,
    log *logger.Logger,
) *MomentPlugin {
    return &MomentPlugin{
        client:      client,
        convManager: convManager,
        momentBus:   momentBus,
        userBus:     userBus,
        log:         log,
    }
}

// Name returns the plugin name.
func (p *MomentPlugin) Name() string {
    return "moments"
}

// Commands returns the commands this plugin handles.
func (p *MomentPlugin) Commands() []string {
    return []string{"/moment"}
}

// Description returns help text.
func (p *MomentPlugin) Description() string {
    return "Create a new moment with guided questions"
}

// HandleCommand handles the /moment command.
func (p *MomentPlugin) HandleCommand(ctx context.Context, msg *tgbotapi.Message) error {
    chatID, err := telegram.ParseChatID(msg.Chat.ID)
    if err != nil {
        return fmt.Errorf("parse chat id: %w", err)
    }

    // Check if user is linked
    userID, err := p.userBus.QueryUserIDByTelegramUserID(ctx, msg.From.ID)
    if err != nil {
        p.log.Warn(ctx, "user not linked", "telegram_user_id", msg.From.ID)
        return p.client.SendMessage(ctx, chatID,
            "⚠️ Your Telegram account is not linked to Rafiki.\n\n"+
                "Please visit https://app.rafiki.lat/settings to connect your account first.")
    }

    // Check if user already has an active conversation
    _, err = p.convManager.GetConversation(ctx, msg.From.ID)
    if err == nil {
        // Already has active conversation
        return p.client.SendMessage(ctx, chatID,
            "⚠️ You already have an active conversation.\n\n"+
                "Please finish your current moment before starting a new one.\n"+
                "The conversation will auto-cancel after 5 minutes of inactivity.")
    }

    // Start new conversation
    conv, err := p.convManager.StartConversation(ctx, msg.From.ID, userID, "moment")
    if err != nil {
        p.log.Error(ctx, "start conversation", "err", err)
        return p.client.SendMessage(ctx, chatID,
            "❌ Sorry, something went wrong. Please try again.")
    }

    p.log.Info(ctx, "moment conversation started",
        "conversation_id", conv.ID,
        "user_id", userID,
        "telegram_user_id", msg.From.ID)

    // Send first question
    return p.client.SendMessage(ctx, chatID,
        "Let's create a moment together. I'll ask you 7 questions.\n\n"+
            "**1/7: Situation**\n"+
            "Describe what happened. Where were you? What was going on?")
}

// HandleMessage handles conversation messages.
func (p *MomentPlugin) HandleMessage(ctx context.Context, msg *tgbotapi.Message) error {
    // Get active conversation
    conv, err := p.convManager.GetConversation(ctx, msg.From.ID)
    if err != nil {
        // No active conversation for this user
        return nil
    }

    // Only handle moment conversations
    if conv.ConversationType != "moment" {
        return nil
    }

    chatID, _ := telegram.ParseChatID(msg.Chat.ID)

    // Route based on current step
    switch conv.CurrentStep {
    case "moment:start":
        return p.handleSituation(ctx, chatID, msg, conv)
    case "moment:awaiting_thoughts":
        return p.handleThoughts(ctx, chatID, msg, conv)
    case "moment:awaiting_physical":
        return p.handlePhysicalSymptoms(ctx, chatID, msg, conv)
    case "moment:awaiting_behavior":
        return p.handleBehavior(ctx, chatID, msg, conv)
    case "moment:awaiting_consequences":
        return p.handleConsequences(ctx, chatID, msg, conv)
    case "moment:awaiting_values":
        return p.handleValuesReflection(ctx, chatID, msg, conv)
    case "moment:awaiting_intensity":
        return p.handleIntensity(ctx, chatID, msg, conv)
    default:
        p.log.Warn(ctx, "unknown conversation step", "step", conv.CurrentStep)
        return nil
    }
}

func (p *MomentPlugin) handleSituation(ctx context.Context, chatID telegram.ChatID, msg *tgbotapi.Message, conv telegram.Conversation) error {
    // Validate input
    situation, err := content.Parse(msg.Text)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            "❌ Please provide a description of the situation.\n\n"+
                "Try again:")
    }

    // Save to conversation data
    conv.Data["situation"] = situation.String()
    conv.CurrentStep = "moment:awaiting_thoughts"

    if err := p.convManager.UpdateConversation(ctx, conv); err != nil {
        p.log.Error(ctx, "update conversation", "err", err)
        return p.client.SendMessage(ctx, chatID,
            "❌ Sorry, something went wrong. Please try again from the beginning with /moment")
    }

    // Send next question
    return p.client.SendMessage(ctx, chatID,
        "**2/7: Thoughts**\n"+
            "What were your thoughts during this situation?")
}

func (p *MomentPlugin) handleThoughts(ctx context.Context, chatID telegram.ChatID, msg *tgbotapi.Message, conv telegram.Conversation) error {
    thoughts, err := content.Parse(msg.Text)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            "❌ Please describe your thoughts.\n\n"+
                "Try again:")
    }

    conv.Data["thoughts"] = thoughts.String()
    conv.CurrentStep = "moment:awaiting_physical"

    if err := p.convManager.UpdateConversation(ctx, conv); err != nil {
        p.log.Error(ctx, "update conversation", "err", err)
        return nil
    }

    return p.client.SendMessage(ctx, chatID,
        "**3/7: Physical Symptoms**\n"+
            "What physical sensations did you notice? (e.g., racing heart, tight chest, sweating)")
}

func (p *MomentPlugin) handlePhysicalSymptoms(ctx context.Context, chatID telegram.ChatID, msg *tgbotapi.Message, conv telegram.Conversation) error {
    physical, err := content.Parse(msg.Text)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            "❌ Please describe the physical symptoms you experienced.\n\n"+
                "Try again:")
    }

    conv.Data["physical_symptoms"] = physical.String()
    conv.CurrentStep = "moment:awaiting_behavior"

    if err := p.convManager.UpdateConversation(ctx, conv); err != nil {
        p.log.Error(ctx, "update conversation", "err", err)
        return nil
    }

    return p.client.SendMessage(ctx, chatID,
        "**4/7: Behavior**\n"+
            "How did you act or respond? What did you do?")
}

func (p *MomentPlugin) handleBehavior(ctx context.Context, chatID telegram.ChatID, msg *tgbotapi.Message, conv telegram.Conversation) error {
    behavior, err := content.Parse(msg.Text)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            "❌ Please describe how you behaved.\n\n"+
                "Try again:")
    }

    conv.Data["behavior"] = behavior.String()
    conv.CurrentStep = "moment:awaiting_consequences"

    if err := p.convManager.UpdateConversation(ctx, conv); err != nil {
        p.log.Error(ctx, "update conversation", "err", err)
        return nil
    }

    return p.client.SendMessage(ctx, chatID,
        "**5/7: Consequences**\n"+
            "What were the consequences of your actions? How did it affect you or others?")
}

func (p *MomentPlugin) handleConsequences(ctx context.Context, chatID telegram.ChatID, msg *tgbotapi.Message, conv telegram.Conversation) error {
    consequences, err := content.Parse(msg.Text)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            "❌ Please describe the consequences.\n\n"+
                "Try again:")
    }

    conv.Data["consequences"] = consequences.String()
    conv.CurrentStep = "moment:awaiting_values"

    if err := p.convManager.UpdateConversation(ctx, conv); err != nil {
        p.log.Error(ctx, "update conversation", "err", err)
        return nil
    }

    return p.client.SendMessage(ctx, chatID,
        "**6/7: Values Reflection**\n"+
            "What personal values were involved? Did this situation bring you closer to or further from what you value?")
}

func (p *MomentPlugin) handleValuesReflection(ctx context.Context, chatID telegram.ChatID, msg *tgbotapi.Message, conv telegram.Conversation) error {
    values, err := content.Parse(msg.Text)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            "❌ Please reflect on the values involved.\n\n"+
                "Try again:")
    }

    conv.Data["values_reflection"] = values.String()
    conv.CurrentStep = "moment:awaiting_intensity"

    if err := p.convManager.UpdateConversation(ctx, conv); err != nil {
        p.log.Error(ctx, "update conversation", "err", err)
        return nil
    }

    return p.client.SendMessage(ctx, chatID,
        "**7/7: Intensity**\n"+
            "On a scale of 0-10, how intense was this experience?\n\n"+
            "• 0 = no distress\n"+
            "• 5 = moderate distress\n"+
            "• 10 = extreme distress\n\n"+
            "Enter a number from 0 to 10:")
}

func (p *MomentPlugin) handleIntensity(ctx context.Context, chatID telegram.ChatID, msg *tgbotapi.Message, conv telegram.Conversation) error {
    // Parse intensity
    intensityVal, err := strconv.Atoi(msg.Text)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            "❌ Please enter a number between 0 and 10.\n\n"+
                "Try again:")
    }

    intensityObj, err := intensity.Parse(intensityVal)
    if err != nil {
        return p.client.SendMessage(ctx, chatID,
            fmt.Sprintf("❌ Intensity must be between 0 and 10, you entered: %d\n\n"+
                "Try again:", intensityVal))
    }

    // All data collected - create moment
    newMoment := momentbus.NewMoment{
        UserID:           conv.UserID,
        MomentDate:       time.Now(),
        Situation:        content.MustParse(conv.Data["situation"].(string)),
        Thoughts:         content.MustParse(conv.Data["thoughts"].(string)),
        PhysicalSymptoms: content.MustParse(conv.Data["physical_symptoms"].(string)),
        Behavior:         content.MustParse(conv.Data["behavior"].(string)),
        Consequences:     content.MustParse(conv.Data["consequences"].(string)),
        ValuesReflection: content.MustParse(conv.Data["values_reflection"].(string)),
        Intensity:        intensityObj,
    }

    moment, err := p.momentBus.Create(ctx, newMoment)
    if err != nil {
        p.log.Error(ctx, "create moment", "err", err)
        return p.client.SendMessage(ctx, chatID,
            "❌ Sorry, something went wrong saving your moment. Please try again later.")
    }

    // TODO: Update moment source to 'telegram' (requires momentbus update)

    // Complete conversation
    if err := p.convManager.CompleteConversation(ctx, conv.ID); err != nil {
        p.log.Error(ctx, "complete conversation", "err", err)
    }

    p.log.Info(ctx, "moment created via telegram",
        "moment_id", moment.ID,
        "user_id", conv.UserID,
        "intensity", intensityVal)

    // Send success message
    return p.client.SendMessage(ctx, chatID,
        fmt.Sprintf("✅ **Moment created successfully!**\n\n"+
            "• **Intensity**: %d/10\n"+
            "• **Time**: %s\n\n"+
            "You can view and edit this moment in the Rafiki app:\n"+
            "https://app.rafiki.lat/momentos\n\n"+
            "Send /moment to create another one.",
            intensityVal,
            time.Now().Format("Jan 2, 3:04 PM")))
}
```

---

## Task: Update MomentBus for Source Tracking

**File**: `business/domain/momentbus/momentbus.go`

Add method to update moment source after creation (temporary workaround):

```go
// SetSource updates the source field for a moment (used by Telegram plugin).
func (b *Business) SetSource(ctx context.Context, momentID uuid.UUID, source string) error {
    // This is a temporary method for Telegram integration
    // Future: Add source to NewMoment struct
    const q = `
    UPDATE moments
    SET source = $1, date_updated = $2
    WHERE moment_id = $3`

    if err := sqldb.ExecContext(ctx, b.log, b.db, q, source, time.Now(), momentID); err != nil {
        return fmt.Errorf("update moment source: %w", err)
    }

    return nil
}
```

Or better: Update `NewMoment` struct to include source:

```go
// NewMoment contains information needed to create a new moment.
type NewMoment struct {
    UserID           uuid.UUID
    MomentDate       time.Time
    Situation        content.Content
    Thoughts         content.Content
    PhysicalSymptoms content.Content
    Behavior         content.Content
    Consequences     content.Content
    ValuesReflection content.Content
    Intensity        intensity.Intensity
    Source           string                 // NEW: "web" or "telegram"
    SourceMetadata   map[string]interface{} // NEW: optional metadata
}
```

Then update the INSERT query to include source.

---

## Tests

**File**: `api/services/partners/plugins/moment_test.go`

```go
package plugins_test

import (
    "context"
    "testing"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/francowini/rafiki/api/services/partners/plugins"
)

func TestMomentPlugin_HandleCommand(t *testing.T) {
    plugin := setupTestMomentPlugin(t)
    ctx := context.Background()

    // Simulate /moment command
    msg := &tgbotapi.Message{
        From: &tgbotapi.User{ID: 123456789},
        Chat: &tgbotapi.Chat{ID: 123456789},
        Text: "/moment",
    }

    err := plugin.HandleCommand(ctx, msg)
    if err != nil {
        t.Fatalf("handle command: %v", err)
    }

    // Verify conversation started
    conv, err := convManager.GetConversation(ctx, 123456789)
    if err != nil {
        t.Fatalf("get conversation: %v", err)
    }

    if conv.ConversationType != "moment" {
        t.Errorf("expected type 'moment', got %s", conv.ConversationType)
    }

    if conv.CurrentStep != "moment:start" {
        t.Errorf("expected step 'moment:start', got %s", conv.CurrentStep)
    }
}

func TestMomentPlugin_CompleteConversation(t *testing.T) {
    plugin := setupTestMomentPlugin(t)
    ctx := context.Background()

    // Start conversation
    msg := &tgbotapi.Message{
        From: &tgbotapi.User{ID: 123456789},
        Chat: &tgbotapi.Chat{ID: 123456789},
        Text: "/moment",
    }
    _ = plugin.HandleCommand(ctx, msg)

    // Answer all 7 questions
    questions := []string{
        "Had panic attack at work",
        "I felt like everyone was judging me",
        "Racing heart, tight chest",
        "I left the room",
        "Missed important decision",
        "Struggled with professionalism vs self-care",
        "7",
    }

    for _, answer := range questions {
        msg := &tgbotapi.Message{
            From: &tgbotapi.User{ID: 123456789},
            Chat: &tgbotapi.Chat{ID: 123456789},
            Text: answer,
        }

        err := plugin.HandleMessage(ctx, msg)
        if err != nil {
            t.Fatalf("handle message: %v", err)
        }
    }

    // Verify moment created
    moments, err := momentBus.Query(ctx, filter)
    if err != nil {
        t.Fatalf("query moments: %v", err)
    }

    if len(moments) != 1 {
        t.Fatalf("expected 1 moment, got %d", len(moments))
    }

    // Verify conversation completed
    _, err = convManager.GetConversation(ctx, 123456789)
    if err == nil {
        t.Error("expected conversation to be completed (deleted)")
    }
}

func TestMomentPlugin_InvalidIntensity(t *testing.T) {
    plugin := setupTestMomentPlugin(t)
    ctx := context.Background()

    // Start conversation and answer first 6 questions...
    // (setup code omitted for brevity)

    // Send invalid intensity
    msg := &tgbotapi.Message{
        From: &tgbotapi.User{ID: 123456789},
        Chat: &tgbotapi.Chat{ID: 123456789},
        Text: "very high",  // Invalid - not a number
    }

    err := plugin.HandleMessage(ctx, msg)
    if err != nil {
        t.Fatalf("handle message: %v", err)
    }

    // Verify conversation still at intensity step
    conv, _ := convManager.GetConversation(ctx, 123456789)
    if conv.CurrentStep != "moment:awaiting_intensity" {
        t.Error("expected conversation to still be at intensity step")
    }

    // Verify no moment created yet
    moments, _ := momentBus.Query(ctx, filter)
    if len(moments) != 0 {
        t.Error("expected no moments created with invalid intensity")
    }
}
```

---

## Checklist

- [ ] Create `plugins/moment.go`
- [ ] Implement `NewMomentPlugin`
- [ ] Implement `Name`, `Commands`, `Description` methods
- [ ] Implement `HandleCommand` (start conversation)
- [ ] Implement `HandleMessage` (route to step handlers)
- [ ] Implement `handleSituation` (step 1/7)
- [ ] Implement `handleThoughts` (step 2/7)
- [ ] Implement `handlePhysicalSymptoms` (step 3/7)
- [ ] Implement `handleBehavior` (step 4/7)
- [ ] Implement `handleConsequences` (step 5/7)
- [ ] Implement `handleValuesReflection` (step 6/7)
- [ ] Implement `handleIntensity` (step 7/7 + create moment)
- [ ] Update `momentbus` to support source field
- [ ] Write tests for all conversation steps
- [ ] Test error handling (invalid inputs)
- [ ] Test user not linked scenario
- [ ] Test active conversation conflict
- [ ] Run tests: `go test ./api/services/partners/plugins/...`

---

## User Experience Notes

### Message Formatting
- Use **bold** for question numbers (1/7, 2/7, etc.)
- Use bullet points for examples and options
- Keep messages concise and friendly
- Use emojis sparingly (✅ ❌ ⚠️)

### Error Handling
- Validate all inputs immediately
- Provide clear error messages
- Don't lose conversation state on validation errors
- Allow users to retry invalid inputs

### Timeout Behavior
- Conversations auto-abandon after 5 minutes
- User can start fresh with `/moment` after timeout
- No partial data saved (all-or-nothing)

---

**Status**: ⏭️ Ready for Implementation
**Next Task**: [06-backend-other-plugins.md](./06-backend-other-plugins.md) - Link & Help plugins

