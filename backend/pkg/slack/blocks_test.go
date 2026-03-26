package slack

import (
	"encoding/json"
	"testing"
)

func TestBuildQuestionResponseBlocks(t *testing.T) {
	aiResponse := "Here is a suggestion for your problem."
	questionID := "q-123"

	blocks := BuildQuestionResponseBlocks(aiResponse, questionID)

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	// Verify the section block contains the AI response.
	section := blocks[0]
	if section.Type != "section" {
		t.Errorf("expected first block type 'section', got %q", section.Type)
	}
	if section.Text == nil {
		t.Fatal("expected section text to be non-nil")
	}
	if section.Text.Type != "mrkdwn" {
		t.Errorf("expected text type 'mrkdwn', got %q", section.Text.Type)
	}
	if section.Text.Text != aiResponse {
		t.Errorf("expected text %q, got %q", aiResponse, section.Text.Text)
	}

	// Verify the actions block has three buttons.
	actions := blocks[1]
	if actions.Type != "actions" {
		t.Errorf("expected second block type 'actions', got %q", actions.Type)
	}
	if len(actions.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(actions.Elements))
	}

	// Verify each button's action_id, style, and value.
	expectedButtons := []struct {
		actionID string
		style    string
		value    string
	}{
		{ActionIDQuestionResolved, ButtonStylePrimary, questionID},
		{ActionIDQuestionContinue, "", questionID},
		{ActionIDQuestionEscalate, ButtonStyleDanger, questionID},
	}

	for i, expected := range expectedButtons {
		btn := actions.Elements[i]
		if btn.Type != "button" {
			t.Errorf("element[%d]: expected type 'button', got %q", i, btn.Type)
		}
		if btn.ActionID != expected.actionID {
			t.Errorf("element[%d]: expected action_id %q, got %q", i, expected.actionID, btn.ActionID)
		}
		if btn.Style != expected.style {
			t.Errorf("element[%d]: expected style %q, got %q", i, expected.style, btn.Style)
		}
		if btn.Value != expected.value {
			t.Errorf("element[%d]: expected value %q, got %q", i, expected.value, btn.Value)
		}
		if btn.Text.Type != "plain_text" {
			t.Errorf("element[%d]: expected text type 'plain_text', got %q", i, btn.Text.Type)
		}
	}
}

func TestBuildQuestionResponseBlocksEmptyResponse(t *testing.T) {
	blocks := BuildQuestionResponseBlocks("", "q-empty")

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].Text.Text != "" {
		t.Errorf("expected empty text, got %q", blocks[0].Text.Text)
	}
}

func TestMarshalBlocks(t *testing.T) {
	blocks := BuildQuestionResponseBlocks("test response", "q-456")

	data, err := MarshalBlocks(blocks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it round-trips as valid JSON.
	var parsed []map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 blocks in JSON, got %d", len(parsed))
	}

	// Verify the style field is omitted when empty.
	actions := parsed[1]
	elements, ok := actions["elements"].([]any)
	if !ok {
		t.Fatal("expected elements array")
	}
	continueBtn, ok := elements[1].(map[string]any)
	if !ok {
		t.Fatal("expected continue button to be object")
	}
	if _, hasStyle := continueBtn["style"]; hasStyle {
		t.Error("continue button should not have 'style' field")
	}
}
