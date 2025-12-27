package telegrammessage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/foundation/anthropic"
)

// AI response status values
const (
	StatusApproved        = "approved"
	StatusNeedsRefinement = "needs_refinement"
)

// Sentinel errors for AI parsing
var (
	ErrMalformedAIResponse = errors.New("AI returned malformed JSON")
	ErrInvalidAIStatus     = errors.New("AI returned invalid status")
	ErrMissingFeedback     = errors.New("AI response missing required feedback")
)

// AIStepResult represents Claude's parsed response.
type AIStepResult struct {
	Status     string            `json:"status"`      // "approved" | "needs_refinement"
	Feedback   string            `json:"feedback"`    // User-facing Spanish message
	ParsedData map[string]string `json:"parsed_data"` // Extracted values
}

// buildPrompt constructs the prompt for Anthropic API.
func (w *Worker) buildPrompt(session telegramsessionbus.Session, userText string) (anthropic.MessageRequest, error) {
	prompts := LoadPrompts()

	stepNum := session.CurrentStep.Value()
	stepPrompt, ok := prompts.Steps[stepNum]
	if !ok {
		return anthropic.MessageRequest{}, fmt.Errorf("no prompt for step %d", stepNum)
	}

	// Determine evaluation criteria based on retry count
	evalCriteria := stepPrompt.EvalFirst
	if session.RetryCount.Value() > 0 {
		evalCriteria = stepPrompt.EvalRetry
	}

	// Build user message with context
	userMessage := w.buildUserMessage(session, userText, stepPrompt, evalCriteria)

	return anthropic.MessageRequest{
		SystemPrompt: prompts.System,
		UserMessage:  userMessage,
	}, nil
}

// buildUserMessage constructs the user message with step context.
func (w *Worker) buildUserMessage(session telegramsessionbus.Session, userText string, stepPrompt StepPrompt, evalCriteria string) string {
	var sb strings.Builder

	// Step header
	sb.WriteString(fmt.Sprintf("## Step %d: %s\n\n", session.CurrentStep.Value(), stepPrompt.Question))

	// Previous context (if not step 1)
	if session.CurrentStep.Value() > 1 {
		sb.WriteString("### Previous Context:\n")
		for i := 1; i < session.CurrentStep.Value(); i++ {
			stepKey := fmt.Sprintf("step_%d", i)
			if data, ok := session.ContextData.GetStep(stepKey); ok {
				response := data.RawResponse
				if response == "" {
					response = "(auto-approved)"
				}
				sb.WriteString(fmt.Sprintf("- Step %d: %s\n", i, response))
			}
		}
		sb.WriteString("\n")
	}

	// Current user response
	sb.WriteString(fmt.Sprintf("### User Response:\n%q\n\n", userText))

	// AI instruction
	sb.WriteString(fmt.Sprintf("### Instruction:\n%s\n\n", stepPrompt.Instruction))

	// Evaluation criteria
	sb.WriteString(fmt.Sprintf("### Evaluation Criteria:\n%s\n\n", evalCriteria))

	// Expected output fields
	sb.WriteString("### Expected Output Fields:\n")
	for _, field := range stepPrompt.ParseFields {
		sb.WriteString(fmt.Sprintf("- %s\n", field))
	}
	sb.WriteString("\n")

	// JSON format reminder
	sb.WriteString(`### Response Format (JSON only):
{
  "status": "approved" | "needs_refinement",
  "feedback": "User-facing message in Spanish",
  "parsed_data": {
    "field_name": "extracted value"
  }
}`)

	return sb.String()
}

// callAnthropicAPI sends the prompt to Claude and returns the response.
func (w *Worker) callAnthropicAPI(ctx context.Context, req anthropic.MessageRequest) (*anthropic.MessageResponse, error) {
	resp, err := w.anthropicClient.SendMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("anthropic send: %w", err)
	}

	w.log.Info(ctx, "telegrammessage.anthropic_call",
		"input_tokens", resp.Usage.InputTokens,
		"output_tokens", resp.Usage.OutputTokens,
	)

	return resp, nil
}

// parseAIResponse parses Claude's JSON response.
func (w *Worker) parseAIResponse(rawContent string) (AIStepResult, error) {
	// Clean response (remove markdown code blocks if present)
	content := strings.TrimSpace(rawContent)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result AIStepResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return AIStepResult{}, fmt.Errorf("%w: %v (content: %s)", ErrMalformedAIResponse, err, truncate(content, 200))
	}

	// Validate required fields
	if result.Status != StatusApproved && result.Status != StatusNeedsRefinement {
		return AIStepResult{}, fmt.Errorf("%w: %s", ErrInvalidAIStatus, result.Status)
	}
	if result.Feedback == "" {
		return AIStepResult{}, ErrMissingFeedback
	}

	return result, nil
}

// isRetryable determines if an error should trigger River retry.
func isRetryable(err error) bool {
	var apiErr *anthropic.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Retryable
	}

	var rateLimitErr *anthropic.RateLimitError
	return errors.As(err, &rateLimitErr)
}

// truncate shortens a string for logging (rune-aware for UTF-8 safety).
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
