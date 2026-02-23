package components

import (
	"testing"
)

func TestNewTextInput(t *testing.T) {
	ti := NewTextInput("Enter name...")
	if ti.Value() != "" {
		t.Errorf("initial Value() = %q, want empty", ti.Value())
	}
	if ti.Focused() {
		t.Error("new TextInput should not be focused")
	}
}

func TestTextInput_FocusBlur(t *testing.T) {
	ti := NewTextInput("placeholder")

	ti.Focus()
	if !ti.Focused() {
		t.Error("Focus() should set focused to true")
	}

	ti.Blur()
	if ti.Focused() {
		t.Error("Blur() should set focused to false")
	}
}

func TestTextInput_SetValue(t *testing.T) {
	ti := NewTextInput("placeholder")
	ti.SetValue("hello world")
	if got := ti.Value(); got != "hello world" {
		t.Errorf("Value() = %q, want %q", got, "hello world")
	}
}

func TestTextInput_Reset(t *testing.T) {
	ti := NewTextInput("placeholder")
	ti.SetValue("some text")
	ti.Reset()
	if got := ti.Value(); got != "" {
		t.Errorf("after Reset(), Value() = %q, want empty", got)
	}
}

func TestTextInput_WithMask(t *testing.T) {
	ti := NewTextInput("password").WithMask('*')
	ti.SetValue("secret")
	if ti.Value() != "secret" {
		t.Errorf("Value() = %q, want %q", ti.Value(), "secret")
	}
	// Mask should affect View() rendering, not the stored value
	ti.Focus()
	view := ti.View()
	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestTextInput_ViewPlaceholder(t *testing.T) {
	ti := NewTextInput("Type here...")
	view := ti.View()
	if view == "" {
		t.Error("View() with placeholder should not be empty")
	}
}

func TestTextInput_ViewFocused(t *testing.T) {
	ti := NewTextInput("Type here...")
	ti.Focus()
	view := ti.View()
	if view == "" {
		t.Error("focused View() should not be empty")
	}
}
