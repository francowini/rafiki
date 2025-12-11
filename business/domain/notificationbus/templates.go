//nolint:misspell // Spanish text is intentional
package notificationbus

import (
	"strings"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/vnotificationbus"
)

// facetEmojis maps facet names to their emoji representations.
var facetEmojis = map[string]string{
	"spirituality":  "🧘",
	"health":        "💪",
	"family":        "👨‍👩‍👧‍👦",
	"relationships": "🤝",
	"career":        "💼",
	"finances":      "💰",
	"growth":        "🌱",
	"fun":           "🎉",
}

// groupedValue represents a value with all its associated life visions (goals).
type groupedValue struct {
	ValueContent string
	ValueFacet   string
	Visions      []string
}

// groupValuesByID groups values with their visions, avoiding duplicates.
func groupValuesByID(values []vnotificationbus.ValueWithVision) []groupedValue {
	// Use ordered map to preserve order and group by value
	valueMap := make(map[uuid.UUID]*groupedValue)
	visionSets := make(map[uuid.UUID]map[string]bool) // Track unique visions per value
	var order []uuid.UUID

	for _, v := range values {
		if _, exists := valueMap[v.ValueID]; !exists {
			valueMap[v.ValueID] = &groupedValue{
				ValueContent: v.ValueContent,
				ValueFacet:   v.ValueFacet,
				Visions:      []string{},
			}
			visionSets[v.ValueID] = make(map[string]bool)
			order = append(order, v.ValueID)
		}
		if v.LifeVisionContent != nil && *v.LifeVisionContent != "" {
			vision := *v.LifeVisionContent
			if !visionSets[v.ValueID][vision] {
				visionSets[v.ValueID][vision] = true
				valueMap[v.ValueID].Visions = append(valueMap[v.ValueID].Visions, vision)
			}
		}
	}

	// Convert to slice preserving order
	result := make([]groupedValue, 0, len(order))
	for _, id := range order {
		result = append(result, *valueMap[id])
	}
	return result
}

// GenerateMorningMessage creates a morning notification message.
func GenerateMorningMessage(values []vnotificationbus.ValueWithVision) string {
	if len(values) == 0 {
		return noValuesMessage()
	}
	return morningTemplate(groupValuesByID(values))
}

// GenerateEveningMessage creates an evening notification message.
func GenerateEveningMessage(values []vnotificationbus.ValueWithVision) string {
	if len(values) == 0 {
		return noValuesMessage()
	}
	return eveningTemplate(groupValuesByID(values))
}

func facetEmoji(facet string) string {
	if emoji, ok := facetEmojis[facet]; ok {
		return emoji
	}
	return "✨"
}

func morningTemplate(values []groupedValue) string {
	var sb strings.Builder
	sb.WriteString("☀️ *Buenos dias!*\n\n")

	for i, v := range values {
		sb.WriteString(facetEmoji(v.ValueFacet))
		sb.WriteString(" ")
		sb.WriteString(v.ValueContent)
		sb.WriteString("\n")
		for _, vision := range v.Visions {
			sb.WriteString("    ↳ _")
			sb.WriteString(vision)
			sb.WriteString("_\n")
		}
		if i < len(values)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func eveningTemplate(values []groupedValue) string {
	var sb strings.Builder
	sb.WriteString("🌙 *Cierre del dia*\n\n")

	for i, v := range values {
		sb.WriteString(facetEmoji(v.ValueFacet))
		sb.WriteString(" ")
		sb.WriteString(v.ValueContent)
		sb.WriteString("\n")
		for _, vision := range v.Visions {
			sb.WriteString("    ↳ _")
			sb.WriteString(vision)
			sb.WriteString("_\n")
		}
		if i < len(values)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n_Descansa bien_ 😴")
	return sb.String()
}

func noValuesMessage() string {
	return `*Tu viaje comienza aqui*

Aun no has definido tus valores en Rafiki.

Los valores son tu brujula interior. Te ayudan a tomar decisiones alineadas con quien quieres ser.

_Entra a la app y define tu primer valor._`
}

// WelcomeMessage returns the welcome message for new users.
func WelcomeMessage() string {
	return `*Bienvenido a las notificaciones de Rafiki!*

A partir de ahora recibiras:
- Mensajes por la manana con tus valores
- Mensajes por la noche para reflexionar

_Que comience tu viaje de desarrollo personal!_`
}

// TestMessage returns a test message.
func TestMessage() string {
	return `*Mensaje de prueba*

Tu conexion con Telegram esta funcionando correctamente.`
}
