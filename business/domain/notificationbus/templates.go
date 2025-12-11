//nolint:misspell // Spanish text is intentional
package notificationbus

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/francowini/rafiki/business/domain/vnotificationbus"
)

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
	var order []uuid.UUID

	for _, v := range values {
		if _, exists := valueMap[v.ValueID]; !exists {
			valueMap[v.ValueID] = &groupedValue{
				ValueContent: v.ValueContent,
				ValueFacet:   v.ValueFacet,
				Visions:      []string{},
			}
			order = append(order, v.ValueID)
		}
		if v.LifeVisionContent != nil && *v.LifeVisionContent != "" {
			valueMap[v.ValueID].Visions = append(valueMap[v.ValueID].Visions, *v.LifeVisionContent)
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

func morningTemplate(values []groupedValue) string {
	var sb strings.Builder
	sb.WriteString("*Buenos dias!*\n\n")
	sb.WriteString("Tus valores te guian hoy:\n\n")

	for _, v := range values {
		sb.WriteString(fmt.Sprintf("*%s* _%s_\n", v.ValueContent, v.ValueFacet))
		for _, vision := range v.Visions {
			sb.WriteString(fmt.Sprintf("  - %s\n", vision))
		}
	}

	return sb.String()
}

func eveningTemplate(values []groupedValue) string {
	var sb strings.Builder
	sb.WriteString("*Cierre del dia*\n\n")
	sb.WriteString("Tus valores te guian:\n\n")

	for _, v := range values {
		sb.WriteString(fmt.Sprintf("*%s* _%s_\n", v.ValueContent, v.ValueFacet))
		for _, vision := range v.Visions {
			sb.WriteString(fmt.Sprintf("  - %s\n", vision))
		}
	}

	sb.WriteString("\n_Descansa bien._")
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
