package telegrammessage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/domain/telegramsessionbus"
	"github.com/francowini/rafiki/business/sdk/page"
	"github.com/francowini/rafiki/business/types/content"
	"github.com/francowini/rafiki/business/types/intensity"
)

// Completion message constants (Spanish - user-facing)
const (
	// Agentic framing + quiet validation
	agenticFraming = "Elegiste tomarte este tiempo para observarte."

	// Completion confirmation
	completionConfirmation = "✓ Momento guardado"

	// Curiosity invitation (ACT-aligned)
	curiosityInvitation = "¿Reconocés este patrón en otros momentos?"
)

// completeMoment creates a moment from the completed session data.
// Implements idempotent retry pattern - safe to call multiple times.
func (w *Worker) completeMoment(ctx context.Context, session telegramsessionbus.Session) error {
	// =========================================================================
	// Idempotency Check
	// =========================================================================
	// Query for moments created by this user in last 1 minute.
	// If found → assume duplicate (retry scenario), skip creation.

	recentMoments, err := w.momentBus.Query(ctx, momentbus.QueryFilter{
		UserID: &session.UserID,
		Page:   page.MustParse("1", "100"),
	})
	if err != nil {
		return fmt.Errorf("idempotency check: %w", err)
	}

	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	for _, m := range recentMoments {
		if m.DateCreated.After(oneMinuteAgo) {
			w.log.Info(ctx, "telegrammessage.completion.duplicate_detected",
				"session_id", session.ID,
				"moment_id", m.ID,
				"created_at", m.DateCreated,
			)
			return nil // Skip creation (idempotent)
		}
	}

	// =========================================================================
	// Build NewMoment from Session Data
	// =========================================================================

	newMoment, err := w.buildNewMoment(ctx, session)
	if err != nil {
		return fmt.Errorf("build moment: %w", err)
	}

	// =========================================================================
	// Create Moment
	// =========================================================================

	moment, err := w.momentBus.Create(ctx, newMoment)
	if err != nil {
		return fmt.Errorf("create moment: %w", err)
	}

	w.log.Info(ctx, "telegrammessage.completion.success",
		"session_id", session.ID,
		"moment_id", moment.ID,
		"user_id", session.UserID,
		"intensity", moment.Intensity.Value(),
	)

	return nil
}

// buildNewMoment maps session context data to momentbus.NewMoment.
func (w *Worker) buildNewMoment(ctx context.Context, session telegramsessionbus.Session) (momentbus.NewMoment, error) {
	// Extract step data
	step1, _ := session.ContextData.GetStep("step_1")
	step2, _ := session.ContextData.GetStep("step_2")
	step3, _ := session.ContextData.GetStep("step_3")
	step4, _ := session.ContextData.GetStep("step_4")
	step5, _ := session.ContextData.GetStep("step_5")
	step6, _ := session.ContextData.GetStep("step_6")

	// =========================================================================
	// Step 1: Situation + Thoughts
	// =========================================================================

	situationStr := getStepValue(step1, "situacion")
	situation, err := content.Parse(situationStr)
	if err != nil {
		situation = content.MustParse("(No especificado)")
		w.log.Warn(ctx, "telegrammessage.completion.empty_situation",
			"session_id", session.ID,
		)
	}

	thoughtsStr := getStepValue(step1, "pensamientos")
	thoughts, err := content.Parse(thoughtsStr)
	if err != nil {
		thoughts = content.MustParse("(No especificado)")
		w.log.Warn(ctx, "telegrammessage.completion.empty_thoughts",
			"session_id", session.ID,
		)
	}

	// =========================================================================
	// Step 2: Physical Symptoms + Emotions (concatenated with |)
	// =========================================================================

	sintomasFisicos := step2.ParsedValues["sintomas_fisicos"]
	emociones := step2.ParsedValues["emociones"]
	physicalStr := mergePhysicalData(sintomasFisicos, emociones)

	physical, err := content.Parse(physicalStr)
	if err != nil {
		if step2.RawResponse != "" {
			physical = content.MustParse(step2.RawResponse)
		} else {
			physical = content.MustParse("(No especificado)")
		}
		w.log.Warn(ctx, "telegrammessage.completion.empty_physical",
			"session_id", session.ID,
		)
	}

	// =========================================================================
	// Step 3: Behavior
	// =========================================================================

	behaviorStr := getStepValue(step3, "conducta")
	behavior, err := content.Parse(behaviorStr)
	if err != nil {
		behavior = content.MustParse("(No especificado)")
		w.log.Warn(ctx, "telegrammessage.completion.empty_behavior",
			"session_id", session.ID,
		)
	}

	// =========================================================================
	// Step 4: Consequences
	// =========================================================================

	consequencesStr := getStepValue(step4, "consecuencias")
	consequences, err := content.Parse(consequencesStr)
	if err != nil {
		consequences = content.MustParse("(No especificado)")
		w.log.Warn(ctx, "telegrammessage.completion.empty_consequences",
			"session_id", session.ID,
		)
	}

	// =========================================================================
	// Step 5: Values Reflection
	// =========================================================================

	valuesStr := getStepValue(step5, "descripcion")
	valuesReflection, err := content.Parse(valuesStr)
	if err != nil {
		valuesReflection = content.MustParse("(No especificado)")
		w.log.Warn(ctx, "telegrammessage.completion.empty_values",
			"session_id", session.ID,
		)
	}

	// =========================================================================
	// Step 6: Intensity (0-10)
	// =========================================================================

	intensityStr := step6.ParsedValues["intensidad"]
	intensityParsed, err := parseIntensity(intensityStr)
	if err != nil {
		intensityParsed = intensity.MustParse(5) // Default to medium
		w.log.Warn(ctx, "telegrammessage.completion.intensity_parse_failed",
			"session_id", session.ID,
			"value", intensityStr,
		)
	}

	return momentbus.NewMoment{
		UserID:           session.UserID,
		MomentDate:       time.Now(),
		Situation:        situation,
		Thoughts:         thoughts,
		PhysicalSymptoms: physical,
		Behavior:         behavior,
		Consequences:     consequences,
		ValuesReflection: valuesReflection,
		Intensity:        intensityParsed,
	}, nil
}

// generateCompletionSummary creates ACT-informed summary with agentic framing.
func (w *Worker) generateCompletionSummary(session telegramsessionbus.Session) string {
	step1Data, _ := session.ContextData.GetStep("step_1")
	step3Data, _ := session.ContextData.GetStep("step_3")
	step5Data, _ := session.ContextData.GetStep("step_5")

	situacion := step1Data.ParsedValues["situacion"]
	conducta := step3Data.ParsedValues["conducta"]
	valoresDescripcion := step5Data.ParsedValues["descripcion"]

	// Build summary with agentic framing first
	summary := agenticFraming + "\n\n" + completionConfirmation + "\n\n"

	// Pattern reflection
	if situacion != "" && conducta != "" {
		summary += fmt.Sprintf("Notaste que cuando %s, tu respuesta fue %s.", situacion, conducta)
	}

	// Values connection
	if valoresDescripcion != "" {
		summary += " " + valoresDescripcion
	}

	// Curiosity invitation
	summary += "\n\n" + curiosityInvitation

	return summary
}

// getStepValue returns raw response if available, otherwise parsed value.
// Preserves user's authentic language for auto-approved steps.
func getStepValue(step telegramsessionbus.StepData, key string) string {
	if strings.TrimSpace(step.RawResponse) != "" {
		return step.RawResponse
	}
	return step.ParsedValues[key]
}

// parseIntensity converts string intensity to intensity.Intensity type.
func parseIntensity(value string) (intensity.Intensity, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return intensity.Intensity{}, fmt.Errorf("empty intensity value")
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return intensity.Intensity{}, fmt.Errorf("invalid intensity format: %w", err)
	}

	return intensity.Parse(intValue)
}

// mergePhysicalData concatenates physical symptoms and emotions with | separator.
func mergePhysicalData(sintomas, emociones string) string {
	sintomas = strings.TrimSpace(sintomas)
	emociones = strings.TrimSpace(emociones)

	if sintomas == "" && emociones == "" {
		return ""
	}
	if sintomas == "" {
		return emociones
	}
	if emociones == "" {
		return sintomas
	}

	return sintomas + " | " + emociones
}
