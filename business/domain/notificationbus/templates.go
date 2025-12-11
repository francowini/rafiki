//nolint:misspell // Spanish text is intentional
package notificationbus

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/francowini/rafiki/business/domain/vnotificationbus"
)

// GenerateMorningMessage creates a morning notification message.
func GenerateMorningMessage(values []vnotificationbus.ValueWithVision) string {
	if len(values) == 0 {
		return noValuesMessage()
	}

	templates := []func([]vnotificationbus.ValueWithVision) string{
		morningTemplate1,
		morningTemplate2,
		morningTemplate3,
	}
	return templates[rand.IntN(len(templates))](values)
}

// GenerateEveningMessage creates an evening notification message.
func GenerateEveningMessage(values []vnotificationbus.ValueWithVision) string {
	if len(values) == 0 {
		return noValuesMessage()
	}

	templates := []func([]vnotificationbus.ValueWithVision) string{
		eveningTemplate1,
		eveningTemplate2,
		eveningTemplate3,
	}
	return templates[rand.IntN(len(templates))](values)
}

func morningTemplate1(values []vnotificationbus.ValueWithVision) string {
	var sb strings.Builder
	sb.WriteString("*Buenos dias!*\n\n")
	sb.WriteString("Hoy es un nuevo dia para vivir segun tus valores:\n\n")

	for i, v := range values {
		sb.WriteString(fmt.Sprintf("*%d. %s*", i+1, v.ValueContent))
		sb.WriteString(fmt.Sprintf(" _%s_\n", v.ValueFacet))

		if v.LifeVisionContent != nil {
			sb.WriteString(fmt.Sprintf("   - %s\n", *v.LifeVisionContent))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("_Que pequena accion puedes hacer hoy alineada con estos valores?_")
	return sb.String()
}

func morningTemplate2(values []vnotificationbus.ValueWithVision) string {
	var sb strings.Builder
	sb.WriteString("*Comienza el dia con intencion*\n\n")
	sb.WriteString("Tus valores fundamentales:\n\n")

	for _, v := range values {
		sb.WriteString(fmt.Sprintf("- *%s* (%s)\n", v.ValueContent, v.ValueFacet))
		if v.LifeVisionContent != nil {
			sb.WriteString(fmt.Sprintf("  _%s_\n", *v.LifeVisionContent))
		}
	}

	sb.WriteString("\n_Recuerda: cada decision es una oportunidad para acercarte a quien quieres ser._")
	return sb.String()
}

func morningTemplate3(values []vnotificationbus.ValueWithVision) string {
	var sb strings.Builder
	sb.WriteString("*Tu brujula de valores*\n\n")

	for _, v := range values {
		sb.WriteString(fmt.Sprintf("*%s*\n", v.ValueContent))
		sb.WriteString(fmt.Sprintf("Faceta: _%s_\n", v.ValueFacet))

		if v.LifeVisionContent != nil {
			sb.WriteString(fmt.Sprintf("Vision: %s\n", *v.LifeVisionContent))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("_Como puedes honrar estos valores hoy?_")
	return sb.String()
}

func eveningTemplate1(values []vnotificationbus.ValueWithVision) string {
	var sb strings.Builder
	sb.WriteString("*Reflexion nocturna*\n\n")
	sb.WriteString("Antes de terminar el dia, revisa tus valores:\n\n")

	for i, v := range values {
		sb.WriteString(fmt.Sprintf("%d. *%s* (%s)\n", i+1, v.ValueContent, v.ValueFacet))
		if v.LifeVisionContent != nil {
			sb.WriteString(fmt.Sprintf("   _%s_\n", *v.LifeVisionContent))
		}
	}

	sb.WriteString("\n_Actuaste hoy de acuerdo con estos valores? Que aprendiste?_")
	return sb.String()
}

func eveningTemplate2(values []vnotificationbus.ValueWithVision) string {
	var sb strings.Builder
	sb.WriteString("*Revision del dia*\n\n")

	for _, v := range values {
		sb.WriteString(fmt.Sprintf("- *%s*\n", v.ValueContent))
		if v.LifeVisionContent != nil {
			sb.WriteString(fmt.Sprintf("  %s\n", *v.LifeVisionContent))
		}
	}

	sb.WriteString("\n_Que decisiones tomaste hoy alineadas con tus valores?_\n")
	sb.WriteString("_Hay algo que quieras ajustar manana?_")
	return sb.String()
}

func eveningTemplate3(values []vnotificationbus.ValueWithVision) string {
	var sb strings.Builder
	sb.WriteString("*Cierre del dia*\n\n")
	sb.WriteString("Tus valores te guian:\n\n")

	for _, v := range values {
		sb.WriteString(fmt.Sprintf("- *%s* _%s_\n", v.ValueContent, v.ValueFacet))
	}

	sb.WriteString("\n_Reflexiona:_\n")
	sb.WriteString("- Que momentos hoy reflejaron tus valores?\n")
	sb.WriteString("- Que oportunidades dejaste pasar?\n")
	sb.WriteString("- Que haras diferente manana?\n")
	sb.WriteString("\n_Descansa bien._")
	return sb.String()
}

func noValuesMessage() string {
	return `*Tu viaje comienza aqui*

Aun no has definido tus valores en Rafiki.

Los valores son tu brujula interior. Te ayudan a tomar decisiones alineadas con quien quieres ser.

_Entra a la app y define tu primer valor. Es el primer paso hacia una vida mas intencional!_`
}

// WelcomeMessage returns the welcome message for new users.
func WelcomeMessage() string {
	return `*Bienvenido a las notificaciones de Rafiki!*

A partir de ahora recibiras:
- Mensajes por la manana con tus valores
- Mensajes por la noche para reflexionar

Estos recordatorios te ayudaran a vivir de acuerdo con lo que mas te importa.

_Que comience tu viaje de desarrollo personal!_`
}

// TestMessage returns a test message.
func TestMessage() string {
	return `*Mensaje de prueba*

Tu conexion con Telegram esta funcionando correctamente.

Recibiras tus recordatorios de valores en los horarios configurados.`
}
