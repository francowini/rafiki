package telegrammessage

import (
	"context"
	"errors"
	"fmt"

	"github.com/francowini/rafiki/app/domain/telegramapp"
	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/business/types/telegramchatid"
)

// processMessage handles a single Telegram message.
func (w *Worker) processMessage(ctx context.Context, args telegramapp.TelegramMessageArgs) error {
	// 1. Load session
	session, err := w.sessionBus.QueryByID(ctx, args.SessionID)
	if err != nil {
		if errors.Is(err, telegramsessionbus.ErrNotFound) {
			return w.sendExpiredMessage(ctx, args.ChatID)
		}
		return fmt.Errorf("load session: %w", err)
	}

	w.log.Info(ctx, "telegrammessage.processing",
		"session_id", session.ID,
		"user_id", session.UserID,
		"step", session.CurrentStep.Value(),
		"retry_count", session.RetryCount.Value(),
	)

	// 2. Build prompt for current step
	prompt, err := w.buildPrompt(session, args.Text)
	if err != nil {
		return fmt.Errorf("build prompt: %w", err)
	}

	// 3. Call Anthropic API
	aiResp, err := w.callAnthropicAPI(ctx, prompt)
	if err != nil {
		// Retryable errors: let River retry
		if isRetryable(err) {
			return err
		}
		// Non-retryable: send error message to user
		return w.sendErrorMessage(ctx, args.ChatID)
	}

	// 4. Parse AI response
	stepResult, err := w.parseAIResponse(aiResp.Content)
	if err != nil {
		w.log.Error(ctx, "telegrammessage.parse_failed",
			"session_id", session.ID,
			"raw_content", aiResp.Content,
			"err", err,
		)
		// Critical error: malformed JSON from AI
		return w.sendErrorMessage(ctx, args.ChatID)
	}

	// 5. Update session state based on AI decision
	updatedSession, reply, err := w.updateSessionState(ctx, session, stepResult, args.Text)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	// 6. Send Telegram reply
	if err := w.sendTelegramMessage(ctx, args.ChatID, reply); err != nil {
		// Telegram failures are retryable
		return fmt.Errorf("send telegram: %w", err)
	}

	w.log.Info(ctx, "telegrammessage.completed",
		"session_id", updatedSession.ID,
		"new_step", updatedSession.CurrentStep.Value(),
		"status", stepResult.Status,
	)

	return nil
}

// updateSessionState handles AI validation result and updates session.
func (w *Worker) updateSessionState(ctx context.Context, session telegramsessionbus.Session, result AIStepResult, userText string) (telegramsessionbus.Session, string, error) {
	stepKey := session.StepKey()

	switch result.Status {
	case StatusApproved:
		return w.handleApproved(ctx, session, stepKey, result, userText)
	case StatusNeedsRefinement:
		return w.handleNeedsRefinement(ctx, session, stepKey, result)
	default:
		return session, "", fmt.Errorf("unknown status: %s", result.Status)
	}
}

// handleApproved advances session to next step.
func (w *Worker) handleApproved(ctx context.Context, session telegramsessionbus.Session, stepKey string, result AIStepResult, userText string) (telegramsessionbus.Session, string, error) {
	// Create step data with parsed values
	stepData := telegramsessionbus.NewStepData(userText)
	for k, v := range result.ParsedData {
		stepData = stepData.WithParsedValue(k, v)
	}

	var updatedSession telegramsessionbus.Session
	var err error

	if session.IsFinalStep() {
		// Store final step data (don't advance)
		updatedSession, err = w.sessionBus.StoreDataForFinalStep(ctx, session, stepKey, stepData)
		if err != nil {
			return session, "", fmt.Errorf("store final step: %w", err)
		}

		// Generate AI summary for completion
		summary := w.generateCompletionSummary(updatedSession)
		return updatedSession, summary, nil
	}

	// Advance to next step
	updatedSession, err = w.sessionBus.AdvanceStepWithData(ctx, session, stepKey, stepData)
	if err != nil {
		return session, "", fmt.Errorf("advance step: %w", err)
	}

	// Build reply: AI feedback + next step prompt
	reply := result.Feedback + "\n\n" + getStepPrompt(updatedSession.CurrentStep.Value())
	return updatedSession, reply, nil
}

// handleNeedsRefinement increments retry or auto-approves after max retries.
func (w *Worker) handleNeedsRefinement(ctx context.Context, session telegramsessionbus.Session, stepKey string, result AIStepResult) (telegramsessionbus.Session, string, error) {
	// Check if at max retries (2)
	if session.RetryCount.IsMaxed() {
		// PRODUCT DECISION: Compassionate auto-approval after 2 retries
		return w.handleCompassionateAutoApproval(ctx, session, stepKey, result)
	}

	// Increment retry count
	updatedSession, err := w.sessionBus.IncrementRetry(ctx, session)
	if err != nil {
		return session, "", fmt.Errorf("increment retry: %w", err)
	}

	// Send AI feedback (asking for clarification)
	return updatedSession, result.Feedback, nil
}

// handleCompassionateAutoApproval auto-approves with validating message after max retries.
// PRODUCT DECISION: Don't punish users for struggling - validate their effort.
func (w *Worker) handleCompassionateAutoApproval(ctx context.Context, session telegramsessionbus.Session, stepKey string, _ AIStepResult) (telegramsessionbus.Session, string, error) {
	stepNum := session.CurrentStep.Value()

	// Get default data for the step
	defaultData := getDefaultDataForStep(stepNum)
	stepData := telegramsessionbus.NewStepData("")
	for k, v := range defaultData {
		stepData = stepData.WithParsedValue(k, v)
	}

	var updatedSession telegramsessionbus.Session
	var err error

	if session.IsFinalStep() {
		updatedSession, err = w.sessionBus.StoreDataForFinalStep(ctx, session, stepKey, stepData)
		if err != nil {
			return session, "", fmt.Errorf("store final step: %w", err)
		}

		// Special message for intensity default
		reply := compassionateIntensityMessage + "\n\n" + w.generateCompletionSummary(updatedSession)
		return updatedSession, reply, nil
	}

	updatedSession, err = w.sessionBus.AdvanceStepWithData(ctx, session, stepKey, stepData)
	if err != nil {
		return session, "", fmt.Errorf("advance step: %w", err)
	}

	// Compassionate auto-approval message + next step prompt
	reply := compassionateAutoApprovalMessage + "\n\n" + getStepPrompt(updatedSession.CurrentStep.Value())
	return updatedSession, reply, nil
}

// generateCompletionSummary creates ACT-informed summary after all steps complete.
// PRODUCT DECISION: Reinforce learning with pattern reflection + curiosity invitation.
func (w *Worker) generateCompletionSummary(session telegramsessionbus.Session) string {
	// Extract key data from session context
	step1Data, _ := session.ContextData.GetStep("step_1")
	step3Data, _ := session.ContextData.GetStep("step_3")
	step5Data, _ := session.ContextData.GetStep("step_5")

	situacion := step1Data.ParsedValues["situacion"]            //nolint:misspell // Spanish word
	conducta := step3Data.ParsedValues["conducta"]              //nolint:misspell // Spanish word
	valoresDescripcion := step5Data.ParsedValues["descripcion"] //nolint:misspell // Spanish word

	// Build summary
	summary := completionConfirmation + "\n\n"

	// Pattern reflection
	if situacion != "" && conducta != "" {
		summary += fmt.Sprintf("Notaste que cuando %s, tu respuesta fue %s.", situacion, conducta)
	}

	// Values connection (if available)
	if valoresDescripcion != "" {
		summary += " " + valoresDescripcion
	}

	// Curiosity invitation
	summary += "\n\n" + curiosityInvitation

	return summary
}

// getDefaultDataForStep returns default parsed values for auto-approval.
func getDefaultDataForStep(stepNum int) map[string]string {
	switch stepNum {
	case 6:
		// PRODUCT DECISION: Default intensity to 5 (not null)
		// Rationale: Self-awareness of intensity is a core ACT skill to develop
		return map[string]string{"intensidad": "5"}
	default:
		// For other steps, store placeholder indicating auto-approved
		return map[string]string{"auto_approved": "true"}
	}
}

// Message constants (Spanish - user-facing)
//
//nolint:misspell // Spanish text - not English misspellings
const (
	// PRODUCT DECISION: Compassionate auto-approval after 2 retries
	compassionateAutoApprovalMessage = "Veo que este paso es difícil ahora mismo, y eso está bien. Vamos a seguir con lo que pudiste identificar. ✓"

	// PRODUCT DECISION: Educational message when defaulting intensity to 5
	compassionateIntensityMessage = "Asigné intensidad 5 (media). Con la práctica, te va a resultar más fácil identificar estas diferencias. ✓"

	// Completion confirmation
	completionConfirmation = "✓ Momento guardado"

	// Curiosity invitation (ACT-aligned)
	curiosityInvitation = "¿Reconocés este patrón en otros momentos?"
)

// sendTelegramMessage sends a message to the user.
func (w *Worker) sendTelegramMessage(ctx context.Context, chatID telegramchatid.TelegramChatID, text string) error {
	_, err := w.telegramClient.SendMessage(ctx, chatID.Value(), text)
	return err
}

// sendExpiredMessage sends session expired notification.
func (w *Worker) sendExpiredMessage(ctx context.Context, chatID telegramchatid.TelegramChatID) error {
	msg := "Tu sesión expiró. Usá /momento para empezar de nuevo." //nolint:misspell // Spanish text
	return w.sendTelegramMessage(ctx, chatID, msg)
}

// sendErrorMessage sends generic error notification.
func (w *Worker) sendErrorMessage(ctx context.Context, chatID telegramchatid.TelegramChatID) error {
	msg := "Hubo un problema técnico. Intentá de nuevo en un momento." //nolint:misspell // Spanish text
	return w.sendTelegramMessage(ctx, chatID, msg)
}
