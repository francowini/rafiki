// Package telegramapp provides HTTP handlers for Telegram webhook integration.
package telegramapp

import (
	_ "embed"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed messages/es.yaml
var messagesYAML []byte

// Messages holds all user-facing message templates loaded from YAML.
type Messages struct {
	Commands CommandMessages `yaml:"commands"`
	Session  SessionMessages `yaml:"session"`
	Steps    StepMessages    `yaml:"steps"`
	Errors   ErrorMessages   `yaml:"errors"`
}

// CommandMessages contains responses for bot commands.
type CommandMessages struct {
	Ayuda   string `yaml:"ayuda"`
	Ejemplo string `yaml:"ejemplo"`
}

// SessionMessages contains session state messages.
type SessionMessages struct {
	Exists    string `yaml:"exists"`
	NoSession string `yaml:"no_session"`
	Canceled  string `yaml:"canceled"`
	Timeout   string `yaml:"timeout"`
	Saved     string `yaml:"saved"`
}

// StepMessages contains prompts for each step in the conversation.
type StepMessages struct {
	Step1 string `yaml:"step_1"`
}

// ErrorMessages contains error response messages.
type ErrorMessages struct {
	UnlinkedUser     string `yaml:"unlinked_user"`
	Technical        string `yaml:"technical"`
	EmptyResponse    string `yaml:"empty_response"`
	OutsideSession   string `yaml:"outside_session"`
	InvalidIntensity string `yaml:"invalid_intensity"`
}

var (
	msgs     Messages
	msgsOnce sync.Once
)

// Msg returns the loaded messages. Panics if YAML is malformed (caught at startup).
func Msg() Messages {
	msgsOnce.Do(func() {
		if err := yaml.Unmarshal(messagesYAML, &msgs); err != nil {
			panic("telegramapp: failed to parse messages/es.yaml: " + err.Error())
		}
	})
	return msgs
}
