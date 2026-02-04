package batch_mail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanUndefinedVariables(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		input    string
		wantKeep []string // substrings that must remain unchanged
		wantGone []string // substrings that must NOT appear in output
		wantHas  []string // substrings that must appear (replacement markers etc.)
	}{
		{
			name:     "preserves Subscriber.Email",
			input:    `Hello {{.Subscriber.Email}}`,
			wantKeep: []string{`{{.Subscriber.Email}}`},
		},
		{
			name:     "preserves Subscriber with whitespace",
			input:    `{{ .Subscriber.Email }}`,
			wantKeep: []string{`{{ .Subscriber.Email }}`},
		},
		{
			name:     "preserves Subscriber.Active",
			input:    `Status: {{.Subscriber.Active}}`,
			wantKeep: []string{`{{.Subscriber.Active}}`},
		},
		{
			name:     "preserves Subscriber custom attrib",
			input:    `Hi {{.Subscriber.FirstName}}`,
			wantKeep: []string{`{{.Subscriber.FirstName}}`},
		},
		{
			name:     "preserves Task variables",
			input:    `Task: {{.Task.TaskName}} ({{.Task.Subject}})`,
			wantKeep: []string{`{{.Task.TaskName}}`, `{{.Task.Subject}}`},
		},
		{
			name:     "preserves API variables",
			input:    `Key: {{.API.CustomKey}}`,
			wantKeep: []string{`{{.API.CustomKey}}`},
		},
		{
			name:     "preserves UnsubscribeURL function",
			input:    `<a href="{{UnsubscribeURL .}}">Unsub</a>`,
			wantKeep: []string{`{{UnsubscribeURL .}}`},
		},
		{
			name:     "replaces unknown variable",
			input:    `Hello {{.Unknown}}`,
			wantGone: []string{`{{.Unknown}}`},
			wantHas:  []string{`[__.Unknown__]`},
		},
		{
			name:     "replaces unknown nested variable",
			input:    `Val: {{.Foo.Bar}}`,
			wantGone: []string{`{{.Foo.Bar}}`},
			wantHas:  []string{`[__.Foo.Bar__]`},
		},
		{
			name:  "mixed known and unknown",
			input: `{{.Subscriber.Email}} {{.Unknown}} {{.Task.Subject}} {{.Nope}}`,
			wantKeep: []string{
				`{{.Subscriber.Email}}`,
				`{{.Task.Subject}}`,
			},
			wantGone: []string{
				`{{.Unknown}}`,
				`{{.Nope}}`,
			},
			wantHas: []string{
				`[__.Unknown__]`,
				`[__.Nope__]`,
			},
		},
		{
			name:     "empty content",
			input:    ``,
			wantKeep: []string{``},
		},
		{
			name:     "no template variables",
			input:    `<html><body>Hello World</body></html>`,
			wantKeep: []string{`<html><body>Hello World</body></html>`},
		},
		{
			name:     "plain text with braces that are not templates",
			input:    `Use { and } in JSON`,
			wantKeep: []string{`Use { and } in JSON`},
		},
		{
			name:     "unknown function call",
			input:    `{{SomeFunc .Arg}}`,
			wantGone: []string{`{{SomeFunc .Arg}}`},
			wantHas:  []string{`[__SomeFunc .Arg__]`},
		},
		{
			name:     "multiple Subscriber vars preserved",
			input:    `{{.Subscriber.Email}} and {{.Subscriber.Status}}`,
			wantKeep: []string{`{{.Subscriber.Email}}`, `{{.Subscriber.Status}}`},
		},
		{
			name:     "API and Task together",
			input:    `{{.API.Token}} / {{.Task.Id}}`,
			wantKeep: []string{`{{.API.Token}}`, `{{.Task.Id}}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)

			for _, s := range tt.wantKeep {
				assert.Contains(t, result, s, "should preserve: %s", s)
			}
			for _, s := range tt.wantGone {
				assert.NotContains(t, result, s, "should replace: %s", s)
			}
			for _, s := range tt.wantHas {
				assert.Contains(t, result, s, "should contain replacement: %s", s)
			}
		})
	}
}

func TestCleanUndefinedVariables_BlankLineRemoval(t *testing.T) {
	engine := NewTemplateEngine()

	input := "line1\n\n\nline2\n\nline3"
	result := engine.cleanUndefinedVariables(input)

	assert.NotContains(t, result, "\n\n", "blank lines should be removed")
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestCleanUndefinedVariables_WhitespaceVariants(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		input    string
		wantKeep bool
	}{
		{"compact subscriber", "{{.Subscriber.Email}}", true},
		{"spaced subscriber", "{{ .Subscriber.Email }}", true},
		{"extra spaces subscriber", "{{  .  Subscriber  .  Email  }}", true},
		{"compact task", "{{.Task.Id}}", true},
		{"spaced task", "{{ .Task.Id }}", true},
		{"compact api", "{{.API.Key}}", true},
		{"spaced api", "{{ .API.Key }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)
			if tt.wantKeep {
				assert.Equal(t, tt.input, result, "known variable should be preserved as-is")
			}
		})
	}
}

func TestCleanUndefinedVariables_UnsubscribeURLVariants(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name  string
		input string
	}{
		{"basic call", `{{UnsubscribeURL .}}`},
		{"with spaces", `{{ UnsubscribeURL . }}`},
		{"extra spaces", `{{  UnsubscribeURL  .  }}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.cleanUndefinedVariables(tt.input)
			assert.Equal(t, tt.input, result, "UnsubscribeURL should be preserved")
		})
	}
}

func TestCleanUndefinedVariables_ReplacementFormat(t *testing.T) {
	engine := NewTemplateEngine()

	// The replacement format is [__<inner>__] where <inner> is the trimmed
	// content between {{ and }}
	result := engine.cleanUndefinedVariables("{{.SomeRandom}}")
	assert.Contains(t, result, "[__")
	assert.Contains(t, result, "__]")
	assert.NotContains(t, result, "{{")
	assert.NotContains(t, result, "}}")
}

func TestNewTemplateEngine_NotNil(t *testing.T) {
	engine := NewTemplateEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.view)
}

func TestGetTemplateEngine_Singleton(t *testing.T) {
	e1 := GetTemplateEngine()
	e2 := GetTemplateEngine()
	assert.Same(t, e1, e2, "should return same instance")
}
