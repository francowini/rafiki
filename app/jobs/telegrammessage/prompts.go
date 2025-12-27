package telegrammessage

import (
	_ "embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed prompts/system.md
var systemPromptMD string

//go:embed prompts/steps.yaml
var stepsYAML []byte

var (
	loadedPrompts Prompts
	promptsOnce   sync.Once
	promptsErr    error
)

// Prompts holds all prompt templates.
type Prompts struct {
	System string
	Steps  map[int]StepPrompt
}

// StepPrompt contains prompts for a single conversation step.
type StepPrompt struct {
	Question    string   `yaml:"question"`     // User-facing (Spanish)
	Instruction string   `yaml:"instruction"`  // AI instruction (English)
	ParseFields []string `yaml:"parse_fields"` // Expected output fields
	EvalFirst   string   `yaml:"eval_first"`   // Strict validation criteria
	EvalRetry   string   `yaml:"eval_retry"`   // Permissive after retry
}

// StepsConfig is the YAML structure for steps.yaml.
type StepsConfig struct {
	Steps map[int]StepPrompt `yaml:"steps"`
}

// LoadPrompts loads and parses all prompt templates.
// Panics on malformed YAML (caught at startup).
func LoadPrompts() Prompts {
	promptsOnce.Do(func() {
		var cfg StepsConfig
		if err := yaml.Unmarshal(stepsYAML, &cfg); err != nil {
			promptsErr = fmt.Errorf("parse steps.yaml: %w", err)
			return
		}

		loadedPrompts = Prompts{
			System: systemPromptMD,
			Steps:  cfg.Steps,
		}
	})

	if promptsErr != nil {
		panic("telegrammessage: failed to load prompts: " + promptsErr.Error())
	}

	return loadedPrompts
}

// getStepPrompt returns the user-facing question for a step.
// Returns "[step not found]" if the step doesn't exist.
func getStepPrompt(stepNum int) string {
	prompts := LoadPrompts()
	step, ok := prompts.Steps[stepNum]
	if !ok {
		// Log warning - step prompts should always exist for valid steps
		// This indicates a configuration issue in steps.yaml
		return fmt.Sprintf("[step %d not found]", stepNum)
	}
	return step.Question
}
